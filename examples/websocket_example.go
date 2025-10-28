package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	versifi "github.com/versifi/versifi-go-sdk"
)

func runWebSocketExample() {
	apiKey := "your-api-key"
	apiSecret := "your-api-secret"

	// Create websocket client
	wsClient := versifi.NewWsClient(apiKey, apiSecret)

	// Set error handler
	wsClient.SetErrorHandler(func(err error) {
		log.Printf("WebSocket Error: %v", err)
	})

	// Connect to websocket (this will automatically authenticate)
	fmt.Println("Connecting to Versifi WebSocket...")
	err := wsClient.Connect()
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer wsClient.Disconnect()

	fmt.Println("✅ Connected and authenticated to Versifi WebSocket")

	// Subscribe to execution_report topic
	err = wsClient.SubscribeExecutionReport(func(message []byte) {
		// Parse execution report
		var execReport versifi.WsExecutionReport
		if err := json.Unmarshal(message, &execReport); err != nil {
			log.Printf("Error parsing execution report: %v", err)
			fmt.Printf("[Raw Message] %s\n", string(message))
			return
		}

		// Display execution report details
		fmt.Printf("\n📊 [Execution Report] Op: %s, Success: %v\n", execReport.Op, execReport.Success)
		fmt.Printf("  Order ID: %d\n", execReport.Message.OrderID)
		fmt.Printf("  Client Order ID: %d\n", execReport.Message.ClientOrderID)
		fmt.Printf("  Order Type: %s\n", execReport.Message.OrderType)
		fmt.Printf("  Status: %s\n", execReport.Message.Status)
		fmt.Printf("  Request Type: %s\n", execReport.Message.RequestOrderType)

		// Handle different order types
		switch execReport.Message.RequestOrderType {
		case "basic":
			handleBasicOrder(execReport.Message.Order)
		case "algo":
			handleAlgoOrder(execReport.Message.Order)
		case "pair":
			handlePairOrder(execReport.Message.Order)
		}
	})

	if err != nil {
		log.Fatalf("Failed to subscribe to execution_report: %v", err)
	}

	fmt.Println("✅ Subscribed to execution_report topic")

	// Optionally subscribe to all messages (catch-all for debugging)
	wsClient.Subscribe("*", func(message []byte) {
		fmt.Printf("[Debug - Raw Message] %s\n", string(message))
	})

	fmt.Println("\n👂 Listening for execution reports... Press Ctrl+C to exit.\n")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Simulate some activity - show connection status
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if wsClient.IsConnected() && wsClient.IsAuthenticated() {
					fmt.Println("💚 WebSocket connection is alive and authenticated...")
				} else if wsClient.IsConnected() {
					fmt.Println("⚠️  WebSocket connected but not authenticated!")
				} else {
					fmt.Println("❌ WebSocket connection lost!")
				}
			case <-sigChan:
				return
			}
		}
	}()

	<-sigChan
	fmt.Println("\n\nShutting down gracefully...")
}

func handleBasicOrder(order interface{}) {
	orderBytes, _ := json.Marshal(order)
	var basicOrder versifi.WsBasicOrderDetail
	if err := json.Unmarshal(orderBytes, &basicOrder); err != nil {
		log.Printf("Error parsing basic order: %v", err)
		return
	}

	fmt.Printf("  📝 Basic Order Details:\n")
	fmt.Printf("    Symbol: %s\n", basicOrder.Symbol)
	fmt.Printf("    Side: %s\n", basicOrder.Side)
	fmt.Printf("    Quantity: %s\n", basicOrder.Quantity)
	fmt.Printf("    Price: %s\n", basicOrder.Price)
	fmt.Printf("    Exchange: %s\n", basicOrder.Exchange)

	// Show trades if any
	if basicOrder.ChildOrder != nil && len(basicOrder.ChildOrder.Trades) > 0 {
		fmt.Printf("  💰 Trades:\n")
		for _, trade := range basicOrder.ChildOrder.Trades {
			fmt.Printf("    Trade ID %d: Executed %s @ %s (Avg: %s, Total Filled: %s)\n",
				trade.TradeID,
				trade.ExecutedQuantity,
				trade.ExecutedPrice,
				trade.AveragePrice,
				trade.CummulativeFilledQuantity)
		}
	}
}

