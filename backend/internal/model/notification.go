package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// UserNotification represents an in-app alert displayed in the user notification center.
type UserNotification struct {
	// ID is the unique database UUID for the notification.
	ID uuid.UUID `db:"id" json:"id"`
	// UserID is the UUID of the user who will receive the notification.
	UserID uuid.UUID `db:"user_id" json:"userId"`
	// Title is the localized title of the notification.
	Title json.RawMessage `db:"title" json:"title"` // Localized map {"en": "...", "ar": "..."}
	// Message is the localized message of the notification.
	Message json.RawMessage `db:"message" json:"message"` // Localized map {"en": "...", "ar": "..."}
	// Category is the category of the notification (e.g. "auth", "order", "system").
	Category string `db:"category" json:"category"`
	// ActionURL is the URL to navigate to when the notification is clicked.
	ActionURL *string `db:"action_url" json:"actionUrl,omitempty"`
	// Metadata is the metadata of the notification.
	Metadata json.RawMessage `db:"metadata" json:"metadata,omitempty"`
	// ReadAt is the timestamp when the notification was read.
	ReadAt *time.Time `db:"read_at" json:"readAt,omitempty"`
	// CreatedAt is the timestamp when the notification was created.
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

// LocalizedTitle resolves the title for a given locale ("ar" or "en").
// If the requested locale and "en" are unavailable, it falls back to any non-empty locale entry in the map
// or an empty string, avoiding leaking raw JSON syntax into user interfaces.
func (n *UserNotification) LocalizedTitle(locale string) string {
	var m map[string]string
	if err := json.Unmarshal(n.Title, &m); err == nil {
		if val, ok := m[locale]; ok && val != "" {
			return val
		}
		if val, ok := m["en"]; ok && val != "" {
			return val
		}
		for _, val := range m {
			if val != "" {
				return val
			}
		}
		return ""
	}
	return string(n.Title)
}

// LocalizedMessage resolves the message for a given locale ("ar" or "en").
// If the requested locale and "en" are unavailable, it falls back to any non-empty locale entry in the map
// or an empty string, avoiding leaking raw JSON syntax into user interfaces.
func (n *UserNotification) LocalizedMessage(locale string) string {
	var m map[string]string
	if err := json.Unmarshal(n.Message, &m); err == nil {
		if val, ok := m[locale]; ok && val != "" {
			return val
		}
		if val, ok := m["en"]; ok && val != "" {
			return val
		}
		for _, val := range m {
			if val != "" {
				return val
			}
		}
		return ""
	}
	return string(n.Message)
}
