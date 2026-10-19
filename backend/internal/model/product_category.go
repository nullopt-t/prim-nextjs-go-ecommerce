package model

import (
	"time"

	"github.com/google/uuid"
)

// ProductCategory represents a taxonomy category node in a hierarchical category tree.
type ProductCategory struct {
	// ID is the unique internal database UUID.
	ID uuid.UUID

	// ParentID is the optional UUID of the parent category (nil for root categories).
	ParentID *uuid.UUID
	// ParentName is the display name of the parent category, resolved on demand
	// for admin listings. It is empty for root categories or when the caller did
	// not request it.
	ParentName string
	// Name is the category display name (e.g. "Electronics", "Laptops").
	Name string
	// ProductCount is the number of non-deleted products assigned directly to
	// this category, resolved on demand for admin listings.
	ProductCount int

	// CreatedAt is the timestamp when the category was created.
	CreatedAt time.Time
	// UpdatedAt is the timestamp when category details were last modified.
	UpdatedAt time.Time
	// DeletedAt is the optional soft-delete timestamp.
	DeletedAt *time.Time
}
