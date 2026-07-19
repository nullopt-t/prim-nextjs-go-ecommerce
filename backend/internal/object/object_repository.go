package object

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

type ObjectRepository struct {
}

func NewRepository() *ObjectRepository {
	return &ObjectRepository{}
}

func (or *ObjectRepository) Create(
	ctx context.Context,
	db database.QueryExecutor,
	object *model.Object,
) error {
	_, err := db.Exec(
		ctx,
		`
		INSERT INTO objects(
			id,
			size,
			status,
			content_type,
			key,
			bucket,
			created_at,
			updated_at
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8
		)
		`,
		object.ID,
		object.Size,
		object.Status,
		object.ContentType,
		object.Key,
		object.Bucket,
		object.CreatedAt,
		object.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf(
			"create a new object :%w",
			err,
		)
	}
	return nil
}

func (or *ObjectRepository) GetByID(
	ctx context.Context,
	db database.QueryExecutor,
	objectID uuid.UUID,
) (*model.Object, error) {
	object := &model.Object{}
	err := db.QueryRow(
		ctx,
		`
		SELECT 
			id,
			key,
			bucket,
			size,
			status,
			content_type,
			created_at,
			updated_at,
			deleted_at
		FROM 
			objects
		WHERE
			id = $1
		`,
		objectID,
	).Scan(
		&object.ID,
		&object.Key,
		&object.Bucket,
		&object.Size,
		&object.Status,
		&object.ContentType,
		&object.CreatedAt,
		&object.UpdatedAt,
		&object.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get by id : %w",
			err,
		)
	}
	return object, nil
}

func (or *ObjectRepository) Delete(
	ctx context.Context,
	db database.QueryExecutor,
	objectID uuid.UUID,
) error {
	_, err := db.Exec(
		ctx,
		`
		UPDATE 
			objects
		SET
			status = 'deleting'
		WHERE
			id = $1
		`,
		objectID,
	)
	if err != nil {
		return fmt.Errorf(
			"delete an object: %w",
			err,
		)
	}
	return nil
}
