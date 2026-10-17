package object

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
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
	err := qe.QueryRow(
		ctx,
		`
		INSERT INTO storage_objects (
			id,
			bucket,
			object_key,
			content_type,
			file_size,
			status,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, now(), now())
		RETURNING created_at, updated_at
		`,
		object.ID,
		object.Bucket,
		object.Key,
		object.ContentType,
		object.FileSize,
		object.Status,
	).Scan(&object.CreatedAt, &object.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create storage object: %w", err)
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
		UPDATE storage_objects
		SET status = $1, updated_at = now()
		WHERE id = $2
		`,
		status,
		objectID,
	)
	if err != nil {
		return fmt.Errorf("update object status: %w", err)
	}
	return nil
}

// MarkDeletingByKey sets the status of a storage_objects row to 'deleting' by bucket+key.
func (or *ObjectRepository) MarkDeletingByKey(
	ctx context.Context,
	qe database.QueryExecutor,
	bucket, key string,
) error {
	_, err := qe.Exec(ctx,
		`UPDATE storage_objects SET status = 'deleting', updated_at = now() WHERE bucket = $1 AND object_key = $2`,
		bucket, key,
	)
	if err != nil {
		return fmt.Errorf("mark deleting by key: %w", err)
	}
	return nil
}

// MarkDeletedByKey sets the status of a storage_objects row to 'deleted' by bucket+key.
func (or *ObjectRepository) MarkDeletedByKey(
	ctx context.Context,
	qe database.QueryExecutor,
	bucket, key string,
) error {
	_, err := qe.Exec(ctx,
		`UPDATE storage_objects SET status = 'deleted', deleted_at = now(), updated_at = now() WHERE bucket = $1 AND object_key = $2`,
		bucket, key,
	)
	if err != nil {
		return fmt.Errorf("mark deleted by key: %w", err)
	}
	return nil
}

func (or *ObjectRepository) GetByID(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
) (*model.Object, error) {
	obj := &model.Object{}
	err := qe.QueryRow(ctx, `
		SELECT id, bucket, object_key, content_type, file_size, status, created_at, updated_at, deleted_at
		FROM storage_objects
		WHERE id = $1
	`, id).Scan(
		&obj.ID, &obj.Bucket, &obj.Key, &obj.ContentType,
		&obj.FileSize, &obj.Status, &obj.CreatedAt, &obj.UpdatedAt, &obj.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get object by id: %w", err)
	}
	return obj, nil
}

func (or *ObjectRepository) GetByIDs(
	ctx context.Context,
	qe database.QueryExecutor,
	ids []uuid.UUID,
) ([]*model.Object, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	query := `
		SELECT id, bucket, object_key, content_type, file_size, status, created_at, updated_at, deleted_at
		FROM storage_objects
		WHERE id = ANY($1)
	`
	rows, err := qe.Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("get objects by ids query: %w", err)
	}
	defer rows.Close()

	var objects []*model.Object
	for rows.Next() {
		obj := &model.Object{}
		err := rows.Scan(
			&obj.ID, &obj.Bucket, &obj.Key, &obj.ContentType,
			&obj.FileSize, &obj.Status, &obj.CreatedAt, &obj.UpdatedAt, &obj.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan object: %w", err)
		}
		objects = append(objects, obj)
	}
	return objects, nil
}

func (or *ObjectRepository) ListMedia(
	ctx context.Context,
	qe database.QueryExecutor,
	limit int,
	cursor *pagination.Keyset,
) ([]*model.Object, *pagination.CursorPage, error) {
	query := `
		SELECT id, bucket, object_key, content_type, file_size, status, created_at, updated_at, deleted_at
		FROM storage_objects
		WHERE deleted_at IS NULL
	`
	args := []any{}
	i := 1

	keysetClause, keysetArgs := pagination.KeysetClause("", cursor, i)
	query += keysetClause
	args = append(args, keysetArgs...)
	i += len(keysetArgs)

	query += pagination.KeysetOrderBy("")
	query += fmt.Sprintf(" LIMIT $%d", i)
	args = append(args, limit+1)

	rows, err := qe.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("list media: %w", err)
	}
	defer rows.Close()

	var objects []*model.Object
	for rows.Next() {
		obj := &model.Object{}
		if err := rows.Scan(
			&obj.ID, &obj.Bucket, &obj.Key, &obj.ContentType,
			&obj.FileSize, &obj.Status, &obj.CreatedAt, &obj.UpdatedAt, &obj.DeletedAt,
		); err != nil {
			return nil, nil, fmt.Errorf("scan media: %w", err)
		}
		objects = append(objects, obj)
	}

	objects, page := pagination.NewCursorPageFrom(objects, limit, func(o *model.Object) pagination.Keyset {
		return pagination.Keyset{CreatedAt: o.CreatedAt, ID: o.ID}
	})
	return objects, page, nil
}