func handleAlgoOrder(order interface{}) {
	orderBytes, _ := json.Marshal(order)
	var algoOrder versifi.WsAlgoOrderDetail
	if err := json.Unmarshal(orderBytes, &algoOrder); err != nil {
		log.Printf("Error parsing algo order: %v", err)
		return
	}

	fmt.Printf("  🤖 Algo Order Details:\n")
	fmt.Printf("    Algorithm: %s\n", algoOrder.OrderType)
	fmt.Printf("    Symbol: %s\n", algoOrder.Symbol)
	fmt.Printf("    Side: %s\n", algoOrder.Side)
	fmt.Printf("    Quantity: %s\n", algoOrder.Quantity)
	fmt.Printf("    Exchange: %s\n", algoOrder.Exchange)

	// Show trades if any
	if algoOrder.ChildOrder != nil && len(algoOrder.ChildOrder.Trades) > 0 {
		fmt.Printf("  💰 Trades:\n")
		for _, trade := range algoOrder.ChildOrder.Trades {
			fmt.Printf("    Trade ID %d: Executed %s @ %s (Avg: %s, Total Filled: %s)\n",
				trade.TradeID,
				trade.ExecutedQuantity,
				trade.ExecutedPrice,
				trade.AveragePrice,
				trade.CummulativeFilledQuantity)
		}
	}
}

func handlePairOrder(order interface{}) {
	orderBytes, _ := json.Marshal(order)
	var pairOrder versifi.WsPairOrderDetail
	if err := json.Unmarshal(orderBytes, &pairOrder); err != nil {
		log.Printf("Error parsing pair order: %v", err)
		return
	}

	fmt.Printf("  🔄 Pair Order Details:\n")

	if pairOrder.LeadLeg != nil {
		fmt.Printf("    Lead Leg:\n")
		fmt.Printf("      Symbol: %s\n", pairOrder.LeadLeg.Symbol)
		fmt.Printf("      Exchange: %s\n", pairOrder.LeadLeg.Exchange)
		fmt.Printf("      Order Type: %s\n", pairOrder.LeadLeg.OrderType)

		if pairOrder.LeadLeg.ChildOrder != nil && len(pairOrder.LeadLeg.ChildOrder.Trades) > 0 {
			fmt.Printf("      Trades:\n")
			for _, trade := range pairOrder.LeadLeg.ChildOrder.Trades {
				fmt.Printf("        Trade ID %d (Leg %d): Executed %s @ %s\n",
					trade.TradeID,
					*trade.LegID,
					trade.ExecutedQuantity,
					trade.ExecutedPrice)
			}
		}
	}

	if pairOrder.Leg != nil {
		fmt.Printf("    Secondary Leg:\n")
		fmt.Printf("      Symbol: %s\n", pairOrder.Leg.Symbol)
		fmt.Printf("      Exchange: %s\n", pairOrder.Leg.Exchange)
		fmt.Printf("      Order Type: %s\n", pairOrder.Leg.OrderType)

		if pairOrder.Leg.ChildOrder != nil && len(pairOrder.Leg.ChildOrder.Trades) > 0 {
			fmt.Printf("      Trades:\n")
			for _, trade := range pairOrder.Leg.ChildOrder.Trades {
				legID := int64(0)
				if trade.LegID != nil {
					legID = *trade.LegID
				}
				fmt.Printf("        Trade ID %d (Leg %d): Executed %s @ %s\n",
					trade.TradeID,
					legID,
					trade.ExecutedQuantity,
					trade.ExecutedPrice)
			}
		}
	}
}

func main() {
	runWebSocketExample()
}
