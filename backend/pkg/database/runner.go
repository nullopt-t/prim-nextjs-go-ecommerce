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
	WithTx(ctx context.Context, fn func(ctx context.Context, tx QueryExecutor) error) error

	// WithDB runs the provided callback directly against the connection pool.
	WithDB(ctx context.Context, fn func(db QueryExecutor) error) error
}

// Executer implements the Runner interface by managing transactions and pool access
// against a target DB instance.
type Executer struct {
	// db points to the active database connection pool.
	db *DB
}

// txContextKey keys the in-flight transaction stored on a request context.
type txContextKey struct{}

// TxFromContext returns the transaction bound to ctx by an enclosing WithTx.
//
// Code that runs inside a transaction can use this to stay on the caller's
// connection; callers that just issue queries can call WithDB, which performs
// the same lookup for them.
func TxFromContext(ctx context.Context) (QueryExecutor, bool) {
	tx, ok := ctx.Value(txContextKey{}).(QueryExecutor)
	return tx, ok
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
//
// If a transaction is already in flight on ctx, the callback joins it instead
// of beginning another: a nested Begin would check out a second pooled
// connection while the outer transaction still holds the first, and concurrent
// requests doing that exhaust the pool. A joined callback neither commits nor
// rolls back on its own - the outermost transaction decides.
func (t *Executer) WithTx(ctx context.Context, fn func(ctx context.Context, tx QueryExecutor) error) error {
	if t == nil || t.db == nil || t.db.Pool == nil {
		return ErrConnectionFailure
	}

	// Join the transaction already in flight for this context, if any.
	if currentTx, ok := TxFromContext(ctx); ok {
		return fn(ctx, currentTx)
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

	// Publish the transaction on the context so that every WithDB call made
	// below - nested service calls included - reuses this connection instead of
	// acquiring another one from the pool.
	wrappedTx := &Tx{tx}
	boundCtx := context.WithValue(ctx, txContextKey{}, QueryExecutor(wrappedTx))

	if err := fn(boundCtx, wrappedTx); err != nil {
		return MapError(err)
	}

	// Commit the transaction.
	if err := tx.Commit(ctx); err != nil {
		return MapError(err)
	}
	return nil
}

// WithDB executes the provided callback against the connection pool without
// starting a transaction.
//
// If ctx already carries a transaction (that is, the caller is inside WithTx),
// the callback runs on that transaction instead of the pool. Checking a second
// connection out while the first is still held is what starves the pool and
// stalls every other request.
func (t *Executer) WithDB(ctx context.Context, fn func(pool QueryExecutor) error) error {
	if t == nil || t.db == nil {
		return ErrConnectionFailure
	}

	if currentTx, ok := TxFromContext(ctx); ok {
		return fn(currentTx)
	}

	return fn(t.db)
}
