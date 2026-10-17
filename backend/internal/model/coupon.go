package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// CouponDiscountType represents whether the coupon applies a percentage or fixed amount discount.
type CouponDiscountType string

const (
	CouponDiscountTypePercentage  CouponDiscountType = "percentage"
	CouponDiscountTypeFixedAmount CouponDiscountType = "fixed_amount"
)

// CouponStatus represents the lifecycle state of a coupon.
type CouponStatus string

const (
	CouponStatusActive   CouponStatus = "active"
	CouponStatusDisabled CouponStatus = "disabled"
	CouponStatusArchived CouponStatus = "archived"
)

var (
	ErrInvalidCouponDiscountType = errors.New("invalid coupon discount type")
	ErrInvalidCouponStatus       = errors.New("invalid coupon status")
)

// Coupon represents a promotional discount voucher in the store catalog.
type Coupon struct {
	ID            uuid.UUID          `json:"id"`
	Code          string             `json:"code"`
	DiscountType  CouponDiscountType `json:"discountType"`
	DiscountValue int64              `json:"discountValue"`
	Currency      string             `json:"currency"`
	MinCartAmount *int64             `json:"minCartAmount,omitempty"`
	UsageLimit    *int               `json:"usageLimit,omitempty"`
	PerUserLimit  *int               `json:"perUserLimit,omitempty"`
	StartsAt      *time.Time         `json:"startsAt,omitempty"`
	ExpiresAt     *time.Time         `json:"expiresAt,omitempty"`
	Status        CouponStatus       `json:"status"`
	CreatedAt     time.Time          `json:"createdAt"`
	UpdatedAt     time.Time          `json:"updatedAt"`

	// UsageCount is the number of times the coupon has been redeemed. It is not a
	// stored column: it is populated by list queries for admin display.
	UsageCount int `json:"usageCount"`
}

// CouponRedemption tracks an individual application of a coupon by a user for an order.
type CouponRedemption struct {
	ID             uuid.UUID `json:"id"`
	CouponID       uuid.UUID `json:"couponId"`
	UserID         uuid.UUID `json:"userId"`
	OrderID        uuid.UUID `json:"orderId"`
	DiscountAmount int64     `json:"discountAmount"`
	CreatedAt      time.Time `json:"createdAt"`
}
