package object

import (
	"context"
	"net/http"
	"time"

	"io"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/storage"
)

type ObjectService struct {
	dr database.Runner
	or *ObjectRepository
	sp storage.StorageProvider
}

func NewService(
	dr database.Runner,
	or *ObjectRepository,
	sp storage.StorageProvider,
) *ObjectService {
	return &ObjectService{
		dr: dr,
		or: or,
		sp: sp,
	}
}

type CreateObjectInput struct {
}

func (os *ObjectService) CreateObject(
	ctx context.Context,
	ContentType string,
	Size int64,
	Bucket string,
	Key string,
	status model.ObjectStatus,
) (*model.Object, error) {
	now := time.Now().UTC()
	object := &model.Object{
		ID:          uuid.New(),
		FileSize:    Size,
		Status:      status,
		ContentType: ContentType,
		Bucket:      Bucket,
		Key:         Key,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err := os.dr.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			return os.or.Create(
				ctx,
				db,
				object,
			)
		},
	)
	if err != nil {
		return nil, apierr.New(
			http.StatusInternalServerError,
			"internal server error",
		).WithCode("INTERNAL_ERROR").
			Wrap(err).
			WithStack()
	}

	return object, nil
}

func (os *ObjectService) CreateObjectWithTx(
	ctx context.Context,
	tx database.QueryExecutor,
	bucket, key string,
	size int64,
	contentType string,
	status model.ObjectStatus,
) (*model.Object, error) {
	now := time.Now().UTC()
	object := &model.Object{
		ID:          uuid.New(),
		FileSize:    size,
		Status:      status,
		ContentType: contentType,
		Bucket:      bucket,
		Key:         key,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err := os.or.Create(
		ctx,
		tx,
		object,
	)
	if err != nil {
		return nil, err
	}

	return object, nil
}

func (os *ObjectService) UploadObject(
	ctx context.Context,
	contentType string,
	size int64,
	bucket string,
	file io.Reader,
) (*model.Object, error) {
	key := uuid.New().String()

	err := os.sp.Upload(ctx, bucket, key, file, size, contentType)
	if err != nil {
		return nil, apierr.New(
			http.StatusInternalServerError,
			"failed to upload file to storage",
		).WithCode("STORAGE_UPLOAD_ERROR").Wrap(err)
	}

	return os.CreateObject(ctx, contentType, size, bucket, key, model.ObjectStatusUploaded)
}

func (os *ObjectService) DeleteObject(
	ctx context.Context,
	bucket, key string,
) error {
	// First mark as deleting in DB
	err := os.dr.WithDB(ctx, func(tx database.QueryExecutor) error {
		return os.or.MarkDeletingByKey(ctx, tx, bucket, key)
	})
	if err != nil {
		return apierr.New(http.StatusInternalServerError, "failed to mark object as deleting").Wrap(err)
	}

	// Delete from storage
	err = os.sp.Delete(ctx, bucket, key)
	if err != nil {
		return apierr.New(http.StatusInternalServerError, "failed to delete object from storage").Wrap(err)
	}

	// Finally mark as deleted in DB
	err = os.dr.WithDB(ctx, func(tx database.QueryExecutor) error {
		return os.or.MarkDeletedByKey(ctx, tx, bucket, key)
	})

	return nil
}

// MarkDeletingByKey marks a storage object as 'deleting' by its bucket and key.
// Safe to call inside a transaction.
func (os *ObjectService) MarkDeletingByKey(
	ctx context.Context,
	tx database.QueryExecutor,
	bucket, key string,
) error {
	return os.or.MarkDeletingByKey(ctx, tx, bucket, key)
}

func (os *ObjectService) GetObjectURL(
	ctx context.Context,
	bucket, key string,
) string {
	return os.sp.GetPublicURL(bucket, key)
}
