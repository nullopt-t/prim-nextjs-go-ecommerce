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

	// CodeSimpleProductVariantLocked indicates that a variant lifecycle operation was rejected
	// because the parent product is a simple product with a fixed default variant.
	CodeSimpleProductVariantLocked = "SIMPLE_PRODUCT_VARIANT_LOCKED"

	// CodeVariationAxesRequired indicates that a product must define variation axes
	// before it can have more than one child variant.
	CodeVariationAxesRequired = "VARIATION_AXES_REQUIRED"

	// CodeVariationAxisMismatch indicates that a variant's attributes do not match
	// the product's chosen variation axes exactly.
	CodeVariationAxisMismatch = "VARIATION_AXIS_MISMATCH"

	// CodeVariationValueInvalid indicates that an axis value is empty or is not one
	// of the allowed values for its category attribute.
	CodeVariationValueInvalid = "VARIATION_VALUE_INVALID"

	// CodeVariantCombinationExists indicates that another child already uses the
	// same variation combination.
	CodeVariantCombinationExists = "VARIANT_COMBINATION_EXISTS"
)
