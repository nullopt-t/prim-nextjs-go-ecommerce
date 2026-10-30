// Package database provides transaction runners and scoped execution helpers
// to ensure ACID transactional boundaries and resource cleanup.
package database

import (
	"context"
	"fmt"
	"time"
)

// Runner defines the contract for executing units of work.
//
// The two methods differ in whether they *ensure* a transaction exists:
//
//   - WithTx ensures a transaction: it begins one when none is in flight, and
//     joins the ambient transaction when the context already carries one.
//   - WithExecutor never begins a transaction: it hands back the ambient
//     transaction if present, otherwise the connection pool.
//
// Both methods only ever hand out a single executor. Implementations are
// expected to be backed by one database pool; a transaction created by one
// runner must not be shared with a runner for a different pool.
//
// WithTx hands out the transaction via ctx so that repository methods that
// need a *Tx can call RequireTx(ctx) to obtain it at the execution boundary.
type Runner interface {
	// WithTx runs the provided callback inside a database transaction.
	// It automatically commits on success or rolls back on error/panic.
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error

	// WithExecutor runs the provided callback without starting a transaction,
	// reusing the ambient one when the context already carries it, or binding
	// the database pool to the context.
	WithExecutor(ctx context.Context, fn func(ctx context.Context) error) error
}

// crunner implements the Runner interface by managing transactions and pool access
// against a target DB instance.
type crunner struct {
	// db points to the active database connection pool.
	db *DB
}

// executorKey keys the active QueryExecutor (pool or *Tx) stored on a context.
type executorKey struct{}

// WithExecutor binds the provided QueryExecutor to ctx using the single executor key
// if no executor is already present in ctx. If an executor already exists on ctx,
// ctx is returned unchanged.
func WithExecutor(ctx context.Context, executor QueryExecutor) context.Context {
	if _, err := Executor(ctx); err == nil {
		return ctx
	}
	return context.WithValue(ctx, executorKey{}, executor)
}

// WithTx binds the active transaction to ctx using the single executor key,
// replacing any non-transaction executor previously bound.
func WithTx(ctx context.Context, tx *Tx) context.Context {
	if _, ok := TxFromContext(ctx); ok {
		return ctx
	}
	return context.WithValue(ctx, executorKey{}, tx)
}

// Executor retrieves the active QueryExecutor from ctx.
// It returns ErrNoExecutor if no executor is found in ctx.
func Executor(ctx context.Context) (QueryExecutor, error) {
	exec, ok := ctx.Value(executorKey{}).(QueryExecutor)
	if !ok || exec == nil {
		return nil, ErrNoExecutor
	}
	return exec, nil
}

// RequireTx retrieves the active *Tx from ctx.
// It rejects missing executors, pool executors, and non-transaction types.
func RequireTx(ctx context.Context) (*Tx, error) {
	exec, err := Executor(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNoActiveTx, err)
	}
	tx, ok := exec.(*Tx)
	if !ok || tx == nil {
		return nil, ErrNoActiveTx
	}
	return tx, nil
}

// TxFromContext returns the transaction bound to ctx by an enclosing WithTx, if present.
func TxFromContext(ctx context.Context) (*Tx, bool) {
	tx, err := RequireTx(ctx)
	return tx, err == nil
}

// ExecutorFromContext returns the active executor from ctx, or fallback if absent.
func ExecutorFromContext(ctx context.Context, fallback QueryExecutor) QueryExecutor {
	if exec, err := Executor(ctx); err == nil {
		return exec
	}
	return fallback
}

// NewTxRunner constructs a new transaction runner backed by the provided DB.
func NewTxRunner(db *DB) *crunner {
	return &crunner{db: db}
}

// WithTx begins a new PostgreSQL transaction, passes the wrapped transaction on ctx to the
// callback fn, and manages commit or rollback.
//
// Concurrency and Transaction Safety:
// If the callback returns an error or panics, the transaction is rolled back.
// If the callback completes successfully, the transaction is committed.
//
// If a transaction is already in flight on ctx, the callback joins it instead
// of beginning another: a nested Begin would check out a second pooled
// connection while the outer transaction still holds the first, and concurrent
// requests doing that exhaust the pool. A joined callback neither commits nor
// rolls back on its own - the outermost transaction decides, so every error
// that invalidates the unit of work must be propagated to the outermost
// callback.
func (t *crunner) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if t == nil || t.db == nil {
		return ErrConnectionFailure
	}

	// Join the transaction already in flight for this context, if any. The
	// ambient transaction owns the connection, so the pool is not consulted.
	if currentTx, ok := TxFromContext(ctx); ok {
		// Verify database instance ownership
		if currentTx.db != nil && t.db != nil && currentTx.db != t.db {
			return ErrMismatchedDatabase
		}
		return fn(ctx)
	}

	if t.db.Pool == nil {
		return ErrConnectionFailure
	}

	tx, err := t.db.Pool.Begin(ctx)
	if err != nil {
		return MapError(err)
	}

	// Defer rollback using a bounded context to ensure rollback is attempted
	// even if the request context was canceled or timed out.
	// If tx.Commit has already succeeded, tx.Rollback is a harmless no-op returning pgx.ErrTxClosed.
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tx.Rollback(cleanupCtx)
	}()

	defer func() {
		if p := recover(); p != nil {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = tx.Rollback(cleanupCtx)
			panic(p) // re-panic after ensuring rollback completed
		}
	}()

	// Publish the transaction on the context using the unified executorKey so that
	// downstream services and repositories reuse this connection instead of
	// acquiring another one from the pool.
	wrappedTx := &Tx{tx: tx, db: t.db}
	boundCtx := WithTx(ctx, wrappedTx)

	if err := fn(boundCtx); err != nil {
		return MapError(err)
	}

	// Commit the transaction.
	if err := tx.Commit(ctx); err != nil {
		return MapError(err)
	}
	return nil
}

// WithExecutor executes the provided callback without starting a transaction.
//
// If ctx already carries an active executor (ambient transaction or pool),
// the callback runs with that context. Otherwise, it binds the runner's DB
// to the context.
func (t *crunner) WithExecutor(ctx context.Context, fn func(ctx context.Context) error) error {
	if t == nil || t.db == nil {
		return ErrConnectionFailure
	}

	// Reuse the ambient executor (tx or pool) before touching the pool directly.
	if currentTx, ok := TxFromContext(ctx); ok {
		if currentTx.db != nil && t.db != nil && currentTx.db != t.db {
			return ErrMismatchedDatabase
		}
		return fn(ctx)
	}
	if _, err := Executor(ctx); err == nil {
		return fn(ctx)
	}

	if t.db.Pool == nil {
		return ErrConnectionFailure
	}

	return fn(WithExecutor(ctx, t.db))
}
