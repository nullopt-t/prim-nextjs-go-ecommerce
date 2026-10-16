// Package database provides error classification and translation mechanisms to map
// low-level PostgreSQL driver errors and context errors into domain-level sentinel errors.
package database

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	// --- Integrity & Constraint Violations (SQLSTATE Class 23) ---

	// ErrNotFound indicates that a requested record does not exist (pgx.ErrNoRows).
	ErrNotFound = errors.New("record not found")

	// ErrConflict indicates a UNIQUE constraint or EXCLUSION violation (SQLSTATE 23505).
	ErrConflict = errors.New("resource already exists")

	// ErrForeignKeyViolation indicates that an insert or update violates a foreign key constraint (SQLSTATE 23503).
	ErrForeignKeyViolation = errors.New("foreign key reference invalid")

	// ErrCheckViolation indicates a table CHECK constraint failed (SQLSTATE 23514).
	ErrCheckViolation = errors.New("check constraint failed")

	// ErrNotNullViolation indicates an attempt to insert NULL into a NOT NULL column (SQLSTATE 23502).
	ErrNotNullViolation = errors.New("required field missing")

	// --- Data & Parsing Exceptions (SQLSTATE Class 22) ---

	// ErrInvalidInput indicates malformed data representations such as invalid UUID syntax or enum values (SQLSTATE 22P02).
	ErrInvalidInput = errors.New("invalid data representation")

	// ErrValueOutOfRange indicates integer/bigint or numeric range overflows (SQLSTATE 22003).
	ErrValueOutOfRange = errors.New("numeric value out of range")

	// ErrValueTooLong indicates string data truncation where input exceeds column length limits (SQLSTATE 22001).
	ErrValueTooLong = errors.New("string length exceeds column limit")

	// --- Concurrency, Locks & Transactions (SQLSTATE Class 40 & 55) ---

	// ErrTransactionFailure indicates serialization failures (40001) or deadlocks (40P01) where transactions must be retried.
	ErrTransactionFailure = errors.New("transaction conflict or deadlock")

	// ErrLockNotAvailable indicates inability to acquire row lock under SELECT ... FOR UPDATE NOWAIT (SQLSTATE 55P03).
	ErrLockNotAvailable = errors.New("row lock could not be acquired")

	// --- Infrastructure & System Exceptions (SQLSTATE Class 08, 53, 57) ---

	// ErrTimeout indicates an operation timeout or query cancellation via statement_timeout (SQLSTATE 57014, context.DeadlineExceeded).
	ErrTimeout = errors.New("database operation timed out")

	// ErrCanceled indicates that an operation was canceled by caller or client disconnect (context.Canceled).
	ErrCanceled = errors.New("database operation canceled")

	// ErrConnectionFailure indicates dropped connections, network unreachable, or pool connection limits exceeded (SQLSTATE 08xxx, 53300).
	ErrConnectionFailure = errors.New("database connection unavailable")

	// --- Authorization & Resource Limits (SQLSTATE Class 28, 54) ---

	// ErrUnauthorized indicates invalid authorization specifications or credentials (SQLSTATE Class 28).
	ErrUnauthorized = errors.New("database authentication failed")

	// ErrLimitExceeded indicates program limit exceeded such as memory or string length (SQLSTATE Class 54).
	ErrLimitExceeded = errors.New("database resource limit exceeded")

	// ErrInternal is a fallback error indicating unclassified internal database failures.
	ErrInternal = errors.New("internal database error")
)

// wrapPgError wraps the domain sentinel error with detailed PostgreSQL server error metadata
// (message, table, column, constraint, detail) to preserve observability while keeping errors.Is matches working.
func wrapPgError(sentinel error, pgErr *pgconn.PgError) error {
	var details []string
	if pgErr.Detail != "" {
		details = append(details, "detail: "+pgErr.Detail)
	}
	if pgErr.TableName != "" {
		details = append(details, "table: "+pgErr.TableName)
	}
	if pgErr.ColumnName != "" {
		details = append(details, "column: "+pgErr.ColumnName)
	}
	if pgErr.ConstraintName != "" {
		details = append(details, "constraint: "+pgErr.ConstraintName)
	}

	if len(details) > 0 {
		return fmt.Errorf("%w: %s (%s)", sentinel, pgErr.Message, strings.Join(details, ", "))
	}
	if pgErr.Message != "" {
		return fmt.Errorf("%w: %s", sentinel, pgErr.Message)
	}
	return sentinel
}

// MapError inspects an error returned from pgx or standard Go libraries,
// translating low-level driver and driver-agnostic errors into domain-specific sentinel errors.
func MapError(err error) error {
	if err == nil {
		return nil
	}

	// 1. Standard Go Context & Driver Errors
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	if errors.Is(err, context.Canceled) {
		return fmt.Errorf("%w: %s", ErrCanceled, err.Error())
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%w: %s", ErrTimeout, err.Error())
	}

	// 2. PostgreSQL Server Error Codes (SQLState)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch {
		// --- Class 23: Integrity Constraint Violation ---
		case pgErr.Code == "23505": // unique_violation
			return wrapPgError(ErrConflict, pgErr)
		case pgErr.Code == "23503": // foreign_key_violation
			return wrapPgError(ErrForeignKeyViolation, pgErr)
		case pgErr.Code == "23514": // check_violation
			return wrapPgError(ErrCheckViolation, pgErr)
		case pgErr.Code == "23502": // not_null_violation
			return wrapPgError(ErrNotNullViolation, pgErr)

		// --- Class 22: Data Exception ---
		case pgErr.Code == "22P02": // invalid_text_representation (bad UUID or text to enum/int conversion)
			return wrapPgError(ErrInvalidInput, pgErr)
		case pgErr.Code == "22003": // numeric_value_out_of_range (integer/numeric overflow)
			return wrapPgError(ErrValueOutOfRange, pgErr)
		case pgErr.Code == "22001": // string_data_right_truncation
			return wrapPgError(ErrValueTooLong, pgErr)

		// --- Class 40: Transaction Rollback ---
		case pgErr.Code == "40001", pgErr.Code == "40P01": // serialization_failure, deadlock_detected
			return wrapPgError(ErrTransactionFailure, pgErr)

		// --- Class 55: Object Not In Prerequisite State ---
		case pgErr.Code == "55P03": // lock_not_available (SELECT ... FOR UPDATE NOWAIT)
			return wrapPgError(ErrLockNotAvailable, pgErr)

		// --- Class 08: Connection Exception & Class 53: Insufficient Resources ---
		case strings.HasPrefix(pgErr.Code, "08"), pgErr.Code == "53300":
			return wrapPgError(ErrConnectionFailure, pgErr)

		// --- Class 57: Operator Intervention ---
		case pgErr.Code == "57014": // query_canceled (e.g., statement_timeout)
			return wrapPgError(ErrTimeout, pgErr)

		// --- Class 28: Invalid Authorization Specification ---
		case strings.HasPrefix(pgErr.Code, "28"):
			return wrapPgError(ErrUnauthorized, pgErr)

		// --- Class 54: Program Limit Exceeded ---
		case strings.HasPrefix(pgErr.Code, "54"):
			return wrapPgError(ErrLimitExceeded, pgErr)
		}
	}

	return err
}
