package model

import (
	"time"

	"github.com/google/uuid"
)

// Cart represents a customer shopping cart container.
// It can belong to an authenticated customer (UserID) or a guest visitor (SessionID).
type Cart struct {
	// ID is the unique database UUID for the cart.
	ID uuid.UUID
	// UserID is the optional UUID of the authenticated user who owns this cart.
	UserID *uuid.UUID
	// SessionID is the guest session string for unauthenticated visitors.
	SessionID *string
	// CreatedAt is the timestamp when the cart was initially created.
	CreatedAt time.Time
	// UpdatedAt is the timestamp when the cart or its items were last modified.
	UpdatedAt time.Time
	// DeletedAt is the optional soft-delete timestamp when the cart was emptied or discarded.
	DeletedAt *time.Time
	// Items contains all line items currently present in the cart.
	Items []CartItem
}

// CartItem represents a specific product variant line item inside a Cart.
type CartItem struct {
	// ID is the unique UUID for this cart line item record.
	ID uuid.UUID
	// CartID is the parent Cart UUID.
	CartID uuid.UUID
	// VariantID is the specific ProductVariant UUID added to the cart.
	VariantID uuid.UUID
	// Quantity is the number of units of this variant in the cart.
	Quantity int
	// PriceAtPurchase is the unit price in cents at the moment this item was added to the cart.
	PriceAtPurchase int64
	// CurrentPrice is the live unit price in cents from the catalog (nil if unchanged or unavailable).
	CurrentPrice *int64
	// Currency is the ISO currency code (e.g. "USD").
	Currency string
	// CartedAt is the timestamp when this variant was added to the cart.
	CartedAt time.Time
	// DeletedAt is the optional timestamp when this item was removed from the cart.
	DeletedAt *time.Time
	// Variant is the eagerly loaded or hydrated ProductVariant entity.
	Variant *ProductVariant
	// Product is the parent Product entity associated with the variant.
	Product *Product
	// ThumbnailURL is the resolved CDN URL for the item's display image.
	ThumbnailURL string
	// InStock is a live flag indicating whether warehouse inventory satisfies the requested quantity.
	InStock bool
}
