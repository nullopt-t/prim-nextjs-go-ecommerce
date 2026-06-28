package model

import (
	"time"

	"github.com/google/uuid"
)

type Challenge struct {
	ID           string
	Identifier   string
	Channel      string
	OtpHash      string
	Status       string
	ResendCount  int
	Attempts     int
	LastResendAt time.Time
	ExpiresAt    time.Time
	CreatedAt    time.Time
}

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

func (c *Challenge) IsExpired() bool {
	return c.Status == "expired"
}
