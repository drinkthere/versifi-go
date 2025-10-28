package versifi

import "fmt"

// APIError represents an error from the Versifi API
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("<APIError> code=%d, message=%s", e.Code, e.Message)
}

// IsAPIError checks if an error is an API error
func IsAPIError(e error) bool {
	_, ok := e.(*APIError)
	return ok
}

// Common types and enums

// SideType represents order side
type SideType string

const (
	SideTypeBuy  SideType = "BUY"
	SideTypeSell SideType = "SELL"
)

// ExchangeType represents the exchange
type ExchangeType string

const (
	ExchangeBinanceSpot    ExchangeType = "BINANCE_SPOT"
	ExchangeBinanceFutures ExchangeType = "BINANCE_FUTURES"
	ExchangeOKXSpot        ExchangeType = "OKX_SPOT"
	ExchangeOKXFutures     ExchangeType = "OKX_FUTURES"
)

// AlgoOrderType represents algorithm order types
type AlgoOrderType string

const (
	AlgoOrderTypeTWAP AlgoOrderType = "TWAP"
	AlgoOrderTypeVWAP AlgoOrderType = "VWAP"
	AlgoOrderTypeIS   AlgoOrderType = "IS"
)

// BasicOrderType represents basic order types
type BasicOrderType string

const (
	BasicOrderTypeMarket          BasicOrderType = "MARKET"
	BasicOrderTypeLimit           BasicOrderType = "LIMIT"
	BasicOrderTypeStop            BasicOrderType = "STOP"
	BasicOrderTypeStopLoss        BasicOrderType = "STOP_LOSS"
	BasicOrderTypeStopLossLimit   BasicOrderType = "STOP_LOSS_LIMIT"
	BasicOrderTypeTakeProfit      BasicOrderType = "TAKE_PROFIT"
	BasicOrderTypeTakeProfitLimit BasicOrderType = "TAKE_PROFIT_LIMIT"
	BasicOrderTypeLimitMaker      BasicOrderType = "LIMIT_MAKER"
)

// PairOrderType represents pair order types
type PairOrderType string

const (
	PairOrderTypeBasis PairOrderType = "BASIS"
)

// TimeInForceType represents time in force
type TimeInForceType string

const (
	TimeInForceFOK     TimeInForceType = "FOK"
	TimeInForceGTC     TimeInForceType = "GTC"
	TimeInForceGTD     TimeInForceType = "GTD"
	TimeInForceIOC     TimeInForceType = "IOC"
	TimeInForceGTX     TimeInForceType = "GTX"
	TimeInForcePostOn  TimeInForceType = "POST_ON"
)

// OrderStatusType represents order status
type OrderStatusType string

const (
	OrderStatusNew             OrderStatusType = "NEW"
	OrderStatusPartiallyFilled OrderStatusType = "PARTIALLY_FILLED"
	OrderStatusFilled          OrderStatusType = "FILLED"
	OrderStatusCanceled        OrderStatusType = "CANCELED"
	OrderStatusRejected        OrderStatusType = "REJECTED"
	OrderStatusExpired         OrderStatusType = "EXPIRED"
)

// PairStyleType represents pair order style
type PairStyleType string

const (
	PairStyleSync  PairStyleType = "SYNC"
	PairStyleAsync PairStyleType = "ASYNC"
	PairStyleTWAP  PairStyleType = "TWAP"
)

// OrderResponse represents the common order response structure
type OrderResponse struct {
	OrderID         int64           `json:"order_id"`
	ClientOrderID   int64           `json:"client_order_id"`
	Status          OrderStatusType `json:"status"`
	Lead            *LegResponse    `json:"lead,omitempty"`
	Secondary       *LegResponse    `json:"secondary,omitempty"`
}

// LegResponse represents a leg in the order response
type LegResponse struct {
	LegID  int64           `json:"leg_id"`
	Status OrderStatusType `json:"status"`
}

// Helper functions for pointer types

// StringPtr returns a pointer to the string value
func StringPtr(s string) *string {
	return &s
}

// Int64Ptr returns a pointer to the int64 value
func Int64Ptr(i int64) *int64 {
	return &i
}

// Float64Ptr returns a pointer to the float64 value
func Float64Ptr(f float64) *float64 {
	return &f
}
