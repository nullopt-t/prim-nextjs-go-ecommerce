package model

import (
	"errors"
	"fmt"
	"mime"
	"strings"
	"time"

	"github.com/google/uuid"
)

// PublicationStatus represents the visibility and publication state of a product in the catalog.
type PublicationStatus string

const (
	// PublicationStatusDraft indicates the product is in progress and hidden from storefront browsing.
	PublicationStatusDraft PublicationStatus = "draft"
	// PublicationStatusPublished indicates the product is active and visible to customer storefronts.
	PublicationStatusPublished PublicationStatus = "published"
	// PublicationStatusArchived indicates the product is discontinued or hidden from public search.
	PublicationStatusArchived PublicationStatus = "archived"
)

// ErrInvalidPublicationStatus is returned when an invalid publication status string is provided.
var ErrInvalidPublicationStatus = errors.New("invalid publication status")

// ProductType distinguishes between simple single-SKU products and configurable variable products.
type ProductType string

const (
	// ProductTypeSimple indicates a product with a single inherent SKU without options.
	ProductTypeSimple ProductType = "simple"
	// ProductTypeVariable indicates a product with multiple configurable variants (sizes, colors, etc.).
	ProductTypeVariable ProductType = "variable"
)

// ErrInvalidProductType is returned when an unrecognized product type string is encountered.
var ErrInvalidProductType = errors.New("invalid product type")

// ParseProductType converts a raw string into a typed ProductType.
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

// String returns the string representation of ProductType.
func (p ProductType) String() string {
	return string(p)
}

// ParsePublicationStatus converts a raw string into a typed PublicationStatus.
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

// String returns the string representation of PublicationStatus.
func (s PublicationStatus) String() string {
	return string(s)
}

// Product represents a master product catalog record.
type Product struct {
	// ID is the unique database UUID for the product.
	ID uuid.UUID `db:"id"`
	// BrandID is the optional UUID of the product's manufacturer/brand.
	BrandID *uuid.UUID `db:"brand_id"`
	// CategoryID is the required taxonomy category UUID.
	CategoryID uuid.UUID `db:"category_id"`

	// Slug is the URL-friendly unique SEO identifier.
	Slug string `db:"slug"`
	// Title is the primary product display name.
	Title string `db:"title"`
	// Description is the full product narrative copy.
	Description *string `db:"description"`
	// Highlights contains key bullet points displayed on product detail pages.
	Highlights []string `db:"highlights"`
	// Status is the draft/published/archived lifecycle state.
	Status PublicationStatus `db:"status"`
	// ProductType indicates simple or variable catalog classification.
	ProductType ProductType `db:"product_type"`
	// ThumbnailObjectID is the optional storage object UUID for the primary cover image.
	ThumbnailObjectID *uuid.UUID `db:"thumbnail_object_id"`

	// CreatedAt is the timestamp when the product was added to the catalog.
	CreatedAt time.Time `db:"created_at"`
	// UpdatedAt is the timestamp when product metadata was last modified.
	UpdatedAt time.Time `db:"updated_at"`
	// DeletedAt is the optional soft-delete timestamp.
	DeletedAt *time.Time `db:"deleted_at"`

	// Brand is the hydrated ProductBrand entity (not stored directly in the products table).
	Brand *ProductBrand `db:"-"`
	// Category is the hydrated ProductCategory entity (not stored directly in the products table).
	Category *ProductCategory `db:"-"`
	// Thumbnail is the hydrated Object entity for the cover image (not stored directly in the products table).
	Thumbnail *Object `db:"-"`
}

// MediaType categorizes binary file assets attached to products or variants.
type MediaType string

const (
	// NoneType indicates an unrecognized or empty media type.
	NoneType MediaType = "none"
	// ImageType indicates image raster/vector assets (e.g. JPEG, PNG, WebP).
	ImageType MediaType = "image"
	// VideoType indicates video stream/clip assets (e.g. MP4, WebM).
	VideoType MediaType = "video"
	// DocumentType indicates document files (e.g. PDF manuals, spec sheets).
	DocumentType MediaType = "document"
)

// ErrUnsupportedMediaType is returned when an uploaded file MIME type is not allowed.
var ErrUnsupportedMediaType = errors.New("unsupported media type")

// String returns the string representation of MediaType.
func (m MediaType) String() string {
	return string(m)
}

// IsNoneMediaType returns true if the media type is NoneType.
func IsNoneMediaType(m MediaType) bool {
	return m == NoneType
}

// ParseMediaType inspects an HTTP Content-Type string and classifies it into a typed MediaType.
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

// ProductMedia associates a stored binary Object with a Product.
type ProductMedia struct {
	// ID is the unique database UUID for the association.
	ID uuid.UUID
	// ObjectID is the underlying binary storage Object UUID.
	ObjectID uuid.UUID
	// ProductID is the target Product UUID.
	ProductID uuid.UUID

	// MediaType classifies the file (image, video, document).
	MediaType MediaType
	// SortOrder defines display order sequence on storefronts.
	SortOrder int

	// CreatedAt is the timestamp when media was linked to the product.
	CreatedAt time.Time
	// UpdatedAt is the timestamp when media sort order or type was modified.
	UpdatedAt time.Time
	// Object is the hydrated binary storage Object entity.
	Object *Object
}
