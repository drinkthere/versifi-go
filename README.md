# Versifi Go SDK

Official Go SDK for the Versifi algorithmic trading platform. This SDK provides a comprehensive interface for interacting with Versifi's REST API and WebSocket streams.

## Features

- **REST API Support**
  - Create Algo Orders (TWAP, VWAP, IS)
  - Create Basic Orders (MARKET, LIMIT, STOP, etc.)
  - Create Pair Orders (BASIS trading)
  - Cancel Orders (single and batch)
  - Get Order Details

- **WebSocket Support**
  - Real-time order updates
  - Real-time trade updates
  - Automatic reconnection
  - Flexible message handling

- **Design Philosophy**
  - Follows go-binance SDK patterns
  - Type-safe with Go enums
  - Fluent API with builder pattern
  - Comprehensive error handling

## Installation

```bash
go get github.com/versifi/versifi-go-sdk
```

## Quick Start

### Initialize Client

```go
import versifi "github.com/versifi/versifi-go-sdk"

client := versifi.NewClient("your-api-key", "your-api-secret")

// Enable debug mode (optional)
client.Debug = true
```

### Initialize Client with Local IP Binding

If your server has multiple IP addresses and only one is whitelisted by Versifi:

```go
// Bind to specific local IP (whitelisted IP)
localIP := "192.168.1.100"  // Your whitelisted IP address
client := versifi.NewClientWithLocalAddr("your-api-key", "your-api-secret", localIP)

// All requests will now originate from 192.168.1.100
```

📖 **See [LOCAL_IP_BINDING.md](LOCAL_IP_BINDING.md) for detailed documentation on IP binding.**

### Create an Algo Order (TWAP)

```go
params := map[string]interface{}{
    "duration":   3600,    // 1 hour in seconds
    "slice_size": 0.1,     // 10% of total quantity per slice
}

response, err := client.NewCreateAlgoOrderService().
    Exchange(versifi.ExchangeBinanceSpot).
    OrderType(versifi.AlgoOrderTypeTWAP).
    Symbol("BTC/USDT").
    Side(versifi.SideTypeBuy).
    Quantity("1.5").
    Params(params).
    ClientOrderID(123456).
    Do(context.Background())

if err != nil {
    log.Fatal(err)
}

fmt.Printf("Order Created: %d\n", response.OrderID)
```

### Create a Basic Order (LIMIT)

```go
response, err := client.NewCreateBasicOrderService().
    Exchange(versifi.ExchangeBinanceSpot).
    OrderType(versifi.BasicOrderTypeLimit).
    Symbol("BTC/USDT").
    Side(versifi.SideTypeBuy).
    Quantity("0.5").
    Price("45000.00").
    TimeInForce(versifi.TimeInForceGTC).
    Do(context.Background())

if err != nil {
    log.Fatal(err)
}
```

### Create a Pair Order (BASIS)

```go
leadLeg := &versifi.PairLeg{
    Exchange:         versifi.ExchangeBinanceSpot,
    Symbol:           "BTC/USDT",
    OrderType:        "LIMIT",
    LegRatio:         versifi.Float64Ptr(1.0),
    MaxPositionLong:  versifi.StringPtr("100"),
    MaxPositionShort: versifi.StringPtr("50"),
}

secondaryLeg := &versifi.PairLeg{
    Exchange: versifi.ExchangeBinanceFutures,
    Symbol:   "BTC/USDT",
    OrderType: "LIMIT",
    LegRatio: versifi.Float64Ptr(1.0),
}

params := map[string]interface{}{
    "entry_spread_threshold": 0.01,
    "exit_spread_threshold":  0.005,
}

response, err := client.NewCreatePairOrderService().
    OrderType(versifi.PairOrderTypeBasis).
    Lead(leadLeg).
    Secondary(secondaryLeg).
    Style(versifi.PairStyleSync).
    Params(params).
    Do(context.Background())
```

### Get Order Details

```go
response, err := client.NewGetOrderService().
    OrderID(12345).
    Do(context.Background())

if err != nil {
    log.Fatal(err)
}

fmt.Printf("Order Status: %s\n", response.Status)
```

### Cancel Order

```go
err := client.NewCancelOrderService().
    OrderID(12345).
    Do(context.Background())
```

### Cancel Batch Orders

```go
orderIDs := []int64{12345, 12346, 12347}

err := client.NewCancelBatchOrderService().
    OrderIDs(orderIDs).
    Do(context.Background())
```

## WebSocket Usage

### Connect and Subscribe

```go
wsClient := versifi.NewWsClient("your-api-key", "your-api-secret")

// Set error handler
wsClient.SetErrorHandler(func(err error) {
    log.Printf("WebSocket Error: %v", err)
})

// Connect
err := wsClient.Connect()
if err != nil {
    log.Fatal(err)
}
defer wsClient.Disconnect()

// Subscribe to execution reports
wsClient.SubscribeExecutionReport(func(message []byte) {
    var execReport versifi.WsExecutionReport
    json.Unmarshal(message, &execReport)

    fmt.Printf("Order %d: %s\n",
        execReport.Message.OrderID,
        execReport.Message.Status)
})
```

