package openapirouter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func getRouterAndServer() (*Router, *httptest.Server) {
	result, err := NewRouter("testdata/test-api.yaml")
	if err != nil {
		panic(err)
	}
	return result, httptest.NewServer(result)
}

type TestData struct {
	Data string `json:"data"`
}

func TestRouter_ValidGETWithoutParams(t *testing.T) {
	// given
	router, server := getRouterAndServer()
	defer server.Close()
	called := false
	router.AddRequestHandler("GET", "/test", func(_ *http.Request, _ map[string]string) (*Response, error) {
		called = true
		return &Response{
			StatusCode: http.StatusOK,
			Body:       TestData{Data: "test"},
		}, nil
	})

	// when
	res, err := server.Client().Get(server.URL + "/test")

	// then
	assert.Nil(t, err)
	assert.True(t, called)
	if assert.NotNil(t, res) {
		assert.Equal(t, http.StatusOK, res.StatusCode)
	}
}

func TestRouter_ValidGETWithQuery(t *testing.T) {
	// given
	router, server := getRouterAndServer()
	defer server.Close()
	called := false
	router.AddRequestHandler("GET", "/test/query", func(r *http.Request, _ map[string]string) (*Response, error) {
		called = true
		assert.Equal(t, "value1", r.URL.Query().Get("param"))
		return &Response{
			StatusCode: http.StatusOK,
			Body:       TestData{Data: "test"},
		}, nil
	})

	// when
	res, err := server.Client().Get(server.URL + "/test/query?param=value1")

	// then
	assert.Nil(t, err)
	assert.True(t, called)
	if assert.NotNil(t, res) {
		assert.Equal(t, http.StatusOK, res.StatusCode)
	}
}

func TestRouter_ValidGETWithPathParam(t *testing.T) {
	// given
	router, server := getRouterAndServer()
	defer server.Close()
	called := false
	router.AddRequestHandler("GET", "/test/pathParams/{param}", func(_ *http.Request, pathParams map[string]string) (*Response, error) {
		called = true
		if assert.Contains(t, pathParams, "param") {
			assert.Equal(t, "value1", pathParams["param"])
		}
		return &Response{
			StatusCode: http.StatusOK,
			Body:       TestData{Data: "test"},
		}, nil
	})

	// when
	res, err := server.Client().Get(server.URL + "/test/pathParams/value1")

	// then
	assert.Nil(t, err)
	assert.True(t, called)
	if assert.NotNil(t, res) {
		assert.Equal(t, http.StatusOK, res.StatusCode)
	}
}

func TestRouter_GETWrongQuery(t *testing.T) {
	// given
	router, server := getRouterAndServer()
	defer server.Close()
	called := false
	router.AddRequestHandler("GET", "/test/query", func(_ *http.Request, _ map[string]string) (*Response, error) {
		called = true
		return &Response{
			StatusCode: http.StatusOK,
			Body:       TestData{Data: "test"},
		}, nil
	})

	// when
	res, err := server.Client().Get(server.URL + "/test/query?param=invalid")

	// then
	assert.Nil(t, err)
	assert.False(t, called)
	if assert.NotNil(t, res) {
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	}
}

func TestRouter_GETInvalidPath(t *testing.T) {
	// given
	router, server := getRouterAndServer()
	defer server.Close()
	called := false
	router.AddRequestHandler("GET", "/test/pathParams/{param}", func(_ *http.Request, pathParams map[string]string) (*Response, error) {
		called = true
		return &Response{
			StatusCode: http.StatusOK,
			Body:       TestData{Data: "test"},
		}, nil
	})

	// when
	res, err := server.Client().Get(server.URL + "/test/pathParams/invalid")

	// then
	assert.Nil(t, err)
	assert.False(t, called)
	if assert.NotNil(t, res) {
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	}
}

func TestRouter_ValidPOST(t *testing.T) {
	// given
	router, server := getRouterAndServer()
	defer server.Close()
	called := false
	sendData := TestData{
		Data: "test",
	}
	dataBytes, _ := json.Marshal(&sendData)
	router.AddRequestHandler("POST", "/test", func(r *http.Request, _ map[string]string) (*Response, error) {
		called = true
		var rcvData TestData
		_ = json.NewDecoder(r.Body).Decode(&rcvData)
		assert.Equal(t, sendData, rcvData)
		return &Response{
			StatusCode: http.StatusNoContent,
		}, nil
	})

	// when
	res, err := server.Client().Post(server.URL+"/test", "application/json", bytes.NewReader(dataBytes))

	// then
	assert.Nil(t, err)
	if assert.NotNil(t, res) {
		assert.True(t, called)
		assert.Equal(t, http.StatusNoContent, res.StatusCode)
	}
}

func TestRouter_POSTInvalidPath(t *testing.T) {
	// given
	router, server := getRouterAndServer()
	defer server.Close()
	called := false
	sendData := TestData{
		Data: "test",
	}
	dataBytes, _ := json.Marshal(&sendData)
	router.AddRequestHandler("GET", "/test/query", func(r *http.Request, _ map[string]string) (*Response, error) {
		called = true
		var rcvData TestData
		_ = json.NewDecoder(r.Body).Decode(&rcvData)
		assert.Equal(t, sendData, rcvData)
		return &Response{
			StatusCode: http.StatusNoContent,
		}, nil
	})

	// when
	res, err := server.Client().Post(server.URL+"/test/query", "application/json", bytes.NewReader(dataBytes))

	// then
	assert.Nil(t, err)
	if assert.NotNil(t, res) {
		assert.False(t, called)
		assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)
	}
}

