package order_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nullopt-t/prim-backend/internal/order"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
)

// refundRow implements pgx.Row for the payment status lookup.
type refundRow struct {
	status string
	err    error
}

func (r refundRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	*dest[0].(*string) = r.status
	return nil
}

type refundExecutor struct{ row pgx.Row }

func (e refundExecutor) Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.NewCommandTag("UPDATE 1"), nil
}

func (e refundExecutor) Query(context.Context, string, ...interface{}) (pgx.Rows, error) {
	return nil, nil
}

func (e refundExecutor) QueryRow(context.Context, string, ...interface{}) pgx.Row { return e.row }

func (e refundExecutor) Begin(context.Context) (pgx.Tx, error) { return e, nil }
func (e refundExecutor) Commit(context.Context) error { return nil }
func (e refundExecutor) Rollback(context.Context) error { return nil }
func (e refundExecutor) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) { return 0, nil }
func (e refundExecutor) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults { return nil }
func (e refundExecutor) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) { return nil, nil }
func (e refundExecutor) Conn() *pgx.Conn { return nil }
func (e refundExecutor) LargeObjects() pgx.LargeObjects { return pgx.LargeObjects{} }

type refundRunner struct{ row pgx.Row }

func (r refundRunner) WithTx(_ context.Context, fn func(context.Context, *database.Tx) error) error {
	return fn(context.Background(), database.NewTx(refundExecutor(r)))
}

func (r refundRunner) WithExecutor(_ context.Context, fn func(database.QueryExecutor) error) error {
	return fn(refundExecutor(r))
}

func TestRefundPaymentStateValidation(t *testing.T) {
	cases := []struct {
		name       string
		row        pgx.Row
		wantStatus int
		wantCode   string
		wantErr    bool
	}{
		{
			name:    "captured payment is refundable",
			row:     refundRow{status: "captured"},
			wantErr: false,
		},
		{
			name:       "already refunded is a conflict",
			row:        refundRow{status: "refunded"},
			wantStatus: 409,
			wantCode:   apierr.CodeResourceConflict,
			wantErr:    true,
		},
		{
			name:       "pending payment is a conflict",
			row:        refundRow{status: "pending"},
			wantStatus: 409,
			wantCode:   apierr.CodeResourceConflict,
			wantErr:    true,
		},
		{
			name:       "missing payment is not found",
			row:        refundRow{err: pgx.ErrNoRows},
			wantStatus: 404,
			wantCode:   apierr.CodeNotFound,
			wantErr:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := order.NewService(refundRunner{row: tc.row}, order.NewRepository(), log.NewNoopLogger(), &fakeInventoryReader{}, &fakeInventoryWriter{})

			err := svc.RefundPayment(context.Background(), uuid.New())
			if !tc.wantErr {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			var apiErr *apierr.APIError
			require.True(t, errors.As(err, &apiErr))
			assert.Equal(t, tc.wantStatus, apiErr.Status)
			assert.Equal(t, tc.wantCode, apiErr.Code)
		})
	}
}
