package api

type ErrorCode string

const (
	ErrInternal   ErrorCode = "internal_error"
	ErrValidation ErrorCode = "validation_error"
	ErrNotFound   ErrorCode = "not_found"
)
