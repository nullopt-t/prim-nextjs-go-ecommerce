package apierr

import (
	"log/slog"
	"net/http"

	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/nullopt-t/errs"
)

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

// ApiError represents a structured application API error embedding nullopt-t/errs.
type ApiError struct {
	*errs.AppError

	// HTTP status code (e.g., 400, 404, 500)
	Status int `json:"status,omitempty"`

	// Machine-readable error code (e.g., "NOT_FOUND")
	Code string `json:"code,omitempty"`

	// Detailed field validation errors
	Fields []api.FieldError `json:"details,omitempty"`
}

// New constructs a new ApiError requiring HTTP status code and message.
func New(status int, message string) *ApiError {
	return &ApiError{
		AppError: errs.New(message),
		Status:   status,
	}
}

func (ae *ApiError) WithCode(code string) *ApiError {
	ae.Code = code
	return ae
}

func (ae *ApiError) WithMessage(m string) *ApiError {
	ae.AppError = errs.New(m)
	return ae
}

func (ae *ApiError) Wrap(err error) *ApiError {
	ae.AppError.Wrap(err)
	return ae
}

func (ae *ApiError) WithFields(fields ...api.FieldError) *ApiError {
	ae.Fields = append(ae.Fields, fields...)
	return ae
}

func (ae *ApiError) WithStack() *ApiError {
	ae.AppError.WithStack()
	return ae
}

func (ae *ApiError) LogValue() slog.Value {
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

// Standard Helpers
func ErrInvalidUUID() *ApiError {
	return New(http.StatusBadRequest, "Invalid UUID format").WithCode(CodeInvalidInput)
}

func ErrInternalError() *ApiError {
	return New(http.StatusInternalServerError, "Internal server error").WithCode(CodeInternalError)
}

func ErrNotFound(message string) *ApiError {
	return New(http.StatusNotFound, message).WithCode(CodeNotFound)
}

func ErrBadRequest(message string) *ApiError {
	return New(http.StatusBadRequest, message).WithCode(CodeBadRequest)
}

func ErrUnauthorized(message string) *ApiError {
	return New(http.StatusUnauthorized, message).WithCode(CodeUnauthorized)
}
