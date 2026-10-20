package model

import (
	"time"

	"github.com/google/uuid"
)

// AuditLog is a single append-only record of a privileged mutation performed
// through the back-office API. It forms the accountability trail that a
// superadmin reviews: who did what, to which entity, and when.
type AuditLog struct {
	// ID is the database-owned identity for the entry.
	ID int64

	// ActorID is the staff account that performed the action (nil for system actions).
	ActorID *uuid.UUID
	// ActorRole is the role the actor held at the time of the action.
	ActorRole string

	// Action is the normalized event name (e.g. "user.role.update").
	Action string
	// EntityType is the affected resource kind (e.g. "user", "payment").
	EntityType string
	// EntityID is the affected resource identifier when known.
	EntityID string

	// Method and Path capture the originating HTTP request.
	Method string
	Path   string
	// StatusCode is the HTTP status returned to the caller.
	StatusCode int

	// IPAddress and UserAgent record the requester's origin.
	IPAddress string
	UserAgent string

	// Before and After are JSON snapshots of the affected state. Before is
	// populated only by explicit recordings; After holds the redacted request
	// payload captured by the audit middleware.
	Before []byte
	After  []byte
	// Metadata holds auxiliary context (e.g. route parameters).
	Metadata []byte

	// CreatedAt is when the action was recorded.
	CreatedAt time.Time
}
