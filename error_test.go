package openapirouter

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"reflect"
	"testing"
)

func TestErrorMapper_ShouldMapHttpErrorToResponseByDefault(t *testing.T) {
	//given
	mapper := errorMapper{errorMapping: make(map[reflect.Type]*HTTPError)}
	err := NewHTTPError(http.StatusNotFound)

	//when
	result := mapper.mapError(err)

	//then
	assert.Equal(t, http.StatusNotFound, result.StatusCode)
	assert.Equal(t, *NewHTTPError(http.StatusNotFound), result.Body)
}

type ExampleError struct {
}

func (e *ExampleError) Error() string {
	return "test error"
}

func TestErrorMapper_ShouldMapKnownErrorToResponse(t *testing.T) {
	//given
	mapper := errorMapper{
		errorMapping: map[reflect.Type]*HTTPError{
			reflect.TypeOf(&ExampleError{}): NewHTTPError(http.StatusBadGateway),
		},
	}
	err := &ExampleError{}

	//when
	result := mapper.mapError(err)

	//then
	assert.Equal(t, http.StatusBadGateway, result.StatusCode)
	assert.Equal(t, *NewHTTPError(http.StatusBadGateway), result.Body)
}

func TestErrorMapper_ShouldMapKnownErrorToResponseWithDetails(t *testing.T) {
	//given
	mapper := errorMapper{
		errorMapping: map[reflect.Type]*HTTPError{
			reflect.TypeOf(&ExampleError{}): NewHTTPError(http.StatusBadGateway, "detail1", "detail2"),
		},
	}
	err := &ExampleError{}

	//when
	result := mapper.mapError(err)

	//then
	assert.Equal(t, http.StatusBadGateway, result.StatusCode)
	assert.Equal(t, *NewHTTPError(http.StatusBadGateway, "detail1", "detail2"), result.Body)
}

func TestErrorMapper_ShouldMapKnownErrorToInternalServerError(t *testing.T) {
	//given
	mapper := errorMapper{errorMapping: make(map[reflect.Type]*HTTPError)}
	err := &ExampleError{}

	//when
	result := mapper.mapError(err)

	//then
	assert.Equal(t, http.StatusInternalServerError, result.StatusCode)
	assert.Equal(t, *NewHTTPError(http.StatusInternalServerError), result.Body)
}
