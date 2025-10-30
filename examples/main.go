package main

import (
	"context"
	"fmt"
	"log"

	versifi "github.com/drinkthere/versifi-go"
)

func main() {
	// Initialize the client
	apiKey := "your-api-key"
	apiSecret := "your-api-secret"
	client := versifi.NewClient(apiKey, apiSecret)

	// Enable debug mode to see request/response details
	client.Debug = true

	// Example 1: Create an Algo Order (TWAP)
	createAlgoOrder(client)

	// Example 2: Create a Basic Order (LIMIT)
	createBasicOrder(client)

	// Example 3: Create a Pair Order (BASIS)
	createPairOrder(client)

	// Example 4: Get Order by ID
	getOrder(client, 12345)

	// Example 5: Cancel an Order
	cancelOrder(client, 12345)

	// Example 6: Cancel Batch Orders
	cancelBatchOrders(client, []int64{12345, 12346, 12347})
}

func createAlgoOrder(client *versifi.Client) {
	fmt.Println("=== Creating Algo Order (TWAP) ===")

	// Create TWAP order
	params := map[string]interface{}{
		"duration":   3600, // 1 hour in seconds
		"slice_size": 0.1,  // 10% of total quantity per slice
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
		log.Printf("Error creating algo order: %v\n", err)
		return
	}

	fmt.Printf("Algo Order Created - OrderID: %d, Status: %s\n",
		response.OrderID, response.Status)
}

func createBasicOrder(client *versifi.Client) {
	fmt.Println("\n=== Creating Basic Order (LIMIT) ===")

	response, err := client.NewCreateBasicOrderService().
		Exchange(versifi.ExchangeBinanceSpot).
		OrderType(versifi.BasicOrderTypeLimit).
		Symbol("BTC/USDT").
		Side(versifi.SideTypeBuy).
		Quantity("0.5").
		Price("45000.00").
		TimeInForce(versifi.TimeInForceGTC).
		ClientOrderID(123457).
		Do(context.Background())

	if err != nil {
		log.Printf("Error creating basic order: %v\n", err)
		return
	}

	fmt.Printf("Basic Order Created - OrderID: %d, Status: %s\n",
		response.OrderID, response.Status)
}

func createPairOrder(client *versifi.Client) {
	fmt.Println("\n=== Creating Pair Order (BASIS) ===")

	// Define lead leg
	leadLeg := &versifi.PairLeg{
		Exchange:         versifi.ExchangeBinanceSpot,
		Symbol:           "BTC/USDT",
		OrderType:        "LIMIT",
		LegRatio:         versifi.Float64Ptr(1.0),
		MaxPositionLong:  versifi.StringPtr("100"),
		MaxPositionShort: versifi.StringPtr("50"),
		MaxNotionalLong:  versifi.StringPtr("5000"),
		MaxNotionalShort: versifi.StringPtr("2500"),
	}

	// Define secondary leg
	secondaryLeg := &versifi.PairLeg{
		Exchange:         versifi.ExchangeBinanceFutures,
		Symbol:           "BTC/USDT",
		OrderType:        "LIMIT",
		LegRatio:         versifi.Float64Ptr(1.0),
		MaxPositionLong:  versifi.StringPtr("100"),
		MaxPositionShort: versifi.StringPtr("50"),
	}

	// Pair order params
	params := map[string]interface{}{
		"entry_spread_threshold": 0.01,
		"exit_spread_threshold":  0.005,
		"max_slippage":           0.002,
	}

	response, err := client.NewCreatePairOrderService().
		OrderType(versifi.PairOrderTypeBasis).
		Lead(leadLeg).
		Secondary(secondaryLeg).
		Style(versifi.PairStyleSync).
		Params(params).
		ClientOrderID(123458).
		Do(context.Background())

	if err != nil {
		log.Printf("Error creating pair order: %v\n", err)
		return
	}

	fmt.Printf("Pair Order Created - OrderID: %d, Status: %s\n",
		response.OrderID, response.Status)
}

func getOrder(client *versifi.Client, orderID int64) {
	fmt.Printf("\n=== Getting Order ID: %d ===\n", orderID)

	response, err := client.NewGetOrderService().
		OrderID(orderID).
		Do(context.Background())

	if err != nil {
		log.Printf("Error getting order: %v\n", err)
		return
	}

	fmt.Printf("Order Details - Type: %s, Status: %s, Timestamp: %d\n",
		response.OrderType, response.Status, response.Timestamp)

	// Check order type and print specific details
	if response.BasicOrder != nil {
		fmt.Printf("  Basic Order - Symbol: %s, Side: %s, Quantity: %s\n",
			response.BasicOrder.Symbol,
			response.BasicOrder.Side,
			response.BasicOrder.Quantity)
	}

	if response.AlgoOrder != nil {
		fmt.Printf("  Algo Order - Symbol: %s, Side: %s, Quantity: %s, Type: %s\n",
			response.AlgoOrder.Symbol,
			response.AlgoOrder.Side,
			response.AlgoOrder.Quantity,
			response.AlgoOrder.OrderType)
	}

	if response.PairOrder != nil {
		fmt.Println("  Pair Order Details:")
		if response.PairOrder.LeadLeg != nil {
			fmt.Printf("    Lead Leg - Symbol: %s, Exchange: %s\n",
				response.PairOrder.LeadLeg.Symbol,
				response.PairOrder.LeadLeg.Exchange)
		}
	}
}

func cancelOrder(client *versifi.Client, orderID int64) {
	fmt.Printf("\n=== Canceling Order ID: %d ===\n", orderID)

	err := client.NewCancelOrderService().
		OrderID(orderID).
		Do(context.Background())

	if err != nil {
		log.Printf("Error canceling order: %v\n", err)
		return
	}

	fmt.Printf("Order %d canceled successfully (status will be sent via WebSocket)\n", orderID)
}

func cancelBatchOrders(client *versifi.Client, orderIDs []int64) {
	fmt.Printf("\n=== Canceling Batch Orders: %v ===\n", orderIDs)

	err := client.NewCancelBatchOrderService().
		OrderIDs(orderIDs).
		Do(context.Background())

	if err != nil {
		log.Printf("Error canceling batch orders: %v\n", err)
		return
	}

	fmt.Println("Batch orders canceled successfully")
}
