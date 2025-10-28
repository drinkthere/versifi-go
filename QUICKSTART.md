# Versifi Go SDK - Quick Start Guide

Get started with the Versifi Go SDK in 5 minutes!

## Prerequisites

- Go 1.21 or later
- Versifi API credentials (API key and secret)

## Installation

```bash
go get github.com/versifi/versifi-go-sdk
```

## Basic Setup

### 1. Import the Package

```go
import versifi "github.com/versifi/versifi-go-sdk"
```

### 2. Initialize the Client

```go
apiKey := "your-api-key"
apiSecret := "your-api-secret"

client := versifi.NewClient(apiKey, apiSecret)
```

### 3. (Optional) Enable Debug Mode

```go
client.Debug = true  // See request/response details
```

## Common Use Cases

### Use Case 1: Execute a TWAP Algorithm

**Goal**: Buy 1.5 BTC over 1 hour using TWAP

```go
package main

import (
    "context"
    "fmt"
    "log"
    versifi "github.com/versifi/versifi-go-sdk"
)

func main() {
    client := versifi.NewClient("your-api-key", "your-api-secret")

    // Configure TWAP parameters
    params := map[string]interface{}{
        "duration":   3600,    // 1 hour
        "slice_size": 0.1,     // 10% per slice
    }

    // Create TWAP order
    response, err := client.NewCreateAlgoOrderService().
        Exchange(versifi.ExchangeBinanceSpot).
        OrderType(versifi.AlgoOrderTypeTWAP).
        Symbol("BTC/USDT").
        Side(versifi.SideTypeBuy).
        Quantity("1.5").
        Params(params).
        Do(context.Background())

    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("✅ TWAP Order Created - ID: %d\n", response.OrderID)
}
```

### Use Case 2: Place a Limit Order

**Goal**: Buy 0.5 BTC at $45,000

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

fmt.Printf("✅ Limit Order Created - ID: %d\n", response.OrderID)
```

### Use Case 3: Basis Trading (Spot-Futures Arbitrage)

**Goal**: Execute a basis trade between spot and futures

```go
// Configure lead leg (spot)
leadLeg := &versifi.PairLeg{
    Exchange:         versifi.ExchangeBinanceSpot,
    Symbol:           "BTC/USDT",
    OrderType:        "LIMIT",
    LegRatio:         versifi.Float64Ptr(1.0),
    MaxPositionLong:  versifi.StringPtr("100"),
    MaxPositionShort: versifi.StringPtr("50"),
}

// Configure secondary leg (futures)
secondaryLeg := &versifi.PairLeg{
    Exchange: versifi.ExchangeBinanceFutures,
    Symbol:   "BTC/USDT",
    OrderType: "LIMIT",
    LegRatio: versifi.Float64Ptr(1.0),
}

// Basis trading parameters
params := map[string]interface{}{
    "entry_spread_threshold": 0.01,   // Enter when spread > 1%
    "exit_spread_threshold":  0.005,  // Exit when spread < 0.5%
    "max_slippage":           0.002,  // Max 0.2% slippage
}

response, err := client.NewCreatePairOrderService().
    OrderType(versifi.PairOrderTypeBasis).
    Lead(leadLeg).
    Secondary(secondaryLeg).
    Style(versifi.PairStyleSync).
    Params(params).
    Do(context.Background())

if err != nil {
    log.Fatal(err)
}

fmt.Printf("✅ Basis Order Created - ID: %d\n", response.OrderID)
```

### Use Case 4: Monitor Orders in Real-Time

**Goal**: Get real-time order updates via WebSocket

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "os/signal"
    versifi "github.com/versifi/versifi-go-sdk"
)

func main() {
    wsClient := versifi.NewWsClient("your-api-key", "your-api-secret")

    // Connect to WebSocket
    err := wsClient.Connect()
    if err != nil {
        log.Fatal(err)
    }
    defer wsClient.Disconnect()

    fmt.Println("✅ Connected to WebSocket")

    // Subscribe to order updates
    wsClient.SubscribeOrders(func(message []byte) {
        var order versifi.WsOrderUpdate
        json.Unmarshal(message, &order)

        fmt.Printf("📊 Order %d: %s (Filled: %s/%s)\n",
            order.OrderID,
            order.Status,
            order.FilledQty,
            order.Quantity)
    })

    // Subscribe to trade executions
    wsClient.SubscribeTrades(func(message []byte) {
        var trade versifi.WsTradeUpdate
        json.Unmarshal(message, &trade)

        fmt.Printf("💰 Trade %d: %s %s @ %s\n",
            trade.TradeID,
            trade.Side,
            trade.Quantity,
            trade.Price)
    })

    fmt.Println("👂 Listening for updates... (Ctrl+C to exit)")

    // Wait for interrupt
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt)
    <-sigChan
}
```

