// Package errcode defines standard machine-readable error codes for the catalog domain.
// These business error codes are returned in API error responses under the "code" field,
// allowing client applications to programmatically identify specific error conditions
// without relying on brittle HTTP status codes or human-readable messages.
package errcode

// Business Domain Error Codes for Product Subdomain.
// These string constants represent discrete failure modes across products, variants,
// brands, categories, tags, inventory, reviews, and wishlists.
const (
	// CodeProductNotFound indicates the requested product UUID or slug does not exist or is soft-deleted.
	CodeProductNotFound = "PRODUCT_NOT_FOUND"

	// CodeProductAlreadyExists indicates a unique constraint violation when attempting to create
	// a product with a conflicting unique identifier or slug.
	CodeProductAlreadyExists = "PRODUCT_ALREADY_EXISTS"

	// CodeVariantNotFound indicates the requested product variant UUID was not found.
	CodeVariantNotFound = "VARIANT_NOT_FOUND"

	// CodeVariantAlreadyExists indicates a conflicting SKU or variant combination already exists.
	CodeVariantAlreadyExists = "VARIANT_ALREADY_EXISTS"

	// CodeVariantProductMismatch indicates that a variant operation was requested against a product
	// that does not own the specified variant.
	CodeVariantProductMismatch = "VARIANT_PRODUCT_MISMATCH"

	// CodeBrandNotFound indicates that the requested brand UUID was not found or is soft-deleted.
	CodeBrandNotFound = "BRAND_NOT_FOUND"

	// CodeBrandAlreadyExists indicates that a brand with the given name already exists in the catalog.
	CodeBrandAlreadyExists = "BRAND_ALREADY_EXISTS"

	// CodeBrandInUse indicates that a brand cannot be deleted because active products are still linked to it.
	CodeBrandInUse = "BRAND_IN_USE"

	// CodeLogoStorageObjectNotFound indicates that the object storage reference (UUID) provided for a brand logo does not exist.
	CodeLogoStorageObjectNotFound = "LOGO_STORAGE_OBJECT_NOT_FOUND"

	// CodeCategoryNotFound indicates that the requested category UUID was not found or is soft-deleted.
	CodeCategoryNotFound = "CATEGORY_NOT_FOUND"

	// CodeCategoryAlreadyExists indicates that a category with the same name already exists under the same parent.
	CodeCategoryAlreadyExists = "CATEGORY_ALREADY_EXISTS"

	// CodeParentCategoryNotFound indicates that the parent category UUID referenced during creation or update does not exist.
	CodeParentCategoryNotFound = "PARENT_CATEGORY_NOT_FOUND"

	// CodeInvalidCategoryHierarchy indicates an invalid tree relationship, such as assigning a category as its own parent.
	CodeInvalidCategoryHierarchy = "INVALID_CATEGORY_HIERARCHY"

	// CodeCyclicCategoryHierarchy indicates that setting a parent would produce an infinite cycle in the category taxonomy tree.
	CodeCyclicCategoryHierarchy = "CYCLIC_CATEGORY_HIERARCHY"

	// CodeTagNotFound indicates that the requested product tag UUID was not found or is soft-deleted.
	CodeTagNotFound = "TAG_NOT_FOUND"

	// CodeTagAlreadyExists indicates that a product tag with the given name already exists.
	CodeTagAlreadyExists = "TAG_ALREADY_EXISTS"

	// CodeTagInUse indicates that a tag cannot be deleted due to active product assignments.
	CodeTagInUse = "TAG_IN_USE"

	// CodeCategoryInUse indicates that a category cannot be deleted due to assigned products or subcategories.
	CodeCategoryInUse = "CATEGORY_IN_USE"

	// CodeMediaNotFound indicates that a requested product media attachment was not found.
	CodeMediaNotFound = "MEDIA_NOT_FOUND"

	// CodeMediaAlreadyAttached indicates that a media object has already been associated with the target variant.
	CodeMediaAlreadyAttached = "MEDIA_ALREADY_ATTACHED"

	// CodePublishFailed indicates that a product cannot be published (e.g. missing variants or required attributes).
	CodePublishFailed = "PUBLISH_FAILED"

	// CodeInsufficientInventory indicates an attempt to reserve or purchase items exceeding available stock.
	CodeInsufficientInventory = "INSUFFICIENT_INVENTORY"

	// CodeInvalidInventoryReason indicates an unrecognized adjustment reason provided during stock level mutation.
	CodeInvalidInventoryReason = "INVALID_INVENTORY_REASON"

	// CodeReservationNotFound indicates that the specified inventory hold/reservation ID does not exist.
	CodeReservationNotFound = "INVENTORY_RESERVATION_NOT_FOUND"

	// CodeReservationExpired indicates that the inventory hold duration has lapsed and stock has been released.
	CodeReservationExpired = "INVENTORY_RESERVATION_EXPIRED"

	// CodeReviewAlreadyExists indicates that the user has already submitted a product review for this item.
	CodeReviewAlreadyExists = "REVIEW_ALREADY_EXISTS"

	// CodeReviewNotFound indicates that the requested review UUID does not exist.
	CodeReviewNotFound = "REVIEW_NOT_FOUND"

	// CodeReviewNotEligible indicates that the user has not purchased the item and cannot post a verified review.
	CodeReviewNotEligible = "REVIEW_NOT_ELIGIBLE"

	// CodeWishlistItemNotFound indicates that the requested item is not present in the customer's wishlist.
	CodeWishlistItemNotFound = "WISHLIST_ITEM_NOT_FOUND"

	// CodeWishlistItemAlreadyExists indicates that the product or variant is already present in the customer's wishlist.
	CodeWishlistItemAlreadyExists = "WISHLIST_ITEM_ALREADY_EXISTS"
)
