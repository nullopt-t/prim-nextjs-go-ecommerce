package coupon

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/shared/errcode"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

type ValidationResult struct {
	ID              uuid.UUID `json:"id"`
	Code            string    `json:"code"`
	DiscountType    string    `json:"discountType"`
	DiscountValue   int64     `json:"discountValue"`
	DiscountAmount  int64     `json:"discountAmount"`
	Subtotal        int64     `json:"subtotal"`
	Shipping        int64     `json:"shipping"`
	Total           int64     `json:"total"`
	DiscountPercent *int      `json:"discountPercent,omitempty"`
	FixedDiscount   *float64  `json:"fixedDiscount,omitempty"`
	MinCartAmount   *int64    `json:"minCartAmount,omitempty"`
	Message         string    `json:"message"`
}

type CreateCouponInput struct {
	Code          string
	DiscountType  model.CouponDiscountType
	DiscountValue int64
	Currency      string
	MinCartAmount *int64
	UsageLimit    *int
	PerUserLimit  *int
	StartsAt      *time.Time
	ExpiresAt     *time.Time
	Status        model.CouponStatus
}

type CouponService struct {
	runner database.Runner
	repo   *CouponRepository
}

func NewService(runner database.Runner, repo *CouponRepository) *CouponService {
	return &CouponService{
		runner: runner,
		repo:   repo,
	}
}

// handleError maps repository errors to standard API errors.
func (s *CouponService) handleError(err error, fallbackMsg string) error {
	if err == nil {
		return nil
	}
	mapped := database.MapError(err)
	switch {
	case errors.Is(mapped, database.ErrNotFound):
		return apierr.NotFoundError("Coupon not found").WithCode(errcode.CodeCouponNotFound)
	case errors.Is(mapped, database.ErrConflict):
		return apierr.ConflictError("A coupon with this code already exists").WithCode(errcode.CodeCouponAlreadyExists)
	default:
		return apierr.InternalServerError(fallbackMsg).Wrap(err)
	}
}

// GetByID finds a coupon by its UUID identifier.
func (s *CouponService) GetByID(ctx context.Context, id uuid.UUID) (*model.Coupon, error) {
	var c *model.Coupon
	err := s.runner.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		c, err = s.repo.GetByID(ctx, db, id)
		return err
	})
	if err != nil {
		return nil, s.handleError(err, "Failed to retrieve coupon")
	}
	return c, nil
}

// GetByCode finds a coupon by code string.
func (s *CouponService) GetByCode(ctx context.Context, code string) (*model.Coupon, error) {
	var c *model.Coupon
	err := s.runner.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		c, err = s.repo.GetByCode(ctx, db, code)
		return err
	})
	if err != nil {
		return nil, s.handleError(err, "Failed to retrieve coupon")
	}
	return c, nil
}

// ListCoupons returns all coupons matching the given filter.
func (s *CouponService) ListCoupons(ctx context.Context, filter ListCouponFilter) ([]*model.Coupon, error) {
	var list []*model.Coupon
	err := s.runner.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		list, err = s.repo.List(ctx, db, filter)
		return err
	})
	if err != nil {
		return nil, s.handleError(err, "Failed to list coupons")
	}
	return list, nil
}

