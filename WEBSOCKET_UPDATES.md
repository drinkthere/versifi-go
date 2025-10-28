# WebSocket Implementation Updates

## 重要更新 - 基于官方WebSocket文档

根据Versifi官方WebSocket文档，我们对WebSocket实现进行了以下关键更新：

---

## 🔄 主要变更

### 1. **连接URL修正**
- **旧实现**: `wss://ws.versifi.io`
- **新实现**: `wss://example.com/v1/ws` (需要更新为生产URL)
- **变更原因**: 文档指定WebSocket端点为 `/v1/ws`

### 2. **认证流程重构**

#### 旧方式（错误）:
```go
// 在HTTP握手时发送认证头
header.Set("X-VERSIFI-API-KEY", c.APIKey)
header.Set("X-VERSIFI-API-SIGN", signature)
```

#### 新方式（正确）:
```go
// 1. 先建立WebSocket连接（无认证头）
conn, _, err := dialer.Dial(c.BaseURL, nil)

// 2. 连接后发送认证消息
authMsg := {
    "op": "auth",
    "args": [api_key, expires, signature]
}
```

### 3. **签名Payload格式修正**

#### 旧方式:
```go
payload := fmt.Sprintf("%d", timestamp)
```

#### 新方式（正确）:
```go
expires := time.Now().Add(5 * time.Minute).Unix()
payload := fmt.Sprintf("GET/realtime%d", expires)
signature := hmac_sha256(payload, apiSecret)
```

**签名格式**: `GET/realtime{expires}`

### 4. **订阅消息格式修正**

#### 旧方式:
```go
subscribeMsg := {
    "action": "subscribe",
    "channel": channel
}
```

#### 新方式（正确）:
```go
subscribeMsg := {
    "op": "subscribe",
    "args": [topic]
}
```

### 5. **Topic名称修正**

#### 旧实现:
- `SubscribeOrders()` → 订阅 `"orders"` topic
- `SubscribeTrades()` → 订阅 `"trades"` topic

#### 新实现（正确）:
- `SubscribeExecutionReport()` → 订阅 `"execution_report"` topic
- `SubscribeAnalytics()` → 订阅 `"analytics"` topic（服务器未实现）

### 6. **消息结构完全重构**

根据文档中的示例消息，更新了所有消息类型：

#### 通用响应格式:
```go
type WsResponse struct {
    Op      string      `json:"op"`       // 操作类型
    Success bool        `json:"success"`  // 是否成功
    Message interface{} `json:"message"`  // 消息内容
}
```

#### Execution Report格式:
```json
{
    "op": "execution_report",
    "success": true,
    "message": {
        "order_id": 1,
        "client_order_id": 987654321,
        "order_type": "LIMIT",
        "status": "FILLED",
        "timestamp": 1677721600,
        "request_order_type": "basic",
        "order": { /* 订单详情 */ }
    }
}
```

#### Trade字段更新:
```go
type WsTrade struct {
    TradeID                   int64  `json:"trade_id"`
    AveragePrice              string `json:"average_price"`              // 新增
    CummulativeFilledQuantity string `json:"cummulative_filled_quantity"` // 新增
    OrderID                   int64  `json:"order_id"`
    LegID                     *int64 `json:"leg_id,omitempty"` // 仅pair order
    ExecutedPrice             string `json:"executed_price"`   // 新增
    ExecutedQuantity          string `json:"executed_quantity"` // 新增
}
```

### 7. **Ping/Pong机制修正**

#### 旧方式（WebSocket级别）:
```go
c.SendMessage(websocket.PingMessage, []byte{})
```

#### 新方式（应用级别）:
```go
// 请求
{ "op": "ping" }

// 响应
{ "op": "ping", "message": object, "success": true }
```

---

## 📝 新增功能

### 1. 认证状态跟踪
```go
func (c *WsClient) IsAuthenticated() bool
```

现在可以检查WebSocket是否已认证。

### 2. 自动认证
`Connect()` 方法现在自动处理认证流程：
```go
wsClient.Connect()  // 自动连接并认证
```

### 3. 认证超时处理
认证过程有10秒超时保护：
```go
case <-time.After(10 * time.Second):
    return fmt.Errorf("authentication timeout")
```

---

## 🔍 消息类型对应

| Request Order Type | WebSocket消息类型 | 结构体 |
|-------------------|------------------|--------|
| `basic` | BasicOrder | `WsBasicOrderDetail` |
| `algo` | AlgoOrder (TWAP/VWAP/IS) | `WsAlgoOrderDetail` |
| `pair` | PairOrder (BASIS) | `WsPairOrderDetail` |

---

## 📚 使用示例

### 完整的WebSocket使用流程

```go
// 1. 创建客户端
wsClient := versifi.NewWsClient(apiKey, apiSecret)

// 2. 设置错误处理
wsClient.SetErrorHandler(func(err error) {
    log.Printf("Error: %v", err)
})

// 3. 连接（自动认证）
err := wsClient.Connect()
if err != nil {
    log.Fatal(err)
}
defer wsClient.Disconnect()

// 4. 订阅execution_report
wsClient.SubscribeExecutionReport(func(message []byte) {
    var execReport versifi.WsExecutionReport
    json.Unmarshal(message, &execReport)

    fmt.Printf("Order %d: %s\n",
        execReport.Message.OrderID,
        execReport.Message.Status)

    // 根据订单类型处理
    switch execReport.Message.RequestOrderType {
    case "basic":
        // 处理基础订单
    case "algo":
        // 处理算法订单
    case "pair":
        // 处理配对订单
    }
})

// 5. 等待消息
select {}
```

