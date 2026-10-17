package errcode

// Business Domain Error Codes for Product Variant Subdomain.
// These constants define the canonical domain failure types returned by variant (SKU) lifecycle operations.
const (
	// CodeVariantNotFound indicates the requested product variant UUID was not found.
	CodeVariantNotFound = "VARIANT_NOT_FOUND"

	// CodeVariantAlreadyExists indicates a conflicting SKU or variant combination already exists.
	CodeVariantAlreadyExists = "VARIANT_ALREADY_EXISTS"

	// CodeVariantProductMismatch indicates that a variant operation was requested against a product
	// that does not own the specified variant.
	CodeVariantProductMismatch = "VARIANT_PRODUCT_MISMATCH"
)
