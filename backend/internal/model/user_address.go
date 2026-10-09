package model

import (
	"time"

	"github.com/google/uuid"
)

// UserAddress represents a physical shipping or billing address saved by a user.
type UserAddress struct {
	ID         uuid.UUID  `db:"id" json:"id"`
	UserID     uuid.UUID  `db:"user_id" json:"userId"`
	Label      *string    `db:"label" json:"label,omitempty"`
	FullName   string     `db:"full_name" json:"fullName"`
	Phone      *string    `db:"phone" json:"phone,omitempty"`
	Line1      string     `db:"line1" json:"line1"`
	Line2      *string    `db:"line2" json:"line2,omitempty"`
	City       string     `db:"city" json:"city"`
	State      *string    `db:"state" json:"state,omitempty"`
	PostalCode *string    `db:"postal_code" json:"postalCode,omitempty"`
	Country    string     `db:"country" json:"country"`
	IsDefault  bool       `db:"is_default" json:"isDefault"`
	CreatedAt  time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt  time.Time  `db:"updated_at" json:"updatedAt"`
	DeletedAt  *time.Time `db:"deleted_at" json:"deletedAt,omitempty"`
}
