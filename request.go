package go_http_client

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HttpRequest interface {
	WithContext(context.Context) HttpRequest
	UserAgent(ua string) HttpRequest
	MaxRedirects(int) HttpRequest
	NoRedirect() HttpRequest
	Timeout(time.Duration) HttpRequest
	Headers(http.Header) HttpRequest
	AddHeader(k string, v string) HttpRequest
	SkipDefaultHeaders() HttpRequest
	Queries(url.Values) HttpRequest
	SkipDefaultQueries() HttpRequest
	Body([]byte) HttpRequest
	Method(string) HttpRequest
	Path(string) HttpRequest
	Send(method string, path string, body []byte, headers http.Header) (HttpResponse, error)
	Get(path ...string) (HttpResponse, error)
	Post(path ...string) (HttpResponse, error)
	Put(path ...string) (HttpResponse, error)
	Patch(path ...string) (HttpResponse, error)
	Delete(path ...string) (HttpResponse, error)
	Submit() (HttpResponse, error)
}

type httpRequest struct {
	client             HttpClient
	userAgent          *string
	ctx                context.Context
	maxRedirects       *int
	timeout            *time.Duration
	headers            http.Header
	queries            url.Values
	skipDefaultQueries bool
	body               []byte
	skipDefaultHeaders bool
	method             string
	path               string
}

func (r *httpRequest) WithContext(ctx context.Context) HttpRequest {
	r.ctx = ctx

	return r
}

func (r *httpRequest) UserAgent(ua string) HttpRequest {
	r.userAgent = &ua

	return r
}

func (r *httpRequest) MaxRedirects(mr int) HttpRequest {
	r.maxRedirects = &mr

	return r
}

func (r *httpRequest) NoRedirect() HttpRequest {
	mr := 0
	r.maxRedirects = &mr

	return r
}

func (r *httpRequest) Timeout(t time.Duration) HttpRequest {
	r.timeout = &t

	return r
}

func (r *httpRequest) NoTimeout() HttpRequest {
	r.timeout = nil

	return r
}

func (r *httpRequest) Headers(headers http.Header) HttpRequest {
	r.headers = headers

	return r
}

func (r *httpRequest) AddHeader(key, value string) HttpRequest {
	r.headers.Add(key, value)

	return r
}

func (r *httpRequest) SkipDefaultHeaders() HttpRequest {
	r.skipDefaultHeaders = true

	return r
}

func (r *httpRequest) Queries(queries url.Values) HttpRequest {
	r.queries = queries

	return r
}

func (r *httpRequest) SkipDefaultQueries() HttpRequest {
	r.skipDefaultQueries = true

	return r
}

func (r *httpRequest) Body(b []byte) HttpRequest {
	r.body = b

	return r
}

func (r *httpRequest) Path(p string) HttpRequest {
	r.path = strings.TrimSpace(p)

	return r
}

func (r *httpRequest) Method(m string) HttpRequest {
	r.method = strings.ToUpper(m)

	return r
}

func (r *httpRequest) Get(path ...string) (HttpResponse, error) {
	r.Method(http.MethodGet)

	if len(path) > 0 {
		r.Path(path[0])
	}

	return r.Submit()
}

func (r *httpRequest) Post(path ...string) (HttpResponse, error) {
	r.Method(http.MethodPost)

	if len(path) > 0 {
		r.Path(path[0])
	}

	return r.Submit()
}

func (r *httpRequest) Put(path ...string) (HttpResponse, error) {
	r.Method(http.MethodPut)

	if len(path) > 0 {
		r.Path(path[0])
	}

	return r.Submit()
}

func (r *httpRequest) Patch(path ...string) (HttpResponse, error) {
	r.Method(http.MethodPatch)

	if len(path) > 0 {
		r.Path(path[0])
	}

	return r.Submit()
}

func (r *httpRequest) Delete(path ...string) (HttpResponse, error) {
	r.Method(http.MethodDelete)

	if len(path) > 0 {
		r.Path(path[0])
	}

	return r.Submit()
}

func (r *httpRequest) Send(method string, path string, body []byte, headers http.Header) (HttpResponse, error) {
	r.Method(method)
	r.Path(path)
	r.Body(body)
	r.Headers(headers)

	return r.Submit()
}

func (r *httpRequest) Submit() (HttpResponse, error) {
	host := r.client.GetHost()
	prefix := r.client.GetUrlPrefix()

	queries := r.queries
	path := ""
	if r.path != "" {
		path = strings.TrimPrefix(r.path, "/")
		u, err := url.Parse(path)
		if err != nil {
			return nil, err
		}

		if u.Host != "" {
			host = fmt.Sprintf("%s://%s", u.Scheme, u.Host)
		}

		for k, v := range u.Query() {
			for _, each := range v {
				queries.Add(k, each)
			}
		}

		path = fmt.Sprintf("/%s", u.Path)
	}

	if !r.skipDefaultQueries {
		for k, v := range r.client.DefaultQueries() {
			for _, each := range v {
				queries.Add(k, each)
			}
		}
	}

	qp := ""
	if len(queries) > 0 {
		qp = fmt.Sprintf("?%s", queries.Encode())
	}

	url := fmt.Sprintf("%s%s%s%s", host, prefix, path, qp)

	req, err := http.NewRequest(r.method, url, bytes.NewBuffer(r.body))
	if err != nil {
		return nil, err
	}

	if r.ctx != nil {
		req = req.WithContext(r.ctx)
	}

	if r.userAgent != nil {
		r.AddHeader("User-Agent", *r.userAgent)
	} else if ua := r.client.GetUserAgent(); ua != "" {
		r.AddHeader("User-Agent", ua)
	}

	headers := r.headers
	if !r.skipDefaultHeaders {
		for k, v := range r.client.DefaultHeaders() {
			for _, each := range v {
				headers.Add(k, each)
			}
		}
	}

	for k, v := range headers {
		for _, each := range v {
			req.Header.Add(k, each)
		}
	}

	timeout := r.client.GetTimeout()
	if r.timeout != nil {
		timeout = *r.timeout
	}

	maxRedirects := r.client.GetMaxRedirects()
	if r.maxRedirects != nil {
		maxRedirects = *r.maxRedirects
	}

	client := &http.Client{
		Transport: r.client.GetTransport(),
		Timeout:   timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if maxRedirects <= 0 {
				return TooManyRedirects
			}

			maxRedirects--

			return nil
		},
	}

	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return newHttpResponse(response)
}

func NewHttpRequest(client HttpClient) HttpRequest {
	return &httpRequest{
		client:  client,
		headers: make(http.Header),
		queries: make(url.Values),
	}
}
