package errcode

// Business Domain Error Codes for Tag Subdomain.
// These constants define the canonical domain failure types returned by product tag operations.
const (
	// CodeTagNotFound indicates that the requested product tag UUID was not found or is soft-deleted.
	CodeTagNotFound = "TAG_NOT_FOUND"

	// CodeTagAlreadyExists indicates that a product tag with the given name already exists.
	CodeTagAlreadyExists = "TAG_ALREADY_EXISTS"

	// CodeTagInUse indicates that a tag cannot be deleted due to active product assignments.
	CodeTagInUse = "TAG_IN_USE"
)