### Use Case 5: Check Order Status

**Goal**: Get detailed information about an order

```go
orderID := int64(12345)

response, err := client.NewGetOrderService().
    OrderID(orderID).
    Do(context.Background())

if err != nil {
    log.Fatal(err)
}

fmt.Printf("Order ID: %d\n", response.OrderID)
fmt.Printf("Status: %s\n", response.Status)
fmt.Printf("Type: %s\n", response.OrderType)

// Check order type and print specific details
if response.AlgoOrder != nil {
    fmt.Printf("Algorithm: %s\n", response.AlgoOrder.OrderType)
    fmt.Printf("Symbol: %s\n", response.AlgoOrder.Symbol)
    fmt.Printf("Side: %s\n", response.AlgoOrder.Side)
    fmt.Printf("Quantity: %s\n", response.AlgoOrder.Quantity)
}
```

### Use Case 6: Cancel Orders

**Goal**: Cancel one or multiple orders

```go
// Cancel a single order
err := client.NewCancelOrderService().
    OrderID(12345).
    Do(context.Background())

if err != nil {
    log.Fatal(err)
}
fmt.Println("✅ Order canceled")

// Cancel multiple orders
orderIDs := []int64{12345, 12346, 12347}

err = client.NewCancelBatchOrderService().
    OrderIDs(orderIDs).
    Do(context.Background())

if err != nil {
    log.Fatal(err)
}
fmt.Printf("✅ Canceled %d orders\n", len(orderIDs))
```

## Error Handling

Always check for errors and handle API errors specifically:

```go
response, err := client.NewCreateBasicOrderService().
    // ... parameters ...
    Do(context.Background())

if err != nil {
    // Check if it's an API error
    if versifi.IsAPIError(err) {
        apiErr := err.(*versifi.APIError)
        fmt.Printf("❌ API Error %d: %s\n", apiErr.Code, apiErr.Message)

        // Handle specific error codes
        switch apiErr.Code {
        case 400:
            fmt.Println("Invalid request parameters")
        case 401:
            fmt.Println("Authentication failed - check your API credentials")
        case 404:
            fmt.Println("Resource not found")
        case 409:
            fmt.Println("Conflict - resource already exists")
        default:
            fmt.Println("Unknown API error")
        }
    } else {
        // Network or other error
        fmt.Printf("❌ Error: %v\n", err)
    }
    return
}

// Success
fmt.Println("✅ Order created successfully")
```

## Configuration Tips

### Using Environment Variables

```go
import "os"

apiKey := os.Getenv("VERSIFI_API_KEY")
apiSecret := os.Getenv("VERSIFI_API_SECRET")

if apiKey == "" || apiSecret == "" {
    log.Fatal("Please set VERSIFI_API_KEY and VERSIFI_API_SECRET")
}

client := versifi.NewClient(apiKey, apiSecret)
```

### Custom HTTP Client

```go
import "net/http"
import "time"

httpClient := &http.Client{
    Timeout: 30 * time.Second,
}

client := versifi.NewClientWithHTTPClient(apiKey, apiSecret, httpClient)
```

### Using Context for Timeouts

```go
import "time"

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

response, err := client.NewCreateAlgoOrderService().
    // ... parameters ...
    Do(ctx)  // Pass context with timeout
```

## Available Algorithms

### TWAP (Time-Weighted Average Price)
Best for: Minimizing market impact over time

```go
params := map[string]interface{}{
    "duration":      3600,  // seconds
    "slice_size":    0.1,   // 10% per slice
    "time_interval": 60,    // slice every 60 seconds
}
```

### VWAP (Volume-Weighted Average Price)
Best for: Matching market volume patterns

```go
params := map[string]interface{}{
    "duration":          3600,  // seconds
    "volume_percentage": 0.05,  // 5% of market volume
    "aggressiveness":    0.5,   // 0=passive, 1=aggressive
}
```

### IS (Implementation Shortfall)
Best for: Minimizing slippage from decision time

```go
params := map[string]interface{}{
    "duration":               3600,  // required
    "urgency_level":          0.7,   // 0=low, 1=high
    "max_participation_rate": 0.1,   // max 10% of volume
}
```

## Supported Exchanges

- `ExchangeBinanceSpot` - Binance Spot Trading
- `ExchangeBinanceFutures` - Binance Futures
- `ExchangeOKXSpot` - OKX Spot Trading
- `ExchangeOKXFutures` - OKX Futures

## Complete Example Program

Here's a complete program that demonstrates multiple features:

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "time"
    versifi "github.com/versifi/versifi-go-sdk"
)

