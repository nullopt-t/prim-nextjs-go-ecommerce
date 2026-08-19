package model

import (
	"time"

	"github.com/google/uuid"
)

type ProductVariant struct {
	ID                uuid.UUID      `db:"id"`
	SKU               string         `db:"sku"`
	ProductID         uuid.UUID      `db:"product_id"`
	IsDefault         bool           `db:"is_default"`
	Title             string         `db:"title"`
	Price             *int64         `db:"price"`
	CrossedOutPrice   *int64         `db:"crossed_out_price"`
	Currency          *string        `db:"currency"`
	Attributes        map[string]any `db:"attributes"`
	ThumbnailObjectID *uuid.UUID     `db:"thumbnail_object_id"`
	CreatedAt         time.Time      `db:"created_at"`
	UpdatedAt         time.Time      `db:"updated_at"`
	DeletedAt         *time.Time     `db:"deleted_at"`

	Thumbnail *Object         `db:"-"`
	Media     []*VariantMedia `db:"-"`
}

type VariantMedia struct {
	ID        uuid.UUID `db:"id"`
	PublicID  uuid.UUID `db:"public_id"`
	VariantID uuid.UUID `db:"variant_id"`
	ObjectID  uuid.UUID `db:"object_id"`
	MediaType string    `db:"media_type"`
	SortOrder int       `db:"sort_order"`

	Object *Object `db:"-"`
}
