package model

import (
	"time"

	"github.com/google/uuid"
)

// WishlistItem represents an item saved in a customer's personal wishlist.
type WishlistItem struct {
	// ID is the unique database UUID for the wishlist item.
	ID uuid.UUID `db:"id"`
	// UserID is the UUID of the customer owning this wishlist item.
	UserID uuid.UUID `db:"user_id"`
	// ProductID is the saved Product UUID.
	ProductID uuid.UUID `db:"product_id"`
	// VariantID is the optional specific variant UUID saved in the wishlist.
	VariantID *uuid.UUID `db:"variant_id"`
	// CreatedAt is the timestamp when the item was added to the wishlist.
	CreatedAt time.Time `db:"created_at"`

	// Populated Product, Variant and Media fields
	ProductTitle       string
	ProductSlug        string
	ProductDescription *string
	ProductType        ProductType
	ProductStatus      PublicationStatus
	VariantTitle       *string
	VariantSKU         *string
	VariantAttributes  []byte
	BrandName          *string
	CategoryName       *string
	ThumbnailBucket    *string
	ThumbnailKey       *string
	ThumbnailURL       *string
	Price              *int64
	CrossedOutPrice    *int64
	Currency           *string
	InStock            bool
}
