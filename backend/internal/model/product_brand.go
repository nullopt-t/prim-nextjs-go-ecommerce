package model

import (
	"time"

	"github.com/google/uuid"
)

type ProductBrandStatus string

const (
	ProductBrandStatusDraft     ProductBrandStatus = "draft"
	ProductBrandStatusPublished ProductBrandStatus = "published"
)

func (s ProductBrandStatus) String() string {
	return string(s)
}

type ProductBrand struct {
	ID uuid.UUID

	Name    string
	PublicationStatus  ProductBrandStatus


	LogoURL *string
	LogoAlt *string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt  *time.Time
}
