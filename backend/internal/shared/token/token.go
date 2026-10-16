// Package token defines identity and claims representation structures across session and token handlers.
package token

import (
	"errors"
	"fmt"

	"github.com/nullopt-t/prim-backend/internal/model"

	"github.com/google/uuid"
)

// UserSubject represents the core authenticated subject identity extracted from tokens or sessions.
type UserSubject struct {
	// UserID is the unique identifier of the user in the database.
	UserID          uuid.UUID      `json:"user_id"`
	// UserRole indicates whether the user is an admin, customer, or other role.
	UserRole        model.UserRole `json:"user_role"`
	// IsEmailVerified flags whether the user's primary email address has completed OTP verification.
	IsEmailVerified bool           `json:"is_email_verified"`
}

// Validate verifies that UserID is non-nil and UserRole is a valid recognized role.
func (s UserSubject) Validate() error {
	if s.UserID == uuid.Nil {
		return errors.New("user ID cannot be nil")
	}
	switch s.UserRole {
	case model.SuperRole, model.AdminRole, model.SupportRole, model.VendorRole, model.CustomerRole:
		return nil
	default:
		return fmt.Errorf("invalid user role: %q", s.UserRole)
	}
}

