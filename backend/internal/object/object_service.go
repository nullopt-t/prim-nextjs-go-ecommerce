package object

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/storage"
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
	if dbRunner == nil {
		panic("object: database runner is required")
	}
	if objectRepo == nil {
		panic("object: object repository is required")
	}
	if storageProvider == nil {
		panic("object: storage provider is required")
	}
	return &ObjectService{
		dbRunner:        dbRunner,
		objectRepo:      objectRepo,
		storageProvider: storageProvider,
	}
}

// handleError maps persistence, driver, and context errors to domain APIErrors.
func (os *ObjectService) handleError(err error, fallbackMsg string) error {
	if err == nil {
		return nil
	}

	var apiErr *apierr.APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}

	mappedErr := database.MapError(err)
	switch {
	case errors.Is(mappedErr, database.ErrNotFound):
		return apierr.NotFoundError("Object not found").
			WithCode(apierr.CodeNotFound).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrConflict):
		return apierr.ConflictError("Object already exists").
			WithCode(apierr.CodeAlreadyExists).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrForeignKeyViolation):
		return apierr.BadRequestError("Referenced record does not exist").
			WithCode(apierr.CodeInvalidInput).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrCheckViolation),
		errors.Is(mappedErr, database.ErrNotNullViolation),
		errors.Is(mappedErr, database.ErrInvalidInput):
		return apierr.BadRequestError("Invalid input").
			WithCode(apierr.CodeInvalidInput).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrTimeout), errors.Is(mappedErr, database.ErrCanceled):
		return apierr.New(http.StatusGatewayTimeout, "database operation timed out or canceled").
			WithCode(apierr.CodeGatewayTimeout).
			Wrap(err)
	default:
		return apierr.InternalServerError(fallbackMsg).
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
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

	err := os.dbRunner.WithTx(
		ctx,
		func(ctx context.Context) error {
			return os.objectRepo.Create(
				ctx,
				object,
			)
		},
	)
	if err != nil {
		return nil, os.handleError(err, "Failed to create storage object")
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
		return nil, apierr.StorageError("failed to upload file to storage").Wrap(err)
	}

	obj, err := os.CreateObject(ctx, contentType, size, bucket, key, model.ObjectStatusUploaded)
	if err != nil {
		_ = os.storageProvider.Delete(ctx, bucket, key)
		return nil, err
	}

	return obj, nil
}

func (os *ObjectService) DeleteObject(
	ctx context.Context,
	bucket, key string,
) error {
	// First mark as deleting in DB within a transaction
	err := os.dbRunner.WithTx(ctx, func(ctx context.Context) error {
		return os.objectRepo.MarkDeletingByKey(ctx, bucket, key)
	})
	if err != nil {
		return os.handleError(err, "Failed to mark object as deleting")
	}

	// Delete from storage
	err = os.storageProvider.Delete(ctx, bucket, key)
	if err != nil {
		return apierr.StorageError("failed to delete object from storage").Wrap(err)
	}

	// Finally mark as deleted in DB within a transaction
	err = os.dbRunner.WithTx(ctx, func(ctx context.Context) error {
		return os.objectRepo.MarkDeletedByKey(ctx, bucket, key)
	})
	if err != nil {
		return os.handleError(err, "Failed to mark object as deleted")
	}

	return nil
}

func (os *ObjectService) DeleteObjectByID(ctx context.Context, id uuid.UUID) error {
	var obj *model.Object
	err := os.dbRunner.WithExecutor(ctx, func(ctx context.Context) error {
		var err error
		obj, err = os.objectRepo.GetByID(ctx, id)
		return err
	})
	if err != nil {
		return os.handleError(err, "Failed to load media")
	}
	// Already soft-deleted objects are gone from the listing, so surface them as not found.
	if obj.DeletedAt != nil || obj.Status == model.ObjectStatusDeleted {
		return apierr.NotFoundError("Media not found").WithCode(apierr.CodeNotFound)
	}
	return os.DeleteObject(ctx, obj.Bucket, obj.Key)
}

// GetByID loads a storage object by id. It aliases GetObjectByID under the
// repository-style name so the service satisfies narrow reader interfaces
// (e.g. vendors.ObjectReader) without receiver changes in consumers.
func (os *ObjectService) GetByID(ctx context.Context, id uuid.UUID) (*model.Object, error) {
	return os.GetObjectByID(ctx, id)
}

