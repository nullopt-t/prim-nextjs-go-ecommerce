// Package api defines standardized HTTP response envelopes, validation error representations,
// and OpenAPI/Swagger schema documentation models for the Prim e-commerce API.
package api

// FieldError represents a single property validation failure,
// detailing which field, constraint rule, and human-readable reason failed.
type FieldError struct {
	// Field is the name of the request property that failed validation (e.g., "quantity", "email").
	Field      string `json:"field,omitempty" example:"quantity"`

	// Constraint is the validation tag or rule that was violated (e.g., "min", "required", "email").
	Constraint string `json:"constraint,omitempty" example:"min"`

	// Message is the user-friendly explanation of why the value is invalid.
	Message    string `json:"message,omitempty" example:"quantity must be at least 1"`
}

// MessageResponse is a simple acknowledgment envelope containing an informational status message.
type MessageResponse struct {
	// Message conveys the completion status or outcome of a mutation.
	Message string `json:"message" example:"Operation completed successfully"`
}

// UnauthorizedResponse represents a 401 Unauthorized API error envelope.
type UnauthorizedResponse struct {
	// Code is the machine-readable error token ("UNAUTHORIZED").
	Code    string `json:"code,omitempty" example:"UNAUTHORIZED"`

	// Message describes the authentication failure.
	Message string `json:"message" example:"Authentication token is missing or invalid"`
}

// SuccessResponse is the standard envelope for successful single-resource or action responses.
type SuccessResponse struct {
	// Message is an optional user-facing confirmation message.
	Message string `json:"message,omitempty"`

	// Data contains the primary response payload (e.g. a resource DTO).
	Data    any    `json:"data"`

	// Meta holds optional contextual information (e.g. processing timestamps or request IDs).
	Meta    any    `json:"meta,omitempty"`
}

// PaginatedResponse is the envelope for collections returning both data and pagination metadata.
type PaginatedResponse struct {
	// Data contains the slice of paginated items.
	Data any `json:"data"`

	// Meta contains the pagination metadata (page, pageSize, totalItems, totalPages).
	Meta any `json:"meta,omitempty"`
}

// DataResponse wraps arbitrary payloads under a mandatory "data" key without omitempty.
type DataResponse struct {
	// Data holds the response payload.
	Data any `json:"data"`
}

// BadReqResponse represents a 400 Bad Request envelope with field validation details.
type BadReqResponse struct {
	// Code is the machine-readable error token ("VALIDATION_FAILED").
	Code    string       `json:"code,omitempty" example:"VALIDATION_FAILED"`

	// Message describes the validation failure summary.
	Message string       `json:"message,omitempty" example:"Invalid request input or malformed payload"`

	// Details contains individual field-level validation errors.
	Details []FieldError `json:"details,omitempty"`
}

// ErrorResponse represents a generic error envelope for 500 Internal Server Errors.
type ErrorResponse struct {
	// Code is the machine-readable error token ("INTERNAL_ERROR").
	Code    string       `json:"code,omitempty" example:"INTERNAL_ERROR"`

	// Message describes the error condition.
	Message string       `json:"message,omitempty" example:"An unexpected internal server error occurred"`

	// Details contains optional validation details if applicable.
	Details []FieldError `json:"details,omitempty"`
}

// --- Dedicated Status Code Error Response Models for Swagger Docs ---

// BadRequestErrorResponse models HTTP 400 responses for Swagger/OpenAPI documentation.
// Aliased to BadReqResponse for schema consolidation.
type BadRequestErrorResponse = BadReqResponse

// UnauthorizedErrorResponse models HTTP 401 responses for Swagger/OpenAPI documentation.
type UnauthorizedErrorResponse struct {
	Code    string `json:"code,omitempty" example:"UNAUTHORIZED"`
	Message string `json:"message,omitempty" example:"Authentication token is missing or expired"`
}

// ForbiddenErrorResponse models HTTP 403 responses for Swagger/OpenAPI documentation.
type ForbiddenErrorResponse struct {
	Code    string `json:"code,omitempty" example:"FORBIDDEN"`
	Message string `json:"message,omitempty" example:"You do not have permission to access this resource"`
}

// NotFoundErrorResponse models HTTP 404 responses for Swagger/OpenAPI documentation.
type NotFoundErrorResponse struct {
	Code    string `json:"code,omitempty" example:"NOT_FOUND"`
	Message string `json:"message,omitempty" example:"The requested resource was not found"`
}

// ConflictErrorResponse models HTTP 409 responses for Swagger/OpenAPI documentation.
type ConflictErrorResponse struct {
	Code    string `json:"code,omitempty" example:"ALREADY_EXISTS"`
	Message string `json:"message,omitempty" example:"A resource with this identifier or name already exists"`
}

// PayloadTooLargeErrorResponse models HTTP 413 responses for Swagger/OpenAPI documentation.
type PayloadTooLargeErrorResponse struct {
	Code    string `json:"code,omitempty" example:"FILE_TOO_LARGE"`
	Message string `json:"message,omitempty" example:"Request payload exceeds the maximum allowed size"`
}

// TooManyRequestsErrorResponse models HTTP 429 responses for Swagger/OpenAPI documentation.
type TooManyRequestsErrorResponse struct {
	Code    string `json:"code,omitempty" example:"RATE_LIMIT_EXCEEDED"`
	Message string `json:"message,omitempty" example:"Rate limit exceeded, please retry later"`
}

// InternalServerErrorResponse models HTTP 500 responses for Swagger/OpenAPI documentation.
type InternalServerErrorResponse struct {
	Code    string `json:"code,omitempty" example:"INTERNAL_ERROR"`
	Message string `json:"message,omitempty" example:"An unexpected internal server error occurred"`
}

// ServiceUnavailableErrorResponse models HTTP 503 responses for Swagger/OpenAPI documentation.
type ServiceUnavailableErrorResponse struct {
	Code    string `json:"code,omitempty" example:"SERVICE_UNAVAILABLE"`
	Message string `json:"message,omitempty" example:"Service temporarily unavailable"`
}

// NewSuccessResponse constructs a SuccessResponse envelope.
func NewSuccessResponse(data any, message ...string) *SuccessResponse {
	var msg string
	if len(message) > 0 {
		msg = message[0]
	}
	return &SuccessResponse{
		Data:    data,
		Message: msg,
	}
}

// NewPaginatedResponse constructs a PaginatedResponse envelope.
// If data is nil, it defaults to an empty slice to ensure valid JSON array serialization.
func NewPaginatedResponse(data any, meta any) *PaginatedResponse {
	if data == nil {
		data = []any{}
	}
	return &PaginatedResponse{
		Data: data,
		Meta: meta,
	}
}

// NewDataResponse constructs a DataResponse envelope wrapping the payload.
func NewDataResponse(data any) *DataResponse {
	return &DataResponse{
		Data: data,
	}
}

// NewMessageResponse constructs an informational MessageResponse envelope.
func NewMessageResponse(message string) *MessageResponse {
	return &MessageResponse{
		Message: message,
	}
}

// NewErrorResponse constructs an ErrorResponse envelope.
func NewErrorResponse(code string, message string, details ...FieldError) *ErrorResponse {
	return &ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// NewBadReqResponse constructs a BadReqResponse envelope with field validation details.
func NewBadReqResponse(code string, message string, details ...FieldError) *BadReqResponse {
	return &BadReqResponse{
		Code:    code,
		Message: message,
		Details: details,
	}
}
