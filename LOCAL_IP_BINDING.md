# Local IP Address Binding

## 概述

当服务器配置了多个IP地址，但只有其中一个IP被Versifi加入白名单时，您需要指定使用该白名单IP作为请求的出口IP。本SDK提供了完整的本地IP绑定支持，适用于REST API和WebSocket连接。

---

## 为什么需要本地IP绑定？

### 常见场景

1. **多IP服务器**
   - 服务器有多个网卡或多个IP地址
   - 只有特定IP被加入Versifi白名单
   - 默认路由可能使用非白名单IP

2. **负载均衡**
   - 多个出口IP用于不同服务
   - 需要确保Versifi请求使用特定IP

3. **网络隔离**
   - 不同业务使用不同IP段
   - 需要明确指定交易系统使用的IP

### 问题示例

```bash
# 服务器有多个IP
eth0: 192.168.1.100  ✅ 已加入Versifi白名单
eth1: 192.168.1.101  ❌ 未加入白名单
eth2: 192.168.1.102  ❌ 未加入白名单

# 默认情况下，出口IP可能是 192.168.1.101
# 这会导致认证失败：Authentication failed: IP not whitelisted
```

---

## REST API - 本地IP绑定

### 基本用法

```go
import versifi "github.com/versifi/versifi-go-sdk"

func main() {
    apiKey := "your-api-key"
    apiSecret := "your-api-secret"
    localIP := "192.168.1.100" // 您的白名单IP

    // 创建绑定本地IP的客户端
    client := versifi.NewClientWithLocalAddr(apiKey, apiSecret, localIP)

    // 现在所有请求都会从 192.168.1.100 发出
    response, err := client.NewCreateAlgoOrderService().
        Exchange(versifi.ExchangeBinanceSpot).
        OrderType(versifi.AlgoOrderTypeTWAP).
        Symbol("BTC/USDT").
        Side(versifi.SideTypeBuy).
        Quantity("1.0").
        Do(context.Background())
}
```

### 完整示例

```go
package main

import (
    "context"
    "fmt"
    "log"
    versifi "github.com/versifi/versifi-go-sdk"
)

func main() {
    // 配置
    apiKey := "your-api-key"
    apiSecret := "your-api-secret"
    whitelistedIP := "192.168.1.100"

    // 创建绑定本地IP的客户端
    client := versifi.NewClientWithLocalAddr(apiKey, apiSecret, whitelistedIP)

    // 可选：启用调试模式查看详细信息
    client.Debug = true

    // 创建订单
    params := map[string]interface{}{
        "duration": 3600,
    }

    response, err := client.NewCreateAlgoOrderService().
        Exchange(versifi.ExchangeBinanceSpot).
        OrderType(versifi.AlgoOrderTypeTWAP).
        Symbol("BTC/USDT").
        Side(versifi.SideTypeBuy).
        Quantity("1.0").
        Params(params).
        Do(context.Background())

    if err != nil {
        log.Fatalf("Error: %v", err)
    }

    fmt.Printf("Order created from IP %s\n", whitelistedIP)
    fmt.Printf("Order ID: %d\n", response.OrderID)
}
```

### API方法

#### `NewClientWithLocalAddr(apiKey, apiSecret, localAddr string) *Client`

创建一个绑定到指定本地IP地址的客户端。

**参数**:
- `apiKey` (string): Versifi API密钥
- `apiSecret` (string): Versifi API密钥
- `localAddr` (string): 本地IP地址（仅IP，不包含端口）

**返回**:
- `*Client`: 配置了本地IP绑定的客户端实例

**示例**:
```go
client := versifi.NewClientWithLocalAddr("key", "secret", "192.168.1.100")
```

---

## WebSocket - 本地IP绑定

### 基本用法

```go
import versifi "github.com/versifi/versifi-go-sdk"

func main() {
    apiKey := "your-api-key"
    apiSecret := "your-api-secret"
    localIP := "192.168.1.100" // 您的白名单IP

    // 创建绑定本地IP的WebSocket客户端
    wsClient := versifi.NewWsClientWithLocalAddr(apiKey, apiSecret, localIP)

    // 连接（自动从指定IP发起）
    err := wsClient.Connect()
    if err != nil {
        log.Fatal(err)
    }
    defer wsClient.Disconnect()

    // 订阅消息
    wsClient.SubscribeExecutionReport(func(message []byte) {
        // 处理消息
    })
}
```

