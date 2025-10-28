package versifi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

// CancelBatchOrderService cancels multiple orders by their IDs
type CancelBatchOrderService struct {
	c        *Client
	orderIDs []int64
}

// OrderIDs sets the order IDs to cancel
func (s *CancelBatchOrderService) OrderIDs(orderIDs []int64) *CancelBatchOrderService {
	s.orderIDs = orderIDs
	return s
}

// AddOrderID adds a single order ID to the batch
func (s *CancelBatchOrderService) AddOrderID(orderID int64) *CancelBatchOrderService {
	s.orderIDs = append(s.orderIDs, orderID)
	return s
}

// CancelBatchRequest represents the request body for batch cancellation
type CancelBatchRequest struct {
	IDs []int64 `json:"ids"`
}

// Do executes the request
// Returns no content on success (HTTP 204)
func (s *CancelBatchOrderService) Do(ctx context.Context, opts ...RequestOption) error {
	r := &request{
		method:   http.MethodDelete,
		endpoint: "/v2/orders/batch",
		secType:  secTypeSigned,
	}

	// Build request body
	body := CancelBatchRequest{
		IDs: s.orderIDs,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	r.body = bytes.NewReader(bodyBytes)

	_, err = s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return err
	}

	return nil
}