func (os *ObjectService) GetObjectByID(ctx context.Context, id uuid.UUID) (*model.Object, error) {
	var obj *model.Object
	err := os.dbRunner.WithExecutor(ctx, func(ctx context.Context) error {
		var err error
		obj, err = os.objectRepo.GetByID(ctx, id)
		return err
	})
	if err != nil {
		return nil, os.handleError(err, "Failed to load object")
	}
	return obj, nil
}

// ResolvePublicURL maps a public object URL back to its stored object. The
// public URL is the only identifier clients ever see; storage object UUIDs
// stay internal. Uploaded keys are single-segment UUIDs, so the last two
// path segments of the public URL are bucket and key. It is deliberately
// asset-agnostic: any caller that accepts an object URL (thumbnails, logos,
// banners, ...) resolves it the same way.
func (os *ObjectService) ResolvePublicURL(ctx context.Context, rawURL string) (*model.Object, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Path == "" {
		return nil, apierr.BadRequestError("Invalid object URL").
			WithCode(apierr.CodeInvalidInput)
	}
	segments := strings.Split(strings.Trim(path.Clean(parsed.Path), "/"), "/")
	if len(segments) < 2 {
		return nil, apierr.BadRequestError("Invalid object URL").
			WithCode(apierr.CodeInvalidInput)
	}
	key, err := url.PathUnescape(segments[len(segments)-1])
	if err != nil || key == "" {
		return nil, apierr.BadRequestError("Invalid object URL").
			WithCode(apierr.CodeInvalidInput)
	}
	bucket := segments[len(segments)-2]

	var obj *model.Object
	err = os.dbRunner.WithExecutor(ctx, func(ctx context.Context) error {
		var err error
		obj, err = os.objectRepo.GetByBucketAndKey(ctx, bucket, key)
		return err
	})
	if err != nil {
		if errors.Is(database.MapError(err), database.ErrNotFound) {
			return nil, apierr.NotFoundError("Object not found").WithCode(apierr.CodeNotFound)
		}
		return nil, os.handleError(err, "Failed to load object")
	}
	if obj.DeletedAt != nil || obj.Status == model.ObjectStatusDeleted {
		return nil, apierr.NotFoundError("Object not found").WithCode(apierr.CodeNotFound)
	}
	return obj, nil
}

func (os *ObjectService) GetObjectsByIDs(ctx context.Context, ids []uuid.UUID) ([]*model.Object, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var objects []*model.Object
	err := os.dbRunner.WithExecutor(ctx, func(ctx context.Context) error {
		var err error
		objects, err = os.objectRepo.GetByIDs(ctx, ids)
		return err
	})
	if err != nil {
		return nil, os.handleError(err, "Failed to load objects")
	}
	return objects, nil
}

func (os *ObjectService) GetObjectURL(
	ctx context.Context,
	bucket, key string,
) string {
	return os.storageProvider.GetPublicURL(bucket, key)
}

type MediaItem struct {
	ID          uuid.UUID `json:"id"`
	Bucket      string    `json:"bucket"`
	ObjectKey   string    `json:"objectKey"`
	ContentType string    `json:"contentType"`
	FileSize    int64     `json:"fileSize"`
	Status      string    `json:"status"`
	URL         string    `json:"url"`
	CreatedAt   string    `json:"createdAt"`
}

func (os *ObjectService) ListMedia(ctx context.Context, limit int, cursor *pagination.Keyset) ([]MediaItem, *pagination.CursorPage, error) {
	var list []MediaItem
	var page *pagination.CursorPage
	err := os.dbRunner.WithExecutor(ctx, func(ctx context.Context) error {
		objects, p, err := os.objectRepo.ListMedia(ctx, limit, cursor)
		if err != nil {
			return err
		}
		page = p
		for _, obj := range objects {
			url := os.storageProvider.GetPublicURL(obj.Bucket, obj.Key)
			list = append(list, MediaItem{
				ID:          obj.ID,
				Bucket:      obj.Bucket,
				ObjectKey:   obj.Key,
				ContentType: obj.ContentType,
				FileSize:    obj.FileSize,
				Status:      string(obj.Status),
				URL:         url,
				CreatedAt:   obj.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			})
		}
		return nil
	})
	if err != nil {
		return nil, nil, os.handleError(err, "Failed to list media")
	}
	return list, page, nil
}
