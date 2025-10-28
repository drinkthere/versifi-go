# Versifi Go SDK - Project Structure

## Overview

This document provides an overview of the Versifi Go SDK project structure, following the design patterns of go-binance.

## Directory Structure

```
versifi-go-sdk/
├── client.go                    # Main client implementation
├── common.go                    # Common types, enums, and error handling
├── request.go                   # HTTP request builder and utilities
├── websocket.go                 # WebSocket client implementation
├── order_algo.go                # Algo order service (TWAP, VWAP, IS)
├── order_basic.go               # Basic order service (MARKET, LIMIT, etc.)
├── order_pair.go                # Pair order service (BASIS trading)
├── order_cancel.go              # Cancel order service
├── order_cancel_batch.go        # Batch cancel order service
├── order_get.go                 # Get order details service
├── go.mod                       # Go module dependencies
├── .gitignore                   # Git ignore rules
├── README.md                    # User documentation
├── PROJECT_STRUCTURE.md         # This file
└── examples/
    ├── main.go                  # REST API usage examples
    └── websocket_example.go     # WebSocket usage examples
```

## File Descriptions

### Core Files

#### `client.go`
- **Purpose**: Main client implementation with authentication and HTTP request handling
- **Key Components**:
  - `Client` struct: Holds API credentials, HTTP client, and configuration
  - `NewClient()`: Client factory function
  - `callAPI()`: Executes HTTP requests with authentication
  - `parseRequest()`: Prepares requests and adds authentication headers
  - `sign()`: Creates HMAC SHA256 signatures
  - Service factory methods: `NewCreateAlgoOrderService()`, etc.

#### `common.go`
- **Purpose**: Shared types, constants, and utilities
- **Key Components**:
  - `APIError`: Error type for API responses
  - Enums: `SideType`, `ExchangeType`, `AlgoOrderType`, `BasicOrderType`, etc.
  - `OrderResponse`: Common response structure
  - Helper functions: `StringPtr()`, `Int64Ptr()`, `Float64Ptr()`

#### `request.go`
- **Purpose**: HTTP request building utilities
- **Key Components**:
  - `request` struct: Represents an HTTP request
  - `params` type: Map for request parameters
  - `RequestOption`: Function type for request customization
  - Helper functions: `setParam()`, `setParams()`, `WithHeader()`

#### `websocket.go`
- **Purpose**: WebSocket client for real-time data streaming
- **Key Components**:
  - `WsClient` struct: WebSocket connection manager
  - `NewWsClient()`: WebSocket client factory
  - `Connect()`, `Disconnect()`: Connection management
  - `Subscribe()`, `Unsubscribe()`: Channel subscription
  - `SubscribeOrders()`, `SubscribeTrades()`: Convenience methods
  - `WsOrderUpdate`, `WsTradeUpdate`: Message types
  - Automatic reconnection logic
  - Keepalive/ping mechanism

### Service Files

#### `order_algo.go`
- **Purpose**: Create algorithmic orders (TWAP, VWAP, IS)
- **Key Components**:
  - `CreateAlgoOrderService`: Service struct with builder pattern
  - Setter methods: `Exchange()`, `OrderType()`, `Symbol()`, `Quantity()`, etc.
  - `Params()`: Sets algorithm-specific parameters
  - `Do()`: Executes the order creation request
  - `AlgoOrderRequest`: Request body structure

#### `order_basic.go`
- **Purpose**: Create basic orders (MARKET, LIMIT, STOP, etc.)
- **Key Components**:
  - `CreateBasicOrderService`: Service struct
  - Setter methods: `Price()`, `StopPrice()`, `TimeInForce()`, `TrailingDelta()`, etc.
  - `Do()`: Executes the order creation request
  - `BasicOrderRequest`: Request body structure

#### `order_pair.go`
- **Purpose**: Create pair orders for basis trading
- **Key Components**:
  - `CreatePairOrderService`: Service struct
  - `PairLeg`: Leg configuration structure
  - Setter methods: `Lead()`, `Secondary()`, `Style()`, `Params()`
  - `Do()`: Executes the pair order creation request
  - `PairOrderRequest`: Request body structure

#### `order_cancel.go`
- **Purpose**: Cancel a single order by ID
- **Key Components**:
  - `CancelOrderService`: Service struct
  - `OrderID()`: Sets the order ID to cancel
  - `Do()`: Executes the cancellation (returns HTTP 204)

#### `order_cancel_batch.go`
- **Purpose**: Cancel multiple orders in a single request
- **Key Components**:
  - `CancelBatchOrderService`: Service struct
  - `OrderIDs()`: Sets multiple order IDs
  - `AddOrderID()`: Adds a single order ID
  - `Do()`: Executes batch cancellation
  - `CancelBatchRequest`: Request body structure

#### `order_get.go`
- **Purpose**: Retrieve order details by ID
- **Key Components**:
  - `GetOrderService`: Service struct
  - `OrderID()`: Sets the order ID to retrieve
  - `Do()`: Executes the get request
  - `GetOrderResponse`: Comprehensive response structure
  - `AlgoOrderDetail`, `BasicOrderDetail`, `PairOrderDetail`: Order-type-specific details
  - `ChildOrder`, `Trade`: Execution details

