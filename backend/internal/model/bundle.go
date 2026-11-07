package model

import "github.com/google/uuid"

// BundleComponent is one catalog item inside a product bundle, flattened with
// the display fields the storefront needs.
//
// StockVariantID is the variant whose inventory gates the kit: the variant the
// merchant pinned when the kit was built, otherwise the product's default (or
// first) child. It is nil only when the component product has no variant left,
// in which case the kit cannot be sold.
type BundleComponent struct {
	// ProductID is the component product UUID.
	ProductID uuid.UUID
	// Slug is the component product's storefront slug.
	Slug string
	// Title is the component product's display title.
	Title string
	// VariantID is the variant the merchant pinned for this component, if any.
	VariantID *uuid.UUID
	// VariantTitle is the pinned variant's title, when one was chosen.
	VariantTitle *string
	// StockVariantID resolves the component to the variant that holds its stock.
	StockVariantID *uuid.UUID
	// Quantity is how many units of the component one kit contains.
	Quantity int
	// ThumbnailBucket and ThumbnailKey locate the component's cover image.
	ThumbnailBucket *string
	ThumbnailKey    *string
}

// BundleComposition is the kit behind a bundle product. A bundle product keeps
// no inventory of its own: its availability is how many complete kits its
// components can build, and buying one reserves the components rather than the
// kit's own variant.
type BundleComposition struct {
	// ID is the product_bundles UUID.
	ID uuid.UUID
	// Price is the bundle's selling price in minor currency units.
	Price int64
	// Currency is the bundle's ISO 4217 currency code.
	Currency string
	// Items is the kit's composition.
	Items []BundleComponent
}

// AvailableQuantity reports how many complete kits the supplied component stock
// can build. A component with no stock record counts as zero, so a kit is never
// advertised as available when any part of it is missing.
func (b *BundleComposition) AvailableQuantity(stock map[uuid.UUID]*InventoryStock) int {
	if b == nil || len(b.Items) == 0 {
		return 0
	}

	available := -1
	for _, item := range b.Items {
		if item.StockVariantID == nil || item.Quantity <= 0 {
			return 0
		}
		onHand := 0
		if componentStock := stock[*item.StockVariantID]; componentStock != nil {
			onHand = componentStock.AvailableQuantity
		}
		if kits := onHand / item.Quantity; available < 0 || kits < available {
			available = kits
		}
	}
	if available < 0 {
		return 0
	}
	return available
}
