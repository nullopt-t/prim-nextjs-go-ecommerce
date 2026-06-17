package security

import (
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"runtime"
	"strconv"
	"strings"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
)

const (
	CodeValidation   = "VALIDATION_ERROR"
	CodeInternal     = "INTERNAL_ERROR"
	CodeConflict     = "CONFLICT_ERROR"
	CodeAuth         = "AUTH_ERROR"
	CodeNotFound     = "NOT_FOUND"
	CodeUnauthorized = "UNAUTHORIZED"
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
	Status   int
	Code     string
	UserMsg  string
	Internal error
	Fields   []api.FieldError
	Stack    string
	MetaData map[string]any
}

func (se *SecureError) Error() string {
	return se.UserMsg
}

func (se *SecureError) Unwrap() error {
	return se.Internal
}

func NewSecureError(
	status int,
	code,
	public string,
	internal error,
) *SecureError {
	return &SecureError{
		Status:   status,
		Code:     code,
		UserMsg:  public,
		Internal: internal,
		Stack:    stack(),
	}
}

func (se *SecureError) WithMetaData(
	meta map[string]any,
) *SecureError {
	se.MetaData = meta
	return se
}

func (se *SecureError) WithFields(
	fields []api.FieldError,
) *SecureError {
	se.Fields = fields
	return se
}
