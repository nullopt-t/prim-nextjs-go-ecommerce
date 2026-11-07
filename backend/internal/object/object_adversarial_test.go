package object

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdversarial_Object_HandleError(t *testing.T) {
	svc := &ObjectService{}
	alreadyWrapped := apierr.BadRequestError("already wrapped").WithCode("CUSTOM_CODE")

	tests := []struct {
		name         string
		err          error
		fallbackMsg  string
		wantStatus   int
		wantCode     string
		wantExactErr error
	}{
		{
			name:        "1. ErrNoRows -> 404",
			err:         pgx.ErrNoRows,
			fallbackMsg: "object",
			wantStatus:  http.StatusNotFound,
			wantCode:    apierr.CodeNotFound,
		},
		{
			name:        "2. Unique constraint violation -> 409",
			err:         &pgconn.PgError{Code: "23505"},
			fallbackMsg: "object",
			wantStatus:  http.StatusConflict,
			wantCode:    apierr.CodeAlreadyExists,
		},
		{
			name:        "3. Foreign key violation -> 400",
			err:         &pgconn.PgError{Code: "23503"},
			fallbackMsg: "object",
			wantStatus:  http.StatusBadRequest,
			wantCode:    apierr.CodeInvalidInput,
		},
		{
			name:        "4a. Check constraint violation -> 400",
			err:         &pgconn.PgError{Code: "23514"},
			fallbackMsg: "object",
			wantStatus:  http.StatusBadRequest,
			wantCode:    apierr.CodeInvalidInput,
		},
		{
			name:        "4b. Not null violation -> 400",
			err:         &pgconn.PgError{Code: "23502"},
			fallbackMsg: "object",
			wantStatus:  http.StatusBadRequest,
			wantCode:    apierr.CodeInvalidInput,
		},
		{
			name:        "5a. Timeout (context.DeadlineExceeded) -> 504 GatewayTimeout",
			err:         context.DeadlineExceeded,
			fallbackMsg: "object",
			wantStatus:  http.StatusGatewayTimeout,
			wantCode:    apierr.CodeGatewayTimeout,
		},
		{
			name:        "5b. Timeout (database.ErrTimeout) -> 504 GatewayTimeout",
			err:         database.ErrTimeout,
			fallbackMsg: "object",
			wantStatus:  http.StatusGatewayTimeout,
			wantCode:    apierr.CodeGatewayTimeout,
		},
		{
			name:        "5c. Cancellation (context.Canceled) -> 504 GatewayTimeout",
			err:         context.Canceled,
			fallbackMsg: "object",
			wantStatus:  http.StatusGatewayTimeout,
			wantCode:    apierr.CodeGatewayTimeout,
		},
		{
			name:        "5d. Cancellation (database.ErrCanceled) -> 504 GatewayTimeout",
			err:         database.ErrCanceled,
			fallbackMsg: "object",
			wantStatus:  http.StatusGatewayTimeout,
			wantCode:    apierr.CodeGatewayTimeout,
		},
		{
			name:         "6. Already-wrapped apierr.APIError -> returned unchanged",
			err:          alreadyWrapped,
			fallbackMsg:  "object",
			wantExactErr: alreadyWrapped,
			wantStatus:   http.StatusBadRequest,
			wantCode:     "CUSTOM_CODE",
		},
		{
			name:        "7. Unclassified error -> 500 with fallback message",
			err:         errors.New("unclassified disk failure"),
			fallbackMsg: "Failed to persist object",
			wantStatus:  http.StatusInternalServerError,
			wantCode:    apierr.CodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := svc.handleError(tt.err, tt.fallbackMsg)
			require.Error(t, res)
			if tt.wantExactErr != nil {
				assert.Equal(t, tt.wantExactErr, res)
			}
			var apiErr *apierr.APIError
			require.True(t, errors.As(res, &apiErr))
			assert.Equal(t, tt.wantStatus, apiErr.Status)
			assert.Equal(t, tt.wantCode, apiErr.Code)
		})
	}
}

