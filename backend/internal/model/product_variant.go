package model

import (
	"time"

	"github.com/google/uuid"
)

type ProductVariant struct {
	ID              uuid.UUID      `db:"id"`
	PublicID        uuid.UUID      `db:"public_id"`
	ProductID       uuid.UUID      `db:"product_id"`
	IsDefault       bool           `db:"is_default"`
	Title           string         `db:"title"`
	Price           *int64         `db:"price"`
	CrossedOutPrice *int64         `db:"crossed_out_price"`
	Currency        *string        `db:"currency"`
	Attributes      map[string]any `db:"attributes"`
	CreatedAt       time.Time      `db:"created_at"`
	UpdatedAt       time.Time      `db:"updated_at"`
	DeletedAt       *time.Time     `db:"deleted_at"`
}

type VariantMedia struct {
	ID              uuid.UUID
	VariantID       uuid.UUID
	StorageObjectID uuid.UUID
	MediaType       string
	SortOrder       int

	Object *Object
}