### 完整示例

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
    // 配置
    apiKey := "your-api-key"
    apiSecret := "your-api-secret"
    whitelistedIP := "192.168.1.100"

    // 创建绑定本地IP的WebSocket客户端
    wsClient := versifi.NewWsClientWithLocalAddr(apiKey, apiSecret, whitelistedIP)

    // 设置错误处理
    wsClient.SetErrorHandler(func(err error) {
        log.Printf("WebSocket Error: %v", err)
    })

    // 连接（会从指定的本地IP发起连接）
    fmt.Printf("Connecting from IP: %s\n", whitelistedIP)
    err := wsClient.Connect()
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer wsClient.Disconnect()

    fmt.Println("Connected successfully!")

    // 订阅执行报告
    wsClient.SubscribeExecutionReport(func(message []byte) {
        var execReport versifi.WsExecutionReport
        json.Unmarshal(message, &execReport)

        fmt.Printf("Order %d: %s\n",
            execReport.Message.OrderID,
            execReport.Message.Status)
    })

    // 等待中断信号
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt)
    <-sigChan

    fmt.Println("Disconnecting...")
}
```

### API方法

#### `NewWsClientWithLocalAddr(apiKey, apiSecret, localAddr string) *WsClient`

创建一个绑定到指定本地IP地址的WebSocket客户端。

**参数**:
- `apiKey` (string): Versifi API密钥
- `apiSecret` (string): Versifi API密钥
- `localAddr` (string): 本地IP地址（仅IP，不包含端口）

**返回**:
- `*WsClient`: 配置了本地IP绑定的WebSocket客户端实例

**示例**:
```go
wsClient := versifi.NewWsClientWithLocalAddr("key", "secret", "192.168.1.100")
```

---

## 技术实现

### REST API实现

```go
func NewClientWithLocalAddr(apiKey, apiSecret, localAddr string) *Client {
    // 解析本地地址
    localTCPAddr, err := net.ResolveTCPAddr("tcp", localAddr+":0")
    if err != nil {
        log.Printf("Warning: failed to resolve local address %s: %v", localAddr, err)
        return NewClient(apiKey, apiSecret) // 回退到普通客户端
    }

    // 创建自定义Dialer，绑定本地地址
    dialer := &net.Dialer{
        LocalAddr: localTCPAddr,
        Timeout:   30 * time.Second,
        KeepAlive: 30 * time.Second,
    }

    // 创建HTTP Transport使用自定义Dialer
    transport := &http.Transport{
        DialContext:           dialer.DialContext,
        // ... 其他配置
    }

    // 创建HTTP Client
    httpClient := &http.Client{
        Transport: transport,
        Timeout:   30 * time.Second,
    }

    return &Client{
        HTTPClient: httpClient,
        // ... 其他字段
    }
}
```

### WebSocket实现

```go
func (c *WsClient) Connect() error {
    dialer := websocket.DefaultDialer

    // 如果指定了本地地址，配置Dialer
    if c.LocalAddr != "" {
        localTCPAddr, _ := net.ResolveTCPAddr("tcp", c.LocalAddr+":0")

        netDialer := &net.Dialer{
            LocalAddr: localTCPAddr,
            Timeout:   30 * time.Second,
            KeepAlive: 30 * time.Second,
        }

        // 设置自定义Dial函数
        dialer.NetDial = netDialer.Dial
    }

    // 建立连接
    conn, _, err := dialer.Dial(c.BaseURL, nil)
    // ...
}
```

---

## 验证本地IP绑定

### 方法1: 使用tcpdump（Linux）

```bash
# 监听指定网卡的流量
sudo tcpdump -i eth0 host api.versifi.io

