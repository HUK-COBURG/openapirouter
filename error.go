package openapirouter

import (
	"net/http"
	"reflect"
	"strconv"
)

const errorTypeUri = "https://developer.mozilla.org/de/docs/Web/HTTP/Status/"

var (
	errorNames = map[int]string{
		400: "Bad Request",
		401: "Unauthorized",
		403: "Forbidden",
		404: "Not found",
		405: "Method not allowed",
		500: "Internal Server Error",
		501: "Not implemented",
		502: "Bad Gateway",
	}
	error500Response = NewHTTPError(http.StatusInternalServerError).ToResponse()
)

// HTTPError implements error and is used to return 4xx/5xx HTTP responses. It can be converted to Response with
// the StatusCode of the HTTPError and HTTPError itself as the Body of the response.
type HTTPError struct {
	// http status code to return
	StatusCode int
	// URL to describe the response code
	Type string `json:"type"`
	// name of the response
	Title string `json:"title"`
	// additional details for the error, e.g. what went wrong
	Details []string `json:"details"`
}

// NewHTTPError creates a new error with a specified statusCode and any number of details. The Title and Type of the
// error is set to correspond with the status code.
func NewHTTPError(statusCode int, details ...string) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Type:       errorTypeUri + strconv.Itoa(statusCode),
		Title:      errorNames[statusCode],
		Details:    details,
	}
}

// implementation of error
func (er *HTTPError) Error() string {
	return er.Title
}

// ToResponse converts the HTTPError to a response to be written.
func (er *HTTPError) ToResponse() *Response {
	return &Response{
		StatusCode: er.StatusCode,
		Body:       *er,
	}
}

// The errorMapper is used to map any error to an HTTP response.
type errorMapper struct {
	errorMapping map[reflect.Type]*HTTPError
}

// mapError receives an error and returns the fitting Response specified by the errorMapping.
func (mapper errorMapper) mapError(err error) *Response {
	var result *HTTPError
	var ok bool
	if result, ok = err.(*HTTPError); !ok {
		result, ok = mapper.errorMapping[reflect.TypeOf(err)]
		if !ok {
			return error500Response
		}
	}
	return result.ToResponse()
}
