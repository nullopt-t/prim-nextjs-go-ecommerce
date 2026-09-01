package model

import (
	"time"

	"github.com/google/uuid"
)

// ProductVariant represents a purchasable SKU option under a master Product (e.g. Size, Color, Capacity).
type ProductVariant struct {
	// ID is the unique database UUID for this variant SKU.
	ID uuid.UUID `db:"id"`
	// SKU is the unique Stock Keeping Unit identifier string.
	SKU string `db:"sku"`
	// ProductID is the master Product UUID.
	ProductID uuid.UUID `db:"product_id"`
	// IsDefault indicates whether this variant is the initial selection on product detail pages.
	IsDefault bool `db:"is_default"`
	// Title is the variant option label (e.g. "Space Black, 18GB RAM, 512GB SSD").
	Title string `db:"title"`
	// Price is the active sales unit price in cents (e.g. 249900 = $2499.00).
	Price *int64 `db:"price"`
	// CrossedOutPrice is the original MSRP / crossed-out price in cents.
	CrossedOutPrice *int64 `db:"crossed_out_price"`
	// Currency is the ISO currency code (e.g. "USD").
	Currency *string `db:"currency"`
	// Attributes contains key-value option dimensions (e.g. {"color":"Black","ram":"18GB"}).
	Attributes map[string]any `db:"attributes"`
	// ThumbnailObjectID is the optional storage object UUID for the variant's specific thumbnail image.
	ThumbnailObjectID *uuid.UUID `db:"thumbnail_object_id"`
	// CreatedAt is the timestamp when the variant was created.
	CreatedAt time.Time `db:"created_at"`
	// UpdatedAt is the timestamp when the variant was last modified.
	UpdatedAt time.Time `db:"updated_at"`
	// DeletedAt is the optional soft-delete timestamp.
	DeletedAt *time.Time `db:"deleted_at"`

	// Thumbnail is the hydrated Object entity for the thumbnail (not stored directly in product_variants table).
	Thumbnail *Object `db:"-"`
	// Media contains all attached gallery media records for this variant.
	Media []*VariantMedia `db:"-"`
}

// VariantMedia links a stored binary Object to a specific ProductVariant.
type VariantMedia struct {
	// ID is the internal database UUID for the association.
	ID uuid.UUID `db:"id"`
	// VariantID is the target ProductVariant UUID.
	VariantID uuid.UUID `db:"variant_id"`
	// ObjectID is the underlying binary storage Object UUID.
	ObjectID uuid.UUID `db:"object_id"`
	// MediaType classifies the file (image or video).
	MediaType string `db:"media_type"`
	// SortOrder defines display order sequence on storefronts.
	SortOrder int `db:"sort_order"`

	// Object is the hydrated binary storage Object entity.
	Object *Object `db:"-"`
}
