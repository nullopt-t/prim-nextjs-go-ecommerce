package database

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotFound            = errors.New("not found")
	ErrConflict            = errors.New("conflict")
	ErrForeignKeyViolation = errors.New("foreign key violation")

	ErrInvalidInput = errors.New("invalid input")

	ErrTimeout            = errors.New("timeout")
	ErrConnectionFailure  = errors.New("connection failure")
	ErrTransactionFailure = errors.New("transaction failure")

	ErrInternal = errors.New("internal database error")
)

func MapError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {

		// Unique violation
		case "23505":
			return ErrConflict

		// Foreign key violation
		case "23503":
			return ErrForeignKeyViolation

		// Invalid UUID, invalid type, etc.
		case "22P02":
			return ErrInvalidInput

		// Deadlock / serialization
		case "40001", "40P01":
			return ErrTransactionFailure

		// Connection issues
		case "08000", "08003", "08006":
			return ErrConnectionFailure

		// Timeout
		case "57014":
			return ErrTimeout
		}
	}
	return err
}
