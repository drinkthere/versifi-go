package versifi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocket configuration
var (
	BaseWSMainURL      = "wss://example.com/v1/ws" // Update with actual production URL
	WebsocketTimeout   = time.Second * 60
	WebsocketKeepalive = true
)

// WsHandler handles websocket messages
type WsHandler func(message []byte)

// ErrHandler handles websocket errors
type ErrHandler func(err error)

// WsClient represents a websocket client
type WsClient struct {
	APIKey         string
	APISecret      string
	BaseURL        string
	LocalAddr      string // Local IP address to bind to (optional)
	conn           *websocket.Conn
	mu             sync.RWMutex
	isConnected    bool
	isAuthenticated bool
	handlers       map[string]WsHandler
	errHandler     ErrHandler
	done           chan struct{}
	reconnect      bool
	reconnectDelay time.Duration
	Logger         *log.Logger
}

// NewWsClient creates a new websocket client
func NewWsClient(apiKey, apiSecret string) *WsClient {
	return &WsClient{
		APIKey:         apiKey,
		APISecret:      apiSecret,
		BaseURL:        getWSEndpoint(),
		handlers:       make(map[string]WsHandler),
		done:           make(chan struct{}),
		reconnect:      true,
		reconnectDelay: 5 * time.Second,
		Logger:         log.Default(),
	}
}

// NewWsClientWithLocalAddr creates a new websocket client that binds to a specific local IP address
// This is useful when the server has multiple IP addresses but only one is whitelisted
func NewWsClientWithLocalAddr(apiKey, apiSecret, localAddr string) *WsClient {
	return &WsClient{
		APIKey:         apiKey,
		APISecret:      apiSecret,
		BaseURL:        getWSEndpoint(),
		LocalAddr:      localAddr,
		handlers:       make(map[string]WsHandler),
		done:           make(chan struct{}),
		reconnect:      true,
		reconnectDelay: 5 * time.Second,
		Logger:         log.Default(),
	}
}

func getWSEndpoint() string {
	if UseTestnet {
		return BaseWSMainURL // Update if testnet has different URL
	}
	return BaseWSMainURL
}

