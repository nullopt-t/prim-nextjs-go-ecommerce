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

type ProductType string

const (
	ProductTypeSimple   ProductType = "simple"
	ProductTypeVariable ProductType = "variable"
)

var ErrInvalidProductType = errors.New("invalid product type")

func ParseProductType(s string) (ProductType, error) {
	switch s {
	case "simple":
		return ProductTypeSimple, nil
	case "variable":
		return ProductTypeVariable, nil
	default:
		return "", ErrInvalidProductType
	}
}

func (p ProductType) String() string {
	return string(p)
}

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

	Slug              string            `db:"slug"`
	Title             string            `db:"title"`
	Description       *string           `db:"description"`
	Highlights        []string          `db:"highlights"`
	Status            PublicationStatus `db:"status"`
	ProductType       ProductType       `db:"product_type"`
	ThumbnailObjectID *uuid.UUID        `db:"thumbnail_object_id"`

	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`

	Brand     *ProductBrand    `db:"-"`
	Category  *ProductCategory `db:"-"`
	Thumbnail *Object          `db:"-"`
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
	ID        uuid.UUID
	ObjectID  uuid.UUID
	ProductID uuid.UUID

	MediaType MediaType
	SortOrder int

	CreatedAt time.Time
	UpdatedAt time.Time
	Object    *Object
}
