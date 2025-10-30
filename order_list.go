package versifi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ListOpenOrdersService retrieves order details by ID
type ListOpenOrdersService struct {
	c      *Client
	limit  int64
	offset int64
	status OrderStatusType
}

func (s *ListOpenOrdersService) Limit(limit int64) *ListOpenOrdersService {
	s.limit = limit
	return s
}

func (s *ListOpenOrdersService) Offset(offset int64) *ListOpenOrdersService {
	s.offset = offset
	return s
}

func (s *ListOpenOrdersService) Status(status OrderStatusType) *ListOpenOrdersService {
	s.status = status
	return s
}

type ListOrderItem struct {
	OrderID          int64  `json:"order_id"`
	ClientOrderID    int64  `json:"client_order_id"`
	Status           string `json:"status"`
	Timestamp        int64  `json:"timestamp"`
	RequestOrderType string `json:"request_order_type"`
	RejectReason     string `json:"reject_reason"`
}

// Do executes the request
func (s *ListOpenOrdersService) Do(ctx context.Context, opts ...RequestOption) (orders []ListOrderItem, err error) {
	r := &request{
		method:   http.MethodGet,
		endpoint: "/v2/orders",
		secType:  secTypeSigned,
	}

	// 设置查询参数
	if s.limit > 0 {
		r.setParam("limit", fmt.Sprintf("%d", s.limit))
	}

	if s.offset > 0 {
		r.setParam("offset", fmt.Sprintf("%d", s.offset))
	}

	if s.status != "" {
		r.setParam("status", string(s.status))
	}

	data, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &orders)
	if err != nil {
		return nil, err
	}

	return orders, nil
}
