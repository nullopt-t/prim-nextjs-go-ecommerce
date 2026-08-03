package errcode

// Business Domain Error Codes for Product Subdomain
const (
	CodeProductNotFound          = "PRODUCT_NOT_FOUND"
	CodeProductAlreadyExists     = "PRODUCT_ALREADY_EXISTS"
	CodeVariantNotFound          = "VARIANT_NOT_FOUND"
	CodeVariantProductMismatch   = "VARIANT_PRODUCT_MISMATCH"
	CodeBrandNotFound            = "BRAND_NOT_FOUND"
	CodeBrandAlreadyExists       = "BRAND_ALREADY_EXISTS"
	CodeLogoStorageObjectNotFound = "LOGO_STORAGE_OBJECT_NOT_FOUND"
	CodeCategoryNotFound         = "CATEGORY_NOT_FOUND"
	CodeCategoryAlreadyExists    = "CATEGORY_ALREADY_EXISTS"
	CodeParentCategoryNotFound   = "PARENT_CATEGORY_NOT_FOUND"
	CodeInvalidCategoryHierarchy  = "INVALID_CATEGORY_HIERARCHY"
	CodeCyclicCategoryHierarchy   = "CYCLIC_CATEGORY_HIERARCHY"
	CodeTagNotFound              = "TAG_NOT_FOUND"
	CodeTagAlreadyExists         = "TAG_ALREADY_EXISTS"
	CodeMediaNotFound            = "MEDIA_NOT_FOUND"
	CodeMediaAlreadyAttached     = "MEDIA_ALREADY_ATTACHED"
	CodePublishFailed            = "PUBLISH_FAILED"
)