func main() {
    // Initialize client
    apiKey := os.Getenv("VERSIFI_API_KEY")
    apiSecret := os.Getenv("VERSIFI_API_SECRET")

    client := versifi.NewClient(apiKey, apiSecret)
    client.Debug = true

    // 1. Create a TWAP order
    fmt.Println("1️⃣ Creating TWAP order...")
    algoOrder, err := createTWAPOrder(client)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✅ Created order ID: %d\n\n", algoOrder.OrderID)

    // 2. Check order status
    fmt.Println("2️⃣ Checking order status...")
    time.Sleep(2 * time.Second)
    checkOrderStatus(client, algoOrder.OrderID)
    fmt.Println()

    // 3. Start WebSocket monitoring
    fmt.Println("3️⃣ Starting WebSocket monitoring...")
    go monitorOrders(apiKey, apiSecret)

    // 4. Wait a bit, then cancel
    fmt.Println("\n4️⃣ Waiting 10 seconds before canceling...")
    time.Sleep(10 * time.Second)

    fmt.Println("5️⃣ Canceling order...")
    cancelOrder(client, algoOrder.OrderID)

    // Wait for WebSocket updates
    time.Sleep(5 * time.Second)
    fmt.Println("\n✅ Demo complete!")
}

func createTWAPOrder(client *versifi.Client) (*versifi.OrderResponse, error) {
    params := map[string]interface{}{
        "duration":   3600,
        "slice_size": 0.1,
    }

    return client.NewCreateAlgoOrderService().
        Exchange(versifi.ExchangeBinanceSpot).
        OrderType(versifi.AlgoOrderTypeTWAP).
        Symbol("BTC/USDT").
        Side(versifi.SideTypeBuy).
        Quantity("1.0").
        Params(params).
        Do(context.Background())
}

func checkOrderStatus(client *versifi.Client, orderID int64) {
    response, err := client.NewGetOrderService().
        OrderID(orderID).
        Do(context.Background())

    if err != nil {
        log.Printf("Error: %v", err)
        return
    }

    fmt.Printf("Order %d: %s\n", response.OrderID, response.Status)
}

func cancelOrder(client *versifi.Client, orderID int64) {
    err := client.NewCancelOrderService().
        OrderID(orderID).
        Do(context.Background())

    if err != nil {
        log.Printf("Error canceling: %v", err)
        return
    }

    fmt.Printf("✅ Canceled order %d\n", orderID)
}

func monitorOrders(apiKey, apiSecret string) {
    wsClient := versifi.NewWsClient(apiKey, apiSecret)

    err := wsClient.Connect()
    if err != nil {
        log.Printf("WebSocket error: %v", err)
        return
    }
    defer wsClient.Disconnect()

    wsClient.SubscribeOrders(func(message []byte) {
        var order versifi.WsOrderUpdate
        json.Unmarshal(message, &order)
        fmt.Printf("📊 [WS] Order %d: %s\n", order.OrderID, order.Status)
    })

    wsClient.SubscribeTrades(func(message []byte) {
        var trade versifi.WsTradeUpdate
        json.Unmarshal(message, &trade)
        fmt.Printf("💰 [WS] Trade %d: %s @ %s\n",
            trade.TradeID, trade.Quantity, trade.Price)
    })

    // Keep running
    select {}
}
```

Run it:

```bash
export VERSIFI_API_KEY="your-key"
export VERSIFI_API_SECRET="your-secret"
go run main.go
```

## Next Steps

1. ✅ **Read the full README**: Check `README.md` for comprehensive documentation
2. 📚 **Explore examples**: See `examples/` directory for more code samples
3. 🏗️ **Build your strategy**: Integrate the SDK into your trading bot
4. 📖 **API Reference**: Visit https://docs.versifi.io for full API docs
5. 🐛 **Report issues**: https://github.com/versifi/versifi-go-sdk/issues

## Common Issues

### Issue: "Authentication failed"
**Solution**: Double-check your API key and secret. Ensure they're not expired.

### Issue: "Invalid signature"
**Solution**: Make sure you're using the correct API secret. The signature is case-sensitive.

### Issue: "Order rejected"
**Solution**: Check order parameters (quantity, price, etc.) and ensure you have sufficient balance.

### Issue: "WebSocket disconnected"
**Solution**: The SDK automatically reconnects. Check your network connection.

## Getting Help

- 📖 **Documentation**: Check `README.md` and `PROJECT_STRUCTURE.md`
- 💻 **Examples**: See `examples/` directory
- 🐛 **Issues**: GitHub Issues
- 📧 **Support**: Contact Versifi support

---

Happy Trading! 🚀
