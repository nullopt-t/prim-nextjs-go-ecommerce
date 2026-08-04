package model

import (
	"time"

	"github.com/google/uuid"
)

type ProductBrand struct {
	ID                  uuid.UUID
	PublicID            string
	Name                string
	Link                *string
	LogoStorageObjectID *uuid.UUID
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           *time.Time
}
