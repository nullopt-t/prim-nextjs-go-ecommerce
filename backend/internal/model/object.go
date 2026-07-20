package model

import (
	"time"

	"github.com/google/uuid"
)

type ObjectStatus string

const (
	ObjectStatusUploading ObjectStatus = "uploading"
	ObjectStatusUploaded  ObjectStatus = "uploaded"
	ObjectStatusDeleting  ObjectStatus = "deleting"
	ObjectStatusDeleted   ObjectStatus = "deleted"
)

func (o ObjectStatus) String() string {
	return string(o)
}

type Object struct {
	ID uuid.UUID

	Size        int64
	Status      ObjectStatus
	ContentType string

	Key       string
	Bucket    string
	PublicURL string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
