package openapirouter

import (
	"context"
	"errors"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"log"
	"net/http"
	"reflect"
)

// The Router which implements the described features. It implements http.Handler to be compatible with existing HTTP
// libraries.
type Router struct {
	baseRouter      routers.Router
	errMapper       *errorMapper
	implementations map[*routers.Route]requestHandler
}

// Creates a new Router with the path of a OpenAPI specification file in YAML or JSON format.
func NewRouter(swaggerPath string) (*Router, error) {
	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromFile(swaggerPath)
	if err != nil {
		return nil, err
	}
	router, err := gorillamux.NewRouter(swagger)
	if err != nil {
		return nil, err
	}
	return &Router{
		baseRouter:      router,
		errMapper:       &errorMapper{errorMapping: make(map[reflect.Type]*HTTPError)},
		implementations: make(map[*routers.Route]requestHandler),
	}, nil
}

// Implementation of http.Handler that finds the requestHandler for an incoming request and validates the requests. It
// also adds the pathParameters to the requests Context so they can be extracted by the requestHandler
func (router *Router) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var response *Response
	route, pathParams, err := router.baseRouter.FindRoute(request)
	if err != nil {
		if err.Error() == routers.ErrMethodNotAllowed.Error() {
			response = NewHTTPError(http.StatusMethodNotAllowed, err.Error()).ToResponse()
		} else {
			response = NewHTTPError(http.StatusNotFound, err.Error()).ToResponse()
		}
		response.write(writer)
		return
	}
	handler, ok := router.implementations[route]
	if ok {
		err = openapi3filter.ValidateRequest(context.Background(), &openapi3filter.RequestValidationInput{
			Request:     request,
			PathParams:  pathParams,
			QueryParams: request.URL.Query(),
			Route:       route,
		})
		if err != nil {
			requestError := &openapi3filter.RequestError{}
			if errors.As(err, &requestError) {
				response = NewHTTPError(http.StatusBadRequest, requestError.Error()).ToResponse()
			} else {
				response = NewHTTPError(http.StatusInternalServerError, "error validating request").ToResponse()
			}
			response.write(writer)
			return
		}
		ctx := context.WithValue(request.Context(), pathParamsKey, pathParams)
		handler.ServeHTTP(writer, request.WithContext(ctx))
	} else {
		response = NewHTTPError(http.StatusNotImplemented).ToResponse()
		response.write(writer)
	}
}

// AddRequestHandler creates a new requestHandler for a specified method and path. It is used to set an implementation
// for an endpoint. The function panics, if the endpoint is not specified in the OpenAPI specification
func (router *Router) AddRequestHandler(method string, path string, handleFunc HandleRequestFunction) {
	request, err := http.NewRequest(method, path, nil)
	if err != nil {
		log.Panicln(err)
	}
	route, _, err := router.baseRouter.FindRoute(request)
	if err != nil {
		log.Panicln(err)
	}

	router.implementations[route] = requestHandler{
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
