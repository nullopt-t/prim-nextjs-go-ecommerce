package model

import (
	"time"

	"github.com/google/uuid"
)

type Cart struct {
	ID        uuid.UUID
	UserID    *uuid.UUID
	SessionID *string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	Items     []CartItem
}

type CartItem struct {
	ID              uuid.UUID
	CartID          uuid.UUID
	VariantID       uuid.UUID
	Quantity        int
	PriceAtPurchase int64
	Currency        string
	CartedAt        time.Time
	DeletedAt       *time.Time
	Variant         *ProductVariant
}
