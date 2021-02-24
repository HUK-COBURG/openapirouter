package openapirouter

import (
	"context"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

const responseText = "test"

func handleRequest(_ *http.Request, _ map[string]string) (*Response, error) {
	return &Response{
		StatusCode: http.StatusOK,
		Body:       responseText,
	}, nil
}

var handler = &requestHandler{handlerFunction: handleRequest}

func TestRequestHandler_ShouldInvokeHandlerFunction(t *testing.T) {
	// given
	ctx := context.WithValue(context.Background(), pathParamsKey, make(map[string]string))
	request := httptest.NewRequest("GET", "/", nil).WithContext(ctx)
	recorder := httptest.NewRecorder()

	// when
	handler.ServeHTTP(recorder, request)

	//then
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, responseText, recorder.Body.String())
}

func TestRequestHandler_ShouldReturnInternalServerError_PathParamsNotMapped(t *testing.T) {
	// given
	request := httptest.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()

	// when
	handler.ServeHTTP(recorder, request)

	//then
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}
