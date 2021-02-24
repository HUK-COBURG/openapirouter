package openapirouter

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteResponse_ShouldWritePlainTextResponse(t *testing.T) {
	// given
	testMessage := "test"
	recorder := httptest.NewRecorder()
	response := &Response{
		Body:       testMessage,
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"X-TEST": testMessage},
	}

	//when
	response.write(recorder)

	//then
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, testMessage, recorder.Body.String())
	assert.Equal(t, testMessage, recorder.Header().Get("X-TEST"))
	assert.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
}

func TestWriteResponse_ShouldWriteJsonResponse(t *testing.T) {
	// given
	testData := TestData{Data: "bla"}
	recorder := httptest.NewRecorder()
	response := &Response{
		Body:       testData,
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"X-TEST": "test"},
	}

	//when
	response.write(recorder)

	//then
	assert.Equal(t, http.StatusOK, recorder.Code)
	expectedBody, _ := json.Marshal(testData)
	assert.Equal(t, strings.TrimSpace(string(expectedBody)), strings.TrimSpace(recorder.Body.String()))
	assert.Equal(t, "test", recorder.Header().Get("X-TEST"))
	assert.Equal(t, "application/json; charset=utf-8", recorder.Header().Get("Content-Type"))
}

func TestWriteResponse_ShouldWriteEmptyResponse(t *testing.T) {
	// given
	recorder := httptest.NewRecorder()
	response := &Response{
		StatusCode: http.StatusNoContent,
	}

	//when
	response.write(recorder)

	//then
	assert.Equal(t, http.StatusNoContent, recorder.Code)
	assert.Empty(t, recorder.Body.String())
	assert.Empty(t, recorder.Header().Get("Content-Type"))
}
