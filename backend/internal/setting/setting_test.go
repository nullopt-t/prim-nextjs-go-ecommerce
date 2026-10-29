package setting

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

type fakeSettingRows struct {
	items []Item
	idx   int
}

func (r *fakeSettingRows) Close()                                       {}
func (r *fakeSettingRows) Err() error                                   { return nil }
func (r *fakeSettingRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (r *fakeSettingRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (r *fakeSettingRows) Next() bool {
	if r.idx < len(r.items) {
		r.idx++
		return true
	}
	return false
}
func (r *fakeSettingRows) Scan(dest ...any) error {
	it := r.items[r.idx-1]
	if len(dest) >= 1 {
		if k, ok := dest[0].(*string); ok {
			*k = it.Key
		}
	}
	if len(dest) >= 2 {
		if v, ok := dest[1].(*interface{}); ok {
			*v = it.Value
		}
	}
	if len(dest) >= 3 {
		if t, ok := dest[2].(*time.Time); ok {
			*t = it.UpdatedAt
		}
	}
	return nil
}
func (r *fakeSettingRows) Values() ([]any, error) { return nil, nil }
func (r *fakeSettingRows) RawValues() [][]byte    { return nil }
func (r *fakeSettingRows) Conn() *pgx.Conn        { return nil }

type fakeSettingExecutor struct {
	items   []Item
	execErr error
}

func (e *fakeSettingExecutor) Exec(_ context.Context, _ string, _ ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.NewCommandTag("INSERT 0 1"), e.execErr
}

func (e *fakeSettingExecutor) Query(_ context.Context, _ string, _ ...interface{}) (pgx.Rows, error) {
	return &fakeSettingRows{items: e.items}, nil
}

func (e *fakeSettingExecutor) QueryRow(_ context.Context, _ string, _ ...interface{}) pgx.Row {
	return nil
}

type fakeSettingTx struct {
	fakeSettingExecutor
}

func (f *fakeSettingTx) Begin(context.Context) (pgx.Tx, error) { return f, nil }
func (f *fakeSettingTx) Commit(context.Context) error { return nil }
func (f *fakeSettingTx) Rollback(context.Context) error { return nil }
func (f *fakeSettingTx) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (f *fakeSettingTx) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults { return nil }
func (f *fakeSettingTx) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (f *fakeSettingTx) Conn() *pgx.Conn { return nil }
func (f *fakeSettingTx) LargeObjects() pgx.LargeObjects { return pgx.LargeObjects{} }

type fakeRunner struct {
	txErr   error
	dbErr   error
	execErr error
	items   []Item
}

func (f *fakeRunner) WithTx(ctx context.Context, fn func(context.Context, *database.Tx) error) error {
	if f.txErr != nil {
		return f.txErr
	}
	return fn(ctx, database.NewTx(&fakeSettingTx{fakeSettingExecutor: fakeSettingExecutor{items: f.items, execErr: f.execErr}}))
}

func (f *fakeRunner) WithExecutor(ctx context.Context, fn func(database.QueryExecutor) error) error {
	if f.dbErr != nil {
		return f.dbErr
	}
	return fn(&fakeSettingExecutor{items: f.items, execErr: f.execErr})
}

func TestSettingService_HandleError(t *testing.T) {
	svc := NewService(&fakeRunner{}, NewRepository())

	tests := []struct {
		name           string
		inputErr       error
		fallbackMsg    string
		expectedStatus int
		expectedCode   string
	}{
		{
			name:        "nil error",
			inputErr:    nil,
			fallbackMsg: "fallback",
		},
		{
			name:           "already mapped APIError",
			inputErr:       apierr.BadRequestError("invalid payload").WithCode(apierr.CodeInvalidPayload),
			fallbackMsg:    "fallback",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   apierr.CodeInvalidPayload,
		},
		{
			name:           "pgx.ErrNoRows to not found",
			inputErr:       pgx.ErrNoRows,
			fallbackMsg:    "setting not found",
			expectedStatus: http.StatusNotFound,
			expectedCode:   apierr.CodeNotFound,
		},
		{
			name:           "unique violation to conflict",
			inputErr:       &pgconn.PgError{Code: "23505"},
			fallbackMsg:    "conflict",
			expectedStatus: http.StatusConflict,
			expectedCode:   apierr.CodeAlreadyExists,
		},
		{
			name:           "foreign key violation to bad request",
			inputErr:       &pgconn.PgError{Code: "23503"},
			fallbackMsg:    "fk error",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   apierr.CodeInvalidInput,
		},
		{
			name:           "check violation to bad request",
			inputErr:       &pgconn.PgError{Code: "23514"},
			fallbackMsg:    "check error",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   apierr.CodeInvalidInput,
		},
		{
			name:           "not null violation to bad request",
			inputErr:       &pgconn.PgError{Code: "23502"},
			fallbackMsg:    "not null error",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   apierr.CodeInvalidInput,
		},
		{
			name:           "invalid input text representation",
			inputErr:       &pgconn.PgError{Code: "22P02"},
			fallbackMsg:    "invalid input",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   apierr.CodeInvalidInput,
		},
		{
			name:           "context deadline exceeded to gateway timeout",
			inputErr:       context.DeadlineExceeded,
			fallbackMsg:    "timeout",
			expectedStatus: http.StatusGatewayTimeout,
			expectedCode:   apierr.CodeGatewayTimeout,
		},
		{
			name:           "context canceled to gateway timeout",
			inputErr:       context.Canceled,
			fallbackMsg:    "canceled",
			expectedStatus: http.StatusGatewayTimeout,
			expectedCode:   apierr.CodeGatewayTimeout,
		},
		{
			name:           "unhandled internal database error",
			inputErr:       errors.New("disk I/O failure"),
			fallbackMsg:    "database operation failed",
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   apierr.CodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.handleError(tt.inputErr, tt.fallbackMsg)
			if tt.inputErr == nil {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			var apiErr *apierr.APIError
			require.True(t, errors.As(err, &apiErr))
			assert.Equal(t, tt.expectedStatus, apiErr.Status)
			assert.Equal(t, tt.expectedCode, apiErr.Code)
		})
	}
}

func TestSettingService_GetSettings_ErrorMapping(t *testing.T) {
	svc := NewService(&fakeRunner{dbErr: pgx.ErrNoRows}, NewRepository())

	settings, err := svc.GetSettings(context.Background())
	require.Error(t, err)
	assert.Nil(t, settings)

	var apiErr *apierr.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusNotFound, apiErr.Status)
	assert.Equal(t, apierr.CodeNotFound, apiErr.Code)
}

func TestSettingService_GetConfig_CachingAndDefaults(t *testing.T) {
	dbCalls := 0
	runner := &fakeRunner{}
	// Intercept WithExecutor to count calls
	countingRunner := &testCountingRunner{
		runner: runner,
		onWithExecutor: func() {
			dbCalls++
		},
	}
	svc := NewService(countingRunner, NewRepository())

	// First call should query DB and populate cache
	cfg1, err := svc.GetConfig(context.Background())
	require.NoError(t, err)
	assert.Equal(t, DefaultConfig().StoreName, cfg1.StoreName)
	assert.Equal(t, 1, dbCalls)

	// Second call within TTL should return from cache without DB call
	cfg2, err := svc.GetConfig(context.Background())
	require.NoError(t, err)
	assert.Equal(t, cfg1.StoreName, cfg2.StoreName)
	assert.Equal(t, 1, dbCalls)
}

func TestSettingService_UpdateSettings_InvalidatesCache(t *testing.T) {
	runner := &fakeRunner{}
	svc := NewService(runner, NewRepository())

	// Seed cache
	_, err := svc.GetConfig(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, svc.cached)

	// Update settings
	err = svc.UpdateSettings(context.Background(), map[string]interface{}{
		"storeName": "Updated Shop Name",
	})
	require.NoError(t, err)

	// Cache must be invalidated
	assert.Nil(t, svc.cached)
}

func TestSettingService_UpdateSettings_ErrorMapping(t *testing.T) {
	svc := NewService(&fakeRunner{txErr: errors.New("db conflict")}, NewRepository())

	err := svc.UpdateSettings(context.Background(), map[string]interface{}{
		"storeName": "Conflict Store",
	})
	require.Error(t, err)

	var apiErr *apierr.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusInternalServerError, apiErr.Status)
}

type testCountingRunner struct {
	runner   *fakeRunner
	onWithExecutor func()
}

func (c *testCountingRunner) WithTx(ctx context.Context, fn func(context.Context, *database.Tx) error) error {
	return c.runner.WithTx(ctx, fn)
}

func (c *testCountingRunner) WithExecutor(ctx context.Context, fn func(database.QueryExecutor) error) error {
	if c.onWithExecutor != nil {
		c.onWithExecutor()
	}
	return c.runner.WithExecutor(ctx, fn)
}
