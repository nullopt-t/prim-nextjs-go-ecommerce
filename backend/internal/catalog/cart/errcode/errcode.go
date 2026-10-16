// Package errcode defines standard machine-readable error codes for shopping cart operations.
// These codes are returned in API error responses to enable client applications to handle specific
// cart failure modes programmatically (e.g., displaying out-of-stock notices or prompting for quantity adjustment).
package errcode

const (
	// CodeCartNotFound indicates that the requested cart does not exist for the specified user or session.
	CodeCartNotFound = "CART_NOT_FOUND"

	// CodeCartItemNotFound indicates that a specific item or product variant was not found within the cart.
	CodeCartItemNotFound = "CART_ITEM_NOT_FOUND"

	// CodeCartItemAlreadyExists indicates an attempt to add an item that already exists in the cart when unique constraint is enforced.
	CodeCartItemAlreadyExists = "CART_ITEM_ALREADY_EXISTS"

	// CodeInvalidQuantity indicates that the requested quantity is invalid (e.g., non-positive, zero, or exceeds upper system bounds).
	CodeInvalidQuantity = "INVALID_CART_QUANTITY"

	// CodeInsufficientInventory indicates that requested item quantity exceeds the currently available warehouse stock.
	CodeInsufficientInventory = "INSUFFICIENT_INVENTORY"

	// CodeVariantNotFound indicates that the product variant ID provided does not exist in the catalog.
	CodeVariantNotFound = "VARIANT_NOT_FOUND"

	// CodeMergeCartFailed indicates that merging an anonymous guest cart into an authenticated customer cart encountered an error.
	CodeMergeCartFailed = "MERGE_CART_FAILED"
)
