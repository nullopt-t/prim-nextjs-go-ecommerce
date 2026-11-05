package vendors

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
// caller's lines, and only those lines. Read-only by design — vendors see
// money and status, fulfillment actions stay with admins.
func (s *Service) ListOwnOrders(
	ctx context.Context,
	userID uuid.UUID,
	status *model.OrderStatus,
	q *pagination.ListQuery,
) (*pagination.PagedValueResult[model.Order], error) {
	var result *pagination.PagedValueResult[model.Order]
	err := s.runner.WithExecutor(ctx, func(ctx context.Context) error {
		vendor, err := s.repo.GetByUserID(ctx, userID)
		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.NotFoundError("Vendor not found").WithCode(errcode.CodeVendorNotFound)
			}
			return s.handleError(err, "Failed to load vendor")
		}

		orders, total, err := s.orders.ListVendorOrders(ctx, vendor.ID, status, q)
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
		vendor, err := s.repo.GetByUserID(ctx, userID)
		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.NotFoundError("Vendor not found").WithCode(errcode.CodeVendorNotFound)
			}
			return s.handleError(err, "Failed to load vendor")
		}

		order, err = s.orders.GetVendorOrder(ctx, vendor.ID, orderID)
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