// CreateCoupon validates and creates a new promotional coupon.
func (s *CouponService) CreateCoupon(ctx context.Context, in CreateCouponInput) (*model.Coupon, error) {
	code := strings.ToUpper(strings.TrimSpace(in.Code))
	if code == "" {
		return nil, apierr.BadRequestError("Validation failed").
			WithCode(apierr.CodeValidationFailed).
			WithFields(api.FieldError{Field: "code", Message: "coupon code is required"})
	}

	if in.DiscountType != model.CouponDiscountTypePercentage && in.DiscountType != model.CouponDiscountTypeFixedAmount {
		return nil, apierr.BadRequestError("Discount type must be either 'percentage' or 'fixed_amount'").
			WithCode(apierr.CodeValidationFailed).
			WithFields(api.FieldError{Field: "discountType", Message: "invalid discount type"})
	}

	if in.DiscountType == model.CouponDiscountTypePercentage && (in.DiscountValue < 1 || in.DiscountValue > 100) {
		return nil, apierr.BadRequestError("Percentage discount value must be between 1 and 100").
			WithCode(apierr.CodeValidationFailed).
			WithFields(api.FieldError{Field: "discountValue", Message: "percentage discount must be 1-100"})
	}

	if in.DiscountType == model.CouponDiscountTypeFixedAmount && in.DiscountValue <= 0 {
		return nil, apierr.BadRequestError("Fixed amount discount value must be greater than zero").
			WithCode(apierr.CodeValidationFailed).
			WithFields(api.FieldError{Field: "discountValue", Message: "fixed discount must be > 0"})
	}

	currency := in.Currency
	if currency == "" {
		currency = "EGP"
	}

	status := in.Status
	if status == "" {
		status = model.CouponStatusActive
	}

	c := &model.Coupon{
		ID:            uuid.New(),
		Code:          code,
		DiscountType:  in.DiscountType,
		DiscountValue: in.DiscountValue,
		Currency:      currency,
		MinCartAmount: in.MinCartAmount,
		UsageLimit:    in.UsageLimit,
		PerUserLimit:  in.PerUserLimit,
		StartsAt:      in.StartsAt,
		ExpiresAt:     in.ExpiresAt,
		Status:        status,
	}

	err := s.runner.WithTx(ctx, func(tx database.QueryExecutor) error {
		return s.repo.Create(ctx, tx, c)
	})
	if err != nil {
		return nil, s.handleError(err, "Failed to create coupon")
	}

	return c, nil
}

// UpdateStatus changes the coupon lifecycle state.
func (s *CouponService) UpdateStatus(ctx context.Context, id uuid.UUID, status model.CouponStatus) error {
	switch status {
	case model.CouponStatusActive, model.CouponStatusDisabled, model.CouponStatusArchived:
	default:
		return apierr.BadRequestError("Invalid coupon status").
			WithCode(apierr.CodeValidationFailed)
	}

	err := s.runner.WithTx(ctx, func(tx database.QueryExecutor) error {
		return s.repo.UpdateStatus(ctx, tx, id, status)
	})
	if err != nil {
		return s.handleError(err, "Failed to update coupon status")
	}
	return nil
}

// Delete permanently removes a coupon.
func (s *CouponService) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.runner.WithTx(ctx, func(tx database.QueryExecutor) error {
		return s.repo.Delete(ctx, tx, id)
	})
	if err != nil {
		return s.handleError(err, "Failed to delete coupon")
	}
	return nil
}

