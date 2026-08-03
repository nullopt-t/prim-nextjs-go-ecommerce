package object

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/security"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

type ObjectService struct {
	dr database.Runner
	or *ObjectRepository
}

func NewService(
	dr database.Runner,
	or *ObjectRepository,
) *ObjectService {
	return &ObjectService{
		dr: dr,
		or: or,
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
		return nil, security.NewSecureError(
			http.StatusInternalServerError,
		).WithCode("INTERNAL_ERROR").
			Wrap(err).
			WithMessage("internal server error").
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

// MarkDeletingByKey marks a storage object as 'deleting' by its bucket and key.
// Safe to call inside a transaction.
func (os *ObjectService) MarkDeletingByKey(
	ctx context.Context,
	tx database.QueryExecutor,
	bucket, key string,
) error {
	return os.or.MarkDeletingByKey(ctx, tx, bucket, key)
}
