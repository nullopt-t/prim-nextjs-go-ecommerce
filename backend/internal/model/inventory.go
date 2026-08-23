package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// InventoryReason describes the business transaction rationale for modifying stock levels.
type InventoryReason string

const (
	// InventoryReasonRestock indicates stock received from a supplier or returned to warehouse inventory.
	InventoryReasonRestock InventoryReason = "restock"
	// InventoryReasonSale indicates stock depleted due to completed customer purchases.
	InventoryReasonSale InventoryReason = "sale"
	// InventoryReasonReturn indicates stock returned by a customer upon an order return.
	InventoryReasonReturn InventoryReason = "return"
	// InventoryReasonAdjustment indicates manual administrative correction (e.g. damaged goods, audit discrepancy).
	InventoryReasonAdjustment InventoryReason = "adjustment"
	// InventoryReasonReservationRelease indicates temporary hold release (e.g. cart expired or checkout cancelled).
	InventoryReasonReservationRelease InventoryReason = "reservation_release"
)

// ErrInvalidInventoryReason is returned when an unrecognized inventory reason string is encountered.
var ErrInvalidInventoryReason = errors.New("invalid inventory reason")

// ParseInventoryReason converts a raw string into a typed InventoryReason.
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

// String returns the string representation of the InventoryReason.
func (r InventoryReason) String() string {
	return string(r)
}

// InventoryLedger represents an immutable append-only transaction event for stock changes.
type InventoryLedger struct {
	// ID is the unique UUID for the ledger transaction entry.
	ID uuid.UUID
	// VariantID is the specific ProductVariant UUID adjusted.
	VariantID uuid.UUID
	// Quantity is the signed unit delta (positive for additions, negative for deductions).
	Quantity int
	// Reason explains the business classification for this stock transaction.
	Reason InventoryReason
	// CreatedAt is the timestamp when the ledger record was inserted.
	CreatedAt time.Time
}

// InventoryReservation represents a temporary hold placed on stock during active checkouts.
type InventoryReservation struct {
	// ID is the unique UUID for the reservation record.
	ID uuid.UUID
	// VariantID is the specific ProductVariant UUID reserved.
	VariantID uuid.UUID
	// CartID is the optional associated Cart UUID holding this reservation.
	CartID *uuid.UUID
	// Quantity is the number of units placed on hold.
	Quantity int
	// ExpiresAt is the timestamp after which unfulfilled reservations are automatically released.
	ExpiresAt time.Time
	// CreatedAt is the timestamp when the reservation was created.
	CreatedAt time.Time
	// ReleasedAt is the optional timestamp when the reservation was released or converted to a sale.
	ReleasedAt *time.Time
}

// InventoryStock represents the aggregated real-time stock balance for a product variant.
type InventoryStock struct {
	// VariantID is the specific ProductVariant UUID.
	VariantID uuid.UUID
	// OnHandQuantity is the total physical unit count located in the warehouse.
	OnHandQuantity int
	// ReservedQuantity is the number of units currently locked in active orders/checkouts.
	ReservedQuantity int
	// AvailableQuantity is the salable stock (OnHandQuantity - ReservedQuantity).
	AvailableQuantity int
	// IsInStock is true when AvailableQuantity is greater than 0.
	IsInStock bool
}
