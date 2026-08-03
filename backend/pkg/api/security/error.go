package security

import (
	"log/slog"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
)

const (
	CodeRateLimit        = "LIMIT_EXCEEDED"
	CodeValidation       = "VALIDATION_ERROR"
	CodeInternal         = "INTERNAL_ERROR"
	CodeConflict         = "CONFLICT_ERROR"
	CodeAuth             = "AUTH_ERROR"
	CodeExpired          = "EXPIRED"
	CodeInvalidOrExpired = "INVALID_OR_EXPIRED"
	CodeNotFound         = "NOT_FOUND"
	CodeUnauthorized     = "UNAUTHORIZED"
)

func stack() string {
	pc := make([]uintptr, 50)
	n := runtime.Callers(3, pc)

	frames := runtime.CallersFrames(pc[:n])

	var b strings.Builder

	for {
		frame, more := frames.Next()
		b.WriteString(frame.File)
		b.WriteString(":")
		b.WriteString(strconv.Itoa(frame.Line))
		b.WriteString("\n")

		if !more {
			break
		}
	}

	return b.String()
}

type SecureError struct {
	// http status code ex(500)
	Status int `json:"status,omitempty"`

	// http error code ex("INTERNAL_ERROR")
	Code string `json:"code,omitempty"`

	// public user message
	Message string `json:"message,omitempty"`

	// internal error
	Internal error `json:"cause,omitempty"`

	// validation error details
	Fields []api.FieldError `json:"details,omitempty"`

	// error stack
	Stack string `json:"stack,omitempty"`
}

func NewSecureError(
	status int,
) *SecureError {
	return &SecureError{
		Status: status,
	}
}

func (se *SecureError) WithCode(code string) *SecureError {
	se.Code = code
	return se
}

func (se *SecureError) WithMessage(m string) *SecureError {
	se.Message = m
	return se
}

func (se *SecureError) Wrap(err error) *SecureError {
	se.Internal = err
	return se
}

func (se *SecureError) Error() string {
	return se.Message
}

func (se *SecureError) Unwrap() error {
	return se.Internal
}

func (se *SecureError) WithFields(
	fields ...api.FieldError,
) *SecureError {
	se.Fields = append(se.Fields, fields...)
	return se
}

func (se *SecureError) WithStack() *SecureError {
	se.Stack = stack()
	return se
}

func (e *SecureError) LogValue() slog.Value {
	attrs := make([]slog.Attr, 0, 6)

	if e.Status != 0 {
		attrs = append(attrs, slog.Int("status", e.Status))
	}

	if e.Code != "" {
		attrs = append(attrs, slog.String("code", e.Code))
	}

	if e.Message != "" {
		attrs = append(attrs, slog.String("message", e.Message))
	}

	if e.Internal != nil {
		attrs = append(attrs, slog.String("internal", e.Internal.Error()))
	}

	if len(e.Fields) > 0 {
		attrs = append(attrs, slog.Any("fields", e.Fields))
	}

	if e.Stack != "" {
		attrs = append(attrs, slog.String("stack", e.Stack))
	}

	return slog.GroupValue(attrs...)
}

func ErrInvalidUUID() *SecureError {
	return NewSecureError(
		http.StatusBadRequest,
	)
}

func ErrInternalError() *SecureError {
	return NewSecureError(
		http.StatusInternalServerError,
	)
}
