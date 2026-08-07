package apierr

import (
	"log/slog"
	"net/http"

	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/nullopt-t/errs"
)

type FieldError = api.FieldError

// Standard API Error Codes (Generic & Infrastructure Concerns)
const (
	CodeBadRequest            = "BAD_REQUEST"
	CodeInvalidInput          = "INVALID_INPUT"
	CodeInvalidPayload        = "INVALID_PAYLOAD"
	CodeValidationFailed      = "VALIDATION_FAILED"
	CodeUnauthorized          = "UNAUTHORIZED"
	CodeForbidden             = "FORBIDDEN"
	CodeNotFound              = "NOT_FOUND"
	CodeAlreadyExists         = "ALREADY_EXISTS"
	CodeResourceConflict      = "RESOURCE_CONFLICT"
	CodeRateLimitExceeded     = "RATE_LIMIT_EXCEEDED"
	CodeExpired               = "EXPIRED"
	CodeInvalidOrExpired      = "INVALID_OR_EXPIRED"
	CodeInternalError         = "INTERNAL_ERROR"
	CodeInvalidReference      = "INVALID_REFERENCE"
	CodeInvalidQueryParameter = "INVALID_QUERY_PARAMETER"

	// Media & File Transfer Codes
	CodeFileRequired         = "FILE_REQUIRED"
	CodeFileTooLarge         = "FILE_TOO_LARGE"
	CodeUnsupportedMediaType = "UNSUPPORTED_MEDIA_TYPE"
	CodeStorageError         = "STORAGE_ERROR"
)

// APIError represents a structured application API error embedding nullopt-t/errs.
type APIError struct {
	*errs.AppError

	// HTTP status code (e.g., 400, 404, 500)
	Status int `json:"status,omitempty"`

	// Machine-readable error code (e.g., "NOT_FOUND")
	Code string `json:"code,omitempty"`

	// Detailed field validation errors
	Fields []FieldError `json:"details,omitempty"`
}

// New constructs a new APIError requiring HTTP status code and message.
func New(status int, message string) *APIError {
	return &APIError{
		AppError: errs.New(message),
		Status:   status,
	}
}

func (ae *APIError) WithCode(code string) *APIError {
	ae.Code = code
	return ae
}

func (ae *APIError) WithMessage(m string) *APIError {
	ae.AppError = errs.New(m)
	return ae
}

func (ae *APIError) Wrap(err error) *APIError {
	_ = ae.AppError.Wrap(err)
	return ae
}

func (ae *APIError) WithFields(fields ...FieldError) *APIError {
	ae.Fields = append(ae.Fields, fields...)
	return ae
}

func (ae *APIError) WithStack() *APIError {
	_ = ae.AppError.WithStack()
	return ae
}

func (ae *APIError) LogValue() slog.Value {
	attrs := make([]slog.Attr, 0, 6)

	if ae.Status != 0 {
		attrs = append(attrs, slog.Int("status", ae.Status))
	}

	if ae.Code != "" {
		attrs = append(attrs, slog.String("code", ae.Code))
	}

	if msg := ae.Error(); msg != "" {
		attrs = append(attrs, slog.String("message", msg))
	}

	if cause := ae.Unwrap(); cause != nil {
		attrs = append(attrs, slog.String("internal", cause.Error()))
	}

	if len(ae.Fields) > 0 {
		attrs = append(attrs, slog.Any("fields", ae.Fields))
	}

	if stack := ae.Stack(); len(stack) > 0 {
		attrs = append(attrs, slog.Any("stack", stack))
	}

	return slog.GroupValue(attrs...)
}

// --- Predefined API Error Constructors ---

// ErrBadRequest creates a 400 Bad Request error.
func ErrBadRequest(message string) *APIError {
	return New(http.StatusBadRequest, message).WithCode(CodeBadRequest)
}

// BadRequestError creates a 400 Bad Request error with default or custom message.
func BadRequestError(message ...string) *APIError {
	msg := "Bad request"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return New(http.StatusBadRequest, msg).WithCode(CodeBadRequest)
}

// ErrNotFound creates a 404 Not Found error.
func ErrNotFound(message string) *APIError {
	return New(http.StatusNotFound, message).WithCode(CodeNotFound)
}

// NotFoundError creates a 404 Not Found error with default or custom message.
func NotFoundError(message ...string) *APIError {
	msg := "Resource not found"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return New(http.StatusNotFound, msg).WithCode(CodeNotFound)
}

// ErrUnauthorized creates a 401 Unauthorized error.
func ErrUnauthorized(message string) *APIError {
	return New(http.StatusUnauthorized, message).WithCode(CodeUnauthorized)
}

// UnauthorizedError creates a 401 Unauthorized error with default or custom message.
func UnauthorizedError(message ...string) *APIError {
	msg := "Unauthorized access"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return New(http.StatusUnauthorized, msg).WithCode(CodeUnauthorized)
}

// ErrForbidden creates a 403 Forbidden error.
func ErrForbidden(message string) *APIError {
	return New(http.StatusForbidden, message).WithCode(CodeForbidden)
}

// ForbiddenError creates a 403 Forbidden error with default or custom message.
func ForbiddenError(message ...string) *APIError {
	msg := "Access forbidden"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return New(http.StatusForbidden, msg).WithCode(CodeForbidden)
}

// ErrConflict creates a 409 Conflict error.
func ErrConflict(message string) *APIError {
	return New(http.StatusConflict, message).WithCode(CodeResourceConflict)
}

// ConflictError creates a 409 Conflict error with default or custom message.
func ConflictError(message ...string) *APIError {
	msg := "Resource conflict"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return New(http.StatusConflict, msg).WithCode(CodeResourceConflict)
}

// ErrInternalError creates a 500 Internal Server Error.
func ErrInternalError(message ...string) *APIError {
	msg := "Internal server error"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return New(http.StatusInternalServerError, msg).WithCode(CodeInternalError)
}

// InternalServerError creates a 500 Internal Server Error with default or custom message.
func InternalServerError(message ...string) *APIError {
	return ErrInternalError(message...)
}

// ErrInvalidUUID creates a 400 Bad Request error for invalid UUID syntax.
func ErrInvalidUUID() *APIError {
	return New(http.StatusBadRequest, "Invalid UUID format").WithCode(CodeInvalidInput)
}

// ErrInvalidPayload creates a 400 Bad Request error for malformed request payloads.
func ErrInvalidPayload(message ...string) *APIError {
	msg := "Invalid request payload"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return New(http.StatusBadRequest, msg).WithCode(CodeInvalidPayload)
}

// ErrValidationFailed creates a 400 Bad Request error for field validation failures.
func ErrValidationFailed(message ...string) *APIError {
	msg := "Validation failed"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return New(http.StatusBadRequest, msg).WithCode(CodeValidationFailed)
}
