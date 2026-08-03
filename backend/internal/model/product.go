package model

import (
	"errors"
	"fmt"
	"mime"
	"strings"
	"time"

	"github.com/google/uuid"
)

type PublicationStatus string

const (
	PublicationStatusDraft     PublicationStatus = "draft"
	PublicationStatusPublished PublicationStatus = "published"
	PublicationStatusArchived  PublicationStatus = "archived"
)

var ErrInvalidPublicationStatus = errors.New("invalid publication status")

func ParsePublicationStatus(s string) (PublicationStatus, error) {
	switch s {
	case "draft":
		return PublicationStatusDraft, nil
	case "published":
		return PublicationStatusPublished, nil
	case "archived":
		return PublicationStatusArchived, nil
	default:
		return "", ErrInvalidPublicationStatus
	}
}

func (s PublicationStatus) String() string {
	return string(s)
}

type Product struct {
	ID         uuid.UUID  `db:"id"`
	BrandID    *uuid.UUID `db:"brand_id"`
	CategoryID uuid.UUID  `db:"category_id"`

	PublicID       string            `db:"public_id"`
	Title          string            `db:"title"`
	Description    string            `db:"description"`
	Highlights     []string          `db:"highlights"`
	Status         PublicationStatus `db:"status"`
	IsConfigurable bool              `db:"is_configurable"`
	DefaultVariant *uuid.UUID        `db:"default_variant_id"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type MediaType string

const (
	NoneType     MediaType = "none"
	ImageType    MediaType = "image"
	VideoType    MediaType = "video"
	DocumentType MediaType = "document"
)

var (
	ErrUnsupportedMediaType = errors.New("unsupported media type")
)

func (m MediaType) String() string {
	return string(m)
}

func IsNoneMediaType(m MediaType) bool {
	return m == NoneType
}

func ParseMediaType(contentType string) (MediaType, error) {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", err
	}

	switch {
	case strings.HasPrefix(mediaType, "image/"):
		return ImageType, nil

	case strings.HasPrefix(mediaType, "video/"):
		return VideoType, nil

	case mediaType == "application/pdf":
		return DocumentType, nil

	default:
		return NoneType, fmt.Errorf(
			"%w: %s",
			ErrUnsupportedMediaType,
			contentType,
		)
	}
}

type ProductMedia struct {
	ID              uuid.UUID
	StorageObjectID uuid.UUID
	ProductID       uuid.UUID

	MediaType MediaType
	SortOrder int

	CreatedAt time.Time
	UpdatedAt time.Time
	Object    *Object
}