type mockObjQE struct{}

func (m *mockObjQE) ExecContext(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	return pgconn.NewCommandTag("UPDATE 1"), nil
}

func (m *mockObjQE) QueryContext(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return nil, nil
}

func (m *mockObjQE) QueryRowContext(ctx context.Context, sql string, args ...any) pgx.Row {
	return &mockObjRow{}
}

type mockObjRow struct{}

func (r *mockObjRow) Scan(dest ...any) error {
	now := time.Now()
	for _, d := range dest {
		if t, ok := d.(*time.Time); ok {
			*t = now
		}
	}
	return nil
}

type mockObjTx struct {
	*mockObjQE
}

func (m *mockObjTx) Begin(context.Context) (pgx.Tx, error) { return m, nil }
func (m *mockObjTx) Commit(context.Context) error          { return nil }
func (m *mockObjTx) Rollback(context.Context) error        { return nil }
func (m *mockObjTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return m.ExecContext(ctx, sql, args...)
}
func (m *mockObjTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return m.QueryContext(ctx, sql, args...)
}
func (m *mockObjTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return m.QueryRowContext(ctx, sql, args...)
}
func (m *mockObjTx) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (m *mockObjTx) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults { return nil }
func (m *mockObjTx) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (m *mockObjTx) Conn() *pgx.Conn                { return nil }
func (m *mockObjTx) LargeObjects() pgx.LargeObjects { return pgx.LargeObjects{} }

type mockObjectRunner struct {
	withDBCalls int
	withTxCalls int
	qe          *mockObjQE
}

func (m *mockObjectRunner) WithExecutor(ctx context.Context, fn func(ctx context.Context) error) error {
	m.withDBCalls++
	return fn(database.WithExecutor(ctx, m.qe))
}

func (m *mockObjectRunner) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	m.withTxCalls++
	return fn(database.WithTx(ctx, database.NewTx(&mockObjTx{mockObjQE: m.qe})))
}

func (m *mockObjectRunner) WithRetry(ctx context.Context, fn func(ctx context.Context) error, n int) error {
	return m.WithTx(ctx, fn)
}

type mockObjStorageForInvariants struct{}

func (m *mockObjStorageForInvariants) Upload(ctx context.Context, bucket, key string, file io.Reader, size int64, contentType string) error {
	return nil
}

func (m *mockObjStorageForInvariants) Delete(ctx context.Context, bucket, key string) error {
	return nil
}

func (m *mockObjStorageForInvariants) GetPresignedURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error) {
	return "", nil
}

func (m *mockObjStorageForInvariants) GetPublicURL(bucket, key string) string {
	return "https://cdn.example.com/" + bucket + "/" + key
}

func TestAdversarial_CreateObject_UsesWithTx(t *testing.T) {
	runner := &mockObjectRunner{qe: &mockObjQE{}}
	repo := NewRepository()
	svc := NewService(runner, repo, &mockObjStorageForInvariants{})

	obj, err := svc.CreateObject(context.Background(), "image/png", 100, "bkt", "key", model.ObjectStatusUploaded)
	require.NoError(t, err)
	require.NotNil(t, obj)

	assert.Equal(t, 0, runner.withDBCalls, "CreateObject must NOT use WithExecutor")
	assert.Equal(t, 1, runner.withTxCalls, "CreateObject must use WithTx")
}

func TestAdversarial_DeleteObject_UsesWithTx(t *testing.T) {
	runner := &mockObjectRunner{qe: &mockObjQE{}}
	repo := NewRepository()
	svc := NewService(runner, repo, &mockObjStorageForInvariants{})

	err := svc.DeleteObject(context.Background(), "bkt", "key")
	require.NoError(t, err)

	assert.Equal(t, 0, runner.withDBCalls, "DeleteObject must NOT use WithExecutor")
	assert.Equal(t, 2, runner.withTxCalls, "DeleteObject must use WithTx for both mark deleting and mark deleted")
}
