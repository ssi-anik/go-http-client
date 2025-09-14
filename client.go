package go_http_client

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HttpClient interface {
	Transport(http.RoundTripper) HttpClient
	GetTransport() http.RoundTripper
	Host(string) HttpClient
	GetHost() string
	UrlPrefix(string) HttpClient
	GetUrlPrefix() string
	MaxRedirects(int) HttpClient
	GetMaxRedirects() int
	Timeout(time.Duration) HttpClient
	GetTimeout() time.Duration
	UserAgent(string) HttpClient
	GetUserAgent() string
	WithDefaultHeaders(map[string][]string) HttpClient
	DefaultHeaders() http.Header
	WithDefaultQueries(map[string][]string) HttpClient
	DefaultQueries() url.Values
	NewHttpRequest() HttpRequest
}

type httpClient struct {
	transport      http.RoundTripper
	host           string
	urlPrefix      string
	maxRedirects   int
	timeout        time.Duration
	userAgent      string
	defaultHeaders http.Header
	defaultQueries url.Values
}

func (c *httpClient) Transport(transport http.RoundTripper) HttpClient {
	c.transport = transport

	return c
}

func (c *httpClient) GetTransport() http.RoundTripper {
	return c.transport
}

func (c *httpClient) Host(host string) HttpClient {
	c.host = strings.TrimSpace(host)

	return c
}

func (c *httpClient) GetHost() string {
	url, err := url.Parse(c.host)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%s://%s", url.Scheme, url.Host)
}

func (c *httpClient) UrlPrefix(prefix string) HttpClient {
	c.urlPrefix = strings.TrimSpace(prefix)

	return c
}

func (c *httpClient) GetUrlPrefix() string {
	prefix := c.urlPrefix
	if prefix == "" {
		return ""
	}

	return fmt.Sprintf("/%s", strings.Trim(prefix, "/"))
}

func (c *httpClient) MaxRedirects(maxRedirects int) HttpClient {
	c.maxRedirects = maxRedirects

	return c
}

func (c *httpClient) GetMaxRedirects() int {
	return c.maxRedirects
}

func (c *httpClient) Timeout(timeout time.Duration) HttpClient {
	c.timeout = timeout

	return c
}

func (c *httpClient) GetTimeout() time.Duration {
	return c.timeout
}

func (c *httpClient) UserAgent(ua string) HttpClient {
	c.userAgent = ua

	return c
}

func (c *httpClient) GetUserAgent() string {
	return c.userAgent
}

func (c *httpClient) WithDefaultHeaders(headers map[string][]string) HttpClient {
	c.defaultHeaders = headers

	return c
}

func (c *httpClient) DefaultHeaders() http.Header {
	return c.defaultHeaders
}

func (c *httpClient) WithDefaultQueries(queries map[string][]string) HttpClient {
	c.defaultQueries = queries

	return c
}

func (c *httpClient) DefaultQueries() url.Values {
	return c.defaultQueries
}

func (c *httpClient) NewHttpRequest() HttpRequest {
	return NewHttpRequest(c)
}

func NewHttpClient(config *ClientConfig) (HttpClient, error) {
	if nil == config {
		return nil, errors.New("config is nil")
	}

	return &httpClient{
		transport:      config.Transport,
		host:           config.Host,
		urlPrefix:      config.UrlPrefix,
		maxRedirects:   config.MaxRedirects,
		timeout:        config.Timeout,
		userAgent:      config.UserAgent,
		defaultHeaders: make(http.Header),
		defaultQueries: make(url.Values),
	}, nil
}

func HttpClientFor(host string, prefixes ...string) (HttpClient, error) {
	client, err := NewHttpClient(DefaultClientConfig)
	if nil != err {
		return nil, err
	}

	client.Host(host)
	if len(prefixes) > 0 {
		client.UrlPrefix(prefixes[0])
	}

	return client, nil
}
