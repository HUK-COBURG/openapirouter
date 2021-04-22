[![Go](https://github.com/HUK-COBURG/openapirouter/actions/workflows/go.yml/badge.svg)](https://github.com/HUK-COBURG/openapirouter/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/huk-coburg/openapirouter)](https://goreportcard.com/report/github.com/huk-coburg/openapirouter)
[![GoDoc](https://godoc.org/github.com/huk-coburg/openapirouter?status.svg)](https://godoc.org/github.com/huk-coburg/openapirouter)

# OpenAPI-Router
The OpenAPI-Router is a "Contract-First" http-Router, specifically designed for JSON-based REST-Services. It takes an
[OpenAPI schema](https://swagger.io/specification/) as input to create the router. The router validates requests and 
maps them to their handler method using [kin-openapi](https://github.com/getkin/kin-openapi/). Additionally, the router
simplifies request writing and error handling for JSON-based REST-Services.

## Features
- HTTP-Router with automatic OpenAPI validation
- Implementation `http.Handler` to be compatible with existing HTTP libraries
- Automatic response writing of JSON or plain-text responses
- ErrorMapper to write helpful responses based on the type of error

## How to use
### Installation
```shell
go get github.com/huk-coburg/openapirouter
```

### Creating the router
In order to create the router, a file with the OpenAPI specification is needed. The file can be in JSON or YAML format.
The router is created using the following `NewRouter` function with the path of the OpenAPI file.

### Handler function
To enable the automatic response writing and error mapping, a custom handler function different from the standard 
`http.HandlerFunc` is used for the implementation of endpoints. The following function signature is used:  
```go
func(request *http.Request, pathParamerters map[string]string) (*openapirouter.Response, error)
```  
The parameters are:
- **request:** The pointer to the http.Request as in the standard `http.HandlerFunc`. It can be used to extract the 
  request body, the headers or query parameters.
- **pathParameters:** A map of the path parameters which are extracted for validation and are populated to the request,
  so they don't need to be extracted manually for the URL.
- **Response:** A struct to depict the response to be returned. It is used to set the response body, the status and the 
  response headers.
- **error:** Standard Go error to indicate that an error occurred.

The `AddRequestHandler` function of the router is used to add a function for a specific path and method to the router.

If an endpoint defines a security requirement, `AddRequestHandlerWithAuthFunc` must be used in order to enable the 
router to check if the user is authorized to access the endpoint. Using the `openapi3filter.NoopAuthenticationFunc` 
as `authFunc` will grant access for any request without further checks. 

### Error handling
If the handler function returns an error, it will be mapped to a corresponding response. Therefore, a custom `HTTPError`
is used and returned as a JSON response. It contains the following fields:
```go
type HTTPError struct {
	// HTTP status code to return
	StatusCode int
	// URL to describe the response code
	Type       string
	// name of the response
	Title      string
	// additional details for the error, e.g. what went wrong
	Details    []string  
}
```  
If an `HTTPError` is returned by the handler function, it will be mapped to a corresponding response by default. The 
`NewError` function is used to create such an error for a status code with any number of details.

It is also possible to map any other error produced by the handler function to a response. Unknown errors are mapped to 
an `Internal Server Error` by default. In order to create a different response, the error needs to be added to the 
routers' error mapper by using the `AddErrorMapping` function to define the `HTTPError` it should be mapped to.

### Full Example

```go
package main

import (
	"github.com/huk-coburg/openapirouter"
	"net/http"
)

// Struct for the data to be returned
type MyOutputData struct {
	Data string `json:"data"`
}

// Custom error to be mapped by error mapper
type MyCustomError struct {
}

func (e *MyCustomError) Error() string {
	return "oops"
}

// Example function which can produce an error
func GetDataForClient(client string) (*MyOutputData, error) {
	var data *MyOutputData
	var success bool
	// ...
	if success {
		return data, nil
	} else {
		return nil, &MyCustomError{}
	}
}

// HandlerFunction to assign to an endpoint
func HandleRequest(_ *http.Request, pathParams map[string]string) (*openapirouter.Response, error) {
	// Extract path parameter
	client := pathParams["client"]
	if client == "unknown" {
		// Return HttpError which is mapped automatically
		return nil, openapirouter.NewHTTPError(http.StatusForbidden, "unknown client must not receive data")
	}
	data, err := GetDataForClient(client)
	if err != nil {
		// Return custom error to be mapped
		return nil, err
	}
	// Return response to be written
	return &openapirouter.Response{
		StatusCode: http.StatusOK,
		Body:       data,
	}, nil

}

func main() {
	router, err := openapirouter.NewRouter("./test-api.yaml")
	if err != nil {
		// Could not read file
		panic(err)
	}
	// Add implementation for endpoint specified in OpenAPI specification
	// Use router.AddRequestHandlerWithAuthFunc if the endpoint defines security requirements
	router.AddRequestHandler(http.MethodGet, "/test/{client}", HandleRequest)
	// Add error mapping for MyCustomError
	router.AddErrorMapping(&MyCustomError{}, http.StatusBadGateway, "could not load data")
	// use router
	_ = http.ListenAndServe(":8080", router)
}
```

## Contribute
Contributions are welcome. You can:
- Submit bugs and feature requests as issues
- Review code and let us know what we can do better
- Submit PullRequests for code or documentation changes or additions