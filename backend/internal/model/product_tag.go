package model

import (
	"time"

	"github.com/google/uuid"
)

type ProductTag struct {
	ID uuid.UUID

	Name string

	CreatedBy uuid.UUID
	UpdatedBy uuid.UUID

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
