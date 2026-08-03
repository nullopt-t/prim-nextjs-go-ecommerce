package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	// --- Integrity & Constraint Violations ---
	ErrNotFound            = errors.New("record not found")
	ErrConflict            = errors.New("resource already exists")       // 23505: UNIQUE violation
	ErrForeignKeyViolation = errors.New("foreign key reference invalid") // 23503: FK violation
	ErrCheckViolation      = errors.New("check constraint failed")       // 23514: CHECK violation (e.g., quantity > 0)
	ErrNotNullViolation    = errors.New("required field missing")        // 23502: NOT NULL violation

	// --- Data & Parsing Exceptions ---
	ErrInvalidInput    = errors.New("invalid data representation")        // 22P02: Bad UUID, invalid enum text, etc.
	ErrValueOutOfRange = errors.New("numeric value out of range")         // 22003: Integer/bigint overflow
	ErrValueTooLong    = errors.New("string length exceeds column limit") // 22001: String truncation

	// --- Concurrency, Locks & Transactions ---
	ErrTransactionFailure = errors.New("transaction conflict or deadlock") // 40001, 40P01: Serialization / Deadlock
	ErrLockNotAvailable   = errors.New("row lock could not be acquired")   // 55P03: FOR UPDATE NOWAIT contention

	// --- Infrastructure & System Exceptions ---
	ErrTimeout           = errors.New("database operation timed out")    // 57014, context.DeadlineExceeded
	ErrConnectionFailure = errors.New("database connection unavailable") // 08xxx class, 53300
	ErrInternal          = errors.New("internal database error")
)

func MapError(err error) error {
	if err == nil {
		return nil
	}

	// 1. Standard Go Context & Driver Errors
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return ErrTimeout
	}

	// 2. PostgreSQL Server Error Codes (SQLState)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {

		// --- Class 23: Integrity Constraint Violation ---
		case "23505":
			return ErrConflict
		case "23503":
			return ErrForeignKeyViolation
		case "23514":
			return ErrCheckViolation
		case "23502":
			return ErrNotNullViolation

		// --- Class 22: Data Exception ---
		case "22P02": // invalid_text_representation (bad UUID or text to enum/int conversion)
			return ErrInvalidInput
		case "22003": // numeric_value_out_of_range
			return ErrValueOutOfRange
		case "22001": // string_data_right_truncation
			return ErrValueTooLong

		// --- Class 40: Transaction Rollback ---
		case "40001": // serialization_failure
		case "40P01": // deadlock_detected
			return ErrTransactionFailure

		// --- Class 55: Object Not In Prerequisite State ---
		case "55P03": // lock_not_available (occurs during SELECT ... FOR UPDATE NOWAIT under high load)
			return ErrLockNotAvailable

		// --- Class 08: Connection Exception & Class 53: Insufficient Resources ---
		case "08000", "08003", "08006", "08001", "08004":
			return ErrConnectionFailure
		case "53300": // too_many_connections
			return ErrConnectionFailure

		// --- Class 57: Operator Intervention ---
		case "57014": // query_canceled (e.g., statement_timeout)
			return ErrTimeout
		}
	}

	return err
}