### Example Files

#### `examples/main.go`
- **Purpose**: Demonstrates REST API usage
- **Examples**:
  - Creating algo orders (TWAP, VWAP)
  - Creating basic orders (LIMIT, MARKET)
  - Creating pair orders (BASIS)
  - Getting order details
  - Canceling orders (single and batch)

#### `examples/websocket_example.go`
- **Purpose**: Demonstrates WebSocket usage
- **Examples**:
  - Connecting to WebSocket
  - Subscribing to order updates
  - Subscribing to trade updates
  - Handling real-time messages
  - Error handling
  - Graceful shutdown

## Design Patterns

### 1. Builder Pattern (Fluent API)
All service structs use method chaining for setting parameters:

```go
client.NewCreateAlgoOrderService().
    Exchange(versifi.ExchangeBinanceSpot).
    OrderType(versifi.AlgoOrderTypeTWAP).
    Symbol("BTC/USDT").
    Quantity("1.5").
    Do(ctx)
```

### 2. Service Factory Pattern
The client creates service instances through factory methods:

```go
func (c *Client) NewCreateAlgoOrderService() *CreateAlgoOrderService {
    return &CreateAlgoOrderService{c: c}
}
```

### 3. Functional Options Pattern
Request options allow flexible customization:

```go
type RequestOption func(*request)

func WithHeader(key, value string) RequestOption {
    return func(r *request) {
        r.header.Set(key, value)
    }
}
```

### 4. Type Safety with Named Types
Enums are implemented as typed strings:

```go
type SideType string

const (
    SideTypeBuy  SideType = "BUY"
    SideTypeSell SideType = "SELL"
)
```

### 5. Pointer-Based Optional Parameters
Optional fields use pointers to distinguish between "not set" and "zero value":

```go
type CreateBasicOrderService struct {
    price     *string  // nil = not set, "0" = explicitly set to zero
    stopPrice *string
}
```

## Authentication Flow

### REST API
1. Build request with method, endpoint, and parameters
2. For GET/DELETE: Sign the query string (without `?`)
3. For POST/PUT: Sign the request body
4. Add headers:
   - `X-VERSIFI-API-KEY`: API key
   - `X-VERSIFI-API-SIGN`: HMAC SHA256 signature (hex-encoded)
5. Execute HTTP request

### WebSocket
1. Generate timestamp and signature
2. Connect with authentication headers:
   - `X-VERSIFI-API-KEY`: API key
   - `X-VERSIFI-API-SIGN`: HMAC SHA256 signature
   - `X-VERSIFI-TIMESTAMP`: Unix timestamp
3. Maintain connection with periodic pings
4. Auto-reconnect on disconnection

## Error Handling

### API Errors
```go
type APIError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}
```

Errors are parsed from HTTP responses with status code >= 400.

### WebSocket Errors
Errors are handled through the `ErrHandler` callback:

```go
wsClient.SetErrorHandler(func(err error) {
    log.Printf("WebSocket Error: %v", err)
})
```

## Testing Strategy

### Unit Tests
- Test individual service methods
- Mock HTTP responses
- Verify request building and authentication

### Integration Tests
- Test against sandbox/testnet environment
- Verify full request/response cycle
- Test WebSocket connections

### Example Tests Structure
```
versifi-go-sdk/
├── client_test.go
├── order_algo_test.go
├── order_basic_test.go
├── order_pair_test.go
├── websocket_test.go
└── common_test.go
```

## Dependencies

### Production Dependencies
- `github.com/gorilla/websocket` - WebSocket client

### Development Dependencies
- Standard Go testing library
- (Add mocking libraries as needed)

## Extension Points

### Adding New Order Types
1. Create `order_<type>.go` file
2. Define service struct with builder methods
3. Add factory method to `client.go`
4. Update `common.go` with new enums/types

### Adding New WebSocket Channels
1. Define message structure in `websocket.go`
2. Add convenience subscription method
3. Update examples

### Supporting Additional Exchanges
1. Add new exchange enum to `common.go`
2. Update validation logic if needed
3. Test with new exchange

## Version History

### v1.0.0 (Current)
- Initial release
- REST API support for all order operations
- WebSocket support for real-time updates
- Complete documentation and examples

## Future Enhancements

### Planned Features
- [ ] Rate limiting management
- [ ] Request retry logic with exponential backoff
- [ ] Context-based timeouts
- [ ] Pagination support for list operations
- [ ] Batch order creation
- [ ] Account information endpoints
- [ ] Position management endpoints
- [ ] Historical data endpoints
- [ ] Comprehensive test coverage

### Under Consideration
- [ ] gRPC support (if Versifi adds gRPC endpoints)
- [ ] Prometheus metrics integration
- [ ] OpenTelemetry tracing
- [ ] Connection pooling for HTTP client
- [ ] Request/response logging middleware

## Contributing

See `README.md` for contribution guidelines.

## License

MIT License - See LICENSE file for details.