// Connect establishes websocket connection and authenticates
func (c *WsClient) Connect() error {
	c.mu.Lock()
	if c.isConnected {
		c.mu.Unlock()
		return fmt.Errorf("already connected")
	}
	c.mu.Unlock()

	// Create websocket dialer
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 45 * time.Second

	// If local address is specified, configure the dialer to bind to it
	if c.LocalAddr != "" {
		localTCPAddr, err := net.ResolveTCPAddr("tcp", c.LocalAddr+":0")
		if err != nil {
			c.Logger.Printf("Warning: failed to resolve local address %s: %v", c.LocalAddr, err)
		} else {
			// Create custom net dialer with local address binding
			netDialer := &net.Dialer{
				LocalAddr: localTCPAddr,
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}
			dialer.NetDial = netDialer.Dial
			c.Logger.Printf("WebSocket binding to local address: %s", c.LocalAddr)
		}
	}

	// Dial websocket (no headers needed for initial connection)
	conn, _, err := dialer.Dial(c.BaseURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.isConnected = true
	c.mu.Unlock()

	// Start reading messages
	go c.readMessages()

	// Authenticate after connection
	if err := c.authenticate(); err != nil {
		c.Disconnect()
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Start keepalive if enabled
	if WebsocketKeepalive {
		go c.keepAlive()
	}

	return nil
}

// authenticate sends authentication message to the server
func (c *WsClient) authenticate() error {
	// Calculate expiration timestamp (e.g., 5 minutes from now)
	expires := time.Now().Add(5 * time.Minute).Unix()

	// Create payload for signature: "GET/realtime{expires}"
	payload := fmt.Sprintf("GET/realtime%d", expires)

	// Generate signature
	signature := c.sign(payload)

	// Send authentication message
	authMsg := map[string]interface{}{
		"op": "auth",
		"args": []interface{}{
			c.APIKey,
			fmt.Sprintf("%d", expires),
			signature,
		},
	}

	c.Logger.Printf("Sending authentication message...")

	if err := c.SendJSON(authMsg); err != nil {
		return err
	}

	// Wait for authentication response (with timeout)
	authResponse := make(chan error, 1)
	tempHandler := func(message []byte) {
		var resp WsResponse
		if err := json.Unmarshal(message, &resp); err != nil {
			authResponse <- fmt.Errorf("failed to parse auth response: %w", err)
			return
		}

		if resp.Op == "auth" {
			if resp.Success {
				c.mu.Lock()
				c.isAuthenticated = true
				c.mu.Unlock()
				c.Logger.Printf("Authentication successful")
				authResponse <- nil
			} else {
				authResponse <- fmt.Errorf("authentication failed: %v", resp.Message)
			}
		}
	}

	// Temporarily store handler for auth response
	c.mu.Lock()
	c.handlers["__auth__"] = tempHandler
	c.mu.Unlock()

	// Wait for auth response or timeout
	select {
	case err := <-authResponse:
		c.mu.Lock()
		delete(c.handlers, "__auth__")
		c.mu.Unlock()
		return err
	case <-time.After(10 * time.Second):
		c.mu.Lock()
		delete(c.handlers, "__auth__")
		c.mu.Unlock()
		return fmt.Errorf("authentication timeout")
	}
}

// Disconnect closes the websocket connection
func (c *WsClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isConnected {
		return nil
	}

	c.reconnect = false
	close(c.done)

	if c.conn != nil {
		err := c.conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		)
		if err != nil {
			c.Logger.Printf("error sending close message: %v", err)
		}

		err = c.conn.Close()
		if err != nil {
			return fmt.Errorf("failed to close connection: %w", err)
		}
	}

	c.isConnected = false
	c.isAuthenticated = false
	c.conn = nil

	return nil
}

// Subscribe subscribes to a specific topic
func (c *WsClient) Subscribe(topic string, handler WsHandler) error {
	c.mu.RLock()
	if !c.isAuthenticated {
		c.mu.RUnlock()
		return fmt.Errorf("not authenticated")
	}
	c.mu.RUnlock()

	c.mu.Lock()
	c.handlers[topic] = handler
	c.mu.Unlock()

	// Send subscription message
	subscribeMsg := map[string]interface{}{
		"op":   "subscribe",
		"args": []string{topic},
	}

	return c.SendJSON(subscribeMsg)
}

// Unsubscribe unsubscribes from a specific topic
func (c *WsClient) Unsubscribe(topic string) error {
	c.mu.Lock()
	delete(c.handlers, topic)
	c.mu.Unlock()

	// Send unsubscription message (if needed)
	// Note: Versifi docs don't specify unsubscribe operation
	return nil
}

// SubscribeExecutionReport subscribes to execution_report topic
func (c *WsClient) SubscribeExecutionReport(handler WsHandler) error {
	return c.Subscribe("execution_report", handler)
}

// SubscribeAnalytics subscribes to analytics topic (not implemented on server yet)
func (c *WsClient) SubscribeAnalytics(handler WsHandler) error {
	return c.Subscribe("analytics", handler)
}

// SetErrorHandler sets the error handler
func (c *WsClient) SetErrorHandler(handler ErrHandler) {
	c.errHandler = handler
}

// SendJSON sends a JSON message
func (c *WsClient) SendJSON(v interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.isConnected || c.conn == nil {
		return fmt.Errorf("not connected")
	}

	return c.conn.WriteJSON(v)
}

// SendPing sends a ping message
func (c *WsClient) SendPing() error {
	pingMsg := map[string]string{
		"op": "ping",
	}
	return c.SendJSON(pingMsg)
}

// readMessages reads messages from websocket
func (c *WsClient) readMessages() {
	defer func() {
		c.mu.Lock()
		c.isConnected = false
		c.isAuthenticated = false
		c.mu.Unlock()

		// Attempt reconnection if enabled
		if c.reconnect {
			c.Logger.Printf("connection lost, attempting to reconnect in %v", c.reconnectDelay)
			time.Sleep(c.reconnectDelay)
			if err := c.Connect(); err != nil {
				c.Logger.Printf("reconnection failed: %v", err)
				if c.errHandler != nil {
					c.errHandler(err)
				}
			}
		}
	}()

	for {
		select {
		case <-c.done:
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				c.Logger.Printf("error reading message: %v", err)
				if c.errHandler != nil {
					c.errHandler(err)
				}
				return
			}

			c.Logger.Printf("Received message: %s", string(message))

			// Parse message to determine operation type
			var wsResp WsResponse
			if err := json.Unmarshal(message, &wsResp); err != nil {
				c.Logger.Printf("error unmarshaling message: %v", err)
				continue
			}

			// Handle special operations
			if wsResp.Op == "auth" {
				// Check if there's an auth handler
				c.mu.RLock()
				handler, exists := c.handlers["__auth__"]
				c.mu.RUnlock()

				if exists && handler != nil {
					handler(message)
				}
				continue
			}

			if wsResp.Op == "ping" {
				c.Logger.Printf("Received pong response")
				continue
			}

			if wsResp.Op == "subscribe" {
				c.Logger.Printf("Subscription confirmed: %v", wsResp.Message)
				continue
			}

			// Handle execution_report messages
			if wsResp.Op == "execution_report" {
				c.mu.RLock()
				handler, exists := c.handlers["execution_report"]
				c.mu.RUnlock()

				if exists && handler != nil {
					handler(message)
				}

				// Also call wildcard handler if exists
				c.mu.RLock()
				wildcardHandler, exists := c.handlers["*"]
				c.mu.RUnlock()

				if exists && wildcardHandler != nil {
					wildcardHandler(message)
				}
				continue
			}

			// Handle other topics
			c.mu.RLock()
			handler, exists := c.handlers[wsResp.Op]
			c.mu.RUnlock()

			if exists && handler != nil {
				handler(message)
			} else {
				// Call wildcard handler
				c.mu.RLock()
				wildcardHandler, exists := c.handlers["*"]
				c.mu.RUnlock()

				if exists && wildcardHandler != nil {
					wildcardHandler(message)
				}
			}
		}
	}
}