// ValidateAndCalculate validates a coupon code against a gross cart subtotal and optional customer ID,
// and computes the authoritative discount and final totals.
func (s *CouponService) ValidateAndCalculate(
	ctx context.Context,
	code string,
	grossTotal int64,
	userID *uuid.UUID,
) (*ValidationResult, error) {
	trimmedCode := strings.TrimSpace(code)
	if trimmedCode == "" {
		return nil, apierr.BadRequestError("Coupon code is required").
			WithCode(apierr.CodeValidationFailed)
	}

	var result *ValidationResult

	err := s.runner.WithTx(ctx, func(db database.QueryExecutor) error {
		c, err := s.repo.GetByCode(ctx, db, trimmedCode)
		if err != nil {
			if errors.Is(err, database.ErrNotFound) {
				return apierr.BadRequestError(fmt.Sprintf("Coupon '%s' not found", trimmedCode)).
					WithCode(errcode.CodeCouponNotFound)
			}
			return apierr.InternalServerError("Failed to validate coupon").Wrap(err)
		}

		if c.Status != model.CouponStatusActive {
			return apierr.BadRequestError(fmt.Sprintf("Coupon '%s' is not active", c.Code)).
				WithCode(errcode.CodeCouponInactive)
		}

		now := time.Now()
		if c.StartsAt != nil && now.Before(*c.StartsAt) {
			return apierr.BadRequestError(fmt.Sprintf("Coupon '%s' has not started yet", c.Code)).
				WithCode(errcode.CodeCouponNotStarted)
		}
		if c.ExpiresAt != nil && now.After(*c.ExpiresAt) {
			return apierr.BadRequestError(fmt.Sprintf("Coupon '%s' has expired", c.Code)).
				WithCode(errcode.CodeCouponExpired)
		}

		// Minimum cart amount check
		if c.MinCartAmount != nil && *c.MinCartAmount > 0 && grossTotal < *c.MinCartAmount {
			minFormatted := float64(*c.MinCartAmount) / 100.0
			return apierr.BadRequestError(
				fmt.Sprintf("Coupon '%s' requires a minimum order amount of %.2f %s", c.Code, minFormatted, c.Currency),
			).WithCode(errcode.CodeCouponMinCartNotMet)
		}

		// Check global usage limit
		if c.UsageLimit != nil && *c.UsageLimit > 0 {
			usedCount, err := s.repo.GetRedemptionCount(ctx, db, c.ID)
			if err != nil {
				return apierr.InternalServerError("Failed to check coupon usage limit").Wrap(err)
			}
			if usedCount >= *c.UsageLimit {
				return apierr.BadRequestError(fmt.Sprintf("Coupon '%s' usage limit has been reached", c.Code)).
					WithCode(errcode.CodeCouponUsageLimitExceeded)
			}
		}

		// Check per-user limit
		if userID != nil {
			userLimit := 1
			if c.PerUserLimit != nil && *c.PerUserLimit > 0 {
				userLimit = *c.PerUserLimit
			}
			userUsedCount, err := s.repo.GetUserRedemptionCount(ctx, db, c.ID, *userID)
			if err != nil {
				return apierr.InternalServerError("Failed to check user coupon usage").Wrap(err)
			}
			if userUsedCount >= userLimit {
				return apierr.BadRequestError(fmt.Sprintf("You have already used coupon '%s' the maximum allowable times", c.Code)).
					WithCode(errcode.CodeCouponUserLimitExceeded)
			}
		}

		res := &ValidationResult{
			ID:            c.ID,
			Code:          c.Code,
			DiscountType:  string(c.DiscountType),
			DiscountValue: c.DiscountValue,
			MinCartAmount: c.MinCartAmount,
			Message:       fmt.Sprintf("Coupon %s applied successfully!", c.Code),
		}

		var calculatedDiscount int64
		if c.DiscountType == model.CouponDiscountTypePercentage {
			pct := int(c.DiscountValue)
			res.DiscountPercent = &pct
			if grossTotal > 0 {
				calculatedDiscount = (grossTotal * c.DiscountValue) / 100
			}
		} else if c.DiscountType == model.CouponDiscountTypeFixedAmount {
			fixed := float64(c.DiscountValue) / 100.0
			res.FixedDiscount = &fixed
			calculatedDiscount = c.DiscountValue
		}

		if calculatedDiscount > grossTotal && grossTotal > 0 {
			calculatedDiscount = grossTotal
		}

		// Standard shipping: free at >= 15000 cents (150 EGP), 999 cents (9.99 EGP) otherwise
		var shippingCost int64 = 0
		if grossTotal > 0 && grossTotal < 15000 {
			shippingCost = 999
		}

		finalTotal := grossTotal - calculatedDiscount + shippingCost
		if finalTotal < 0 {
			finalTotal = 0
		}

		res.Subtotal = grossTotal
		res.DiscountAmount = calculatedDiscount
		res.Shipping = shippingCost
		res.Total = finalTotal

		result = res
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// ValidateAndCalculateByID validates a coupon by its UUID identifier against a gross cart subtotal and customer ID.
func (s *CouponService) ValidateAndCalculateByID(
	ctx context.Context,
	couponID uuid.UUID,
	grossTotal int64,
	userID *uuid.UUID,
) (*ValidationResult, error) {
	c, err := s.GetByID(ctx, couponID)
	if err != nil {
		return nil, err
	}
	return s.ValidateAndCalculate(ctx, c.Code, grossTotal, userID)
}

// RecordRedemption records a coupon redemption during checkout.
func (s *CouponService) RecordRedemption(
	ctx context.Context,
	couponID uuid.UUID,
	userID uuid.UUID,
	orderID uuid.UUID,
	discountAmount int64,
) error {
	red := &model.CouponRedemption{
		ID:             uuid.New(),
		CouponID:       couponID,
		UserID:         userID,
		OrderID:        orderID,
		DiscountAmount: discountAmount,
	}
	return s.runner.WithTx(ctx, func(tx database.QueryExecutor) error {
		return s.repo.RecordRedemption(ctx, tx, red)
	})
}
