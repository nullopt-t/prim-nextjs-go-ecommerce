package model

import (
	"time"

	"github.com/google/uuid"
)

type SessionType string

const (
	SessionTypeGuest         SessionType = "guest"
	SessionTypeAuthenticated SessionType = "authenticated"
)

type UserSession struct {
	ID         string      `json:"id"`
	UserID     uuid.UUID   `json:"userId"`
	UserRole   string      `json:"userRole,omitempty"`
	Type       SessionType `json:"type"`
	UserAgent  string      `json:"userAgent"`
	IPAddress  string      `json:"ipAddress"`
	CreatedAt  time.Time   `json:"createdAt"`
	LastActive time.Time   `json:"lastActive"`
}
