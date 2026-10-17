package errcode

// Business Domain Error Codes for Cart Subdomain.
// These constants define the canonical domain failure types returned by shopping cart operations.
const (
	// CodeCartNotFound indicates that the requested cart does not exist for the specified user or session.
	CodeCartNotFound = "CART_NOT_FOUND"

	// CodeCartItemNotFound indicates that a specific item or product variant was not found within the cart.
	CodeCartItemNotFound = "CART_ITEM_NOT_FOUND"

	// CodeCartItemAlreadyExists indicates an attempt to add an item that already exists in the cart when unique constraint is enforced.
	CodeCartItemAlreadyExists = "CART_ITEM_ALREADY_EXISTS"

	// CodeInvalidQuantity indicates that the requested quantity is invalid (e.g., non-positive, zero, or exceeds upper system bounds).
	CodeInvalidQuantity = "INVALID_CART_QUANTITY"

	// CodeMergeCartFailed indicates that merging an anonymous guest cart into an authenticated customer cart encountered an error.
	CodeMergeCartFailed = "MERGE_CART_FAILED"
)
