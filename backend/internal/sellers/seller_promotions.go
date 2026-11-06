package sellers

import (
	"context"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/catalog/coupon"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
)

// CouponCatalog is the seller portal's window into coupon persistence.
// Coupon rules (discount math, usage limits, active windows) stay in the
// coupon domain; the seller service only scopes by owning merchant.
type CouponCatalog interface {
	ListSellerCoupons(ctx context.Context, sellerID uuid.UUID, limit int, cursor *pagination.Keyset) ([]*model.Coupon, *pagination.CursorPage, error)
	CreateSellerCoupon(ctx context.Context, sellerID uuid.UUID, in coupon.CreateCouponInput) (*model.Coupon, error)
	SetSellerCouponStatus(ctx context.Context, sellerID, couponID uuid.UUID, status model.CouponStatus) error
	DeleteSellerCoupon(ctx context.Context, sellerID, couponID uuid.UUID) error
}

// ListPromotions lists the merchant's coupons.
func (s *Service) ListPromotions(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
	cursor *pagination.Keyset,
) ([]*model.Coupon, *pagination.CursorPage, error) {
	var list []*model.Coupon
	var page *pagination.CursorPage
	err := s.runner.WithExecutor(ctx, func(ctx context.Context) error {
		seller, err := s.requireActiveSeller(ctx, userID)
		if err != nil {
			return err
		}
		list, page, err = s.coupons.ListSellerCoupons(ctx, seller.ID, limit, cursor)
		return err
	})
	if err != nil {
		return nil, nil, s.handleError(err, "Failed to list promotions")
	}
	return list, page, nil
}

// CreatePromotion creates a coupon owned by the merchant.
func (s *Service) CreatePromotion(
	ctx context.Context,
	userID uuid.UUID,
	in coupon.CreateCouponInput,
) (*model.Coupon, error) {
	var created *model.Coupon
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		seller, err := s.requireActiveSeller(ctx, userID)
		if err != nil {
			return err
		}
		created, err = s.coupons.CreateSellerCoupon(ctx, seller.ID, in)
		return err
	})
	if err != nil {
		return nil, s.handleError(err, "Failed to create promotion")
	}
	return created, nil
}

// SetPromotionStatus enables or disables a merchant-owned coupon.
func (s *Service) SetPromotionStatus(
	ctx context.Context,
	userID, couponID uuid.UUID,
	status model.CouponStatus,
) error {
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		seller, err := s.requireActiveSeller(ctx, userID)
		if err != nil {
			return err
		}
		return s.coupons.SetSellerCouponStatus(ctx, seller.ID, couponID, status)
	})
	if err != nil {
		return s.handleError(err, "Failed to update promotion")
	}
	return nil
}

// DeletePromotion permanently removes a merchant-owned coupon.
func (s *Service) DeletePromotion(
	ctx context.Context,
	userID, couponID uuid.UUID,
) error {
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		seller, err := s.requireActiveSeller(ctx, userID)
		if err != nil {
			return err
		}
		return s.coupons.DeleteSellerCoupon(ctx, seller.ID, couponID)
	})
	if err != nil {
		return s.handleError(err, "Failed to delete promotion")
	}
	return nil
}
