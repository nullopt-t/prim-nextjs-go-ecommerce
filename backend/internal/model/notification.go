package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// UserNotification represents an in-app alert displayed in the user notification center.
type UserNotification struct {
	ID        uuid.UUID       `db:"id" json:"id"`
	UserID    uuid.UUID       `db:"user_id" json:"userId"`
	Title     json.RawMessage `db:"title" json:"title"`         // Localized map {"en": "...", "ar": "..."}
	Message   json.RawMessage `db:"message" json:"message"`     // Localized map {"en": "...", "ar": "..."}
	Category  string          `db:"category" json:"category"`   // e.g. "auth", "order", "system"
	ActionURL *string         `db:"action_url" json:"actionUrl,omitempty"`
	Metadata  json.RawMessage `db:"metadata" json:"metadata,omitempty"`
	ReadAt    *time.Time      `db:"read_at" json:"readAt,omitempty"`
	CreatedAt time.Time       `db:"created_at" json:"createdAt"`
}

// LocalizedTitle resolves the title for a given locale ("ar" or "en").
func (n *UserNotification) LocalizedTitle(locale string) string {
	var m map[string]string
	if err := json.Unmarshal(n.Title, &m); err == nil {
		if val, ok := m[locale]; ok && val != "" {
			return val
		}
		if val, ok := m["en"]; ok && val != "" {
			return val
		}
	}
	return string(n.Title)
}

// LocalizedMessage resolves the message for a given locale ("ar" or "en").
func (n *UserNotification) LocalizedMessage(locale string) string {
	var m map[string]string
	if err := json.Unmarshal(n.Message, &m); err == nil {
		if val, ok := m[locale]; ok && val != "" {
			return val
		}
		if val, ok := m["en"]; ok && val != "" {
			return val
		}
	}
	return string(n.Message)
}
