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
	return context.WithValue(context.Background(), txContextKey{}, tx)
}

func TestWithExecutor_ReturnsConnectionFailureWhenPoolMissing(t *testing.T) {
	called := false

	err := (&Executor{db: &DB{}}).WithExecutor(context.Background(), func(QueryExecutor) error {
		called = true
		return nil
	})

	require.ErrorIs(t, err, ErrConnectionFailure)
	assert.False(t, called, "callback must not run against a DB with a nil pool")
}

func TestWithExecutor_RejectsNilExecutor(t *testing.T) {
	var runner *Executor

	err := runner.WithExecutor(context.Background(), func(QueryExecutor) error { return nil })

	require.ErrorIs(t, err, ErrConnectionFailure)
}

func TestWithExecutor_RejectsNilDB(t *testing.T) {
	err := (&Executor{}).WithExecutor(context.Background(), func(QueryExecutor) error { return nil })

	require.ErrorIs(t, err, ErrConnectionFailure)
}

func TestWithTx_RejectsNilExecutor(t *testing.T) {
	var runner *Executor

	err := runner.WithTx(context.Background(), func(context.Context, *Tx) error { return nil })

	require.ErrorIs(t, err, ErrConnectionFailure)
}

func TestWithExecutor_JoinsAmbientTransaction(t *testing.T) {
	ambient := &Tx{}
	ctx := contextWithTx(ambient)

	// Joining must not require the runner's own pool: the ambient transaction
	// already owns the connection.
	err := (&Executor{db: &DB{}}).WithExecutor(ctx, func(exec QueryExecutor) error {
		require.Same(t, ambient, exec, "callback must receive the ambient transaction")
		return nil
	})

	require.NoError(t, err)
}

func TestWithTx_JoinsAmbientTransaction(t *testing.T) {
	ambient := &Tx{}
	ctx := contextWithTx(ambient)
	joined := false

	err := (&Executor{db: &DB{}}).WithTx(ctx, func(_ context.Context, tx *Tx) error {
		joined = true
		require.Same(t, ambient, tx, "inner call must reuse the ambient transaction")
		return nil
	})

	require.NoError(t, err)
	assert.True(t, joined, "callback must run when joining an ambient transaction")
}

func TestWithExecutor_JoinReturnsCallbackErrorUnchanged(t *testing.T) {
	sentinel := errors.New("business rule violated")
	ctx := contextWithTx(&Tx{})

	err := (&Executor{db: &DB{}}).WithExecutor(ctx, func(QueryExecutor) error {
		return sentinel
	})

	require.ErrorIs(t, err, sentinel)
}
