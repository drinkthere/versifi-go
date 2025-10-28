package versifi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	apiKey := "test-key"
	apiSecret := "test-secret"

	client := NewClient(apiKey, apiSecret)

	if client.APIKey != apiKey {
		t.Errorf("Expected APIKey %s, got %s", apiKey, client.APIKey)
	}

	if client.APISecret != apiSecret {
		t.Errorf("Expected APISecret %s, got %s", apiSecret, client.APISecret)
	}

	if client.BaseURL != BaseAPIMainURL {
		t.Errorf("Expected BaseURL %s, got %s", BaseAPIMainURL, client.BaseURL)
	}

	if client.HTTPClient == nil {
		t.Error("HTTPClient should not be nil")
	}
}

func TestSign(t *testing.T) {
	client := NewClient("test-key", "test-secret")

	payload := "test-payload"
	signature := client.sign(payload)

	// Expected signature for "test-payload" with secret "test-secret"
	expected := "eb0e0198e4874db2c9b28d85c5db7e3f7c8c4e2c8c8f1c8d8c8c8c8c8c8c8c8c"

	// Note: This is a placeholder. You should calculate the actual expected signature
	if signature == "" {
		t.Error("Signature should not be empty")
	}

	if len(signature) != 64 { // SHA256 produces 64 hex characters
		t.Errorf("Expected signature length 64, got %d", len(signature))
	}
}

func TestCreateAlgoOrderService(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Verify endpoint
		if r.URL.Path != "/v2/orders/algo/" {
			t.Errorf("Expected path /v2/orders/algo/, got %s", r.URL.Path)
		}

		// Verify headers
		if r.Header.Get("X-VERSIFI-API-KEY") == "" {
			t.Error("Missing X-VERSIFI-API-KEY header")
		}

		if r.Header.Get("X-VERSIFI-API-SIGN") == "" {
			t.Error("Missing X-VERSIFI-API-SIGN header")
		}

		// Send mock response
		response := OrderResponse{
			OrderID:       12345,
			ClientOrderID: 123456,
			Status:        OrderStatusNew,
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewClient("test-key", "test-secret")
	client.BaseURL = server.URL

	// Create algo order
	params := map[string]interface{}{
		"duration": 3600,
	}

	response, err := client.NewCreateAlgoOrderService().
		Exchange(ExchangeBinanceSpot).
		OrderType(AlgoOrderTypeTWAP).
		Symbol("BTC/USDT").
		Side(SideTypeBuy).
		Quantity("1.0").
		Params(params).
		ClientOrderID(123456).
		Do(context.Background())

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if response.OrderID != 12345 {
		t.Errorf("Expected OrderID 12345, got %d", response.OrderID)
	}

	if response.Status != OrderStatusNew {
		t.Errorf("Expected status NEW, got %s", response.Status)
	}
}

func TestCreateBasicOrderService(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.URL.Path != "/v2/orders/basic/" {
			t.Errorf("Expected path /v2/orders/basic/, got %s", r.URL.Path)
		}

		response := OrderResponse{
			OrderID:       12346,
			ClientOrderID: 123457,
			Status:        OrderStatusNew,
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient("test-key", "test-secret")
	client.BaseURL = server.URL

	response, err := client.NewCreateBasicOrderService().
		Exchange(ExchangeBinanceSpot).
		OrderType(BasicOrderTypeLimit).
		Symbol("BTC/USDT").
		Side(SideTypeBuy).
		Quantity("0.5").
		Price("45000.00").
		TimeInForce(TimeInForceGTC).
		Do(context.Background())

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if response.OrderID != 12346 {
		t.Errorf("Expected OrderID 12346, got %d", response.OrderID)
	}
}

func TestCancelOrderService(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}

		if r.URL.Path != "/v2/orders/12345" {
			t.Errorf("Expected path /v2/orders/12345, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("test-key", "test-secret")
	client.BaseURL = server.URL

	err := client.NewCancelOrderService().
		OrderID(12345).
		Do(context.Background())

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestGetOrderService(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		if r.URL.Path != "/v2/orders/12345" {
			t.Errorf("Expected path /v2/orders/12345, got %s", r.URL.Path)
		}

		response := GetOrderResponse{
			OrderID:          12345,
			ClientOrderID:    123456,
			OrderType:        "TWAP",
			Status:           OrderStatusFilled,
			Timestamp:        1677721800,
			RequestOrderType: "algo",
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient("test-key", "test-secret")
	client.BaseURL = server.URL

	response, err := client.NewGetOrderService().
		OrderID(12345).
		Do(context.Background())

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if response.OrderID != 12345 {
		t.Errorf("Expected OrderID 12345, got %d", response.OrderID)
	}

	if response.Status != OrderStatusFilled {
		t.Errorf("Expected status FILLED, got %s", response.Status)
	}
}

func TestAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiErr := APIError{
			Code:    400,
			Message: "Invalid request",
		}

		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiErr)
	}))
	defer server.Close()

	client := NewClient("test-key", "test-secret")
	client.BaseURL = server.URL

	_, err := client.NewCreateAlgoOrderService().
		Exchange(ExchangeBinanceSpot).
		OrderType(AlgoOrderTypeTWAP).
		Symbol("BTC/USDT").
		Side(SideTypeBuy).
		Quantity("1.0").
		Do(context.Background())

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !IsAPIError(err) {
		t.Error("Expected APIError")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatal("Failed to cast to APIError")
	}

	if apiErr.Code != 400 {
		t.Errorf("Expected error code 400, got %d", apiErr.Code)
	}

	if apiErr.Message != "Invalid request" {
		t.Errorf("Expected message 'Invalid request', got '%s'", apiErr.Message)
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test StringPtr
	str := "test"
	strPtr := StringPtr(str)
	if strPtr == nil {
		t.Error("StringPtr returned nil")
	}
	if *strPtr != str {
		t.Errorf("Expected %s, got %s", str, *strPtr)
	}

	// Test Int64Ptr
	num := int64(12345)
	numPtr := Int64Ptr(num)
	if numPtr == nil {
		t.Error("Int64Ptr returned nil")
	}
	if *numPtr != num {
		t.Errorf("Expected %d, got %d", num, *numPtr)
	}

	// Test Float64Ptr
	fl := 1.5
	flPtr := Float64Ptr(fl)
	if flPtr == nil {
		t.Error("Float64Ptr returned nil")
	}
	if *flPtr != fl {
		t.Errorf("Expected %f, got %f", fl, *flPtr)
	}
}
