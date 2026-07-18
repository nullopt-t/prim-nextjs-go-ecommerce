package model

import (
	"time"

	"github.com/google/uuid"
)

type ProductVariant struct {
	ID        uuid.UUID
	ProductID uuid.UUID

	SKU      *string
	Price    int64
	Currency string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
