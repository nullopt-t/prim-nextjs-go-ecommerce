package model

import (
	"time"

	"github.com/google/uuid"
)

// PaymentStatus represents the transaction processing status of a payment.
type PaymentStatus string

const (
	// StatusPending indicates the payment has been initialized and is awaiting authorization.
	StatusPending PaymentStatus = "pending"
	// StatusProcessing indicates the payment processor is currently settling the charge.
	StatusProcessing PaymentStatus = "processing"
	// StatusSucceeded indicates the payment charge was successfully captured.
	StatusSucceeded PaymentStatus = "succeeded"
	// StatusFailed indicates the payment attempt was rejected or failed.
	StatusFailed PaymentStatus = "failed"
	// StatusCancelled indicates the customer or system aborted the payment session.
	StatusCancelled PaymentStatus = "cancelled"
)

// Payment represents a customer payment transaction record or stored payment method.
type Payment struct {
	// ID is the unique database UUID for the payment record.
	ID uuid.UUID
	// UserID is the associated Customer UUID.
	UserID uuid.UUID
	// Type is the payment method category (e.g. "card", "wallet", "bank_transfer").
	Type string
	// Provider is the payment gateway provider name (e.g. "stripe", "paypal").
	Provider string
	// Token is the external processor transaction reference or gateway token.
	Token string
	// CreatedAt is the timestamp when the payment record was created.
	CreatedAt time.Time
	// UpdatedAt is the timestamp when payment status was last updated.
	UpdatedAt time.Time
}
