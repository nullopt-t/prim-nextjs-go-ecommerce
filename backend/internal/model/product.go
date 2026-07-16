package model

import (
	"time"

	"github.com/google/uuid"
)

type ProductVariant struct {
	ID        uuid.UUID
	ProductID uuid.UUID

	SKU      string
	Price    int64 // minor units
	Currency string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type ProductStatus string

const (
	ProductStatusActive   ProductStatus = "active"
	ProductStatusDraft    ProductStatus = "draft"
	ProductStatusArchived ProductStatus = "archived"
)

func (p ProductStatus) String() string {
	return string(p)
}

type Product struct {
	ID      uuid.UUID
	BrandID *uuid.UUID

	Slug             string
	Title            string
	ShortDescription string
	Description      string
	Status           ProductStatus

	CreatedBy uuid.UUID
	UpdatedBy uuid.UUID

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
