package model

import (
	"time"

	"github.com/google/uuid"
)

type ProductVariant struct {
	ID              uuid.UUID      `db:"id"               json:"id"`
	PublicID        string         `db:"public_id"        json:"public_id"`
	ProductID       uuid.UUID      `db:"product_id"       json:"product_id"`
	IsDefault       bool           `db:"is_default"       json:"is_default"`
	Title           string         `db:"title"            json:"title"`
	Price           *int64         `db:"price"            json:"price"`
	CrossedOutPrice *int64         `db:"crossed_out_price" json:"crossed_out_price,omitempty"`
	Currency        *string        `db:"currency"         json:"currency,omitempty"`
	Attributes      map[string]any `db:"attributes"       json:"attributes,omitempty"`
	CreatedAt       time.Time      `db:"created_at"       json:"created_at"`
	UpdatedAt       time.Time      `db:"updated_at"       json:"updated_at"`
	DeletedAt       *time.Time     `db:"deleted_at"       json:"deleted_at,omitempty"`
}

type VariantMedia struct {
	ID              uuid.UUID `json:"id"`
	VariantID       uuid.UUID `json:"variant_id"`
	StorageObjectID uuid.UUID `json:"storage_object_id"`
	MediaType       string    `json:"media_type"`
	SortOrder       int       `json:"sort_order"`

	Object *Object `json:"object,omitempty"`
}
