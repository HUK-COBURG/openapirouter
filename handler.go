package openapirouter

import (
	"github.com/getkin/kin-openapi/openapi3filter"
	"net/http"
)

type contextKey int

const pathParamsKey contextKey = iota

// HandleRequestFunction is a custom function to specify the implementation of an HTTP endpoint. It does not receive the
// http.ResponseWriter for the request since it is written by the requestHandler. The content of this response is
// specified by the returned Response or the error. If an error is returned, it is mapped to a response by the
// errorMapper of the Router the function was added to.
// The Function does receive the path parameters, because they are already extracted for validation purposes.
type HandleRequestFunction = func(*http.Request, map[string]string) (*Response, error)

// requestHandler implements http.Handler and contains the implementation of an endpoint. It is set to a Route in the
// openapi3filter
type requestHandler struct {
	errMapper       *errorMapper
	handlerFunction HandleRequestFunction
	options         *openapi3filter.Options
}

// implementation of http.Handler that extracts the pathParameters from the request's context and invokes the
// handlerFunction. If an error occurs calling the handlerFunction, it is mapped by the Router's errorMapper
func (handler *requestHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	pathParams, ok := request.Context().Value(pathParamsKey).(map[string]string)
	response := error500Response
	var err error
	if ok {
		response, err = handler.handlerFunction(request, pathParams)
		if err != nil {
			response = handler.errMapper.mapError(err)
		}
	}
	response.write(writer)
}
