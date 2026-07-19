package model

import (
	"time"

	"github.com/google/uuid"
)

type ProductStatus string

const (
	ProductStatusPublished ProductStatus = "published"
	ProductStatusDraft     ProductStatus = "draft"
	ProductStatusArchived  ProductStatus = "archived"
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
