package sellers

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
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
