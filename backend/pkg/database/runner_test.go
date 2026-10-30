package database

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeExecutor is an inert QueryExecutor used to stand in for an ambient
// transaction stored on a context.
type fakeExecutor struct {
	name string
}

func (f *fakeExecutor) Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (f *fakeExecutor) Query(context.Context, string, ...interface{}) (pgx.Rows, error) {
	return nil, nil
}

func (f *fakeExecutor) QueryRow(context.Context, string, ...interface{}) pgx.Row {
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

	runner2 := NewTxRunner(db2)
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
