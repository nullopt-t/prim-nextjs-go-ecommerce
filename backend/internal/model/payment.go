package model

import (
	"time"

	"github.com/google/uuid"
)

// PaymentStatus represents the transaction processing status of a payment.
// Aligned with the database ENUM ('pending', 'authorized', 'captured', 'failed', 'refunded').
type PaymentStatus string

const (
	// StatusPending indicates the payment has been initialized and is awaiting authorization.
	StatusPending PaymentStatus = "pending"
	// StatusAuthorized indicates the payment has been pre-authorized by the provider.
	StatusAuthorized PaymentStatus = "authorized"
	// StatusCaptured indicates the payment charge was successfully captured.
	StatusCaptured PaymentStatus = "captured"
	// StatusFailed indicates the payment attempt was rejected or failed.
	StatusFailed PaymentStatus = "failed"
	// StatusRefunded indicates the payment has been refunded.
	StatusRefunded PaymentStatus = "refunded"
)

// Payment represents a customer payment transaction record matching the 'payments' table schema.
type Payment struct {
	// ID is the unique database UUID for the payment record.
	ID uuid.UUID
	// OrderID is the associated Order UUID.
	OrderID uuid.UUID
	// Amount is the payment amount in the smallest currency unit (e.g. cents).
	Amount int64
	// Currency is the ISO currency code (e.g. "USD").
	Currency string
	// Status is the payment status.
	Status PaymentStatus
	// Type is the payment method category (e.g. "card", "wallet", "bank_transfer").
	Type string
	// Provider is the payment gateway provider name (e.g. "stripe", "paypal").
	Provider string
	// ProviderTransactionID is the external processor transaction reference or gateway token.
	ProviderTransactionID *string
	// ErrorMessage stores failure reason or decline description if payment failed.
	ErrorMessage *string
	// CreatedAt is the timestamp when the payment record was created.
	CreatedAt time.Time
	// UpdatedAt is the timestamp when payment status was last updated.
	UpdatedAt time.Time
}