# 或者监听特定IP
sudo tcpdump -i any src 192.168.1.100 and host api.versifi.io
```

### 方法2: 使用netstat

```bash
# 查看活动连接的本地地址
netstat -an | grep ESTABLISHED | grep 443
```

### 方法3: 代码中启用调试

```go
client := versifi.NewClientWithLocalAddr(apiKey, apiSecret, localIP)
client.Debug = true  // 启用调试日志

// 日志会显示请求详情
```

### 方法4: 使用测试程序

```go
package main

import (
    "fmt"
    "io"
    "net/http"
    versifi "github.com/versifi/versifi-go-sdk"
)

func main() {
    localIP := "192.168.1.100"

    // 创建绑定本地IP的客户端
    client := versifi.NewClientWithLocalAddr("", "", localIP)

    // 访问IP检测服务
    req, _ := http.NewRequest("GET", "https://api.ipify.org", nil)
    resp, err := client.HTTPClient.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)
    fmt.Printf("Request sent from IP: %s\n", string(body))
    fmt.Printf("Expected IP: %s\n", localIP)
}
```

---

## 故障排查

### 问题1: "failed to resolve local address"

**错误信息**:
```
Warning: failed to resolve local address 192.168.1.100: ...
```

**原因**: 指定的IP地址在本机不存在

**解决方案**:
```bash
# 1. 检查本机IP地址
ip addr show  # Linux
ifconfig      # macOS/BSD

# 2. 确认IP地址格式正确
# ✅ 正确: "192.168.1.100"
# ❌ 错误: "192.168.1.100:8080"  (包含端口)
# ❌ 错误: "http://192.168.1.100" (包含协议)
```

### 问题2: "Authentication failed" 或 "IP not whitelisted"

**原因**:
1. 指定的IP未加入白名单
2. 请求没有从指定IP发出

**解决方案**:
```go
// 1. 验证IP是否正确
client := versifi.NewClientWithLocalAddr(apiKey, apiSecret, "192.168.1.100")
client.Debug = true  // 启用调试查看详细信息

// 2. 联系Versifi确认白名单状态
// 3. 使用tcpdump验证实际发送IP
```

### 问题3: "connection refused" 或 "network unreachable"

**原因**: 指定的本地IP无法路由到目标地址

**解决方案**:
```bash
# 测试从指定IP是否能连通
ping -I 192.168.1.100 api.versifi.io

# 检查路由表
ip route show  # Linux
netstat -rn    # macOS/BSD
```

### 问题4: WebSocket连接失败

**错误信息**:
```
failed to connect: dial tcp: ...
```

**解决方案**:
```go
// 确保本地地址格式正确（不包含端口）
wsClient := versifi.NewWsClientWithLocalAddr(
    apiKey,
    apiSecret,
    "192.168.1.100"  // ✅ 正确
    // "192.168.1.100:0"  // ❌ 错误
)
```

---

## 高级用法

### 动态选择本地IP

```go
package main

import (
    "fmt"
    "net"
    versifi "github.com/versifi/versifi-go-sdk"
)

// 获取第一个非回环IPv4地址
func getFirstIPv4() string {
    addrs, err := net.InterfaceAddrs()
    if err != nil {
        return ""
    }

    for _, addr := range addrs {
        if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
            if ipnet.IP.To4() != nil {
                return ipnet.IP.String()
            }
        }
    }
    return ""
}

// 获取指定网卡的IP
func getIPByInterface(ifaceName string) string {
    iface, err := net.InterfaceByName(ifaceName)
    if err != nil {
        return ""
    }

    addrs, err := iface.Addrs()
    if err != nil {
        return ""
    }

    for _, addr := range addrs {
        if ipnet, ok := addr.(*net.IPNet); ok {
            if ipnet.IP.To4() != nil {
                return ipnet.IP.String()
            }
        }
    }
    return ""
}

func main() {
    // 方法1: 使用配置文件
    localIP := "192.168.1.100"

    // 方法2: 自动检测
    // localIP := getFirstIPv4()

    // 方法3: 指定网卡
    // localIP := getIPByInterface("eth0")

    client := versifi.NewClientWithLocalAddr("key", "secret", localIP)
    fmt.Printf("Using IP: %s\n", localIP)
}
```

### 环境变量配置

```go
package main

