package model

import (
	"time"

	"github.com/google/uuid"
)

// Challenge represents an OTP authentication or verification challenge.
type Challenge struct {
	// ID is the unique challenge identifier.
	ID string
	// Identifier is the target destination (e.g. email address or phone number).
	Identifier string
	// Channel is the transmission medium (e.g. "email", "sms").
	Channel string
	// OtpHash is the cryptographically hashed OTP secret.
	OtpHash string
	// Status is the challenge lifecycle state ("pending", "verified", "expired", "failed").
	Status string
	// ResendCount tracks the number of times this challenge was resent to the user.
	ResendCount int
	// Attempts tracks the number of failed verification attempts against this challenge.
	Attempts int
	// LastResendAt is the timestamp of the most recent OTP resend.
	LastResendAt time.Time
	// ExpiresAt is the timestamp after which this challenge is invalid.
	ExpiresAt time.Time
	// CreatedAt is the timestamp when this challenge was created.
	CreatedAt time.Time
}

// NewChallenge creates a new pending authentication challenge with the provided TTL.
func NewChallenge(
	identifier,
	channel,
	otpHash string,
	ttl time.Duration,
) *Challenge {
	return &Challenge{
		ID:          uuid.NewString(),
		Identifier:  identifier,
		Channel:     channel,
		OtpHash:     otpHash,
		Status:      "pending",
		ResendCount: 1,
		Attempts:    0,
		ExpiresAt:   time.Now().Add(ttl),
		CreatedAt:   time.Now(),
	}
}

// IsExpired returns true if the challenge status is marked as expired.
func (c *Challenge) IsExpired() bool {
	return c.Status == "expired"
}
