package api

type FieldError struct {
	Field   string `json:"field,omitempty" example:"name"`
	Tags    string `json:"tags,omitempty" example:"string"`
	Message string `json:"message,omitempty" example:"name is required and cannot be empty"`
}

type MessageResponse struct {
	Message string `json:"message" example:"Operation completed successfully"`
}

type UnauthorizedResponse struct {
	Code    string `json:"code,omitempty" example:"UNAUTHORIZED"`
	Message string `json:"message" example:"Authentication token is missing or invalid"`
}

type SuccessResponse struct {
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Meta    any    `json:"meta,omitempty"`
}

type PaginatedResponse struct {
	Data any `json:"data,omitempty"`
	Meta any `json:"meta,omitempty"`
}

type DataResponse struct {
	Data any `json:"data"`
}

type BadReqResponse struct {
	Code    string       `json:"code,omitempty" example:"VALIDATION_FAILED"`
	Message string       `json:"message,omitempty" example:"Invalid input or payload parameters"`
	Details []FieldError `json:"details,omitempty"`
}

type ErrorResponse struct {
	Code    string       `json:"code,omitempty" example:"INTERNAL_ERROR"`
	Message string       `json:"message,omitempty" example:"An unexpected internal server error occurred"`
	Details []FieldError `json:"details,omitempty"`
}

// --- Dedicated Status Code Error Response Models for Swagger Docs ---

type BadRequestErrorResponse struct {
	Code    string       `json:"code,omitempty" example:"VALIDATION_FAILED"`
	Message string       `json:"message,omitempty" example:"Invalid request input or malformed payload"`
	Details []FieldError `json:"details,omitempty"`
}

type UnauthorizedErrorResponse struct {
	Code    string       `json:"code,omitempty" example:"UNAUTHORIZED"`
	Message string       `json:"message,omitempty" example:"Authentication token is missing or expired"`
	Details []FieldError `json:"details,omitempty"`
}

type ForbiddenErrorResponse struct {
	Code    string       `json:"code,omitempty" example:"FORBIDDEN"`
	Message string       `json:"message,omitempty" example:"You do not have permission to access this resource"`
	Details []FieldError `json:"details,omitempty"`
}

type NotFoundErrorResponse struct {
	Code    string       `json:"code,omitempty" example:"NOT_FOUND"`
	Message string       `json:"message,omitempty" example:"The requested resource was not found"`
	Details []FieldError `json:"details,omitempty"`
}

type ConflictErrorResponse struct {
	Code    string       `json:"code,omitempty" example:"ALREADY_EXISTS"`
	Message string       `json:"message,omitempty" example:"A resource with this identifier or name already exists"`
	Details []FieldError `json:"details,omitempty"`
}

type InternalServerErrorResponse struct {
	Code    string       `json:"code,omitempty" example:"INTERNAL_ERROR"`
	Message string       `json:"message,omitempty" example:"An unexpected internal server error occurred"`
	Details []FieldError `json:"details,omitempty"`
}
