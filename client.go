package versifi

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

// Global configuration
var (
	BaseAPIMainURL = "https://api.versifi.io"
	UseTestnet     = false
)

// Security type
type secType int

const (
	secTypeNone secType = iota
	secTypeAPIKey
	secTypeSigned
)

// Client represents the Versifi API client
type Client struct {
	APIKey     string
	APISecret  string
	BaseURL    string
	UserAgent  string
	HTTPClient *http.Client
	Debug      bool
	Logger     *log.Logger
	do         doFunc
}

type doFunc func(req *http.Request) (*http.Response, error)

// NewClient creates a new Versifi client
func NewClient(apiKey, apiSecret string) *Client {
	return &Client{
		APIKey:     apiKey,
		APISecret:  apiSecret,
		BaseURL:    getAPIEndpoint(),
		UserAgent:  "Versifi/go",
		HTTPClient: http.DefaultClient,
		Logger:     log.New(os.Stderr, "Versifi-go ", log.LstdFlags),
	}
}

// NewClientWithHTTPClient creates a new client with custom HTTP client
func NewClientWithHTTPClient(apiKey, apiSecret string, httpClient *http.Client) *Client {
	return &Client{
		APIKey:     apiKey,
		APISecret:  apiSecret,
		BaseURL:    getAPIEndpoint(),
		UserAgent:  "Versifi/go",
		HTTPClient: httpClient,
		Logger:     log.New(os.Stderr, "Versifi-go ", log.LstdFlags),
	}
}

// NewClientWithLocalAddr creates a new client that binds to a specific local IP address
// This is useful when the server has multiple IP addresses but only one is whitelisted
func NewClientWithLocalAddr(apiKey, apiSecret, localAddr string) *Client {
	// Parse the local address
	localTCPAddr, err := net.ResolveTCPAddr("tcp", localAddr+":0")
	if err != nil {
		// If resolution fails, fall back to standard client
		log.Printf("Warning: failed to resolve local address %s: %v", localAddr, err)
		return NewClient(apiKey, apiSecret)
	}

	// Create a custom dialer that binds to the specified local address
	dialer := &net.Dialer{
		LocalAddr: localTCPAddr,
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	// Create HTTP transport with the custom dialer
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// Create HTTP client with custom transport
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	return &Client{
		APIKey:     apiKey,
		APISecret:  apiSecret,
		BaseURL:    getAPIEndpoint(),
		UserAgent:  "Versifi/go",
		HTTPClient: httpClient,
		Logger:     log.New(os.Stderr, "Versifi-go ", log.LstdFlags),
	}
}

func getAPIEndpoint() string {
	if UseTestnet {
		return BaseAPIMainURL // Versifi doesn't have separate testnet, adjust if needed
	}
	return BaseAPIMainURL
}

// callAPI executes the HTTP request
func (c *Client) callAPI(ctx context.Context, r *request, opts ...RequestOption) (data []byte, err error) {
	err = c.parseRequest(r, opts...)
	if err != nil {
		return []byte{}, err
	}

	req, err := http.NewRequest(r.method, r.fullURL, r.body)
	if err != nil {
		return []byte{}, err
	}

	req = req.WithContext(ctx)
	req.Header = r.header

	c.debug("request: %#v", req)

	f := c.do
	if f == nil {
		f = c.HTTPClient.Do
	}

	res, err := f(req)
	if err != nil {
		return []byte{}, err
	}

	data, err = io.ReadAll(res.Body)
	if err != nil {
		return []byte{}, err
	}
	defer func() {
		closeErr := res.Body.Close()
		if closeErr != nil {
			c.debug("failed to close response body: %v", closeErr)
		}
	}()

	c.debug("response: %#v", res)
	c.debug("response body: %s", string(data))
	c.debug("response status code: %d", res.StatusCode)

	if res.StatusCode >= http.StatusBadRequest {
		apiErr := new(APIError)
		e := json.Unmarshal(data, apiErr)
		if e != nil {
			c.debug("failed to unmarshal json: %s", e)
		}
		return nil, apiErr
	}

	return data, nil
}

// parseRequest parses the request and sets authentication headers
func (c *Client) parseRequest(r *request, opts ...RequestOption) (err error) {
	// Set request options
	for _, opt := range opts {
		opt(r)
	}

	// Build full URL
	if r.query != nil && len(r.query) > 0 {
		r.fullURL = fmt.Sprintf("%s%s?%s", c.BaseURL, r.endpoint, r.query.Encode())
	} else {
		r.fullURL = fmt.Sprintf("%s%s", c.BaseURL, r.endpoint)
	}

	// Set headers
	if r.header == nil {
		r.header = http.Header{}
	}

	r.header.Set("User-Agent", c.UserAgent)
	r.header.Set("Content-Type", "application/json")

	// Authentication
	if r.secType == secTypeAPIKey || r.secType == secTypeSigned {
		r.header.Set("X-VERSIFI-API-KEY", c.APIKey)
	}

	if r.secType == secTypeSigned {
		var payload string

		// For GET and DELETE requests, payload is the query string (without "?")
		if r.method == http.MethodGet || r.method == http.MethodDelete {
			if r.query != nil && len(r.query) > 0 {
				payload = r.query.Encode()
			}
		} else {
			// For POST and PUT requests, payload is the body
			if r.body != nil {
				bodyBytes, err := io.ReadAll(r.body)
				if err != nil {
					return err
				}
				payload = string(bodyBytes)
				// Reset the body reader
				r.body = bytes.NewReader(bodyBytes)
			}
		}

		// Create signature
		signature := c.sign(payload)
		r.header.Set("X-VERSIFI-API-SIGN", signature)
	}

	return nil
}

// sign creates HMAC SHA256 signature
func (c *Client) sign(payload string) string {
	key := []byte(c.APISecret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

func (c *Client) debug(format string, v ...interface{}) {
	if c.Debug {
		c.Logger.Printf(format, v...)
	}
}

// Service factory methods

// NewCreateAlgoOrderService creates a new CreateAlgoOrderService
func (c *Client) NewCreateAlgoOrderService() *CreateAlgoOrderService {
	return &CreateAlgoOrderService{c: c}
}

// NewCreateBasicOrderService creates a new CreateBasicOrderService
func (c *Client) NewCreateBasicOrderService() *CreateBasicOrderService {
	return &CreateBasicOrderService{c: c}
}

// NewCreatePairOrderService creates a new CreatePairOrderService
func (c *Client) NewCreatePairOrderService() *CreatePairOrderService {
	return &CreatePairOrderService{c: c}
}

// NewCancelOrderService creates a new CancelOrderService
func (c *Client) NewCancelOrderService() *CancelOrderService {
	return &CancelOrderService{c: c}
}

// NewGetOrderService creates a new GetOrderService
func (c *Client) NewGetOrderService() *GetOrderService {
	return &GetOrderService{c: c}
}

// NewGetOrderService creates a new GetOrderService
func (c *Client) NewListOpenOrdersService() *ListOpenOrdersService {
	return &ListOpenOrdersService{c: c}
}

// NewCancelBatchOrderService creates a new CancelBatchOrderService
func (c *Client) NewCancelBatchOrderService() *CancelBatchOrderService {
	return &CancelBatchOrderService{c: c}
}
