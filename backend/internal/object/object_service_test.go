package object

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

type fakeRow struct {
	err     error
	status  model.ObjectStatus
	deleted bool
}

func (r fakeRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	now := time.Now()
	if len(dest) >= 1 {
		if idPtr, ok := dest[0].(*uuid.UUID); ok {
			*idPtr = uuid.New()
		}
	}
	if len(dest) >= 2 {
		if bktPtr, ok := dest[1].(*string); ok {
			*bktPtr = "catalog"
		}
	}
	if len(dest) >= 3 {
		if keyPtr, ok := dest[2].(*string); ok {
			*keyPtr = "products/example.webp"
		}
	}
	if len(dest) >= 4 {
		if ctPtr, ok := dest[3].(*string); ok {
			*ctPtr = "image/webp"
		}
	}
	if len(dest) >= 5 {
		if sizePtr, ok := dest[4].(*int64); ok {
			*sizePtr = 1024
		}
	}
	if len(dest) >= 6 {
		if stPtr, ok := dest[5].(*model.ObjectStatus); ok {
			*stPtr = r.status
		}
	}
	if len(dest) >= 7 {
		if tPtr, ok := dest[6].(*time.Time); ok {
			*tPtr = now
		}
	}
	if len(dest) >= 8 {
		if tPtr, ok := dest[7].(*time.Time); ok {
			*tPtr = now
		}
	}
	if len(dest) >= 9 {
		if dtPtr, ok := dest[8].(**time.Time); ok {
			if r.deleted {
				*dtPtr = &now
			} else {
				*dtPtr = nil
			}
		}
	}
	return nil
}

type fakeExecutor struct {
	row     pgx.Row
	execErr error
}

func (f fakeExecutor) ExecContext(context.Context, string, ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, f.execErr
}

func (f fakeExecutor) QueryContext(context.Context, string, ...interface{}) (pgx.Rows, error) {
	return nil, nil
}

func (f fakeExecutor) QueryRowContext(context.Context, string, ...interface{}) pgx.Row { return f.row }

type fakeTx struct {
	fakeExecutor
}

func (f *fakeTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return f.ExecContext(ctx, sql, args...)
}
func (f *fakeTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return f.QueryContext(ctx, sql, args...)
}
func (f *fakeTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return f.QueryRowContext(ctx, sql, args...)
}

func (f *fakeTx) Begin(context.Context) (pgx.Tx, error) { return f, nil }
func (f *fakeTx) Commit(context.Context) error          { return nil }
func (f *fakeTx) Rollback(context.Context) error        { return nil }
func (f *fakeTx) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (f *fakeTx) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults { return nil }
func (f *fakeTx) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (f *fakeTx) Conn() *pgx.Conn                { return nil }
func (f *fakeTx) LargeObjects() pgx.LargeObjects { return pgx.LargeObjects{} }

type fakeRunner struct {
	row     pgx.Row
	txErr   error
	dbErr   error
	execErr error
}

func (f fakeRunner) WithTx(ctx context.Context, fn func(context.Context) error) error {
	if f.txErr != nil {
		return f.txErr
	}
	tx := database.NewTx(&fakeTx{fakeExecutor: fakeExecutor{row: f.row, execErr: f.execErr}})
	return fn(database.WithTx(ctx, tx))
}

func (f fakeRunner) WithExecutor(ctx context.Context, fn func(context.Context) error) error {
	if f.dbErr != nil {
		return f.dbErr
	}
	exec := fakeExecutor{row: f.row, execErr: f.execErr}
	return fn(database.WithExecutor(ctx, exec))
}

type fakeStorageProvider struct {
	uploadErr   error
	deleteErr   error
	deleteCalls int
	deletedBkt  string
	deletedKey  string
}

func (f *fakeStorageProvider) Upload(ctx context.Context, bucket, key string, file io.Reader, size int64, contentType string) error {
	return f.uploadErr
}

