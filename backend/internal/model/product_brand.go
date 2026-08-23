package model

import (
	"time"

	"github.com/google/uuid"
)

// ProductBrand represents a manufacturer or commercial brand associated with catalog products.
type ProductBrand struct {
	// ID is the internal database UUID.
	ID uuid.UUID
	// PublicID is the customer-facing public identifier string.
	PublicID string
	// Name is the unique brand name (e.g. "Apple", "Nike").
	Name string
	// Link is the optional official website URL.
	Link *string
	// LogoObjectID is the optional storage object UUID for the brand logo.
	LogoObjectID *uuid.UUID
	// LogoURL is the resolved public CDN URL for the brand logo image.
	LogoURL *string
	// CreatedAt is the timestamp when the brand record was created.
	CreatedAt time.Time
	// UpdatedAt is the timestamp when brand details were last modified.
	UpdatedAt time.Time
	// DeletedAt is the optional soft-delete timestamp.
	DeletedAt *time.Time
}
