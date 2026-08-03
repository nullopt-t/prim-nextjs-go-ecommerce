package model

import (
	"time"

	"github.com/google/uuid"
)

type ProductCategory struct {
	ID uuid.UUID

	ParentID *uuid.UUID
	Name     string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
