package versifi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// GetOrderService retrieves order details by ID
type GetOrderService struct {
	c       *Client
	orderID int64
}

// OrderID sets the order ID to retrieve
func (s *GetOrderService) OrderID(orderID int64) *GetOrderService {
	s.orderID = orderID
	return s
}

// GetOrderResponse represents the response structure for getting an order
type GetOrderResponse struct {
	OrderID          int64           `json:"order_id"`
	ClientOrderID    int64           `json:"client_order_id"`
	OrderType        string          `json:"order_type"`
	Status           OrderStatusType `json:"status"`
	Timestamp        int64           `json:"timestamp"`
	RequestOrderType string          `json:"request_order_type"`
	AlgoOrder        *AlgoOrderDetail `json:"algo_order,omitempty"`
	BasicOrder       *BasicOrderDetail `json:"basic_order,omitempty"`
	PairOrder        *PairOrderDetail `json:"pair_order,omitempty"`
}

// AlgoOrderDetail represents algo order details
type AlgoOrderDetail struct {
	Exchange            ExchangeType    `json:"exchange"`
	OrderType           AlgoOrderType   `json:"order_type"`
	Quantity            string          `json:"quantity"`
	QuoteOrderQuantity  string          `json:"quote_order_quantity,omitempty"`
	Side                SideType        `json:"side"`
	Symbol              string          `json:"symbol"`
	OrderParams         json.RawMessage `json:"order_params,omitempty"`
	AveragePrice        string          `json:"average_price,omitempty"`
	FilledQuantity      string          `json:"filled_quantity,omitempty"`
	RejectReason        string          `json:"reject_reason,omitempty"`
	TIF                 TimeInForceType `json:"tif,omitempty"`
	ChildOrders         []ChildOrder    `json:"child_orders,omitempty"`
}

// BasicOrderDetail represents basic order details
type BasicOrderDetail struct {
	Exchange            ExchangeType    `json:"exchange"`
	OrderType           BasicOrderType  `json:"order_type"`
	Price               string          `json:"price,omitempty"`
	Quantity            string          `json:"quantity"`
	QuoteOrderQuantity  string          `json:"quote_order_quantity,omitempty"`
	Side                SideType        `json:"side"`
	StopPrice           string          `json:"stop_price,omitempty"`
	Symbol              string          `json:"symbol"`
	TIF                 TimeInForceType `json:"tif,omitempty"`
	TrailingDelta       string          `json:"trailing_delta,omitempty"`
	AveragePrice        string          `json:"average_price,omitempty"`
	FilledQuantity      string          `json:"filled_quantity,omitempty"`
	RejectReason        string          `json:"reject_reason,omitempty"`
	ChildOrders         []ChildOrder    `json:"child_orders,omitempty"`
}

// PairOrderDetail represents pair order details
type PairOrderDetail struct {
	LeadLeg       *PairLegDetail     `json:"lead_leg,omitempty"`
	Secondary     *PairLegDetail     `json:"leg,omitempty"`
	Params        json.RawMessage    `json:"params,omitempty"`
	RejectReason  string             `json:"reject_reason,omitempty"`
	Style         PairStyleType      `json:"style,omitempty"`
}

// PairLegDetail represents details of a pair leg
type PairLegDetail struct {
	Symbol           string          `json:"symbol"`
	Exchange         ExchangeType    `json:"exchange"`
	OrderType        string          `json:"order_type"`
	LegRatio         float64         `json:"leg_ratio"`
	MaxPositionLong  string          `json:"max_position_long,omitempty"`
	MaxPositionShort string          `json:"max_position_short,omitempty"`
	MaxNotionalLong  string          `json:"max_notional_long,omitempty"`
	MaxNotionalShort string          `json:"max_notional_short,omitempty"`
	ChildOrders      []ChildOrder    `json:"child_order,omitempty"`
}

// ChildOrder represents a child order and its trades
type ChildOrder struct {
	ID                 int64           `json:"id,omitempty"`
	ChildOrderID       int64           `json:"child_order_id,omitempty"`
	OrderID            int64           `json:"order_id,omitempty"`
	Exchange           ExchangeType    `json:"exchange,omitempty"`
	ExchangeOrderID    string          `json:"exchange_order_id,omitempty"`
	Symbol             string          `json:"symbol,omitempty"`
	OrderType          string          `json:"order_type,omitempty"`
	Price              string          `json:"price,omitempty"`
	Quantity           string          `json:"quantity,omitempty"`
	Side               SideType        `json:"side,omitempty"`
	OrderStatus        OrderStatusType `json:"order_status,omitempty"`
	AveragePrice       string          `json:"average_price,omitempty"`
	FilledQuantity     string          `json:"filled_quantity,omitempty"`
	RejectReason       string          `json:"reject_reason,omitempty"`
	LegID              int64           `json:"leg_id,omitempty"`
	Trades             []Trade         `json:"trades,omitempty"`
}

// Trade represents a trade execution
type Trade struct {
	TradeID         int64        `json:"trade_id"`
	OrderID         int64        `json:"order_id"`
	ChildOrderID    int64        `json:"child_order_id"`
	ExchangeTradeID string       `json:"exchange_trade_id"`
	Exchange        ExchangeType `json:"exchange"`
	Symbol          string       `json:"symbol"`
	Price           string       `json:"price"`
	Quantity        string       `json:"quantity"`
	Side            SideType     `json:"side"`
	Fee             string       `json:"fee"`
	LegID           int64        `json:"leg_id"`
}

// Do executes the request
func (s *GetOrderService) Do(ctx context.Context, opts ...RequestOption) (res *GetOrderResponse, err error) {
	r := &request{
		method:   http.MethodGet,
		endpoint: fmt.Sprintf("/v2/orders/%d", s.orderID),
		secType:  secTypeSigned,
	}

	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}

	res = new(GetOrderResponse)
	err = json.Unmarshal(data, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}
