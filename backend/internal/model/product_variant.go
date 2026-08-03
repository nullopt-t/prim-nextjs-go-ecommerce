package model

import (
	"time"

	"github.com/google/uuid"
)

type ProductVariant struct {
	ID              uuid.UUID
	ProductID       uuid.UUID
	IsDefault       bool
	Title           string
	Price           *int64
	CrossedOutPrice *int64
	Currency        *string
	Attributes      map[string]any
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

type VariantMedia struct {
	ID              uuid.UUID
	VariantID       uuid.UUID
	StorageObjectID uuid.UUID
	MediaType       string
	SortOrder       int

	Object *Object
}
