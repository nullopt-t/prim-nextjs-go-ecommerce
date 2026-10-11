package token

import (
	"github.com/nullopt-t/prim-backend/internal/model"

	"github.com/google/uuid"
)

type UserSubject struct {
	UserID          uuid.UUID      `json:"user_id"`
	UserRole        model.UserRole `json:"user_role"`
	IsEmailVerified bool           `json:"is_email_verified"`
}
