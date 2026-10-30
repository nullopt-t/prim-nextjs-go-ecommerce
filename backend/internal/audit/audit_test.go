package audit

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

type mockRunner struct {
	withTxCalled bool
	withDBCalled bool
	withTxErr    error
	withDBErr    error
	qe           database.QueryExecutor
	tx           *database.Tx
}

func (m *mockRunner) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	m.withTxCalled = true
	if m.withTxErr != nil {
		return m.withTxErr
	}
	tx := m.tx
	if tx == nil {
		tx = database.NewTx(&fakeAuditTx{})
	}
	return fn(database.WithTx(ctx, tx))
}

func (m *mockRunner) WithExecutor(ctx context.Context, fn func(ctx context.Context) error) error {
	m.withDBCalled = true
	if m.withDBErr != nil {
		return m.withDBErr
	}
	qe := m.qe
	if qe == nil {
		qe = &fakeAuditExecutor{}
	}
	return fn(database.WithExecutor(ctx, qe))
}

type fakeAuditTx struct {
	fakeAuditExecutor
}

func (f *fakeAuditTx) Begin(context.Context) (pgx.Tx, error) { return f, nil }
func (f *fakeAuditTx) Commit(context.Context) error { return nil }
func (f *fakeAuditTx) Rollback(context.Context) error { return nil }
func (f *fakeAuditTx) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) { return 0, nil }
func (f *fakeAuditTx) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults { return nil }
func (f *fakeAuditTx) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) { return nil, nil }
func (f *fakeAuditTx) Conn() *pgx.Conn { return nil }
func (f *fakeAuditTx) LargeObjects() pgx.LargeObjects { return pgx.LargeObjects{} }

type fakeAuditExecutor struct {
	row pgx.Row
}

func (f *fakeAuditExecutor) Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (f *fakeAuditExecutor) Query(context.Context, string, ...interface{}) (pgx.Rows, error) {
	return nil, nil
}

func (f *fakeAuditExecutor) QueryRow(context.Context, string, ...interface{}) pgx.Row {
	if f.row != nil {
		return f.row
	}
	return fakeAuditRow{id: 42}
}

type fakeAuditRow struct {
	id  int64
	err error
}

func (r fakeAuditRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) > 0 {
		if p, ok := dest[0].(*int64); ok {
			*p = r.id
		}
	}
	return nil
}

