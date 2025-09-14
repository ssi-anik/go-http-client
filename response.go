package go_http_client

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
)

type HttpResponse interface {
	Original() *http.Response
	StatusCode() int
	IsSuccess() bool
	IsServerError() bool
	IsClientError() bool
	Headers() http.Header
	HasHeader(key string) bool
	GetHeader(key string) (string, bool)
	IsJsonResponse() bool
	Content() []byte
	ParseJson() (map[string]interface{}, error)
	ParseAs(dest interface{}) error
}

type httpResponse struct {
	original   *http.Response
	statusCode int
	body       []byte
	headers    http.Header
}

func (r *httpResponse) Original() *http.Response {
	return r.original
}

func (r *httpResponse) StatusCode() int {
	return r.statusCode
}

func (r *httpResponse) IsSuccess() bool {
	return r.statusCode >= 200 && r.statusCode < 300
}

func (r *httpResponse) IsServerError() bool {
	return r.statusCode >= 500
}

func (r *httpResponse) IsClientError() bool {
	return r.statusCode >= 400 && r.statusCode < 500
}

func (r *httpResponse) Headers() http.Header {
	return r.headers
}

func (r *httpResponse) HasHeader(key string) bool {
	_, ok := r.GetHeader(key)

	return ok
}

func (r *httpResponse) GetHeader(key string) (string, bool) {
	h, ok := r.headers[strings.ToLower(key)]
	if !ok {
		return "", false
	}

	return h[0], true
}

func (r *httpResponse) IsJsonResponse() bool {
	v, ok := r.GetHeader("content-type")
	if !ok || v != "application/json" {
		return false
	}

	return true
}

func (r *httpResponse) Content() []byte {
	return r.body
}

func (r *httpResponse) ParseJson() (map[string]interface{}, error) {
	if !r.IsJsonResponse() {
		return nil, errors.New("not a json response")
	}

	var m map[string]interface{}
	err := json.Unmarshal(r.body, &m)

	return m, err
}

func (r *httpResponse) ParseAs(dest interface{}) error {
	if nil == dest {
		return errors.New("dest is nil")
	}

	if len(r.body) == 0 {
		return errors.New("body is empty")
	}

	if reflect.ValueOf(dest).Kind() != reflect.Ptr {
		return errors.New("dest is not a pointer")
	}

	if !r.IsJsonResponse() {
		return errors.New("not a json response")
	}

	if _, ok := dest.(json.Unmarshaler); !ok {
		return errors.New("dest does not implement json.Unmarshaler")
	}

	return json.Unmarshal(r.body, dest)
}

func newHttpResponse(response *http.Response) (HttpResponse, error) {
	statusCode := response.StatusCode

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	headers := make(http.Header)
	for k, v := range response.Header {
		headers[strings.ToLower(k)] = v
	}

	return &httpResponse{
		original:   response,
		statusCode: statusCode,
		body:       body,
		headers:    headers,
	}, nil
}
