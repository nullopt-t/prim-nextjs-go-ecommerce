package coupon

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/errcode"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

type ValidationResult struct {
	ID              uuid.UUID `json:"id"`
	Code            string    `json:"code"`
	DiscountType    string    `json:"discountType"`
	DiscountValue   int64     `json:"discountValue"`
	DiscountAmount  int64     `json:"discountAmount"`
	Currency        string    `json:"currency"`
	Subtotal        int64     `json:"subtotal"`
	Shipping        int64     `json:"shipping"`
	Total           int64     `json:"total"`
	DiscountPercent *int      `json:"discountPercent,omitempty"`
	FixedDiscount   *float64  `json:"fixedDiscount,omitempty"`
	MinCartAmount   *int64    `json:"minCartAmount,omitempty"`
}

// AppliedCoupon is the authoritative coupon eligibility and discount result for
// a given gross amount. It intentionally excludes shipping and tax: those are
// the caller's pricing rules, not the coupon domain's. Consumers that only need
// to enforce coupon rules (checkout, cart previews) should use this instead of
// reimplementing them against the coupon tables.
type AppliedCoupon struct {
	ID              uuid.UUID
	Code            string
	DiscountType    string
	DiscountValue   int64
	DiscountAmount  int64
	Currency        string
	DiscountPercent *int
	FixedDiscount   *float64
	MinCartAmount   *int64
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
	if runner == nil {
		panic("coupon: database runner is required")
	}
	if repo == nil {
		panic("coupon: coupon repository is required")
	}
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
	var apiErr *apierr.APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}
	mapped := database.MapError(err)
	switch {
	case errors.Is(mapped, database.ErrNotFound):
		return apierr.NotFoundError("Coupon not found").WithCode(errcode.CodeCouponNotFound)
	case errors.Is(mapped, database.ErrConflict):
		return apierr.ConflictError("A coupon with this code already exists").WithCode(errcode.CodeCouponAlreadyExists)
	case errors.Is(mapped, database.ErrForeignKeyViolation):
		return apierr.BadRequestError("Referenced record does not exist").WithCode(apierr.CodeInvalidInput)
	default:
		return apierr.InternalServerError(fallbackMsg).WithCode(apierr.CodeInternalError).Wrap(err).WithStack()
	}
}

// GetByID finds a coupon by its UUID identifier.
func (s *CouponService) GetByID(ctx context.Context, id uuid.UUID) (*model.Coupon, error) {
	var c *model.Coupon
	err := s.runner.WithExecutor(ctx, func(exec database.QueryExecutor) error {
		var err error
		c, err = s.repo.GetByID(ctx, exec, id)
		return err
	})
	if err != nil {
		return nil, s.handleError(err, "Failed to retrieve coupon")
	}
	return c, nil
}

// GetCodeByID resolves a coupon's code by its UUID identifier on the provided executor.
func (s *CouponService) GetCodeByID(ctx context.Context, qe database.QueryExecutor, id uuid.UUID) (string, error) {
	c, err := s.repo.GetByID(ctx, qe, id)
	if err != nil {
		return "", err
	}
	return c.Code, nil
}

// GetByCode finds a coupon by code string.
func (s *CouponService) GetByCode(ctx context.Context, code string) (*model.Coupon, error) {
	var c *model.Coupon
	err := s.runner.WithExecutor(ctx, func(exec database.QueryExecutor) error {
		var err error
		c, err = s.repo.GetByCode(ctx, exec, code)
		return err
	})
	if err != nil {
		return nil, s.handleError(err, "Failed to retrieve coupon")
	}
	return c, nil
}

// ListCoupons returns a keyset-paginated set of coupons matching the given filter.
func (s *CouponService) ListCoupons(ctx context.Context, filter ListCouponFilter, limit int, cursor *pagination.Keyset) ([]*model.Coupon, *pagination.CursorPage, error) {
	var list []*model.Coupon
	var page *pagination.CursorPage
	filter.Limit = limit
	filter.Cursor = cursor
	err := s.runner.WithExecutor(ctx, func(exec database.QueryExecutor) error {
		var err error
		list, page, err = s.repo.List(ctx, exec, filter)
		return err
	})
	if err != nil {
		return nil, nil, s.handleError(err, "Failed to list coupons")
	}
	return list, page, nil
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

	err := s.runner.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
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

	err := s.runner.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		return s.repo.UpdateStatus(ctx, tx, id, status)
	})
	if err != nil {
		return s.handleError(err, "Failed to update coupon status")
	}
	return nil
}

