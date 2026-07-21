package model

import (
	"time"

	"github.com/google/uuid"
)

type ProductTag struct {
	ID uuid.UUID

	Name              string
	PublicationStatus PublicationStatus

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