func TestService_HandleError(t *testing.T) {
	svc := &Service{}
	existingAPIErr := apierr.BadRequestError("existing error").WithCode("CUSTOM_CODE")

	tests := []struct {
		name         string
		err          error
		fallbackMsg  string
		wantStatus   int
		wantCode     string
		wantNil      bool
		wantExactErr error
	}{
		{
			name:        "nil error returns nil",
			err:         nil,
			fallbackMsg: "fallback",
			wantNil:     true,
		},
		{
			name:         "existing APIError returned unchanged",
			err:          existingAPIErr,
			fallbackMsg:  "fallback",
			wantExactErr: existingAPIErr,
			wantStatus:   http.StatusBadRequest,
			wantCode:     "CUSTOM_CODE",
		},
		{
			name:        "pgx.ErrNoRows maps to 404 NOT_FOUND",
			err:         pgx.ErrNoRows,
			fallbackMsg: "failed to fetch",
			wantStatus:  http.StatusNotFound,
			wantCode:    apierr.CodeNotFound,
		},
		{
			name:        "database.ErrNotFound maps to 404 NOT_FOUND",
			err:         database.ErrNotFound,
			fallbackMsg: "failed to fetch",
			wantStatus:  http.StatusNotFound,
			wantCode:    apierr.CodeNotFound,
		},
		{
			name:        "database.ErrConflict maps to 409 RESOURCE_CONFLICT",
			err:         database.ErrConflict,
			fallbackMsg: "failed to insert",
			wantStatus:  http.StatusConflict,
			wantCode:    apierr.CodeResourceConflict,
		},
		{
			name:        "pgconn unique violation (23505) maps to 409 RESOURCE_CONFLICT",
			err:         &pgconn.PgError{Code: "23505", Message: "duplicate key"},
			fallbackMsg: "failed to insert",
			wantStatus:  http.StatusConflict,
			wantCode:    apierr.CodeResourceConflict,
		},
		{
			name:        "database.ErrForeignKeyViolation maps to 400 INVALID_REFERENCE",
			err:         database.ErrForeignKeyViolation,
			fallbackMsg: "failed to save",
			wantStatus:  http.StatusBadRequest,
			wantCode:    apierr.CodeInvalidReference,
		},
		{
			name:        "pgconn foreign key violation (23503) maps to 400 INVALID_REFERENCE",
			err:         &pgconn.PgError{Code: "23503", Message: "violates foreign key"},
			fallbackMsg: "failed to save",
			wantStatus:  http.StatusBadRequest,
			wantCode:    apierr.CodeInvalidReference,
		},
		{
			name:        "database.ErrCheckViolation maps to 400 INVALID_INPUT",
			err:         database.ErrCheckViolation,
			fallbackMsg: "failed to save",
			wantStatus:  http.StatusBadRequest,
			wantCode:    apierr.CodeInvalidInput,
		},
		{
			name:        "database.ErrNotNullViolation maps to 400 INVALID_INPUT",
			err:         database.ErrNotNullViolation,
			fallbackMsg: "failed to save",
			wantStatus:  http.StatusBadRequest,
			wantCode:    apierr.CodeInvalidInput,
		},
		{
			name:        "database.ErrInvalidInput maps to 400 INVALID_INPUT",
			err:         database.ErrInvalidInput,
			fallbackMsg: "failed to save",
			wantStatus:  http.StatusBadRequest,
			wantCode:    apierr.CodeInvalidInput,
		},
		{
			name:        "pgconn check violation (23514) maps to 400 INVALID_INPUT",
			err:         &pgconn.PgError{Code: "23514", Message: "check violation"},
			fallbackMsg: "failed to save",
			wantStatus:  http.StatusBadRequest,
			wantCode:    apierr.CodeInvalidInput,
		},
		{
			name:        "pgconn not null violation (23502) maps to 400 INVALID_INPUT",
			err:         &pgconn.PgError{Code: "23502", Message: "null value"},
			fallbackMsg: "failed to save",
			wantStatus:  http.StatusBadRequest,
			wantCode:    apierr.CodeInvalidInput,
		},
		{
			name:        "pgconn invalid input (22P02) maps to 400 INVALID_INPUT",
			err:         &pgconn.PgError{Code: "22P02", Message: "bad uuid"},
			fallbackMsg: "failed to save",
			wantStatus:  http.StatusBadRequest,
			wantCode:    apierr.CodeInvalidInput,
		},
		{
			name:        "unclassified error maps to 500 INTERNAL_ERROR with fallback message",
			err:         errors.New("unexpected disk corruption"),
			fallbackMsg: "Custom fallback failure",
			wantStatus:  http.StatusInternalServerError,
			wantCode:    apierr.CodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.handleError(tt.err, tt.fallbackMsg)
			if tt.wantNil {
				assert.NoError(t, got)
				return
			}
			require.Error(t, got)
			if tt.wantExactErr != nil {
				assert.Same(t, tt.wantExactErr, got)
			}
			var apiErr *apierr.APIError
			require.True(t, errors.As(got, &apiErr))
			assert.Equal(t, tt.wantStatus, apiErr.Status)
			assert.Equal(t, tt.wantCode, apiErr.Code)
			if tt.err != nil && tt.wantExactErr == nil && tt.wantStatus == http.StatusInternalServerError {
				assert.Equal(t, tt.fallbackMsg, apiErr.Error())
			}
		})
	}
}

func TestService_Record(t *testing.T) {
	t.Run("success uses WithTx", func(t *testing.T) {
		runner := &mockRunner{}
		svc := NewService(runner, NewRepository())
		entry := &model.AuditLog{Action: "user.role.update"}

		err := svc.Record(context.Background(), entry)
		require.NoError(t, err)
		assert.True(t, runner.withTxCalled, "Record must execute inside WithTx")
		assert.False(t, runner.withDBCalled, "Record must NOT execute inside WithExecutor")
		assert.Equal(t, int64(42), entry.ID)
	})

	t.Run("runner error translated via handleError", func(t *testing.T) {
		runner := &mockRunner{withTxErr: database.ErrForeignKeyViolation}
		svc := NewService(runner, NewRepository())
		entry := &model.AuditLog{Action: "user.role.update"}

		err := svc.Record(context.Background(), entry)
		require.Error(t, err)
		var apiErr *apierr.APIError
		require.True(t, errors.As(err, &apiErr))
		assert.Equal(t, http.StatusBadRequest, apiErr.Status)
		assert.Equal(t, apierr.CodeInvalidReference, apiErr.Code)
		assert.True(t, runner.withTxCalled)
	})
}

func TestService_List(t *testing.T) {
	t.Run("runner error translated via handleError", func(t *testing.T) {
		runner := &mockRunner{withDBErr: errors.New("connection reset by peer")}
		svc := NewService(runner, NewRepository())

		entries, page, err := svc.List(context.Background(), Filter{}, 10, nil)
		require.Error(t, err)
		assert.Nil(t, entries)
		assert.Nil(t, page)

		var apiErr *apierr.APIError
		require.True(t, errors.As(err, &apiErr))
		assert.Equal(t, http.StatusInternalServerError, apiErr.Status)
		assert.Equal(t, apierr.CodeInternalError, apiErr.Code)
		assert.True(t, runner.withDBCalled, "List must execute inside WithExecutor")
		assert.False(t, runner.withTxCalled, "List must NOT execute inside WithTx")
	})
}