import (
    "os"
    versifi "github.com/versifi/versifi-go-sdk"
)

func main() {
    apiKey := os.Getenv("VERSIFI_API_KEY")
    apiSecret := os.Getenv("VERSIFI_API_SECRET")
    localIP := os.Getenv("VERSIFI_LOCAL_IP") // 从环境变量读取

    var client *versifi.Client
    if localIP != "" {
        // 使用本地IP绑定
        client = versifi.NewClientWithLocalAddr(apiKey, apiSecret, localIP)
    } else {
        // 使用默认客户端
        client = versifi.NewClient(apiKey, apiSecret)
    }

    // 使用client...
}
```

运行时指定：
```bash
export VERSIFI_LOCAL_IP=192.168.1.100
go run main.go
```

### 多客户端不同IP

```go
package main

import (
    versifi "github.com/versifi/versifi-go-sdk"
)

func main() {
    // 不同业务使用不同IP
    tradingClient := versifi.NewClientWithLocalAddr(
        "trading-key",
        "trading-secret",
        "192.168.1.100",  // 交易专用IP
    )

    marketDataClient := versifi.NewClientWithLocalAddr(
        "market-key",
        "market-secret",
        "192.168.1.101",  // 行情专用IP
    )

    // 分别使用
    // tradingClient.NewCreateAlgoOrderService()...
    // marketDataClient.NewGetOrderService()...
}
```

---

## 性能考虑

### 连接复用

本地IP绑定不影响HTTP连接复用：

```go
client := versifi.NewClientWithLocalAddr(apiKey, apiSecret, localIP)

// HTTP Transport会复用连接
for i := 0; i < 100; i++ {
    // 所有请求共享连接池，都从同一个本地IP发出
    client.NewGetOrderService().OrderID(int64(i)).Do(ctx)
}
```

### 并发安全

```go
client := versifi.NewClientWithLocalAddr(apiKey, apiSecret, localIP)

// 客户端是并发安全的
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(orderID int64) {
        defer wg.Done()
        client.NewGetOrderService().OrderID(orderID).Do(ctx)
    }(int64(i))
}
wg.Wait()
```

---

## 最佳实践

### 1. 配置管理

```go
// config.yaml
versifi:
  api_key: "your-key"
  api_secret: "your-secret"
  local_ip: "192.168.1.100"  # 白名单IP
```

### 2. 错误处理

```go
client := versifi.NewClientWithLocalAddr(apiKey, apiSecret, localIP)

response, err := client.NewCreateAlgoOrderService().
    // ... parameters ...
    Do(ctx)

if err != nil {
    if versifi.IsAPIError(err) {
        apiErr := err.(*versifi.APIError)
        if apiErr.Code == 401 {
            log.Printf("Authentication failed - check IP whitelist")
            log.Printf("Local IP used: %s", localIP)
        }
    }
}
```

### 3. 健康检查

```go
func checkConnection(client *versifi.Client) error {
    // 尝试简单的API调用
    _, err := client.NewGetOrderService().
        OrderID(1).  // 不存在的订单
        Do(context.Background())

    // 如果是404，说明连接正常（订单不存在）
    // 如果是401，说明IP认证失败
    if err != nil {
        if versifi.IsAPIError(err) {
            apiErr := err.(*versifi.APIError)
            if apiErr.Code == 404 {
                return nil // 连接正常
            }
        }
    }
    return err
}
```

---

## 示例代码

完整示例请参考：
- `examples/local_addr_example.go` - 完整的本地IP绑定示例

---

## 总结

本地IP绑定功能让您可以：

✅ **REST API**: 使用 `NewClientWithLocalAddr()` 绑定本地IP
✅ **WebSocket**: 使用 `NewWsClientWithLocalAddr()` 绑定本地IP
✅ **自动回退**: IP解析失败时自动使用默认客户端
✅ **连接复用**: 不影响HTTP连接池性能
✅ **并发安全**: 多goroutine安全使用

这个功能解决了多IP服务器的白名单认证问题，确保所有请求都从正确的IP地址发出。