// Delete permanently removes a coupon.
func (s *CouponService) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.runner.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		return s.repo.Delete(ctx, tx, id)
	})
	if err != nil {
		return s.handleError(err, "Failed to delete coupon")
	}
	return nil
}

// ListRedemptions returns the usage log for a coupon, newest first. The coupon
// is verified first on the same connection so an unknown ID yields a 404 rather than an empty log.
func (s *CouponService) ListRedemptions(ctx context.Context, couponID uuid.UUID, limit int, cursor *pagination.Keyset) ([]*RedemptionEntry, *pagination.CursorPage, error) {
	if couponID == uuid.Nil {
		return nil, nil, apierr.BadRequestError("Coupon ID is required").WithCode(apierr.CodeInvalidInput)
	}

	var list []*RedemptionEntry
	var page *pagination.CursorPage
	err := s.runner.WithExecutor(ctx, func(exec database.QueryExecutor) error {
		if _, err := s.repo.GetByID(ctx, exec, couponID); err != nil {
			return err
		}
		var err error
		list, page, err = s.repo.ListRedemptions(ctx, exec, couponID, limit, cursor)
		return err
	})
	if err != nil {
		return nil, nil, s.handleError(err, "Failed to list coupon redemptions")
	}
	return list, page, nil
}



// ApplyByCode validates and prices a coupon by code on the caller's executor.
// It locks the coupon row so concurrent redemptions are serialized in transactions.
func (s *CouponService) ApplyByCode(
	ctx context.Context,
	tx *database.Tx,
	code string,
	grossTotal int64,
	userID *uuid.UUID,
) (*AppliedCoupon, error) {
	trimmedCode := strings.TrimSpace(code)
	if trimmedCode == "" {
		return nil, apierr.BadRequestError("Coupon code is required").
			WithCode(apierr.CodeValidationFailed)
	}

	c, err := s.repo.GetByCodeAndLock(ctx, tx, trimmedCode)
	if err != nil {
		return nil, err
	}
	
	if err := s.validateCoupon(ctx, tx, c, grossTotal, userID); err != nil {
		return nil, err
	}
	return s.computeDiscount(c, grossTotal), nil
}

// validateCoupon enforces the coupon invariants and computes the discount on the
// provided executor. Lookup errors are returned raw so callers can map them to
// their own presentation; rule violations are typed APIErrors.
func (s *CouponService) validateCoupon(
	ctx context.Context,
	qe database.QueryExecutor,
	c *model.Coupon,
	grossTotal int64,
	userID *uuid.UUID,
) error {
	if err := s.validateActiveWindow(c); err != nil {
		return err
	}
	if err := s.validateMinCartAmount(c, grossTotal); err != nil {
		return err
	}
	if err := s.validateGlobalUsageLimit(ctx, qe, c); err != nil {
		return err
	}
	if err := s.validateUserUsageLimit(ctx, qe, c, userID); err != nil {
		return err
	}
	return nil
}

func (s *CouponService) validateActiveWindow(c *model.Coupon) error {
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

	return nil
}

func (s *CouponService) validateMinCartAmount(c *model.Coupon, grossTotal int64) error {
	if c.MinCartAmount != nil && *c.MinCartAmount > 0 && grossTotal < *c.MinCartAmount {
		minFormatted := float64(*c.MinCartAmount) / 100.0
		return apierr.BadRequestError(
			fmt.Sprintf("Coupon '%s' requires a minimum order amount of %.2f %s", c.Code, minFormatted, c.Currency),
		).WithCode(errcode.CodeCouponMinCartNotMet)
	}
	return nil
}

func (s *CouponService) validateGlobalUsageLimit(ctx context.Context, qe database.QueryExecutor, c *model.Coupon) error {
	if c.UsageLimit == nil || *c.UsageLimit <= 0 {
		return nil
	}

	usedCount, err := s.repo.GetRedemptionCount(ctx, qe, c.ID)
	if err != nil {
		return err
	}
	if usedCount >= *c.UsageLimit {
		return apierr.BadRequestError(fmt.Sprintf("Coupon '%s' usage limit has been reached", c.Code)).
			WithCode(errcode.CodeCouponUsageLimitExceeded)
	}
	return nil
}

