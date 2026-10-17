package errcode

// Business Domain Error Codes for Wishlist Subdomain.
// These constants define the canonical domain failure types returned by wishlist operations.
const (
	// CodeWishlistItemNotFound indicates that the requested item is not present in the customer's wishlist.
	CodeWishlistItemNotFound = "WISHLIST_ITEM_NOT_FOUND"

	// CodeWishlistItemAlreadyExists indicates that the product or variant is already present in the customer's wishlist.
	CodeWishlistItemAlreadyExists = "WISHLIST_ITEM_ALREADY_EXISTS"
)
