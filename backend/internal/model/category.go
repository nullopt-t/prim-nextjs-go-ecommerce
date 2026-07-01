package model

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID        uuid.UUID
	Name      string
	Slug      string
	ParentID  *uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewCategory(
	name,
	slug string,
	parentID *uuid.UUID,
) *Category {
	now := time.Now().UTC()
	return &Category{
		ID:        uuid.New(),
		Name:      name,
		Slug:      slug,
		ParentID:  parentID,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
