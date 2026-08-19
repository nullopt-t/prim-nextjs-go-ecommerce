package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type InventoryReason string

const (
	InventoryReasonRestock            InventoryReason = "restock"
	InventoryReasonSale               InventoryReason = "sale"
	InventoryReasonReturn             InventoryReason = "return"
	InventoryReasonAdjustment         InventoryReason = "adjustment"
	InventoryReasonReservationRelease InventoryReason = "reservation_release"
)

var ErrInvalidInventoryReason = errors.New("invalid inventory reason")

func ParseInventoryReason(s string) (InventoryReason, error) {
	switch s {
	case "restock":
		return InventoryReasonRestock, nil
	case "sale":
		return InventoryReasonSale, nil
	case "return":
		return InventoryReasonReturn, nil
	case "adjustment":
		return InventoryReasonAdjustment, nil
	case "reservation_release":
		return InventoryReasonReservationRelease, nil
	default:
		return "", ErrInvalidInventoryReason
	}
}

func (r InventoryReason) String() string {
	return string(r)
}

type InventoryLedger struct {
	ID          uuid.UUID       `db:"id" json:"id"`
	VariantID   uuid.UUID       `db:"variant_id" json:"variant_id"`
	Quantity    int             `db:"quantity" json:"quantity"`
	Reason      InventoryReason `db:"reason" json:"reason"`
	ReferenceID *string         `db:"reference_id" json:"reference_id,omitempty"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
}

type InventoryReservation struct {
	ID         uuid.UUID  `db:"id" json:"id"`
	VariantID  uuid.UUID  `db:"variant_id" json:"variant_id"`
	CartID     *uuid.UUID `db:"cart_id" json:"cart_id,omitempty"`
	Quantity   int        `db:"quantity" json:"quantity"`
	ExpiresAt  time.Time  `db:"expires_at" json:"expires_at"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
	ReleasedAt *time.Time `db:"released_at" json:"released_at,omitempty"`
}

type InventoryStock struct {
	VariantID         uuid.UUID `json:"variant_id"`
	OnHandQuantity    int       `json:"on_hand_quantity"`
	ReservedQuantity  int       `json:"reserved_quantity"`
	AvailableQuantity int       `json:"available_quantity"`
	IsInStock         bool      `json:"is_in_stock"`
}
