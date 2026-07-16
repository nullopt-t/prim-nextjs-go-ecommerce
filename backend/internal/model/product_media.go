package model

import (
	"time"

	"github.com/google/uuid"
)

type MediaType string

const (
	ImageType    MediaType = "image"
	VideoType    MediaType = "video"
	DocumentType MediaType = "document"
)

type Media struct {
	ID uuid.UUID

	Alt      string
	Type     MediaType
	MimeType string
	FileSize int64
	Checksum string

	Width  int
	Height int

	CreatedBy uuid.UUID
	UpdatedBy uuid.UUID

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type ProductMedia struct {
	ProductID uuid.UUID
	MediaID   uuid.UUID

	SortOrder int
	IsPrimary bool
}

type VariantMedia struct {
	ProductID uuid.UUID
	MediaID   uuid.UUID

	SortOrder int
	IsPrimary bool
}
