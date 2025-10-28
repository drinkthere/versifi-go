package versifi

import (
	"io"
	"net/http"
	"net/url"
)

type request struct {
	method   string
	endpoint string
	query    url.Values
	header   http.Header
	body     io.Reader
	fullURL  string
	secType  secType
}

// setParam sets a query parameter
func (r *request) setParam(key string, value string) *request {
	if r.query == nil {
		r.query = url.Values{}
	}
	r.query.Set(key, value)
	return r
}

// setParams sets multiple query parameters
func (r *request) setParams(m params) *request {
	for k, v := range m {
		r.setParam(k, v)
	}
	return r
}

type params map[string]interface{}

// RequestOption defines a function that modifies a request
type RequestOption func(*request)

// WithHeader sets a custom header
func WithHeader(key, value string) RequestOption {
	return func(r *request) {
		if r.header == nil {
			r.header = http.Header{}
		}
		r.header.Set(key, value)
	}
}

// WithHeaders sets multiple custom headers
func WithHeaders(headers map[string]string) RequestOption {
	return func(r *request) {
		if r.header == nil {
			r.header = http.Header{}
		}
		for k, v := range headers {
			r.header.Set(k, v)
		}
	}
}
