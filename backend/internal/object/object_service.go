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
	ContentType string
	Size        int64
	Status      string
	Bucket      string
	Key         string
}

func (os *ObjectService) CreateObject(
	ctx context.Context,
	input CreateObjectInput,
) (*model.Object, error) {
	now := time.Now().UTC()
	object := &model.Object{
		ID:          uuid.New(),
		Size:        input.Size,
		Status:      input.Status,
		ContentType: input.ContentType,
		Bucket:      input.Bucket,
		Key:         input.Key,
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
			security.CodeInternal,
			"failed to create a new object",
			err,
		)
	}

	return object, nil
}

func (os *ObjectService) GetObjectByID(
	ctx context.Context,
	objectID uuid.UUID,
) (*model.Object, error) {
	var object *model.Object
	err := os.dr.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			o, err := os.or.GetByID(
				ctx,
				db,
				objectID,
			)
			if err != nil {
				return err
			}
			object = o
			return nil
		},
	)
	if err != nil {
		return nil, security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to create a new object",
			err,
		)
	}
	return object, nil
}

func (os *ObjectService) DeleteObject(
	ctx context.Context,
	objectID uuid.UUID,
) error {
	err := os.dr.WithDB(
		ctx,
		func(db database.QueryExecutor) error {
			return os.or.Delete(
				ctx,
				db,
				objectID,
			)
		},
	)
	if err != nil {
		return security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to delete and object",
			err,
		)
	}
	return nil
}