func (f *fakeStorageProvider) Delete(ctx context.Context, bucket, key string) error {
	f.deleteCalls++
	f.deletedBkt = bucket
	f.deletedKey = key
	return f.deleteErr
}

func (f *fakeStorageProvider) GetPresignedURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error) {
	return "", nil
}

func (f *fakeStorageProvider) GetPublicURL(bucket, key string) string {
	return "https://cdn.example.com/" + bucket + "/" + key
}

func TestDeleteObjectByIDNotFoundMapsTo404(t *testing.T) {
	svc := NewService(fakeRunner{row: fakeRow{err: pgx.ErrNoRows}}, NewRepository(), &fakeStorageProvider{})

	err := svc.DeleteObjectByID(context.Background(), uuid.New())
	require.Error(t, err)

	var apiErr *apierr.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, 404, apiErr.Status)
	assert.Equal(t, apierr.CodeNotFound, apiErr.Code)
}

func TestDeleteObjectByIDAlreadyDeletedMapsTo404(t *testing.T) {
	svc := NewService(
		fakeRunner{row: fakeRow{status: model.ObjectStatusDeleted, deleted: true}},
		NewRepository(),
		&fakeStorageProvider{},
	)

	err := svc.DeleteObjectByID(context.Background(), uuid.New())
	require.Error(t, err)

	var apiErr *apierr.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, 404, apiErr.Status)
	assert.Equal(t, apierr.CodeNotFound, apiErr.Code)
}

func TestObjectService_HandleError(t *testing.T) {
	svc := NewService(fakeRunner{}, NewRepository(), &fakeStorageProvider{})

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
			fallbackMsg:    "not found",
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
			inputErr:       errors.New("unexpected disk crash"),
			fallbackMsg:    "database query failed",
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

func TestObjectService_UploadObject_CompensatingDeleteOnCreateFailure(t *testing.T) {
	storage := &fakeStorageProvider{}
	svc := NewService(
		fakeRunner{txErr: errors.New("db insert failure")},
		NewRepository(),
		storage,
	)

	_, err := svc.UploadObject(
		context.Background(),
		"image/webp",
		1024,
		"catalog",
		bytes.NewReader([]byte("fake image content")),
	)
	require.Error(t, err)

	// Compensating delete MUST be called once for cleanup
	assert.Equal(t, 1, storage.deleteCalls)
	assert.Equal(t, "catalog", storage.deletedBkt)
	assert.NotEmpty(t, storage.deletedKey)

	var apiErr *apierr.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusInternalServerError, apiErr.Status)
}

func TestObjectService_UploadObject_StorageUploadFails(t *testing.T) {
	storage := &fakeStorageProvider{uploadErr: errors.New("s3 unreachable")}
	svc := NewService(fakeRunner{}, NewRepository(), storage)

	_, err := svc.UploadObject(
		context.Background(),
		"image/webp",
		1024,
		"catalog",
		bytes.NewReader([]byte("fake image content")),
	)
	require.Error(t, err)
	assert.Equal(t, 0, storage.deleteCalls)

	var apiErr *apierr.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, apierr.CodeStorageError, apiErr.Code)
}

func TestObjectService_DeleteObject_Success(t *testing.T) {
	storage := &fakeStorageProvider{}
	svc := NewService(fakeRunner{}, NewRepository(), storage)

	err := svc.DeleteObject(context.Background(), "catalog", "products/img.webp")
	require.NoError(t, err)
	assert.Equal(t, 1, storage.deleteCalls)
	assert.Equal(t, "catalog", storage.deletedBkt)
	assert.Equal(t, "products/img.webp", storage.deletedKey)
}

func TestObjectService_DeleteObject_StorageDeleteFailure(t *testing.T) {
	storage := &fakeStorageProvider{deleteErr: errors.New("s3 delete denied")}
	svc := NewService(fakeRunner{}, NewRepository(), storage)

	err := svc.DeleteObject(context.Background(), "catalog", "products/img.webp")
	require.Error(t, err)

	var apiErr *apierr.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, apierr.CodeStorageError, apiErr.Code)
}
