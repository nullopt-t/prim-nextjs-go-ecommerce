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

type fakeInventoryReader struct {
	orderIDs []uuid.UUID
}

func (f *fakeInventoryReader) ExpiredReservationOrderIDs(ctx context.Context, qe database.QueryExecutor) ([]uuid.UUID, error) {
	return f.orderIDs, nil
}

// paymentStatusExecutor returns a caller-controlled command tag so tests can
// simulate both an updated row and a missing payment (UPDATE 0).
type paymentStatusExecutor struct{ tag pgconn.CommandTag }

func (e paymentStatusExecutor) Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error) {
	return e.tag, nil
}

func (e paymentStatusExecutor) Query(context.Context, string, ...interface{}) (pgx.Rows, error) {
	return nil, nil
}

// paymentStatusRow is a minimal pgx.Row that satisfies the SELECTs made before
// the payment UPDATE (currently: resolving the owning order id).
type paymentStatusRow struct{ orderID uuid.UUID }

func (r paymentStatusRow) Scan(dest ...any) error {
	if len(dest) > 0 {
		if p, ok := dest[0].(*uuid.UUID); ok {
			*p = r.orderID
		}
	}
	return nil
}

func (e paymentStatusExecutor) QueryRow(context.Context, string, ...interface{}) pgx.Row {
	return paymentStatusRow{orderID: uuid.New()}
}

func (e paymentStatusExecutor) Begin(context.Context) (pgx.Tx, error) { return e, nil }
func (e paymentStatusExecutor) Commit(context.Context) error          { return nil }
func (e paymentStatusExecutor) Rollback(context.Context) error        { return nil }
func (e paymentStatusExecutor) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (e paymentStatusExecutor) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults { return nil }
func (e paymentStatusExecutor) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (e paymentStatusExecutor) Conn() *pgx.Conn                { return nil }
func (e paymentStatusExecutor) LargeObjects() pgx.LargeObjects { return pgx.LargeObjects{} }

type paymentStatusRunner struct{ tag pgconn.CommandTag }

func (r paymentStatusRunner) WithTx(_ context.Context, fn func(context.Context, *database.Tx) error) error {
	return fn(context.Background(), database.NewTx(paymentStatusExecutor(r)))
}

func (r paymentStatusRunner) WithExecutor(_ context.Context, fn func(database.QueryExecutor) error) error {
	return fn(paymentStatusExecutor(r))
}

func TestUpdatePaymentStatusValidation(t *testing.T) {
	updated := pgconn.NewCommandTag("UPDATE 1")
	missing := pgconn.NewCommandTag("UPDATE 0")

	cases := []struct {
		name       string
		tag        pgconn.CommandTag
		status     string
		wantStatus int
		wantCode   string
		wantErr    bool
	}{
		{
			name:   "valid status updates the row",
			tag:    updated,
			status: "captured",
		},
		{
			name:       "unknown status is a bad request",
			tag:        updated,
			status:     "bogus",
			wantStatus: 400,
			wantCode:   apierr.CodeInvalidInput,
			wantErr:    true,
		},
		{
			name:       "empty status is a bad request",
			tag:        updated,
			status:     "",
			wantStatus: 400,
			wantCode:   apierr.CodeInvalidInput,
			wantErr:    true,
		},
		{
			name:       "missing payment is not found",
			tag:        missing,
			status:     "captured",
			wantStatus: 404,
			wantCode:   apierr.CodeNotFound,
			wantErr:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			writer := &fakeInventoryWriter{}
			svc := order.NewService(paymentStatusRunner{tag: tc.tag}, order.NewRepository(), log.NewNoopLogger(), &fakeInventoryReader{}, writer, &fakeUserReader{}, &fakeSettingsProvider{}, &fakeOrderNotifier{})

			err := svc.UpdatePaymentStatus(context.Background(), uuid.New(), tc.status)
			if !tc.wantErr {
				require.NoError(t, err)
				// Capturing a payment is a sale: the holds must be committed
				// through the inventory writer, on the order's transaction.
				assert.Equal(t, 1, writer.commitCalls)
				return
			}

			require.Error(t, err)
			var apiErr *apierr.APIError
			require.True(t, errors.As(err, &apiErr))
			assert.Equal(t, tc.wantStatus, apiErr.Status)
			assert.Equal(t, tc.wantCode, apiErr.Code)
			assert.Zero(t, writer.commitCalls, "a rejected update must not touch stock")
		})
	}
}
