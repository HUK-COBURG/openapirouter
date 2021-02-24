package openapirouter

import (
	"context"
	"errors"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"log"
	"net/http"
	"net/url"
	"reflect"
)

// The Router which implements the described features. It implements http.Handler to be compatible with existing HTTP
// libraries.
type Router struct {
	baseRouter *openapi3filter.Router
	errMapper  *errorMapper
}

// Creates a new Router with the path of a OpenAPI specification file in YAML or JSON format.
func NewRouter(swaggerPath string) (*Router, error) {
	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromFile(swaggerPath)
	if err != nil {
		return nil, err
	}
	return &Router{
		baseRouter: openapi3filter.NewRouter().WithSwagger(swagger),
		errMapper:  &errorMapper{errorMapping: make(map[reflect.Type]*HTTPError)},
	}, nil
}

// Implementation of http.Handler that finds the requestHandler for an incoming request and validates the requests. It
// also adds the pathParameters to the requests Context so they can be extracted by the requestHandler
func (router *Router) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var response *Response
	route, pathParams, err := router.baseRouter.FindRoute(request.Method, request.URL)
	if err != nil {
		if err.Error() == "Path doesn't support the HTTP method" {
			response = NewHTTPError(http.StatusMethodNotAllowed, err.Error()).ToResponse()
		} else {
			response = NewHTTPError(http.StatusNotFound, err.Error()).ToResponse()
		}
		response.write(writer)
		return
	}
	switch handler := route.Handler.(type) {
	case *requestHandler:
		err = openapi3filter.ValidateRequest(context.Background(), &openapi3filter.RequestValidationInput{
			Request:     request,
			PathParams:  pathParams,
			QueryParams: request.URL.Query(),
			Route:       route,
		})
		if err != nil {
			requestError := &openapi3filter.RequestError{}
			if errors.As(err, &requestError) {
				response = NewHTTPError(requestError.HTTPStatus(), requestError.Error()).ToResponse()
			} else {
				response = NewHTTPError(http.StatusInternalServerError, "error validating request").ToResponse()
			}
			response.write(writer)
			return
		}
		ctx := context.WithValue(request.Context(), pathParamsKey, pathParams)
		handler.ServeHTTP(writer, request.WithContext(ctx))
	default:
		response = NewHTTPError(http.StatusNotImplemented).ToResponse()
		response.write(writer)
		return
	}
}

// AddRequestHandler creates a new requestHandler for a specified method and path. It is used to set an implementation
// for an endpoint. The function panics, if the endpoint is not specified in the OpenAPI specification
func (router *Router) AddRequestHandler(method string, path string, handleFunc HandleRequestFunction) {
	pathUrl, err := url.Parse(path)
	if err != nil {
		log.Panicln(err)
	}
	route, _, err := router.baseRouter.FindRoute(method, pathUrl)
	if err != nil {
		log.Panicln(err)
	}
	route.Handler = &requestHandler{
		errMapper:       router.errMapper,
		handlerFunction: handleFunc,
	}
}

// AddErrorMapping adds a custom error that should be mapped to an error response. It uses the HTTPError to create the
// response.
// It takes an error and the response code this error should be mapped to. Additionally, any number of details can
// be specified.
func (router *Router) AddErrorMapping(err error, responseCode int, details ...string) {
	router.errMapper.errorMapping[reflect.TypeOf(err)] = NewHTTPError(responseCode, details...)
}
