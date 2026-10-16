// Package apierr provides structured application error models, standard HTTP error
// constructors, and slog.LogValuer integration for the Prim e-commerce API.
package apierr

import (
	"log/slog"
	"net/http"

	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/errs"
)

// FieldError is an alias for api.FieldError, capturing individual field-level validation failures.
type FieldError = api.FieldError

// Standard API Error Codes (Generic & Infrastructure Concerns)
const (
	// CodeBadRequest represents general malformed HTTP requests.
	CodeBadRequest            = "BAD_REQUEST"
	// CodeInvalidInput represents malformed data parameters (e.g. invalid UUIDs or wrong types).
	CodeInvalidInput          = "INVALID_INPUT"
	// CodeInvalidPayload represents unparseable request body JSON/form data.
	CodeInvalidPayload        = "INVALID_PAYLOAD"
	// CodeValidationFailed represents domain or schema constraint validation failures.
	CodeValidationFailed      = "VALIDATION_FAILED"
	// CodeUnauthorized represents unauthenticated requests lacking valid credentials.
	CodeUnauthorized          = "UNAUTHORIZED"
	// CodeForbidden represents authenticated requests lacking required permissions or role.
	CodeForbidden             = "FORBIDDEN"
	// CodeNotFound represents requests targeting non-existent resources.
	CodeNotFound              = "NOT_FOUND"
	// CodeAlreadyExists represents uniqueness constraint violations during creation.
	CodeAlreadyExists         = "ALREADY_EXISTS"
	// CodeResourceConflict represents concurrent state conflicts or optimistic locking failures.
	CodeResourceConflict      = "RESOURCE_CONFLICT"
	// CodeRateLimitExceeded represents client request rate limit throttling.
	CodeRateLimitExceeded     = "RATE_LIMIT_EXCEEDED"
	// CodeExpired represents token, challenge, or entity expiration.
	CodeExpired               = "EXPIRED"
	// CodeInvalidOrExpired represents auth tokens that are either invalid or expired.
	CodeInvalidOrExpired      = "INVALID_OR_EXPIRED"
	// CodeInternalError represents unexpected backend panics or system failures.
	CodeInternalError         = "INTERNAL_ERROR"
	// CodeInvalidReference represents invalid foreign key relationships.
	CodeInvalidReference      = "INVALID_REFERENCE"
	// CodeInvalidQueryParameter represents unparseable or out-of-range URL query parameters.
	CodeInvalidQueryParameter = "INVALID_QUERY_PARAMETER"

	// Media & File Transfer Codes

	// CodeFileRequired indicates a required multipart file upload is missing.
	CodeFileRequired         = "FILE_REQUIRED"
	// CodeFileTooLarge indicates uploaded payload exceeds configured byte limits.
	CodeFileTooLarge         = "FILE_TOO_LARGE"
	// CodeUnsupportedMediaType indicates uploaded Content-Type or MIME type is rejected.
	CodeUnsupportedMediaType = "UNSUPPORTED_MEDIA_TYPE"
	// CodeStorageError indicates an internal error during S3/MinIO upload/download.
	CodeStorageError         = "STORAGE_ERROR"
)

// APIError represents a structured application API error embedding nullopt-t/errs.
// It carries HTTP status codes, machine-readable error codes, and field validation details.
type APIError struct {
	// AppError embeds message formatting, wrapped error cause, and stack traces.
	*errs.AppError

	// Status is the HTTP status code (e.g., 400, 404, 500) sent to the client.
	Status int `json:"status,omitempty"`

	// Code is the machine-readable error code string (e.g., "NOT_FOUND").
	Code string `json:"code,omitempty"`

	// Fields contains detailed field validation error details.
	Fields []FieldError `json:"details,omitempty"`
}

// New constructs a new APIError requiring HTTP status code and human-readable message.
func New(status int, message string) *APIError {
	return &APIError{
		AppError: errs.New(message),
		Status:   status,
	}
}

// WithCode assigns a machine-readable error code to the APIError.
func (ae *APIError) WithCode(code string) *APIError {
	ae.Code = code
	return ae
}

// WithMessage updates the error message string, preserving any underlying cause and stack trace.
func (ae *APIError) WithMessage(m string) *APIError {
	if ae.AppError == nil {
		ae.AppError = errs.New(m)
	} else {
		ae.AppError.Message = m
	}
	return ae
}

// Wrap chains an underlying root cause error into the APIError for diagnostic unwrapping.
func (ae *APIError) Wrap(err error) *APIError {
	if ae.AppError == nil {
		ae.AppError = errs.New("")
	}
	_ = ae.AppError.Wrap(err)
	return ae
}

// WithFields appends field-level validation failures to the error.
func (ae *APIError) WithFields(fields ...FieldError) *APIError {
	ae.Fields = append(ae.Fields, fields...)
	return ae
}

// WithStack captures a stack trace at the call site for debugging.
func (ae *APIError) WithStack() *APIError {
	if ae.AppError == nil {
		ae.AppError = errs.New("")
	}
	_ = ae.AppError.WithStack()
	return ae
}

// Error returns the human-readable error message.
func (ae *APIError) Error() string {
	if ae.AppError == nil {
		return ""
	}
	return ae.AppError.Error()
}

