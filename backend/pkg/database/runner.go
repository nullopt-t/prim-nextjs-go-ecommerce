// Package database provides transaction runners and scoped execution helpers
// to ensure ACID transactional boundaries and resource cleanup.
package database

import (
	"context"
)

// Runner defines the contract for executing units of work either within
// a managed database transaction (WithTx) or against the connection pool directly (WithDB).
type Runner interface {
	// WithTx runs the provided callback inside a database transaction.
	// It automatically commits on success or rolls back on error/panic.
	WithTx(ctx context.Context, fn func(tx QueryExecutor) error) error

	// WithDB runs the provided callback directly against the connection pool.
	WithDB(ctx context.Context, fn func(db QueryExecutor) error) error
}

// Executer implements the Runner interface by managing transactions and pool access
// against a target DB instance.
type Executer struct {
	// db points to the active database connection pool.
	db *DB
}

// NewTxRunner constructs a new transaction runner backed by the provided DB.
func NewTxRunner(db *DB) *Executer {
	return &Executer{db: db}
}

// WithTx begins a new PostgreSQL transaction, passes the wrapped transaction to the
// callback fn, and manages commit or rollback.
//
// Concurrency and Transaction Safety:
// If the callback returns an error or panics, the transaction is rolled back.
// If the callback completes successfully, the transaction is committed.
func (t *Executer) WithTx(ctx context.Context, fn func(tx QueryExecutor) error) error {
	if t == nil || t.db == nil || t.db.Pool == nil {
		return ErrConnectionFailure
	}

	tx, err := t.db.Pool.Begin(ctx)
	if err != nil {
		return MapError(err)
	}

	// Defer rollback using context.Background() to ensure the transaction is aborted
	// even if the request context was cancelled or timed out.
	// If tx.Commit has already succeeded, tx.Rollback is a harmless no-op returning pgx.ErrTxClosed.
	defer func() {
		_ = tx.Rollback(context.Background())
	}()

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(context.Background())
			panic(p) // re-panic after ensuring rollback completed
		}
	}()

	if err := fn(&Tx{tx}); err != nil {
		return MapError(err)
	}

	// Commit the transaction.
	if err := tx.Commit(ctx); err != nil {
		return MapError(err)
	}
	return nil
}

// WithDB executes the provided callback against the connection pool without starting a transaction.
func (t *Executer) WithDB(ctx context.Context, fn func(pool QueryExecutor) error) error {
	if t == nil || t.db == nil {
		return ErrConnectionFailure
	}
	return fn(t.db)
}
