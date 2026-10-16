// Package model defines the core domain entities, data transfer objects, and value objects
// used throughout the Prim backend services.
package model

import (
	"time"

	"github.com/google/uuid"
)

// ChallengeChannel represents the transmission channel used to deliver the challenge OTP.
type ChallengeChannel string

const (
	// ChallengeChannelEmail delivers the OTP via email.
	ChallengeChannelEmail ChallengeChannel = "email"
	// ChallengeChannelSMS delivers the OTP via SMS.
	ChallengeChannelSMS ChallengeChannel = "sms"
)

// ChallengeStatus represents the lifecycle state of an authentication challenge.
type ChallengeStatus string

const (
	// ChallengeStatusPending indicates the challenge is awaiting verification.
	ChallengeStatusPending ChallengeStatus = "pending"
	// ChallengeStatusVerified indicates the challenge OTP was successfully verified.
	ChallengeStatusVerified ChallengeStatus = "verified"
	// ChallengeStatusExpired indicates the challenge lifetime has elapsed.
	ChallengeStatusExpired ChallengeStatus = "expired"
	// ChallengeStatusFailed indicates the challenge exceeded maximum verification attempts.
	ChallengeStatusFailed ChallengeStatus = "failed"
)

// Challenge represents an OTP authentication or verification challenge issued to a user.
// It tracks delivery medium, cryptographic hash of the generated secret, expiration timestamps,
// and retry/resend rate-limiting counters.
type Challenge struct {
	// ID is the unique challenge identifier (UUID string).
	ID string
	// Identifier is the target destination (e.g. normalized email address or E.164 phone number).
	Identifier string
	// Channel is the transmission medium (e.g. ChallengeChannelEmail, ChallengeChannelSMS).
	Channel ChallengeChannel
	// OtpHash is the cryptographically hashed OTP secret (e.g. bcrypt/Argon2 or HMAC-SHA256).
	OtpHash string
	// Status is the challenge lifecycle state (pending, verified, expired, failed).
	Status ChallengeStatus
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
	identifier string,
	channel ChallengeChannel,
	otpHash string,
	ttl time.Duration,
) *Challenge {
	now := time.Now()
	return &Challenge{
		ID:          uuid.NewString(),
		Identifier:  identifier,
		Channel:     channel,
		OtpHash:     otpHash,
		Status:      ChallengeStatusPending,
		ResendCount: 0,
		Attempts:    0,
		ExpiresAt:   now.Add(ttl),
		CreatedAt:   now,
	}
}

// IsExpired returns true if the challenge status is marked as expired or the expiration timestamp has elapsed.
func (c *Challenge) IsExpired() bool {
	return c.Status == ChallengeStatusExpired || (!c.ExpiresAt.IsZero() && time.Now().After(c.ExpiresAt))
}
