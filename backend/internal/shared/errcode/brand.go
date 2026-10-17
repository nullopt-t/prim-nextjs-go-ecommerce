package errcode

// Business Domain Error Codes for Brand Subdomain.
// These constants define the canonical domain failure types returned by brand lifecycle operations.
const (
	// CodeBrandNotFound indicates that the requested brand UUID was not found or is soft-deleted.
	CodeBrandNotFound = "BRAND_NOT_FOUND"

	// CodeBrandAlreadyExists indicates that a brand with the given name already exists in the catalog.
	CodeBrandAlreadyExists = "BRAND_ALREADY_EXISTS"

	// CodeBrandInUse indicates that a brand cannot be deleted because active products are still linked to it.
	CodeBrandInUse = "BRAND_IN_USE"

	// CodeLogoStorageObjectNotFound indicates that the object storage reference (UUID) provided for a brand logo does not exist.
	CodeLogoStorageObjectNotFound = "LOGO_STORAGE_OBJECT_NOT_FOUND"
)
