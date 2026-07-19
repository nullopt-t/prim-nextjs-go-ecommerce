package model

import (
	"errors"
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

func ValidateMediaType(s string) (MediaType, error) {
	switch s {
	case ImageType.String():
		return ImageType, nil
	case VideoType.String():
		return VideoType, nil
	case DocumentType.String():
		return DocumentType, nil
	default:
		return "", errors.New("unsupported media type")
	}
}

type ProductMedia struct {
	ID        uuid.UUID
	VariantID uuid.UUID
	ObjectID  uuid.UUID

	Type      MediaType
	SortOrder int

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
