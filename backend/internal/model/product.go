package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type PublicationStatus string

const (
	PublicationStatusDraft     PublicationStatus = "draft"
	PublicationStatusPublished PublicationStatus = "published"
)

var ErrInvalidPublicationStatus = errors.New("invalid publication status")

func ParsePublicationStatus(s string) (PublicationStatus, error) {
	switch s {
	case "draft":
		return PublicationStatusDraft, nil
	case "published":
		return PublicationStatusPublished, nil
	default:
		return "", ErrInvalidPublicationStatus
	}
}

func (s PublicationStatus) String() string {
	return string(s)
}

type ProductStatus string

const (
	ProductStatusPublished ProductStatus = "published"
	ProductStatusDraft     ProductStatus = "draft"
	ProductStatusArchived  ProductStatus = "archived"
)

func (p ProductStatus) String() string {
	return string(p)
}

type Product struct {
	ID      uuid.UUID
	BrandID *uuid.UUID

	Slug             string
	Title            string
	ShortDescription string
	Description      string
	Status           ProductStatus

	CreatedBy uuid.UUID
	UpdatedBy uuid.UUID

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
