package model

import (
	"time"

	"github.com/google/uuid"
)

type ProductCategory struct {
	ID uuid.UUID

	Name string
	Slug string

	ParentID          *uuid.UUID
	PublicationStatus PublicationStatus

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
