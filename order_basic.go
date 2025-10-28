package versifi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

// CreateBasicOrderService creates a basic order (MARKET, LIMIT, STOP, etc.)
type CreateBasicOrderService struct {
	c               *Client
	clientOrderID   *int64
	exchange        ExchangeType
	orderType       BasicOrderType
	price           *string
	quantity        string
	side            SideType
	startTime       *int64
	stopPrice       *string
	symbol          string
	tif             *TimeInForceType
	trailingDelta   *string
}

// ClientOrderID sets the client order ID
func (s *CreateBasicOrderService) ClientOrderID(clientOrderID int64) *CreateBasicOrderService {
	s.clientOrderID = &clientOrderID
	return s
}

// Exchange sets the exchange
func (s *CreateBasicOrderService) Exchange(exchange ExchangeType) *CreateBasicOrderService {
	s.exchange = exchange
	return s
}

// OrderType sets the basic order type
func (s *CreateBasicOrderService) OrderType(orderType BasicOrderType) *CreateBasicOrderService {
	s.orderType = orderType
	return s
}

// Price sets the price (required for LIMIT orders)
func (s *CreateBasicOrderService) Price(price string) *CreateBasicOrderService {
	s.price = &price
	return s
}

// Quantity sets the quantity
func (s *CreateBasicOrderService) Quantity(quantity string) *CreateBasicOrderService {
	s.quantity = quantity
	return s
}

// Side sets the order side
func (s *CreateBasicOrderService) Side(side SideType) *CreateBasicOrderService {
	s.side = side
	return s
}

// StartTime sets the start timestamp (UTC Epoch Microseconds)
func (s *CreateBasicOrderService) StartTime(startTime int64) *CreateBasicOrderService {
	s.startTime = &startTime
	return s
}

// StopPrice sets the stop price (for STOP_LOSS, STOP_LOSS_LIMIT, TAKE_PROFIT, TAKE_PROFIT_LIMIT)
func (s *CreateBasicOrderService) StopPrice(stopPrice string) *CreateBasicOrderService {
	s.stopPrice = &stopPrice
	return s
}

// Symbol sets the trading symbol (format: Asset/Currency, e.g., BTC/USD)
func (s *CreateBasicOrderService) Symbol(symbol string) *CreateBasicOrderService {
	s.symbol = symbol
	return s
}

// TimeInForce sets the time in force (defaults to GTC if not specified)
func (s *CreateBasicOrderService) TimeInForce(tif TimeInForceType) *CreateBasicOrderService {
	s.tif = &tif
	return s
}

// TrailingDelta sets the trailing delta for trailing stop orders
func (s *CreateBasicOrderService) TrailingDelta(trailingDelta string) *CreateBasicOrderService {
	s.trailingDelta = &trailingDelta
	return s
}

// BasicOrderRequest represents the request body for creating a basic order
type BasicOrderRequest struct {
	ClientOrderID *int64          `json:"client_order_id,omitempty"`
	Exchange      ExchangeType    `json:"exchange"`
	OrderType     BasicOrderType  `json:"order_type"`
	Price         *string         `json:"price,omitempty"`
	Quantity      string          `json:"quantity"`
	Side          SideType        `json:"side"`
	StartTime     *int64          `json:"start_time,omitempty"`
	StopPrice     *string         `json:"stop_price,omitempty"`
	Symbol        string          `json:"symbol"`
	TIF           *TimeInForceType `json:"tif,omitempty"`
	TrailingDelta *string         `json:"trailing_delta,omitempty"`
}

// Do executes the request
func (s *CreateBasicOrderService) Do(ctx context.Context, opts ...RequestOption) (res *OrderResponse, err error) {
	r := &request{
		method:   http.MethodPost,
		endpoint: "/v2/orders/basic/",
		secType:  secTypeSigned,
	}

	// Build request body
	body := BasicOrderRequest{
		ClientOrderID: s.clientOrderID,
		Exchange:      s.exchange,
		OrderType:     s.orderType,
		Price:         s.price,
		Quantity:      s.quantity,
		Side:          s.side,
		StartTime:     s.startTime,
		StopPrice:     s.stopPrice,
		Symbol:        s.symbol,
		TIF:           s.tif,
		TrailingDelta: s.trailingDelta,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	r.body = bytes.NewReader(bodyBytes)

	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}

	res = new(OrderResponse)
	err = json.Unmarshal(data, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}
