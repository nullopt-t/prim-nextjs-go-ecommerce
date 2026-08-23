package model

import (
	"time"

	"github.com/google/uuid"
)

// ObjectStatus represents the lifecycle state of a stored file/binary object.
type ObjectStatus string

const (
	// ObjectStatusUploading indicates an upload has been initialized/presigned but not finalized.
	ObjectStatusUploading ObjectStatus = "uploading"
	// ObjectStatusUploaded indicates the file was successfully written and verified in object storage.
	ObjectStatusUploaded ObjectStatus = "uploaded"
	// ObjectStatusDeleting indicates the file has been scheduled for asynchronous cleanup.
	ObjectStatusDeleting ObjectStatus = "deleting"
	// ObjectStatusDeleted indicates the object has been permanently deleted from storage.
	ObjectStatusDeleted ObjectStatus = "deleted"
)

// String returns the string representation of ObjectStatus.
func (o ObjectStatus) String() string {
	return string(o)
}

// Object represents a binary media asset or file stored in MinIO/S3.
type Object struct {
	// ID is the unique database UUID for the storage object.
	ID uuid.UUID

	// FileSize is the file length in bytes.
	FileSize int64
	// Status is the current storage lifecycle state.
	Status ObjectStatus
	// ContentType is the MIME type (e.g. "image/webp", "application/pdf").
	ContentType string

	// Key is the object path key inside the bucket (e.g. "products/uuid/image.webp").
	Key string
	// Bucket is the target S3/MinIO bucket name.
	Bucket string
	// PublicURL is the publicly accessible CDN or presigned HTTP endpoint for the file.
	PublicURL string

	// CreatedAt is the timestamp when the object record was created.
	CreatedAt time.Time
	// UpdatedAt is the timestamp when metadata was last modified.
	UpdatedAt time.Time
	// DeletedAt is the optional soft-delete timestamp.
	DeletedAt *time.Time
}