### WebSocket with Local IP Binding

```go
// Bind WebSocket to specific local IP (whitelisted IP)
localIP := "192.168.1.100"
wsClient := versifi.NewWsClientWithLocalAddr("your-api-key", "your-api-secret", localIP)

// All WebSocket connections will now originate from 192.168.1.100
err := wsClient.Connect()
if err != nil {
    log.Fatal(err)
}
defer wsClient.Disconnect()

wsClient.SubscribeExecutionReport(func(message []byte) {
    // Handle execution reports
})
```

## Authentication

The SDK handles authentication automatically using HMAC SHA256 signatures:

- **REST API**: Signs request payload with `X-VERSIFI-API-KEY` and `X-VERSIFI-API-SIGN` headers
- **WebSocket**: Authenticates during connection establishment

### Signature Generation

For **GET/DELETE** requests: Signs the query string (without `?`)
For **POST/PUT** requests: Signs the request body

## API Reference

### Order Types

#### Algo Order Types
- `AlgoOrderTypeTWAP` - Time-Weighted Average Price
- `AlgoOrderTypeVWAP` - Volume-Weighted Average Price
- `AlgoOrderTypeIS` - Implementation Shortfall

#### Basic Order Types
- `BasicOrderTypeMarket` - Market order
- `BasicOrderTypeLimit` - Limit order
- `BasicOrderTypeStop` - Stop order
- `BasicOrderTypeStopLoss` - Stop loss order
- `BasicOrderTypeStopLossLimit` - Stop loss limit order
- `BasicOrderTypeTakeProfit` - Take profit order
- `BasicOrderTypeTakeProfitLimit` - Take profit limit order

#### Pair Order Types
- `PairOrderTypeBasis` - Basis trading (spot-futures arbitrage)

### Exchanges

- `ExchangeBinanceSpot` - Binance Spot
- `ExchangeBinanceFutures` - Binance Futures
- `ExchangeOKXSpot` - OKX Spot
- `ExchangeOKXFutures` - OKX Futures

### Order Sides

- `SideTypeBuy` - Buy order
- `SideTypeSell` - Sell order

### Time In Force

- `TimeInForceFOK` - Fill or Kill
- `TimeInForceGTC` - Good Till Cancel
- `TimeInForceGTD` - Good Till Date
- `TimeInForceIOC` - Immediate or Cancel
- `TimeInForceGTX` - Good Till Crossing
- `TimeInForcePostOn` - Post Only

### Order Status

- `OrderStatusNew` - Order created
- `OrderStatusPartiallyFilled` - Partially filled
- `OrderStatusFilled` - Completely filled
- `OrderStatusCanceled` - Canceled
- `OrderStatusRejected` - Rejected
- `OrderStatusExpired` - Expired

### Pair Order Styles

- `PairStyleSync` - Both legs executed synchronously
- `PairStyleAsync` - First leg passive, second leg aggressive on fill
- `PairStyleTWAP` - Each leg executed independently using TWAP

## Algorithm Details

### TWAP (Time-Weighted Average Price)

Distributes orders evenly over time to minimize market impact.

**Parameters:**
- `duration` - Total execution window in seconds
- `slice_size` - Percentage of total quantity per slice
- `time_interval` - Frequency of order placement

### VWAP (Volume-Weighted Average Price)

Executes orders in proportion to market volume.

**Parameters:**
- `duration` - Execution window
- `volume_percentage` - Target percentage of market volume
- `aggressiveness` - Execution speed vs. market impact

### IS (Implementation Shortfall)

Minimizes slippage from decision time to execution.

**Parameters:**
- `duration` - Required execution duration
- `urgency_level` - Speed of execution
- `max_participation_rate` - Maximum market volume percentage

### BASIS Trading

Exploits price differences between spot and futures markets.

**Parameters:**
- `entry_spread_threshold` - Minimum spread to enter position
- `exit_spread_threshold` - Spread level to exit position
- `max_slippage` - Maximum acceptable slippage
- `max_drawdown` - Maximum drawdown before stopping

## Error Handling

```go
response, err := client.NewCreateAlgoOrderService().
    // ... parameters ...
    Do(context.Background())

if err != nil {
    if versifi.IsAPIError(err) {
        apiErr := err.(*versifi.APIError)
        fmt.Printf("API Error %d: %s\n", apiErr.Code, apiErr.Message)
    } else {
        fmt.Printf("Error: %v\n", err)
    }
    return
}
```

## Examples

Complete examples are available in the `examples/` directory:

- `examples/main.go` - REST API examples
- `examples/websocket_example.go` - WebSocket examples

Run examples:

```bash
cd examples
go run main.go
go run websocket_example.go
```

## Testing

```bash
go test ./...
```

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## Support

For issues and questions:

- GitHub Issues: https://github.com/versifi/versifi-go-sdk/issues
- Documentation: https://docs.versifi.io

## License

MIT License - see LICENSE file for details

## Acknowledgments

This SDK follows the design patterns established by the excellent [go-binance](https://github.com/adshao/go-binance) library.
