package database

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeExecutor is an inert QueryExecutor used to stand in for an ambient
// transaction stored on a context.
type fakeExecutor struct {
	name string
}

func (f *fakeExecutor) ExecContext(context.Context, string, ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (f *fakeExecutor) QueryContext(context.Context, string, ...interface{}) (pgx.Rows, error) {
	return nil, nil
}

func (f *fakeExecutor) QueryRowContext(context.Context, string, ...interface{}) pgx.Row {
	return nil
}

func contextWithTx(tx *Tx) context.Context {
	return WithTx(context.Background(), tx)
}

func TestWithExecutor_StoresPoolAndWithTxReplacesSameKey(t *testing.T) {
	pool := &fakeExecutor{name: "pool"}
	ctx := WithExecutor(context.Background(), pool)

	exec, err := Executor(ctx)
	require.NoError(t, err)
	require.Same(t, pool, exec)

	// RequireTx should reject the pool
	_, err = RequireTx(ctx)
	require.ErrorIs(t, err, ErrNoActiveTx)

	// WithTx replaces the executor under the same key
	tx := &Tx{}
	txCtx := WithTx(ctx, tx)

	gotExec, err := Executor(txCtx)
	require.NoError(t, err)
	require.Same(t, tx, gotExec)

	gotTx, err := RequireTx(txCtx)
	require.NoError(t, err)
	require.Same(t, tx, gotTx)
}

func TestExecutor_ReturnsErrorWhenMissing(t *testing.T) {
	_, err := Executor(context.Background())
	require.ErrorIs(t, err, ErrNoExecutor)
}

func TestWithTx_RejectsMismatchedDatabase(t *testing.T) {
	db1 := &DB{}
	db2 := &DB{}
	tx := NewTxWithDB(nil, db1)
	ctx := WithTx(context.Background(), tx)

	runner2 := NewTxRunner(db2, log.NewNoopLogger())
	err := runner2.WithTx(ctx, func(ctx context.Context) error {
		return nil
	})
	require.ErrorIs(t, err, ErrMismatchedDatabase)
}

func TestWithExecutor_ReturnsConnectionFailureWhenPoolMissing(t *testing.T) {
	called := false

	err := (&crunner{db: &DB{}}).WithExecutor(context.Background(), func(context.Context) error {
		called = true
		return nil
	})

	require.ErrorIs(t, err, ErrConnectionFailure)
	assert.False(t, called, "callback must not run against a DB with a nil pool")
}

func TestWithExecutor_RejectsNilExecutor(t *testing.T) {
	var runner *crunner

	err := runner.WithExecutor(context.Background(), func(context.Context) error { return nil })

	require.ErrorIs(t, err, ErrConnectionFailure)
}

func TestWithExecutor_RejectsNilDB(t *testing.T) {
	err := (&crunner{}).WithExecutor(context.Background(), func(context.Context) error { return nil })

	require.ErrorIs(t, err, ErrConnectionFailure)
}

func TestWithTx_RejectsNilExecutor(t *testing.T) {
	var runner *crunner

	err := runner.WithTx(context.Background(), func(context.Context) error { return nil })

	require.ErrorIs(t, err, ErrConnectionFailure)
}

func TestWithExecutor_JoinsAmbientTransaction(t *testing.T) {
	ambient := &Tx{}
	ctx := contextWithTx(ambient)

	// Joining must not require the runner's own pool: the ambient transaction
	// already owns the connection.
	err := (&crunner{db: &DB{}}).WithExecutor(ctx, func(ctx context.Context) error {
		exec, err := Executor(ctx)
		require.NoError(t, err)
		require.Same(t, ambient, exec, "callback must receive the ambient transaction")
		return nil
	})

	require.NoError(t, err)
}

func TestWithTx_JoinsAmbientTransaction(t *testing.T) {
	ambient := &Tx{}
	ctx := contextWithTx(ambient)
	joined := false

	err := (&crunner{db: &DB{}}).WithTx(ctx, func(ctx context.Context) error {
		joined = true
		got, ok := TxFromContext(ctx)
		require.True(t, ok, "tx must be present in context")
		require.Same(t, ambient, got, "inner call must reuse the ambient transaction")
		return nil
	})

	require.NoError(t, err)
	assert.True(t, joined, "callback must run when joining an ambient transaction")
}

func TestWithExecutor_JoinReturnsCallbackErrorUnchanged(t *testing.T) {
	sentinel := errors.New("business rule violated")
	ctx := contextWithTx(&Tx{})

	err := (&crunner{db: &DB{}}).WithExecutor(ctx, func(context.Context) error {
		return sentinel
	})

	require.ErrorIs(t, err, sentinel)
}

func TestRequireTx_ReturnsErrorWhenNoTx(t *testing.T) {
	_, err := RequireTx(context.Background())
	require.ErrorIs(t, err, ErrNoActiveTx)
}

func TestRequireTx_ReturnsTxWhenPresent(t *testing.T) {
	ambient := &Tx{}
	ctx := contextWithTx(ambient)
	got, err := RequireTx(ctx)
	require.NoError(t, err)
	require.Same(t, ambient, got)
}

func TestExecutorFromContext_ReturnsPoolWhenNoTx(t *testing.T) {
	pool := &fakeExecutor{name: "pool"}
	got := ExecutorFromContext(context.Background(), pool)
	require.Same(t, pool, got)
}

func TestExecutorFromContext_ReturnsTxWhenPresent(t *testing.T) {
	ambient := &Tx{}
	ctx := contextWithTx(ambient)
	pool := &fakeExecutor{name: "pool"}
	got := ExecutorFromContext(ctx, pool)
	require.Same(t, ambient, got)
}

func TestWithRetry_SuccessOnFirstAttempt(t *testing.T) {
	ambient := &Tx{}
	ctx := contextWithTx(ambient)
	runner := NewTxRunner(&DB{}, log.NewNoopLogger())

	err := runner.WithRetry(ctx, func(ctx context.Context) error {
		return nil
	}, 3)

	require.NoError(t, err)
}

func TestWithRetry_RetryableError(t *testing.T) {
	ambient := &Tx{}
	ctx := contextWithTx(ambient)
	runner := NewTxRunner(&DB{}, log.NewNoopLogger())

	callCount := 0
	err := runner.WithRetry(ctx, func(ctx context.Context) error {
		callCount++
		// Simulate a retryable error on first call, success on second
		if callCount == 1 {
			return &pgconn.PgError{Code: "40001"} // serialization_failure
		}
		return nil
	}, 3)

	require.NoError(t, err)
	require.Equal(t, 2, callCount, "should have retried once")
}

func TestWithRetry_NonRetryableError(t *testing.T) {
	ambient := &Tx{}
	ctx := contextWithTx(ambient)
	runner := NewTxRunner(&DB{}, log.NewNoopLogger())

	sentinel := errors.New("some other error")
	callCount := 0
	err := runner.WithRetry(ctx, func(ctx context.Context) error {
		callCount++
		return sentinel
	}, 3)

	require.ErrorIs(t, err, sentinel)
	require.Equal(t, 1, callCount, "should not retry on non-retryable error")
}

func TestWithRetry_ExhaustsRetries(t *testing.T) {
	ambient := &Tx{}
	ctx := contextWithTx(ambient)
	runner := NewTxRunner(&DB{}, log.NewNoopLogger())

	retryErr := &pgconn.PgError{Code: "40P01"} // deadlock_detected
	callCount := 0
	err := runner.WithRetry(ctx, func(ctx context.Context) error {
		callCount++
		return retryErr
	}, 3)

	require.ErrorIs(t, err, retryErr)
	require.Equal(t, 3, callCount, "should have used all 3 attempts")
}

func TestWithRetry_ContextCancellation(t *testing.T) {
	ambient := &Tx{}
	ctx, cancel := context.WithCancel(contextWithTx(ambient))
	runner := NewTxRunner(&DB{}, log.NewNoopLogger())

	callCount := 0
	err := runner.WithRetry(ctx, func(ctx context.Context) error {
		callCount++
		cancel() // cancel context during execution
		return &pgconn.PgError{Code: "40001"}
	}, 3)

	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, 1, callCount, "should have attempted once before cancellation")
}

func TestWithRetry_ZeroRetries(t *testing.T) {
	ambient := &Tx{}
	ctx := contextWithTx(ambient)
	runner := NewTxRunner(&DB{}, log.NewNoopLogger())

	sentinel := errors.New("test error")
	err := runner.WithRetry(ctx, func(ctx context.Context) error {
		return sentinel
	}, 0)

	require.ErrorIs(t, err, sentinel)
}

func TestWithRetry_SuccessAfterRetries(t *testing.T) {
	ambient := &Tx{}
	ctx := contextWithTx(ambient)
	runner := NewTxRunner(&DB{}, log.NewNoopLogger())

	callCount := 0
	err := runner.WithRetry(ctx, func(ctx context.Context) error {
		callCount++
		// Fail first two, succeed on third
		if callCount < 3 {
			return &pgconn.PgError{Code: "40001"}
		}
		return nil
	}, 5)

	require.NoError(t, err)
	require.Equal(t, 3, callCount, "should have succeeded on 3rd attempt")
}
