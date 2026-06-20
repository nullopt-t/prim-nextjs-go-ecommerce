package model

import (
	"time"

	"github.com/google/uuid"
)

type Role struct {
	ID        int
	Code      string
	CreatedAt time.Time
}

type UserRole struct {
	UserID uuid.UUID
	RoleID int
}
