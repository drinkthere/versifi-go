package versifi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

// CreateAlgoOrderService creates an algorithmic order (TWAP, VWAP, IS)
type CreateAlgoOrderService struct {
	c               *Client
	clientOrderID   *int64
	exchange        ExchangeType
	orderType       AlgoOrderType
	params          map[string]interface{}
	quantity        string
	side            SideType
	symbol          string
}

// ClientOrderID sets the client order ID
func (s *CreateAlgoOrderService) ClientOrderID(clientOrderID int64) *CreateAlgoOrderService {
	s.clientOrderID = &clientOrderID
	return s
}

// Exchange sets the exchange
func (s *CreateAlgoOrderService) Exchange(exchange ExchangeType) *CreateAlgoOrderService {
	s.exchange = exchange
	return s
}

// OrderType sets the algo order type (TWAP, VWAP, IS)
func (s *CreateAlgoOrderService) OrderType(orderType AlgoOrderType) *CreateAlgoOrderService {
	s.orderType = orderType
	return s
}

// Params sets the algorithm parameters
// For TWAP/VWAP: can include duration, slice_size, volume_percentage, etc.
// For IS: duration is required
func (s *CreateAlgoOrderService) Params(params map[string]interface{}) *CreateAlgoOrderService {
	s.params = params
	return s
}

// Quantity sets the quantity
func (s *CreateAlgoOrderService) Quantity(quantity string) *CreateAlgoOrderService {
	s.quantity = quantity
	return s
}

// Side sets the order side
func (s *CreateAlgoOrderService) Side(side SideType) *CreateAlgoOrderService {
	s.side = side
	return s
}

// Symbol sets the trading symbol (format: Asset/Currency, e.g., BTC/USD)
func (s *CreateAlgoOrderService) Symbol(symbol string) *CreateAlgoOrderService {
	s.symbol = symbol
	return s
}

// AlgoOrderRequest represents the request body for creating an algo order
type AlgoOrderRequest struct {
	ClientOrderID *int64                 `json:"client_order_id,omitempty"`
	Exchange      ExchangeType           `json:"exchange"`
	OrderType     AlgoOrderType          `json:"order_type"`
	Params        map[string]interface{} `json:"params,omitempty"`
	Quantity      string                 `json:"quantity"`
	Side          SideType               `json:"side"`
	Symbol        string                 `json:"symbol"`
}

// Do executes the request
func (s *CreateAlgoOrderService) Do(ctx context.Context, opts ...RequestOption) (res *OrderResponse, err error) {
	r := &request{
		method:   http.MethodPost,
		endpoint: "/v2/orders/algo/",
		secType:  secTypeSigned,
	}

	// Build request body
	body := AlgoOrderRequest{
		ClientOrderID: s.clientOrderID,
		Exchange:      s.exchange,
		OrderType:     s.orderType,
		Params:        s.params,
		Quantity:      s.quantity,
		Side:          s.side,
		Symbol:        s.symbol,
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
