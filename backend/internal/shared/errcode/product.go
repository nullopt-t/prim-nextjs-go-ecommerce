package errcode

// Business Domain Error Codes for Product Subdomain.
// These constants define the canonical domain failure types returned by product lifecycle operations.
const (
	// CodeProductNotFound indicates the requested product UUID or slug does not exist or is soft-deleted.
	CodeProductNotFound = "PRODUCT_NOT_FOUND"

	// CodeProductAlreadyExists indicates a unique constraint violation when attempting to create
	// a product with a conflicting unique identifier or slug.
	CodeProductAlreadyExists = "PRODUCT_ALREADY_EXISTS"

	// CodePublishFailed indicates that a product cannot be published (e.g. missing variants or required attributes).
	CodePublishFailed = "PUBLISH_FAILED"

	// CodeVariationAxesInvalid indicates that one or more requested variation axes
	// are not defined attributes for the product's category.
	CodeVariationAxesInvalid = "VARIATION_AXES_INVALID"

	// CodeVariationAxesLocked indicates that variation axes cannot be changed while
	// a product already has multiple child variants.
	CodeVariationAxesLocked = "VARIATION_AXES_LOCKED"
)
