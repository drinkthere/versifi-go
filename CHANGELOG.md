# Changelog

All notable changes to the Versifi Go SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2025-01-XX

### Added

#### Local IP Address Binding Support
- **REST API**: Added `NewClientWithLocalAddr()` function to create clients that bind to specific local IP addresses
- **WebSocket**: Added `NewWsClientWithLocalAddr()` function to create WebSocket clients that bind to specific local IP addresses
- New documentation file: `LOCAL_IP_BINDING.md` with comprehensive guide on using local IP binding
- New example file: `examples/local_addr_example.go` demonstrating both REST and WebSocket with local IP binding

**Use Cases:**
- Servers with multiple IP addresses where only one is whitelisted
- Ensure all requests originate from the whitelisted IP
- Network isolation and routing control

**Example Usage:**
```go
// REST API
client := versifi.NewClientWithLocalAddr(apiKey, apiSecret, "192.168.1.100")

// WebSocket
wsClient := versifi.NewWsClientWithLocalAddr(apiKey, apiSecret, "192.168.1.100")
```

### Changed

#### WebSocket Implementation - Complete Rewrite Based on Official Documentation
- **Authentication Flow**: Changed from header-based auth to message-based auth
  - Now connects first, then sends authentication message
  - Proper `{"op": "auth", "args": [...]}` format
- **Signature Format**: Updated to correct format `GET/realtime{expires}`
- **Subscription Format**: Changed to `{"op": "subscribe", "args": [topic]}`
- **Topic Names**: Updated to use official topic names
  - `execution_report` instead of generic `orders`/`trades`
  - Added `analytics` topic (not yet implemented on server)
- **Message Structures**: Complete rewrite of all WebSocket message types
  - `WsExecutionReport` with proper structure
  - `WsTrade` with new fields: `average_price`, `cummulative_filled_quantity`, `executed_price`, `executed_quantity`
  - Support for Basic, Algo, and Pair order types
- **Ping Mechanism**: Changed from WebSocket-level to application-level ping
  - Request: `{"op": "ping"}`
  - Response: `{"op": "ping", "message": object, "success": true}`
- **Connection URL**: Updated to `/v1/ws` endpoint

**Breaking Changes:**
- `SubscribeOrders()` removed → use `SubscribeExecutionReport()` instead
- `SubscribeTrades()` removed → execution reports include all trade information
- `WsOrderUpdate` type removed → use `WsExecutionReport` instead
- `WsTradeUpdate` type changed → new fields added

**Migration Guide:**
```go
// Old code (v1.0.0)
wsClient.SubscribeOrders(handler)
wsClient.SubscribeTrades(handler)

// New code (v1.1.0)
wsClient.SubscribeExecutionReport(handler)
```

### Documentation

- Added `WEBSOCKET_UPDATES.md` - Detailed explanation of WebSocket changes
- Added `LOCAL_IP_BINDING.md` - Comprehensive guide on local IP binding
- Updated `README.md` with local IP binding examples
- Updated `examples/websocket_example.go` to match new WebSocket implementation

### Technical Details

#### REST API Local IP Binding Implementation
- Uses custom `net.Dialer` with `LocalAddr` set to specified IP
- Custom `http.Transport` with custom dialer
- Automatic fallback to standard client if IP resolution fails
- Full HTTP/2 support maintained

#### WebSocket Local IP Binding Implementation
- Uses `websocket.Dialer.NetDial` with custom `net.Dialer`
- Binds to local address before establishing WebSocket connection
- Maintains all existing features (auto-reconnect, keepalive, etc.)

## [1.0.0] - 2025-01-XX

### Added

#### Initial Release

**REST API Support:**
- ✅ Create Algo Orders (TWAP, VWAP, IS)
- ✅ Create Basic Orders (MARKET, LIMIT, STOP, etc.)
- ✅ Create Pair Orders (BASIS trading)
- ✅ Cancel Orders (single and batch)
- ✅ Get Order Details by ID

