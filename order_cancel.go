package versifi

import (
	"context"
	"fmt"
	"net/http"
)

// CancelOrderService cancels a specific order by ID
type CancelOrderService struct {
	c       *Client
	orderID int64
}

// OrderID sets the order ID to cancel
func (s *CancelOrderService) OrderID(orderID int64) *CancelOrderService {
	s.orderID = orderID
	return s
}

// Do executes the request
// Returns no content on success (HTTP 204), cancellation status sent via WebSocket
func (s *CancelOrderService) Do(ctx context.Context, opts ...RequestOption) error {
	r := &request{
		method:   http.MethodDelete,
		endpoint: fmt.Sprintf("/v2/orders/%d", s.orderID),
		secType:  secTypeSigned,
	}

	_, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return err
	}

	return nil
}
