package vendors

import (
	"context"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/catalog/coupon"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
)

// CouponCatalog is the vendor portal's window into coupon persistence.
// Coupon rules (discount math, usage limits, active windows) stay in the
// coupon domain; the vendor service only scopes by owning merchant.
type CouponCatalog interface {
	ListVendorCoupons(ctx context.Context, vendorID uuid.UUID, limit int, cursor *pagination.Keyset) ([]*model.Coupon, *pagination.CursorPage, error)
	CreateVendorCoupon(ctx context.Context, vendorID uuid.UUID, in coupon.CreateCouponInput) (*model.Coupon, error)
	SetVendorCouponStatus(ctx context.Context, vendorID, couponID uuid.UUID, status model.CouponStatus) error
	DeleteVendorCoupon(ctx context.Context, vendorID, couponID uuid.UUID) error
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
		vendor, err := s.requireActiveVendor(ctx, userID)
		if err != nil {
			return err
		}
		list, page, err = s.coupons.ListVendorCoupons(ctx, vendor.ID, limit, cursor)
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
		vendor, err := s.requireActiveVendor(ctx, userID)
		if err != nil {
			return err
		}
		created, err = s.coupons.CreateVendorCoupon(ctx, vendor.ID, in)
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
		vendor, err := s.requireActiveVendor(ctx, userID)
		if err != nil {
			return err
		}
		return s.coupons.SetVendorCouponStatus(ctx, vendor.ID, couponID, status)
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
		vendor, err := s.requireActiveVendor(ctx, userID)
		if err != nil {
			return err
		}
		return s.coupons.DeleteVendorCoupon(ctx, vendor.ID, couponID)
	})
	if err != nil {
		return s.handleError(err, "Failed to delete promotion")
	}
	return nil
}
