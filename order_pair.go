package versifi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

// CreatePairOrderService creates a pair order (BASIS algo)
type CreatePairOrderService struct {
	c             *Client
	clientOrderID *int64
	lead          *PairLeg
	orderType     PairOrderType
	params        map[string]interface{}
	secondary     *PairLeg
	style         *PairStyleType
}

// PairLeg represents a leg in a pair order
type PairLeg struct {
	Exchange         ExchangeType           `json:"exchange"`
	Symbol           string                 `json:"symbol"`
	OrderType        string                 `json:"order_type,omitempty"`
	LegRatio         *float64               `json:"leg_ratio,omitempty"`
	MaxPositionLong  *string                `json:"max_position_long,omitempty"`
	MaxPositionShort *string                `json:"max_position_short,omitempty"`
	MaxNotionalLong  *string                `json:"max_notional_long,omitempty"`
	MaxNotionalShort *string                `json:"max_notional_short,omitempty"`
	Params           map[string]interface{} `json:"params,omitempty"`
}

// ClientOrderID sets the client order ID
func (s *CreatePairOrderService) ClientOrderID(clientOrderID int64) *CreatePairOrderService {
	s.clientOrderID = &clientOrderID
	return s
}

// Lead sets the lead leg configuration
func (s *CreatePairOrderService) Lead(lead *PairLeg) *CreatePairOrderService {
	s.lead = lead
	return s
}

// OrderType sets the pair order type (BASIS)
func (s *CreatePairOrderService) OrderType(orderType PairOrderType) *CreatePairOrderService {
	s.orderType = orderType
	return s
}

// Params sets the algorithm parameters for pair trading
func (s *CreatePairOrderService) Params(params map[string]interface{}) *CreatePairOrderService {
	s.params = params
	return s
}

// Secondary sets the secondary leg configuration
func (s *CreatePairOrderService) Secondary(secondary *PairLeg) *CreatePairOrderService {
	s.secondary = secondary
	return s
}

// Style sets the pair order style (SYNC, ASYNC, TWAP)
func (s *CreatePairOrderService) Style(style PairStyleType) *CreatePairOrderService {
	s.style = &style
	return s
}

// PairOrderRequest represents the request body for creating a pair order
type PairOrderRequest struct {
	ClientOrderID *int64                 `json:"client_order_id,omitempty"`
	Lead          *PairOrderLead         `json:"lead"`
	Style         *PairStyleType         `json:"style,omitempty"`
}

// PairOrderLead represents the lead configuration in pair order request
type PairOrderLead struct {
	OrderType PairOrderType          `json:"order_type"`
	Params    map[string]interface{} `json:"params,omitempty"`
}

// PairOrderRequestWithLegs represents the full pair order request structure
type PairOrderRequestFull struct {
	ClientOrderID *int64                 `json:"client_order_id,omitempty"`
	Lead          *PairOrderLeadFull     `json:"lead"`
	Secondary     *PairLeg               `json:"secondary,omitempty"`
	Style         *PairStyleType         `json:"style,omitempty"`
}

// PairOrderLeadFull represents the lead leg with all parameters
type PairOrderLeadFull struct {
	OrderType PairOrderType          `json:"order_type"`
	Params    map[string]interface{} `json:"params,omitempty"`
	Exchange  ExchangeType           `json:"exchange,omitempty"`
	Symbol    string                 `json:"symbol,omitempty"`
	LegRatio  *float64               `json:"leg_ratio,omitempty"`
}

// Do executes the request
func (s *CreatePairOrderService) Do(ctx context.Context, opts ...RequestOption) (res *OrderResponse, err error) {
	r := &request{
		method:   http.MethodPost,
		endpoint: "/v2/orders/pair/",
		secType:  secTypeSigned,
	}

	// Build request body based on API documentation structure
	// The lead object contains order_type and params
	leadConfig := &PairOrderLeadFull{
		OrderType: s.orderType,
		Params:    s.params,
	}

	// If lead leg is provided, add its details to params or as separate fields
	if s.lead != nil {
		leadConfig.Exchange = s.lead.Exchange
		leadConfig.Symbol = s.lead.Symbol
		leadConfig.LegRatio = s.lead.LegRatio

		// Merge lead leg params if they exist
		if s.lead.Params != nil {
			if leadConfig.Params == nil {
				leadConfig.Params = make(map[string]interface{})
			}
			for k, v := range s.lead.Params {
				leadConfig.Params[k] = v
			}
		}
	}

	body := PairOrderRequestFull{
		ClientOrderID: s.clientOrderID,
		Lead:          leadConfig,
		Secondary:     s.secondary,
		Style:         s.style,
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
