package database

import (
	"context"
)

type Runner interface {
	WithTx(ctx context.Context, fn func(tx QueryExecutor) error) error
	WithDB(ctx context.Context, fn func(db QueryExecutor) error) error
}

type Executer struct {
	db *DB
}

func NewTxRunner(db *DB) *Executer {
	return &Executer{db: db}
}

func (t *Executer) WithTx(ctx context.Context, fn func(tx QueryExecutor) error) error {
	tx, err := t.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback(ctx) // safe rollback if commit never happens
	}()

	if err := fn(&Tx{tx}); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (t *Executer) WithDB(ctx context.Context, fn func(pool QueryExecutor) error) error {
	return fn(t.db)
}
