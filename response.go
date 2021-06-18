package openapirouter

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

// Response is a data struct that depicts the http response to be returned.
type Response struct {
	// http StatusCode to return
	StatusCode int
	// Body of the http request to return. If it is set to string, a plain text response is return. If it is anything
	// else, the response is returned in JSON format.
	Body interface{}
	// http Headers to add to the response
	Headers map[string]string
}

// write is used by the requestHandler and writes the result of the request as an http response.
func (response *Response) write(writer http.ResponseWriter) {
	var err error
	for key, value := range response.Headers {
		writer.Header().Set(key, value)
	}
	switch data := response.Body.(type) {
	case string:
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.WriteHeader(response.StatusCode)
		_, err = io.WriteString(writer, data)
	default:
		if data != nil {
			writer.Header().Set("Content-Type", "application/json; charset=utf-8")
			writer.WriteHeader(response.StatusCode)
			err = json.NewEncoder(writer).Encode(response.Body)
		} else {
			writer.WriteHeader(response.StatusCode)
			_, err = io.WriteString(writer, "")
		}
	}
	if err != nil {
		log.Println("Could not write response", err)
		if response.StatusCode != http.StatusInternalServerError {
			error500Response.write(writer)
		}
	}
}
