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

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