func (s *CouponService) validateUserUsageLimit(ctx context.Context, qe database.QueryExecutor, c *model.Coupon, userID *uuid.UUID) error {
	if userID == nil {
		return nil
	}

	userLimit := 1
	if c.PerUserLimit != nil && *c.PerUserLimit > 0 {
		userLimit = *c.PerUserLimit
	}
	userUsedCount, err := s.repo.GetUserRedemptionCount(ctx, qe, c.ID, *userID)
	if err != nil {
		return err
	}
	if userUsedCount >= userLimit {
		return apierr.BadRequestError(fmt.Sprintf("You have already used coupon '%s' the maximum allowable times", c.Code)).
			WithCode(errcode.CodeCouponUserLimitExceeded)
	}
	return nil
}

func (s *CouponService) computeDiscount(c *model.Coupon, grossTotal int64) *AppliedCoupon {
	currency := c.Currency
	if currency == "" {
		currency = "EGP"
	}

	applied := &AppliedCoupon{
		ID:            c.ID,
		Code:          c.Code,
		DiscountType:  string(c.DiscountType),
		DiscountValue: c.DiscountValue,
		MinCartAmount: c.MinCartAmount,
		Currency:      currency,
	}

	switch c.DiscountType {
	case model.CouponDiscountTypePercentage:
		pct := int(c.DiscountValue)
		applied.DiscountPercent = &pct
		if grossTotal > 0 {
			applied.DiscountAmount = (grossTotal * c.DiscountValue) / 100
		}
	case model.CouponDiscountTypeFixedAmount:
		fixed := float64(c.DiscountValue) / 100.0
		applied.FixedDiscount = &fixed
		applied.DiscountAmount = c.DiscountValue
	}

	if applied.DiscountAmount > grossTotal && grossTotal > 0 {
		applied.DiscountAmount = grossTotal
	}

	return applied
}

// calculateDiscount evaluates coupon constraints and calculates discount and totals on the provided connection.
func (s *CouponService) calculateDiscount(
	ctx context.Context,
	qe database.QueryExecutor,
	c *model.Coupon,
	grossTotal int64,
	userID *uuid.UUID,
) (*ValidationResult, error) {
	if err := s.validateCoupon(ctx, qe, c, grossTotal, userID); err != nil {
		return nil, err
	}

	applied := s.computeDiscount(c, grossTotal)

	// Standard shipping: free at >= 15000 cents (150 EGP), 999 cents (9.99 EGP) otherwise
	var shippingCost int64
	if grossTotal > 0 && grossTotal < 15000 {
		shippingCost = 999
	}

	finalTotal := grossTotal - applied.DiscountAmount + shippingCost
	if finalTotal < 0 {
		finalTotal = 0
	}

	return &ValidationResult{
		ID:              applied.ID,
		Code:            applied.Code,
		DiscountType:    applied.DiscountType,
		DiscountValue:   applied.DiscountValue,
		DiscountAmount:  applied.DiscountAmount,
		Currency:        applied.Currency,
		Subtotal:        grossTotal,
		Shipping:        shippingCost,
		Total:           finalTotal,
		DiscountPercent: applied.DiscountPercent,
		FixedDiscount:   applied.FixedDiscount,
		MinCartAmount:   applied.MinCartAmount,
	}, nil
}

// ValidateAndCalculate validates a coupon code against a gross cart subtotal and optional customer ID,
// and computes the authoritative discount and final totals over a single database connection.
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
	err := s.runner.WithExecutor(ctx, func(exec database.QueryExecutor) error {
		c, err := s.repo.GetByCode(ctx, exec, trimmedCode)
		if err != nil {
			return err
		}
		var calcErr error
		result, calcErr = s.calculateDiscount(ctx, exec, c, grossTotal, userID)
		return calcErr
	})
	if err != nil {
		return nil, s.handleError(err, "Failed to validate coupon")
	}

	return result, nil
}

func (s *CouponService) RecordRedemption(
	ctx context.Context,
	tx *database.Tx,
	couponID uuid.UUID,
	userID uuid.UUID,
	orderID uuid.UUID,
	discountAmount int64,
	currency string,
) error {
	if couponID == uuid.Nil || userID == uuid.Nil || orderID == uuid.Nil {
		return apierr.BadRequestError("Coupon ID, User ID, and Order ID are required").
			WithCode(apierr.CodeInvalidInput)
	}
	return s.repo.RecordRedemption(ctx, tx, &model.CouponRedemption{
		CouponID:       couponID,
		UserID:         userID,
		OrderID:        orderID,
		DiscountAmount: discountAmount,
		Currency:       currency,
	})
}
