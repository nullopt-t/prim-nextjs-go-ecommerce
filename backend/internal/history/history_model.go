package history

import (
	"time"

	"github.com/google/uuid"
)

// MaxItems is the fixed rolling window kept for each account. Anything older
// falls off after a view is recorded or guest history is merged.
const MaxItems = 50

// DefaultListLimit is the default number of products returned by the list
// endpoint when the caller does not specify a limit.
const DefaultListLimit = 10

// MaxSyncItems bounds the post-authentication merge payload so a malformed or
// hostile client cannot force an unbounded product lookup.
const MaxSyncItems = 100

// ViewedProductItem is the merchant-independent product projection returned to
// the storefront for a customer's recent views.
type ViewedProductItem struct {
	// ProductID is the catalog product UUID.
	ProductID uuid.UUID `json:"productId" example:"60000000-0000-0000-0000-000000000001"`
	// Title is the product display title.
	Title string `json:"title" example:"MacBook Pro 16"`
	// Slug is the public product URL slug.
	Slug string `json:"slug" example:"macbook-pro-16"`
	// BasePrice is the default variant price in minor units.
	BasePrice int64 `json:"basePrice" example:"249900"`
	// Currency is the ISO currency code for BasePrice.
	Currency string `json:"currency" example:"EGP"`
	// ThumbnailURL is the public cover image URL when the product has one.
	ThumbnailURL *string `json:"thumbnailUrl,omitempty" example:"https://cdn.example.com/catalog/products/cover.webp"`
	// ViewedAt is the most recent view timestamp (RFC3339).
	ViewedAt time.Time `json:"viewedAt" example:"2026-11-06T10:00:00Z"`

	thumbnailBucket *string
	thumbnailKey    *string
}
