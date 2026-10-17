package errcode

// Business Domain Error Codes for Category Subdomain.
// These constants define the canonical domain failure types returned by category taxonomy operations.
const (
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

	// CodeCategoryInUse indicates that a category cannot be deleted due to assigned products or subcategories.
	CodeCategoryInUse = "CATEGORY_IN_USE"
)
