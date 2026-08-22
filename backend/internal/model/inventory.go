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
	ID        uuid.UUID
	VariantID uuid.UUID
	Quantity  int
	Reason    InventoryReason
	CreatedAt time.Time
}

type InventoryReservation struct {
	ID         uuid.UUID
	VariantID  uuid.UUID
	CartID     *uuid.UUID
	Quantity   int
	ExpiresAt  time.Time
	CreatedAt  time.Time
	ReleasedAt *time.Time
}

type InventoryStock struct {
	VariantID         uuid.UUID
	OnHandQuantity    int
	ReservedQuantity  int
	AvailableQuantity int
	IsInStock         bool
}
