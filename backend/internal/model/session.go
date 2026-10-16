// Package model defines core business entities, validation methods, and domain states.
package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// SessionType defines whether a session belongs to an anonymous storefront visitor or an authenticated account.
type SessionType string

const (
	// SessionTypeGuest indicates an unauthenticated session used for guest cart and browsing tracking.
	SessionTypeGuest         SessionType = "guest"
	// SessionTypeAuthenticated indicates an active session authenticated with user credentials.
	SessionTypeAuthenticated SessionType = "authenticated"
)

// UserSession represents an active user login session stored in Redis.
type UserSession struct {
	// ID is the unique random session token stored in client cookies.
	ID         string      `json:"id"`
	// UserID is the database UUID of the authenticated user (uuid.Nil for guest sessions).
	UserID     uuid.UUID   `json:"userId"`
	// UserRole specifies the authorization role (e.g., "customer", "admin").
	UserRole   UserRole    `json:"userRole,omitempty"`
	// Type indicates guest or authenticated session lifecycle state.
	Type       SessionType `json:"type"`
	// UserAgent records the client browser or device signature.
	UserAgent  string      `json:"userAgent"`
	// IPAddress records the client network IP address at session creation.
	IPAddress  string      `json:"ipAddress"`
	// CreatedAt is the timestamp when the session was created.
	CreatedAt  time.Time   `json:"createdAt"`
	// LastActive is the timestamp of the latest request made with this session.
	LastActive time.Time   `json:"lastActive"`
}

// Validate ensures structural integrity and authorization invariants for guest vs authenticated sessions.
func (s *UserSession) Validate() error {
	switch s.Type {
	case SessionTypeGuest:
		s.UserID = uuid.Nil
		s.UserRole = ""
	case SessionTypeAuthenticated:
		if s.UserID == uuid.Nil {
			return errors.New("authenticated session requires non-nil user ID")
		}
		if s.UserRole == "" {
			s.UserRole = CustomerRole
		}
	}
	return nil
}