### 处理不同类型的订单

```go
wsClient.SubscribeExecutionReport(func(message []byte) {
    var execReport versifi.WsExecutionReport
    json.Unmarshal(message, &execReport)

    switch execReport.Message.RequestOrderType {
    case "basic":
        orderBytes, _ := json.Marshal(execReport.Message.Order)
        var basicOrder versifi.WsBasicOrderDetail
        json.Unmarshal(orderBytes, &basicOrder)

        // 访问交易信息
        for _, trade := range basicOrder.ChildOrder.Trades {
            fmt.Printf("Trade: %s @ %s\n",
                trade.ExecutedQuantity,
                trade.ExecutedPrice)
            fmt.Printf("Average Price: %s\n", trade.AveragePrice)
            fmt.Printf("Total Filled: %s\n",
                trade.CummulativeFilledQuantity)
        }

    case "algo":
        // 类似处理...

    case "pair":
        // 配对订单有leg_id字段
        // 类似处理...
    }
})
```

---

## ⚙️ 配置选项

### 更新生产URL

在使用前，需要将示例URL更新为实际的生产URL：

```go
// 在 websocket.go 中
var BaseWSMainURL = "wss://actual-production-url.versifi.io/v1/ws"
```

### 调整超时时间

```go
versifi.WebsocketTimeout = time.Second * 120  // 2分钟
```

### 禁用Keepalive

```go
versifi.WebsocketKeepalive = false
```

---

## 🐛 调试技巧

### 1. 启用详细日志

WebSocket客户端会自动记录所有接收的消息：

```go
wsClient.Logger.SetOutput(os.Stdout)  // 输出到stdout
```

### 2. 订阅所有消息

用于调试，捕获所有消息：

```go
wsClient.Subscribe("*", func(message []byte) {
    fmt.Printf("[Debug] %s\n", string(message))
})
```

### 3. 检查连接状态

```go
if !wsClient.IsConnected() {
    log.Println("Not connected!")
}

if !wsClient.IsAuthenticated() {
    log.Println("Not authenticated!")
}
```

---

## 🔐 安全注意事项

### 1. Expires时间戳

认证的expires应该是未来的时间戳：
```go
expires := time.Now().Add(5 * time.Minute).Unix()
```

### 2. 签名格式

必须严格遵循格式：`GET/realtime{expires}`

```go
// ✅ 正确
payload := fmt.Sprintf("GET/realtime%d", expires)

// ❌ 错误
payload := fmt.Sprintf("GET/realtime/%d", expires)  // 多了斜杠
payload := fmt.Sprintf("get/realtime%d", expires)   // 小写
```

### 3. 密钥保护

不要在代码中硬编码API密钥：
```go
apiKey := os.Getenv("VERSIFI_API_KEY")
apiSecret := os.Getenv("VERSIFI_API_SECRET")
```

---

## 📊 性能优化

### 1. 消息处理

消息处理应该是非阻塞的：

```go
wsClient.SubscribeExecutionReport(func(message []byte) {
    // 快速处理或发送到channel
    go processMessage(message)  // 异步处理
})
```

### 2. 重连策略

自动重连已实现，延迟为5秒：

```go
wsClient.reconnectDelay = 5 * time.Second
```

可以根据需要调整。

---

## ✅ 测试清单

在部署前，确保测试以下场景：

- [ ] 成功连接并认证
- [ ] 认证失败处理
- [ ] 订阅execution_report topic
- [ ] 接收basic订单更新
- [ ] 接收algo订单更新
- [ ] 接收pair订单更新
- [ ] 处理FILLED状态
- [ ] 处理PARTIALLY_FILLED状态
- [ ] 处理CANCELED状态
- [ ] 网络断开后自动重连
- [ ] 优雅关闭连接
- [ ] Ping/Pong保活机制

---

## 📖 相关文档

- Versifi WebSocket官方文档
- `examples/websocket_example.go` - 完整示例
- `README.md` - 用户指南
- `PROJECT_STRUCTURE.md` - 架构说明

---

## 🔄 从旧版本迁移

如果你正在使用旧版本的WebSocket实现，需要进行以下更改：

### 1. 更新订阅方法

```go
// 旧代码
wsClient.SubscribeOrders(handler)
wsClient.SubscribeTrades(handler)

// 新代码
wsClient.SubscribeExecutionReport(handler)
```

### 2. 更新消息解析

```go
// 旧代码
var update WsOrderUpdate  // 已过时

// 新代码
var execReport WsExecutionReport
json.Unmarshal(message, &execReport)
```

### 3. 处理新的字段

Trade消息现在包含更多字段：

```go
trade.AveragePrice               // 平均价格
trade.CummulativeFilledQuantity  // 累计成交量
trade.ExecutedPrice              // 执行价格
trade.ExecutedQuantity           // 执行数量
```

---

## 🎯 总结

这次更新完全按照Versifi官方WebSocket文档重构了实现，主要改进包括：

1. ✅ **正确的认证流程** - 先连接后认证
2. ✅ **正确的签名格式** - `GET/realtime{expires}`
3. ✅ **正确的消息格式** - 使用`op`和`args`
4. ✅ **完整的消息类型** - 支持所有订单类型
5. ✅ **健壮的错误处理** - 超时、重连等
6. ✅ **完整的示例代码** - 展示所有功能

现在SDK完全符合Versifi WebSocket API规范！🎉
