package model

import (
	"time"

	"github.com/google/uuid"
)

type ProductVariantStatus string

const (
	ProductVariantActive       ProductVariantStatus = "active"
	ProductVariantInactive     ProductVariantStatus = "inactive"
	ProductVariantDiscontinued ProductVariantStatus = "discontinued"
)

type ProductVariant struct {
	ID        uuid.UUID
	ProductID uuid.UUID

	SKU    string
	Status ProductVariantStatus
	Price  int64 // minor units

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

type Product struct {
	ID      uuid.UUID
	BrandID uuid.UUID

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

type Inventory struct {
	ID        uuid.UUID
	VariantID uuid.UUID

	Quantity         int64
	ReservedQuantity int64

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
