package model

import (
	"time"

	"github.com/google/uuid"
)

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	// StatusPending indicates the payment is pending
	StatusPending PaymentStatus = "pending"
	// StatusProcessing indicates the payment is being processed
	StatusProcessing PaymentStatus = "processing"
	// StatusSucceeded indicates the payment was successful
	StatusSucceeded PaymentStatus = "succeeded"
	// StatusFailed indicates the payment failed
	StatusFailed PaymentStatus = "failed"
	// StatusCancelled indicates the payment was cancelled
	StatusCancelled PaymentStatus = "cancelled"
)

// Payment represents a payment record
type Payment struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Type      string
	Provider  string
	Token     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
