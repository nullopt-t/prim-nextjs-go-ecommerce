package model

import (
	"time"

	"github.com/google/uuid"
)

// ProductCategory represents a taxonomy category node in a hierarchical category tree.
type ProductCategory struct {
	// ID is the unique internal database UUID.
	ID uuid.UUID
	// PublicID is the customer-facing public identifier string.
	PublicID string

	// ParentID is the optional UUID of the parent category (nil for root categories).
	ParentID *uuid.UUID
	// Name is the category display name (e.g. "Electronics", "Laptops").
	Name string

	// CreatedAt is the timestamp when the category was created.
	CreatedAt time.Time
	// UpdatedAt is the timestamp when category details were last modified.
	UpdatedAt time.Time
	// DeletedAt is the optional soft-delete timestamp.
	DeletedAt *time.Time
}
