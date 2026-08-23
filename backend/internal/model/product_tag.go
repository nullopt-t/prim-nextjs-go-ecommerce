package model

import (
	"time"

	"github.com/google/uuid"
)

// ProductTag represents a searchable promotional label or tag attached to products (e.g. "Featured", "Sale").
type ProductTag struct {
	// ID is the unique database UUID for the tag.
	ID uuid.UUID

	// Name is the unique tag label text.
	Name string

	// CreatedAt is the timestamp when the tag was created.
	CreatedAt time.Time
	// UpdatedAt is the timestamp when the tag was last modified.
	UpdatedAt time.Time
	// DeletedAt is the optional soft-delete timestamp.
	DeletedAt *time.Time
}
