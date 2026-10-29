// Package database provides transaction runners and scoped execution helpers
// to ensure ACID transactional boundaries and resource cleanup.
package database

import (
	"context"
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
// WithTx hands out *Tx rather than QueryExecutor so that write paths - services
// that participate in a unit of work and repository methods that mutate - can
// require a transaction by type. Passing the pool where a *Tx is expected does
// not compile.
type Runner interface {
	// WithTx runs the provided callback inside a database transaction.
	// It automatically commits on success or rolls back on error/panic.
	WithTx(ctx context.Context, fn func(ctx context.Context, tx *Tx) error) error

	// WithExecutor runs the provided callback without starting a transaction,
	// reusing the ambient one when the context already carries it.
	WithExecutor(ctx context.Context, fn func(exec QueryExecutor) error) error
}

// Executor implements the Runner interface by managing transactions and pool access
// against a target DB instance.
type Executor struct {
	// db points to the active database connection pool.
	db *DB
}

// txContextKey keys the in-flight transaction stored on a request context.
type txContextKey struct{}

// TxFromContext returns the transaction bound to ctx by an enclosing WithTx.
//
// Code that runs inside a transaction can use this to stay on the caller's
// connection; callers that just issue queries can call WithExecutor, which
// performs the same lookup for them.
func TxFromContext(ctx context.Context) (*Tx, bool) {
	tx, ok := ctx.Value(txContextKey{}).(*Tx)
	return tx, ok
}

// NewTxRunner constructs a new transaction runner backed by the provided DB.
func NewTxRunner(db *DB) *Executor {
	return &Executor{db: db}
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
// rolls back on its own - the outermost transaction decides, so every error
// that invalidates the unit of work must be propagated to the outermost
// callback.
func (t *Executor) WithTx(ctx context.Context, fn func(ctx context.Context, tx *Tx) error) error {
	if t == nil || t.db == nil {
		return ErrConnectionFailure
	}

	// Join the transaction already in flight for this context, if any. The
	// ambient transaction owns the connection, so the pool is not consulted.
	if currentTx, ok := TxFromContext(ctx); ok {
		return fn(ctx, currentTx)
	}

	if t.db.Pool == nil {
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

	// Publish the transaction on the context so that every WithExecutor call made
	// below - nested service calls included - reuses this connection instead of
	// acquiring another one from the pool.
	wrappedTx := &Tx{tx}
	boundCtx := context.WithValue(ctx, txContextKey{}, wrappedTx)

	if err := fn(boundCtx, wrappedTx); err != nil {
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
// If ctx already carries a transaction (that is, the caller is inside WithTx),
// the callback runs on that transaction instead of the pool. Checking a second
// connection out while the first is still held is what starves the pool and
// stalls every other request.
func (t *Executor) WithExecutor(ctx context.Context, fn func(exec QueryExecutor) error) error {
	if t == nil || t.db == nil {
		return ErrConnectionFailure
	}

	// Reuse the ambient transaction before touching the pool: it owns the
	// connection, so a runner without a usable pool can still join.
	if currentTx, ok := TxFromContext(ctx); ok {
		return fn(currentTx)
	}

	if t.db.Pool == nil {
		return ErrConnectionFailure
	}

	return fn(t.db)
}
