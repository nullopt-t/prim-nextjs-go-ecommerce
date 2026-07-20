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
	qe database.QueryExecutor,
	object *model.Object,
) error {
	_, err := qe.Exec(
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

func (or *ObjectRepository) UpdateStatus(
	ctx context.Context,
	db database.QueryExecutor,
	objectID uuid.UUID,
	status model.ObjectStatus,
) error {
	_, err := db.Exec(
		ctx,
		`
		UPDATE
			objects
		SET
			status = $1
		WHERE
			id = $2
		`,
		status,
		objectID,
	)
	if err != nil {
		return fmt.Errorf(
			"update status: %w",
			err,
		)
	}
	return nil
}
