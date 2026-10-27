package report

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

type mockReportRunner struct {
	withTxCalled bool
	withDBCalled bool
	withTxErr    error
	withDBErr    error
	qe           database.QueryExecutor
}

func (m *mockReportRunner) WithTx(ctx context.Context, fn func(ctx context.Context, tx database.QueryExecutor) error) error {
	m.withTxCalled = true
	if m.withTxErr != nil {
		return m.withTxErr
	}
	qe := m.qe
	if qe == nil {
		qe = &fakeReportExecutor{}
	}
	return fn(ctx, qe)
}

func (m *mockReportRunner) WithExecutor(ctx context.Context, fn func(exec database.QueryExecutor) error) error {
	m.withDBCalled = true
	if m.withDBErr != nil {
		return m.withDBErr
	}
	qe := m.qe
	if qe == nil {
		qe = &fakeReportExecutor{}
	}
	return fn(qe)
}

type fakeReportExecutor struct{}

func (f *fakeReportExecutor) Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (f *fakeReportExecutor) Query(context.Context, string, ...interface{}) (pgx.Rows, error) {
	return nil, nil
}

func (f *fakeReportExecutor) QueryRow(context.Context, string, ...interface{}) pgx.Row {
	return fakeReportRow{}
}

type fakeReportRow struct{}

func (r fakeReportRow) Scan(...any) error {
	return nil
}

func TestReportService_HandleError(t *testing.T) {
	svc := &Service{}
	existingAPIErr := apierr.BadRequestError("already mapped").WithCode("CUSTOM_CODE")

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
			fallbackMsg: "failed to fetch report",
			wantStatus:  http.StatusNotFound,
			wantCode:    apierr.CodeNotFound,
		},
		{
			name:        "database.ErrNotFound maps to 404 NOT_FOUND",
			err:         database.ErrNotFound,
			fallbackMsg: "failed to fetch report",
			wantStatus:  http.StatusNotFound,
			wantCode:    apierr.CodeNotFound,
		},
		{
			name:        "database.ErrConflict maps to 409 RESOURCE_CONFLICT",
			err:         database.ErrConflict,
			fallbackMsg: "failed to save report",
			wantStatus:  http.StatusConflict,
			wantCode:    apierr.CodeResourceConflict,
		},
		{
			name:        "pgconn unique violation (23505) maps to 409 RESOURCE_CONFLICT",
			err:         &pgconn.PgError{Code: "23505"},
			fallbackMsg: "failed to save report",
			wantStatus:  http.StatusConflict,
			wantCode:    apierr.CodeResourceConflict,
		},
		{
			name:        "database.ErrForeignKeyViolation maps to 400 INVALID_REFERENCE",
			err:         database.ErrForeignKeyViolation,
			fallbackMsg: "failed to save report",
			wantStatus:  http.StatusBadRequest,
			wantCode:    apierr.CodeInvalidReference,
		},
		{
			name:        "pgconn foreign key violation (23503) maps to 400 INVALID_REFERENCE",
			err:         &pgconn.PgError{Code: "23503"},
			fallbackMsg: "failed to save report",
			wantStatus:  http.StatusBadRequest,
			wantCode:    apierr.CodeInvalidReference,
		},
		{
			name:        "database.ErrCheckViolation maps to 400 INVALID_INPUT",
			err:         database.ErrCheckViolation,
			fallbackMsg: "failed to save report",
			wantStatus:  http.StatusBadRequest,
			wantCode:    apierr.CodeInvalidInput,
		},
		{
			name:        "database.ErrNotNullViolation maps to 400 INVALID_INPUT",
			err:         database.ErrNotNullViolation,
			fallbackMsg: "failed to save report",
			wantStatus:  http.StatusBadRequest,
			wantCode:    apierr.CodeInvalidInput,
		},
		{
			name:        "database.ErrInvalidInput maps to 400 INVALID_INPUT",
			err:         database.ErrInvalidInput,
			fallbackMsg: "failed to save report",
			wantStatus:  http.StatusBadRequest,
			wantCode:    apierr.CodeInvalidInput,
		},
		{
			name:        "pgconn check violation (23514) maps to 400 INVALID_INPUT",
			err:         &pgconn.PgError{Code: "23514"},
			fallbackMsg: "failed to save report",
			wantStatus:  http.StatusBadRequest,
			wantCode:    apierr.CodeInvalidInput,
		},
		{
			name:        "pgconn not null violation (23502) maps to 400 INVALID_INPUT",
			err:         &pgconn.PgError{Code: "23502"},
			fallbackMsg: "failed to save report",
			wantStatus:  http.StatusBadRequest,
			wantCode:    apierr.CodeInvalidInput,
		},
		{
			name:        "pgconn invalid input (22P02) maps to 400 INVALID_INPUT",
			err:         &pgconn.PgError{Code: "22P02"},
			fallbackMsg: "failed to save report",
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

func TestReportService_GetSalesReport_Validation(t *testing.T) {
	runner := &mockReportRunner{}
	svc := NewService(runner, NewRepository())

	tests := []struct {
		name      string
		groupBy   string
		wantErr   bool
		wantCode  string
		wantCalls bool
	}{
		{
			name:      "invalid groupBy year returns 400",
			groupBy:   "year",
			wantErr:   true,
			wantCode:  apierr.CodeInvalidInput,
			wantCalls: false,
		},
		{
			name:      "invalid groupBy week returns 400",
			groupBy:   "week",
			wantErr:   true,
			wantCode:  apierr.CodeInvalidInput,
			wantCalls: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner.withDBCalled = false
			rep, err := svc.GetSalesReport(context.Background(), ReportFilter{}, tt.groupBy)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, rep)
				var apiErr *apierr.APIError
				require.True(t, errors.As(err, &apiErr))
				assert.Equal(t, http.StatusBadRequest, apiErr.Status)
				assert.Equal(t, tt.wantCode, apiErr.Code)
			}
			assert.Equal(t, tt.wantCalls, runner.withDBCalled)
		})
	}
}

func TestReportService_GetSalesReport_Runner(t *testing.T) {
	t.Run("runner error routed through handleError", func(t *testing.T) {
		runner := &mockReportRunner{withDBErr: pgx.ErrNoRows}
		svc := NewService(runner, NewRepository())

		rep, err := svc.GetSalesReport(context.Background(), ReportFilter{}, "day")
		require.Error(t, err)
		assert.Nil(t, rep)

		var apiErr *apierr.APIError
		require.True(t, errors.As(err, &apiErr))
		assert.Equal(t, http.StatusNotFound, apiErr.Status)
		assert.Equal(t, apierr.CodeNotFound, apiErr.Code)
		assert.True(t, runner.withDBCalled)
	})
}

func TestReportService_BuildCSV_Runner(t *testing.T) {
	t.Run("runner error routed through handleError", func(t *testing.T) {
		runner := &mockReportRunner{withDBErr: database.ErrTimeout}
		svc := NewService(runner, NewRepository())

		bytes, err := svc.BuildCSV(context.Background(), ExportOrders, ReportFilter{})
		require.Error(t, err)
		assert.Nil(t, bytes)

		var apiErr *apierr.APIError
		require.True(t, errors.As(err, &apiErr))
		assert.Equal(t, http.StatusGatewayTimeout, apiErr.Status)
		assert.Equal(t, apierr.CodeGatewayTimeout, apiErr.Code)
		assert.True(t, runner.withDBCalled)
	})
}
