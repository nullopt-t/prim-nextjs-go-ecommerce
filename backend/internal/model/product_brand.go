package model

import (
	"time"

	"github.com/google/uuid"
)

type ProductBrand struct {
	ID uuid.UUID

	Name    string
	Slug    string
	LogoURL string
	LogoAlt string

	CreatedBy uuid.UUID
	UpdatedBy uuid.UUID

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