// Unwrap returns the underlying wrapped error cause.
func (ae *APIError) Unwrap() error {
	if ae.AppError == nil {
		return nil
	}
	return ae.AppError.Unwrap()
}


// LogValue implements slog.LogValuer to produce structured slog attribute groups
// containing error status, code, message, underlying cause, validation fields, and stack trace.
func (ae *APIError) LogValue() slog.Value {
	if ae == nil {
		return slog.Value{}
	}

	attrs := make([]slog.Attr, 0, 6)

	if ae.Status != 0 {
		attrs = append(attrs, slog.Int("status", ae.Status))
	}

	if ae.Code != "" {
		attrs = append(attrs, slog.String("code", ae.Code))
	}

	if ae.AppError != nil {
		if msg := ae.Error(); msg != "" {
			attrs = append(attrs, slog.String("message", msg))
		}

		if cause := ae.Unwrap(); cause != nil {
			attrs = append(attrs, slog.String("internal", cause.Error()))
		}

		if stack := ae.Stack(); len(stack) > 0 {
			attrs = append(attrs, slog.Any("stack", stack))
		}
	}

	if len(ae.Fields) > 0 {
		attrs = append(attrs, slog.Any("fields", ae.Fields))
	}

	return slog.GroupValue(attrs...)
}

// BadRequestError creates a 400 Bad Request error with default or custom message.
func BadRequestError(message ...string) *APIError {
	msg := "Bad request"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return New(http.StatusBadRequest, msg).WithCode(CodeBadRequest)
}

// NotFoundError creates a 404 Not Found error with default or custom message.
func NotFoundError(message ...string) *APIError {
	msg := "Resource not found"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return New(http.StatusNotFound, msg).WithCode(CodeNotFound)
}

// UnauthorizedError creates a 401 Unauthorized error with default or custom message.
func UnauthorizedError(message ...string) *APIError {
	msg := "Unauthorized access"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return New(http.StatusUnauthorized, msg).WithCode(CodeUnauthorized)
}

// ForbiddenError creates a 403 Forbidden error with default or custom message.
func ForbiddenError(message ...string) *APIError {
	msg := "Access forbidden"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return New(http.StatusForbidden, msg).WithCode(CodeForbidden)
}

// ConflictError creates a 409 Conflict error with default or custom message.
func ConflictError(message ...string) *APIError {
	msg := "Resource conflict"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return New(http.StatusConflict, msg).WithCode(CodeResourceConflict)
}

// InternalServerError creates a 500 Internal Server Error with default or custom message.
func InternalServerError(message ...string) *APIError {
	msg := "Internal server error"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return New(http.StatusInternalServerError, msg).WithCode(CodeInternalError)
}

// TooManyRequestsError creates a 429 Too Many Requests error with default or custom message.
func TooManyRequestsError(message ...string) *APIError {
	msg := "Too many requests"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return New(http.StatusTooManyRequests, msg).WithCode(CodeRateLimitExceeded)
}

// InvalidPayloadError creates a 400 Bad Request error for malformed request payloads.
func InvalidPayloadError(message ...string) *APIError {
	msg := "Invalid request payload"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return New(http.StatusBadRequest, msg).WithCode(CodeInvalidPayload)
}

// ValidationFailedError creates a 400 Bad Request error for field validation failures.
func ValidationFailedError(message ...string) *APIError {
	msg := "Validation failed"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return New(http.StatusBadRequest, msg).WithCode(CodeValidationFailed)
}

// InvalidUUIDError creates a 400 Bad Request error for invalid UUID syntax.
func InvalidUUIDError(message ...string) *APIError {
	msg := "Invalid UUID format"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return New(http.StatusBadRequest, msg).WithCode(CodeInvalidInput)
}

// InvalidReferenceError creates a 400 Bad Request error for foreign key constraint or reference failures.
func InvalidReferenceError(message ...string) *APIError {
	msg := "Invalid reference"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return New(http.StatusBadRequest, msg).WithCode(CodeInvalidReference)
}

// InvalidQueryParameterError creates a 400 Bad Request error for unparseable or out-of-range URL query parameters.
func InvalidQueryParameterError(message ...string) *APIError {
	msg := "Invalid query parameter"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return New(http.StatusBadRequest, msg).WithCode(CodeInvalidQueryParameter)
}

// FileRequiredError creates a 400 Bad Request error when a required multipart file upload is missing.
func FileRequiredError(message ...string) *APIError {
	msg := "File is required"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return New(http.StatusBadRequest, msg).WithCode(CodeFileRequired)
}

// FileTooLargeError creates a 413 Payload Too Large error when uploaded content exceeds size limits.
func FileTooLargeError(message ...string) *APIError {
	msg := "File too large"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return New(http.StatusRequestEntityTooLarge, msg).WithCode(CodeFileTooLarge)
}

// UnsupportedMediaTypeError creates a 415 Unsupported Media Type error for rejected Content-Type or MIME type.
func UnsupportedMediaTypeError(message ...string) *APIError {
	msg := "Unsupported media type"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return New(http.StatusUnsupportedMediaType, msg).WithCode(CodeUnsupportedMediaType)
}

// StorageError creates a 500 Internal Server Error for object storage upload/download failures.
func StorageError(message ...string) *APIError {
	msg := "Storage error"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return New(http.StatusInternalServerError, msg).WithCode(CodeStorageError)
}

