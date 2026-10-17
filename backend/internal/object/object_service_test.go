package object_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/object"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

// fakeRow implements pgx.Row for the storage_objects GetByID column order:
// id, bucket, object_key, content_type, file_size, status, created_at, updated_at, deleted_at.
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
	*dest[0].(*uuid.UUID) = uuid.New()
	*dest[1].(*string) = "catalog"
	*dest[2].(*string) = "products/example.webp"
	*dest[3].(*string) = "image/webp"
	*dest[4].(*int64) = 1024
	*dest[5].(*model.ObjectStatus) = r.status
	*dest[6].(*time.Time) = now
	*dest[7].(*time.Time) = now
	*dest[8].(**time.Time) = nil
	if r.deleted {
		*dest[8].(**time.Time) = &now
	}
	return nil
}

type fakeExecutor struct{ row pgx.Row }

func (f fakeExecutor) Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (f fakeExecutor) Query(context.Context, string, ...interface{}) (pgx.Rows, error) {
	return nil, nil
}

func (f fakeExecutor) QueryRow(context.Context, string, ...interface{}) pgx.Row { return f.row }

type fakeRunner struct{ row pgx.Row }

func (f fakeRunner) WithTx(_ context.Context, fn func(context.Context, database.QueryExecutor) error) error {
	return fn(context.Background(), fakeExecutor{row: f.row})
}

func (f fakeRunner) WithDB(_ context.Context, fn func(database.QueryExecutor) error) error {
	return fn(fakeExecutor{row: f.row})
}

func TestDeleteObjectByIDNotFoundMapsTo404(t *testing.T) {
	svc := object.NewService(fakeRunner{row: fakeRow{err: pgx.ErrNoRows}}, object.NewRepository(), nil)

	err := svc.DeleteObjectByID(context.Background(), uuid.New())
	require.Error(t, err)

	var apiErr *apierr.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, 404, apiErr.Status)
	assert.Equal(t, apierr.CodeNotFound, apiErr.Code)
}

func TestDeleteObjectByIDAlreadyDeletedMapsTo404(t *testing.T) {
	svc := object.NewService(
		fakeRunner{row: fakeRow{status: model.ObjectStatusDeleted, deleted: true}},
		object.NewRepository(),
		nil,
	)

	err := svc.DeleteObjectByID(context.Background(), uuid.New())
	require.Error(t, err)

	var apiErr *apierr.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, 404, apiErr.Status)
	assert.Equal(t, apierr.CodeNotFound, apiErr.Code)
}