**WebSocket Support:**
- ✅ Real-time order updates
- ✅ Real-time trade updates
- ✅ Automatic reconnection
- ✅ Flexible message handling

**Design Features:**
- Follows go-binance SDK patterns
- Builder pattern for fluent API
- Type-safe enums
- Comprehensive error handling
- HMAC SHA256 authentication
- Context support for timeouts
- Debug logging

**Documentation:**
- Complete README with examples
- API reference documentation
- Project structure documentation
- Quick start guide
- Implementation summary

**Examples:**
- REST API usage examples
- WebSocket usage examples
- All order types demonstrated

**Testing:**
- Unit tests for client initialization
- Unit tests for signature generation
- Unit tests for all order services
- Unit tests for error handling

---

## Version Comparison

### v1.1.0 vs v1.0.0

| Feature | v1.0.0 | v1.1.0 |
|---------|--------|--------|
| **REST API - Local IP Binding** | ❌ Not supported | ✅ Supported |
| **WebSocket - Local IP Binding** | ❌ Not supported | ✅ Supported |
| **WebSocket Authentication** | Header-based (incorrect) | Message-based (correct) |
| **WebSocket Topics** | Generic topics | Official topics (`execution_report`) |
| **WebSocket Signature** | Simple timestamp | `GET/realtime{expires}` format |
| **Trade Fields** | Basic fields | Extended fields (avg price, cumulative filled, etc.) |
| **Ping/Pong** | WebSocket protocol level | Application level |

---

## Upgrade Guide

### From v1.0.0 to v1.1.0

#### WebSocket Changes (Breaking)

1. **Update subscription method:**
   ```go
   // Before (v1.0.0)
   wsClient.SubscribeOrders(handler)
   wsClient.SubscribeTrades(handler)

   // After (v1.1.0)
   wsClient.SubscribeExecutionReport(handler)
   ```

2. **Update message parsing:**
   ```go
   // Before (v1.0.0)
   var update WsOrderUpdate  // Removed
   json.Unmarshal(message, &update)

   // After (v1.1.0)
   var execReport WsExecutionReport
   json.Unmarshal(message, &execReport)

   // Access order details
   orderID := execReport.Message.OrderID
   status := execReport.Message.Status
   ```

3. **Handle new trade fields:**
   ```go
   // New fields available in v1.1.0
   trade.AveragePrice
   trade.CummulativeFilledQuantity
   trade.ExecutedPrice
   trade.ExecutedQuantity
   ```

#### Local IP Binding (New Feature)

If you need to bind to a specific local IP:

1. **REST API:**
   ```go
   // Before (v1.0.0)
   client := versifi.NewClient(apiKey, apiSecret)

   // After (v1.1.0) - with IP binding
   client := versifi.NewClientWithLocalAddr(apiKey, apiSecret, "192.168.1.100")
   ```

2. **WebSocket:**
   ```go
   // Before (v1.0.0)
   wsClient := versifi.NewWsClient(apiKey, apiSecret)

   // After (v1.1.0) - with IP binding
   wsClient := versifi.NewWsClientWithLocalAddr(apiKey, apiSecret, "192.168.1.100")
   ```

---

## Future Roadmap

### Planned for v1.2.0
- [ ] Rate limiting management
- [ ] Request retry logic with exponential backoff
- [ ] Batch order creation
- [ ] Account information endpoints
- [ ] Position management endpoints

### Under Consideration
- [ ] Historical data endpoints
- [ ] Advanced analytics support
- [ ] Connection pooling
- [ ] Prometheus metrics integration
- [ ] OpenTelemetry tracing

---

## Links

- **Documentation**: [README.md](README.md)
- **Local IP Binding Guide**: [LOCAL_IP_BINDING.md](LOCAL_IP_BINDING.md)
- **WebSocket Updates**: [WEBSOCKET_UPDATES.md](WEBSOCKET_UPDATES.md)
- **Project Structure**: [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md)
- **Examples**: [examples/](examples/)

---

## License

MIT License - See [LICENSE](LICENSE) file for details
