package sellers

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/order"
	"github.com/nullopt-t/prim-backend/internal/shared/errcode"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

// ListOwnOrders returns the merchant's orders: only orders containing the
// caller's lines, and only those lines. Read-only by design — sellers see
// money and status, fulfillment actions stay with admins.
func (s *Service) ListOwnOrders(
	ctx context.Context,
	userID uuid.UUID,
	status *model.OrderStatus,
	q *pagination.ListQuery,
) (*pagination.PagedValueResult[model.Order], error) {
	var result *pagination.PagedValueResult[model.Order]
	err := s.runner.WithExecutor(ctx, func(ctx context.Context) error {
		seller, err := s.requireSellerWithStatus(ctx, userID, model.SellerStatusActive, model.SellerStatusSuspended)
		if err != nil {
			return err
		}

		orders, total, err := s.orders.ListSellerOrders(ctx, seller.ID, status, q)
		if err != nil {
			return s.handleError(err, "Failed to list orders")
		}

		result = pagination.NewPagedValueResult(orders, pagination.NewPage(q.Page, q.PageSize, int(total)))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetOwnOrder returns a single order when it contains the caller's line.
// Foreign and platform orders answer 404 so their existence is not leaked.
func (s *Service) GetOwnOrder(ctx context.Context, userID, orderID uuid.UUID) (*model.Order, error) {
	var order *model.Order
	err := s.runner.WithExecutor(ctx, func(ctx context.Context) error {
		seller, err := s.requireSellerWithStatus(ctx, userID, model.SellerStatusActive, model.SellerStatusSuspended)
		if err != nil {
			return err
		}

		order, err = s.orders.GetSellerOrder(ctx, seller.ID, orderID)
		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.NotFoundError("Order not found").WithCode(errcode.CodeOrderNotFound)
			}
			return s.handleError(err, "Failed to load order")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return order, nil
}

// FulfillOwnOrderInput carries the merchant's fulfillment advance.
type FulfillOwnOrderInput struct {
	Status         model.OrderStatus
	TrackingNumber string
	Carrier        *string
	TrackingURL    *string
	Notes          *string
}

// FulfillOwnOrder advances an order the caller owns to processing or
// shipped. Sellers cannot cancel, refund, or mark delivered — those remain
// admin/customer transitions.
func (s *Service) FulfillOwnOrder(ctx context.Context, userID, orderID uuid.UUID, in FulfillOwnOrderInput) (*model.Order, error) {
	trackingNumber := strings.TrimSpace(in.TrackingNumber)
	var carrier, trackingURL *string
	if in.Carrier != nil {
		if c := strings.TrimSpace(*in.Carrier); c != "" {
			carrier = &c
		}
	}
	if in.TrackingURL != nil {
		if u := strings.TrimSpace(*in.TrackingURL); u != "" {
			if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
				return nil, apierr.BadRequestError("Tracking URL must start with http:// or https://").
					WithCode(apierr.CodeInvalidInput)
			}
			trackingURL = &u
		}
	}
	var notes *string
	if in.Notes != nil {
		if n := strings.TrimSpace(*in.Notes); n != "" {
			notes = &n
		}
	}

	var result *model.Order
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		seller, err := s.requireActiveSeller(ctx, userID)
		if err != nil {
			return err
		}

		// Pre-check against the current state for a typed 400; the
		// repository re-checks under the row lock.
		current, err := s.orders.GetSellerOrder(ctx, seller.ID, orderID)
		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.NotFoundError("Order not found").WithCode(errcode.CodeOrderNotFound)
			}
			return s.handleError(err, "Failed to load order")
		}
		if !order.CanTransitionOrderStatus(current.Status, in.Status) {
			return apierr.BadRequestError(
				"Cannot transition order status from '"+string(current.Status)+"' to '"+string(in.Status)+"'",
			).WithCode(errcode.CodeInvalidOrderStatus)
		}
		if in.Status == model.OrderStatusShipped && trackingNumber == "" {
			return apierr.BadRequestError("Tracking number is required when shipping").
				WithCode(apierr.CodeInvalidInput)
		}

		result, err = s.orders.FulfillSellerOrder(ctx, seller.ID, orderID, in.Status, trackingNumber, carrier, trackingURL, notes)
		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.NotFoundError("Order not found").WithCode(errcode.CodeOrderNotFound)
			}
			return s.handleError(err, "Failed to fulfill order")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// FulfillOwnOrderItems fulfills specific line items belonging to the authenticated seller.
// When all line items across the entire order are fulfilled/shipped, the parent order transitions to Shipped.
func (s *Service) FulfillOwnOrderItems(
	ctx context.Context,
	userID, orderID uuid.UUID,
	itemIDs []uuid.UUID,
	trackingNumber, carrier string,
) (*model.Order, error) {
	trackingNumber = strings.TrimSpace(trackingNumber)
	if trackingNumber == "" {
		return nil, apierr.BadRequestError("Tracking number is required").
			WithCode(apierr.CodeInvalidInput)
	}
	carrier = strings.TrimSpace(carrier)
	if carrier == "" {
		return nil, apierr.BadRequestError("Carrier is required").
			WithCode(apierr.CodeInvalidInput)
	}

	var result *model.Order
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		seller, err := s.requireActiveSeller(ctx, userID)
		if err != nil {
			return err
		}

		result, err = s.orders.FulfillSellerOrderItems(ctx, seller.ID, orderID, itemIDs, trackingNumber, carrier)
		if err != nil {
			var apiErr *apierr.APIError
			if errors.As(err, &apiErr) {
				return apiErr
			}
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.NotFoundError("Order not found").WithCode(errcode.CodeOrderNotFound)
			}
			return s.handleError(err, "Failed to fulfill order items")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
