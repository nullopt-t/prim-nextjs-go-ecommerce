package model

import (
	"time"

	"github.com/google/uuid"
)

type Inventory struct {
	ID        uuid.UUID
	VariantID uuid.UUID

	Quantity         int64
	ReservedQuantity int64

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
