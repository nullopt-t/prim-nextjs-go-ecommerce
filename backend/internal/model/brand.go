package model

import (
	"time"

	"github.com/google/uuid"
)

type Brand struct {
	ID        uuid.UUID
	Name      string
	LogoURL   string
	LogoLabel string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewBrand(
	name,
	logoURL,
	logoLabel string,
) *Brand {
	now := time.Now().UTC()
	return &Brand{
		ID: uuid.New(),
		Name: name,
		LogoURL: logoURL,
		LogoLabel: logoLabel,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
