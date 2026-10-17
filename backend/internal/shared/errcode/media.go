package errcode

// Business Domain Error Codes for Product Media Subdomain.
// These constants define the canonical domain failure types returned when attaching media to variants.
const (
	// CodeMediaNotFound indicates that a requested product media attachment was not found.
	CodeMediaNotFound = "MEDIA_NOT_FOUND"

	// CodeMediaAlreadyAttached indicates that a media object has already been associated with the target variant.
	CodeMediaAlreadyAttached = "MEDIA_ALREADY_ATTACHED"
)
