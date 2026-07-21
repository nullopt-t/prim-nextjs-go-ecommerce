package model

import (
	"time"

	"github.com/google/uuid"
)

type ProductBrandStatus string

const (
	ProductBrandStatusActive   ProductBrandStatus = "active"
	ProductBrandStatusArchived ProductBrandStatus = "archived"
)

func (s ProductBrandStatus) String() string {
	return string(s)
}

type ProductBrand struct {
	ID uuid.UUID

	Name    string
	Slug    string
	LogoURL string
	LogoAlt string
	Status  ProductBrandStatus

	CreatedAt time.Time
	UpdatedAt time.Time

	DeletedAt  *time.Time
	ArchivedAt *time.Time
}
