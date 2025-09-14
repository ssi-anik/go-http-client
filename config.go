package go_http_client

import (
	"net/http"
	"time"
)

var (
	httpClientVersion = "0.9"
	defaultUserAgent  = "anik;go-http-client/" + httpClientVersion
)

type ClientConfig struct {
	Transport    http.RoundTripper
	Host         string
	UrlPrefix    string
	MaxRedirects int
	Timeout      time.Duration
	UserAgent    string
}

var DefaultClientConfig = &ClientConfig{
	Transport:    nil,
	Host:         "",
	UrlPrefix:    "",
	MaxRedirects: 10,
	Timeout:      60 * time.Second,
	UserAgent:    defaultUserAgent,
}
