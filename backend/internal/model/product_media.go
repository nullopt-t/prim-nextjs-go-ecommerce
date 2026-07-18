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

func (m MediaType) String() string {
	return string(m)
}

type Media struct {
	ID uuid.UUID

	Alt      string
	Type     MediaType
	MimeType string
	FileSize int64
	Checksum string

	Width  int
	Height int

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
