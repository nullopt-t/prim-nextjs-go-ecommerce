package model

import (
	"time"

	"github.com/google/uuid"
)

type Object struct {
	ID uuid.UUID

	Size        int64
	Status      string
	ContentType string

	Key    string
	Bucket string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
