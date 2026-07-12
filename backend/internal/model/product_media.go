package model

import (
	"time"

	"github.com/google/uuid"
)

type ProductMedia struct {
	ID        uuid.UUID
	VariantID uuid.UUID

	URL      string
	Alt      string
	MimeType string

	// the first media always the main
	SortOrder int

	Width    int
	Height   int
	FileSize int64

	CreatedBy uuid.UUID
	UpdatedBy uuid.UUID

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
