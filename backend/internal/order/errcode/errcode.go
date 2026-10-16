// Package errcode defines standard business domain error codes for the order and checkout subdomains.
// These error constants are mapped to API error responses to provide structured, machine-readable
// failure categories to client applications and upstream consumers.
package errcode

// Business Domain Error Codes for Order Subdomain.
// These constants define the canonical domain failure types returned by order lifecycle operations.
const (
	// CodeOrderNotFound is returned when an order requested by UUID does not exist or has been deleted.
	CodeOrderNotFound = "ORDER_NOT_FOUND"

	// CodeOrderAlreadyExists is returned when attempting to insert an order with an identifier that conflicts.
	CodeOrderAlreadyExists = "ORDER_ALREADY_EXISTS"

	// CodeInvalidOrderStatus is returned when an order status transition violates state machine rules.
	CodeInvalidOrderStatus = "INVALID_ORDER_STATUS"

	// CodeEmptyOrderItems is returned when an order creation payload contains no line items.
	CodeEmptyOrderItems = "EMPTY_ORDER_ITEMS"

	// CodeOrderItemNotFound is returned when a specific order line item is not located in persistence.
	CodeOrderItemNotFound = "ORDER_ITEM_NOT_FOUND"
)