// keepAlive sends periodic ping messages
func (c *WsClient) keepAlive() {
	ticker := time.NewTicker(WebsocketTimeout / 2)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			c.mu.RLock()
			isConnected := c.isConnected
			c.mu.RUnlock()

			if !isConnected {
				return
			}

			// Send application-level ping (not WebSocket ping)
			if err := c.SendPing(); err != nil {
				c.Logger.Printf("error sending ping: %v", err)
				if c.errHandler != nil {
					c.errHandler(err)
				}
				return
			}
		}
	}
}

// sign creates HMAC SHA256 signature
func (c *WsClient) sign(payload string) string {
	key := []byte(c.APISecret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

// IsConnected returns the connection status
func (c *WsClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isConnected
}

// IsAuthenticated returns the authentication status
func (c *WsClient) IsAuthenticated() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isAuthenticated
}

// WebSocket message types

// WsResponse represents a general websocket response
type WsResponse struct {
	Op      string      `json:"op"`
	Success bool        `json:"success"`
	Message interface{} `json:"message,omitempty"`
}

// WsExecutionReport represents the execution_report message
type WsExecutionReport struct {
	Op      string                  `json:"op"`
	Success bool                    `json:"success"`
	Message WsExecutionReportDetail `json:"message"`
}

// WsExecutionReportDetail represents the detail of execution report
type WsExecutionReportDetail struct {
	OrderID          int64           `json:"order_id"`
	ClientOrderID    int64           `json:"client_order_id"`
	OrderType        string          `json:"order_type"`
	Status           OrderStatusType `json:"status"`
	Timestamp        int64           `json:"timestamp"`
	RequestOrderType string          `json:"request_order_type"`
	Order            interface{}     `json:"order"` // Can be BasicOrder, AlgoOrder, or PairOrder
}

// WsBasicOrderDetail represents a basic order in execution report
type WsBasicOrderDetail struct {
	QuoteOrderQuantity string         `json:"quote_order_quantity,omitempty"`
	Symbol             string         `json:"symbol"`
	ClientOrderID      int64          `json:"client_order_id"`
	StopPrice          string         `json:"stop_price,omitempty"`
	Exchange           ExchangeType   `json:"exchange"`
	Price              string         `json:"price,omitempty"`
	Quantity           string         `json:"quantity"`
	Side               SideType       `json:"side"`
	OrderType          BasicOrderType `json:"order_type"`
	ChildOrder         *WsChildOrder  `json:"child_order,omitempty"`
}

// WsAlgoOrderDetail represents an algo order in execution report
type WsAlgoOrderDetail struct {
	ID                 int64          `json:"id"`
	Exchange           ExchangeType   `json:"exchange"`
	OrderType          AlgoOrderType  `json:"order_type"`
	Quantity           string         `json:"quantity"`
	QuoteOrderQuantity string         `json:"quote_order_quantity,omitempty"`
	Side               SideType       `json:"side"`
	Symbol             string         `json:"symbol"`
	OrderParams        interface{}    `json:"order_params,omitempty"`
	ChildOrder         *WsChildOrder  `json:"child_order,omitempty"`
}

// WsPairOrderDetail represents a pair order in execution report
type WsPairOrderDetail struct {
	Params   interface{}    `json:"params,omitempty"`
	LeadLeg  *WsPairLeg     `json:"lead_leg,omitempty"`
	Leg      *WsPairLeg     `json:"leg,omitempty"`
}

// WsPairLeg represents a leg in pair order
type WsPairLeg struct {
	Symbol           string         `json:"symbol"`
	Exchange         ExchangeType   `json:"exchange"`
	OrderType        string         `json:"order_type"`
	LegRatio         float64        `json:"leg_ratio"`
	MaxPositionLong  string         `json:"max_position_long,omitempty"`
	MaxPositionShort string         `json:"max_position_short,omitempty"`
	MaxNotionalLong  string         `json:"max_notional_long,omitempty"`
	MaxNotionalShort string         `json:"max_notional_short,omitempty"`
	ChildOrder       *WsChildOrder  `json:"child_order,omitempty"`
}

// WsChildOrder represents child order with trades
type WsChildOrder struct {
	ID     int64     `json:"id"`
	Trades []WsTrade `json:"trades"`
}

// WsTrade represents a trade execution with extended fields
type WsTrade struct {
	TradeID                    int64  `json:"trade_id"`
	AveragePrice               string `json:"average_price,omitempty"`
	CummulativeFilledQuantity  string `json:"cummulative_filled_quantity,omitempty"`
	OrderID                    int64  `json:"order_id"`
	LegID                      *int64 `json:"leg_id,omitempty"` // Only for pair orders
	ExecutedPrice              string `json:"executed_price"`
	ExecutedQuantity           string `json:"executed_quantity"`
}
