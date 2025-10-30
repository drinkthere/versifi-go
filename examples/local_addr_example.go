package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"

	versifi "github.com/drinkthere/versifi-go"
)

func main() {
	apiKey := "your-api-key"
	apiSecret := "your-api-secret"

	// Specify the whitelisted local IP address
	// This is the IP that is added to Versifi's whitelist
	localIP := "192.168.1.100" // Replace with your actual whitelisted IP

	fmt.Println("=== Versifi Go SDK - Local IP Binding Example ===\n")

	// Example 1: REST API with Local IP Binding
	restAPIExample(apiKey, apiSecret, localIP)

	// Example 2: WebSocket with Local IP Binding
	websocketExample(apiKey, apiSecret, localIP)
}

// REST API Example with Local IP Binding
func restAPIExample(apiKey, apiSecret, localIP string) {
	fmt.Println("1️⃣ REST API with Local IP Binding")
	fmt.Printf("   Binding to local IP: %s\n\n", localIP)

	// Create client with local IP binding
	client := versifi.NewClientWithLocalAddr(apiKey, apiSecret, localIP)
	client.Debug = true

	// Create a TWAP order
	params := map[string]interface{}{
		"duration":   3600,
		"slice_size": 0.1,
	}

	fmt.Println("   Creating TWAP order...")
	response, err := client.NewCreateAlgoOrderService().
		Exchange(versifi.ExchangeBinanceSpot).
		OrderType(versifi.AlgoOrderTypeTWAP).
		Symbol("BTC/USDT").
		Side(versifi.SideTypeBuy).
		Quantity("1.0").
		Params(params).
		Do(context.Background())

	if err != nil {
		log.Printf("   ❌ Error: %v\n", err)
		return
	}

	fmt.Printf("   ✅ Order created successfully!\n")
	fmt.Printf("   Order ID: %d\n", response.OrderID)
	fmt.Printf("   Status: %s\n\n", response.Status)

	// Get order details
	fmt.Println("   Fetching order details...")
	orderDetails, err := client.NewGetOrderService().
		OrderID(response.OrderID).
		Do(context.Background())

	if err != nil {
		log.Printf("   ❌ Error: %v\n", err)
		return
	}

	fmt.Printf("   ✅ Order details retrieved!\n")
	fmt.Printf("   Order Type: %s\n", orderDetails.OrderType)
	fmt.Printf("   Status: %s\n\n", orderDetails.Status)
}

// WebSocket Example with Local IP Binding
func websocketExample(apiKey, apiSecret, localIP string) {
	fmt.Println("2️⃣ WebSocket with Local IP Binding")
	fmt.Printf("   Binding to local IP: %s\n\n", localIP)

	// Create WebSocket client with local IP binding
	wsClient := versifi.NewWsClientWithLocalAddr(apiKey, apiSecret, localIP)

	// Set error handler
	wsClient.SetErrorHandler(func(err error) {
		log.Printf("   WebSocket Error: %v\n", err)
	})

	// Connect to WebSocket (will bind to specified local IP)
	fmt.Println("   Connecting to WebSocket...")
	err := wsClient.Connect()
	if err != nil {
		log.Fatalf("   ❌ Failed to connect: %v\n", err)
	}
	defer wsClient.Disconnect()

	fmt.Println("   ✅ Connected and authenticated!")
	fmt.Printf("   Local IP: %s\n\n", localIP)

	// Subscribe to execution reports
	err = wsClient.SubscribeExecutionReport(func(message []byte) {
		var execReport versifi.WsExecutionReport
		if err := json.Unmarshal(message, &execReport); err != nil {
			log.Printf("   Error parsing: %v\n", err)
			return
		}

		fmt.Printf("\n   📊 Execution Report Received:\n")
		fmt.Printf("      Order ID: %d\n", execReport.Message.OrderID)
		fmt.Printf("      Status: %s\n", execReport.Message.Status)
		fmt.Printf("      Type: %s\n", execReport.Message.OrderType)
	})

	if err != nil {
		log.Fatalf("   ❌ Failed to subscribe: %v\n", err)
	}

	fmt.Println("   ✅ Subscribed to execution_report")
	fmt.Println("   👂 Listening for updates...\n")
	fmt.Println("   Press Ctrl+C to exit\n")

	// Wait for interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan

	fmt.Println("\n   Disconnecting...")
}

/*
Output example:

=== Versifi Go SDK - Local IP Binding Example ===

1️⃣ REST API with Local IP Binding
   Binding to local IP: 192.168.1.100

   Creating TWAP order...
   ✅ Order created successfully!
   Order ID: 12345
   Status: NEW

   Fetching order details...
   ✅ Order details retrieved!
   Order Type: TWAP
   Status: NEW

2️⃣ WebSocket with Local IP Binding
   Binding to local IP: 192.168.1.100

   Connecting to WebSocket...
   WebSocket binding to local address: 192.168.1.100
   ✅ Connected and authenticated!
   Local IP: 192.168.1.100

   ✅ Subscribed to execution_report
   👂 Listening for updates...

   Press Ctrl+C to exit

   📊 Execution Report Received:
      Order ID: 12345
      Status: FILLED
      Type: TWAP

   Disconnecting...
*/
