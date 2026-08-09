package object

import (
	"context"
	"net/http"

	"io"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/storage"
)

type ObjectService struct {
	dbRunner        database.Runner
	objectRepo      *ObjectRepository
	storageProvider storage.StorageProvider
}

func NewService(
	dbRunner database.Runner,
	objectRepo *ObjectRepository,
	storageProvider storage.StorageProvider,
) *ObjectService {
	return &ObjectService{
		dbRunner:        dbRunner,
		objectRepo:      objectRepo,
		storageProvider: storageProvider,
	}
}

type CreateObjectInput struct {
}

func (os *ObjectService) CreateObject(
	ctx context.Context,
	contentType string,
	size int64,
	bucket string,
	key string,
	status model.ObjectStatus,
) (*model.Object, error) {
	object := &model.Object{
		ID:          uuid.New(),
		FileSize:    size,
		Status:      status,
		ContentType: contentType,
		Bucket:      bucket,
		Key:         key,
	}

	err := os.dbRunner.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			return os.objectRepo.Create(
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
	object := &model.Object{
		ID:          uuid.New(),
		FileSize:    size,
		Status:      status,
		ContentType: contentType,
		Bucket:      bucket,
		Key:         key,
	}

	err := os.objectRepo.Create(
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

	err := os.storageProvider.Upload(ctx, bucket, key, file, size, contentType)
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
	err := os.dbRunner.WithDB(ctx, func(tx database.QueryExecutor) error {
		return os.objectRepo.MarkDeletingByKey(ctx, tx, bucket, key)
	})
	if err != nil {
		return apierr.New(http.StatusInternalServerError, "failed to mark object as deleting").Wrap(err)
	}

	// Delete from storage
	err = os.storageProvider.Delete(ctx, bucket, key)
	if err != nil {
		return apierr.New(http.StatusInternalServerError, "failed to delete object from storage").Wrap(err)
	}

	// Finally mark as deleted in DB
	return os.dbRunner.WithDB(ctx, func(tx database.QueryExecutor) error {
		return os.objectRepo.MarkDeletedByKey(ctx, tx, bucket, key)
	})
}

func (os *ObjectService) DeleteObjectByID(ctx context.Context, id uuid.UUID) error {
	var obj *model.Object
	err := os.dbRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		obj, err = os.objectRepo.GetByID(ctx, db, id)
		return err
	})
	if err != nil {
		return apierr.New(http.StatusInternalServerError, "failed to get object by id").Wrap(err)
	}
	return os.DeleteObject(ctx, obj.Bucket, obj.Key)
}

func (os *ObjectService) GetObjectByID(ctx context.Context, id uuid.UUID) (*model.Object, error) {
	var obj *model.Object
	err := os.dbRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		obj, err = os.objectRepo.GetByID(ctx, db, id)
		return err
	})
	return obj, err
}

func (os *ObjectService) GetObjectsByIDs(ctx context.Context, ids []uuid.UUID) ([]*model.Object, error) {
	var objects []*model.Object
	err := os.dbRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		objects, err = os.objectRepo.GetByIDs(ctx, db, ids)
		return err
	})
	return objects, err
}

// MarkDeletingByKey marks a storage object as 'deleting' by its bucket and key.
// Safe to call inside a transaction.
func (os *ObjectService) MarkDeletingByKey(
	ctx context.Context,
	tx database.QueryExecutor,
	bucket, key string,
) error {
	return os.objectRepo.MarkDeletingByKey(ctx, tx, bucket, key)
}

func (os *ObjectService) GetObjectURL(
	ctx context.Context,
	bucket, key string,
) string {
	return os.storageProvider.GetPublicURL(bucket, key)
}
