package model

import (
	"time"

	"github.com/google/uuid"
)

type Cart struct {
	ID        uuid.UUID  `json:"id"`
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	SessionID *string    `json:"session_id,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	Items     []CartItem `json:"items,omitempty"`
}

type CartItem struct {
	ID              uuid.UUID       `json:"id"`
	CartID          uuid.UUID       `json:"cart_id"`
	VariantID       uuid.UUID       `json:"variant_id"`
	Quantity        int             `json:"quantity"`
	PriceAtPurchase int64           `json:"price_at_purchase"`
	Currency        string          `json:"currency"`
	CartedAt        time.Time       `json:"carted_at"`
	DeletedAt       *time.Time      `json:"deleted_at,omitempty"`
	Variant         *ProductVariant `json:"variant,omitempty"`
}
