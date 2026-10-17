package errcode

// Business Domain Error Codes for Inventory Subdomain.
// These constants define the canonical domain failure types returned by stock level and reservation operations.
const (
	// CodeInsufficientInventory indicates an attempt to reserve or purchase items exceeding available stock.
	CodeInsufficientInventory = "INSUFFICIENT_INVENTORY"

	// CodeInvalidInventoryReason indicates an unrecognized adjustment reason provided during stock level mutation.
	CodeInvalidInventoryReason = "INVALID_INVENTORY_REASON"

	// CodeReservationNotFound indicates that the specified inventory hold/reservation ID does not exist.
	CodeReservationNotFound = "INVENTORY_RESERVATION_NOT_FOUND"

	// CodeReservationExpired indicates that the inventory hold duration has lapsed and stock has been released.
	CodeReservationExpired = "INVENTORY_RESERVATION_EXPIRED"
)