type InvalidData struct {
	Invalid string `json:"invalid"`
}

func TestRouter_POSTInvalidData(t *testing.T) {
	// given
	router, server := getRouterAndServer()
	defer server.Close()
	called := false
	sendData := InvalidData{
		Invalid: "test",
	}
	dataBytes, _ := json.Marshal(&sendData)
	router.AddRequestHandler("POST", "/test", func(r *http.Request, _ map[string]string) (*Response, error) {
		called = true
		return &Response{
			StatusCode: http.StatusNoContent,
		}, nil
	})

	// when
	res, err := server.Client().Post(server.URL+"/test", "application/json", bytes.NewReader(dataBytes))

	// then
	assert.Nil(t, err)
	if assert.NotNil(t, res) {
		assert.False(t, called)
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	}
}

func TestRouter_POSTandGETSamePath(t *testing.T) {
	//given
	router, server := getRouterAndServer()
	defer server.Close()
	postCalled := false
	getCalled := false
	sendData := TestData{
		Data: "test",
	}
	dataBytes, _ := json.Marshal(&sendData)
	router.AddRequestHandler(http.MethodPost, "/test", func(_ *http.Request, _ map[string]string) (*Response, error) {
		postCalled = true
		return &Response{
			StatusCode: http.StatusNoContent,
		}, nil
	})
	router.AddRequestHandler(http.MethodGet, "/test", func(_ *http.Request, _ map[string]string) (*Response, error) {
		getCalled = true
		return &Response{
			StatusCode: http.StatusOK,
			Body: &TestData{
				Data: "test",
			},
		}, nil
	})

	// when
	postResponse, postErr := server.Client().Post(server.URL+"/test", "application/json", bytes.NewReader(dataBytes))
	getResponse, getErr := server.Client().Get(server.URL + "/test")

	// then
	assert.Nil(t, postErr)
	if assert.NotNil(t, postResponse) {
		assert.True(t, postCalled)
		assert.Equal(t, http.StatusNoContent, postResponse.StatusCode)
	}

	assert.Nil(t, getErr)
	if assert.NotNil(t, getResponse) {
		assert.True(t, getCalled)
		assert.Equal(t, http.StatusOK, getResponse.StatusCode)
	}
}

func TestRouter_InvalidPath(t *testing.T) {
	// given
	_, server := getRouterAndServer()
	defer server.Close()

	// when
	res, err := server.Client().Get(server.URL + "/invalid")

	// then
	assert.Nil(t, err)
	if assert.NotNil(t, res) {
		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	}
}

func TestRouter_NotImplementedPath(t *testing.T) {
	// given
	_, server := getRouterAndServer()
	defer server.Close()

	// when
	res, err := server.Client().Get(server.URL + "/test")

	// then
	assert.Nil(t, err)
	if assert.NotNil(t, res) {
		assert.Equal(t, http.StatusNotImplemented, res.StatusCode)
	}
}

func TestRouter_SecuredPath_AllowingAuthFunc(t *testing.T) {
	// given
	router, server := getRouterAndServer()
	defer server.Close()
	router.AddRequestHandlerWithAuthFunc("GET", "/test/secured", func(_ *http.Request, _ map[string]string) (*Response, error) {
		return &Response{
			StatusCode: http.StatusOK,
		}, nil
	}, openapi3filter.NoopAuthenticationFunc)
	// when
	res, err := server.Client().Get(server.URL + "/test/secured")

	// then
	assert.Nil(t, err)
	if assert.NotNil(t, res) {
		assert.Equal(t, http.StatusOK, res.StatusCode)
	}
}

func TestRouter_SecuredPath_DenyingAuthFunc(t *testing.T) {
	// given
	router, server := getRouterAndServer()
	defer server.Close()
	called := false
	router.AddRequestHandlerWithAuthFunc("GET", "/test/secured", func(_ *http.Request, _ map[string]string) (*Response, error) {
		called = true
		return &Response{
			StatusCode: http.StatusOK,
		}, nil
	}, func(_x context.Context, _ *openapi3filter.AuthenticationInput) error {
		return errors.New("not authorized")
	})
	// when
	res, err := server.Client().Get(server.URL + "/test/secured")

	// then
	assert.Nil(t, err)
	assert.False(t, called)
	if assert.NotNil(t, res) {
		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	}
}

func TestRouter_SecuredPath_NoAuthFunc(t *testing.T) {
	// given
	router, server := getRouterAndServer()
	defer server.Close()
	called := false
	router.AddRequestHandler("GET", "/test/secured", func(_ *http.Request, _ map[string]string) (*Response, error) {
		called = true
		return &Response{
			StatusCode: http.StatusOK,
		}, nil
	})
	// when
	res, err := server.Client().Get(server.URL + "/test/secured")

	// then
	assert.Nil(t, err)
	assert.False(t, called)
	if assert.NotNil(t, res) {
		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	}
}

func TestAddRoute_RouteNotDocumented(t *testing.T) {
	// given
	router, server := getRouterAndServer()
	defer server.Close()

	// when + then
	assert.Panics(t, func() {
		router.AddRequestHandler("GET", "/invalid", func(_ *http.Request, _ map[string]string) (*Response, error) {
			return error500Response, nil
		})
	})
}
