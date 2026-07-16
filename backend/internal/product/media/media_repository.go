package media

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

type MediaRepository struct {
}

func (mr *MediaRepository) Create(
	ctx context.Context,
	qe database.QueryExecutor,
	media *model.Media,
) error {
	_, err := qe.Exec(
		ctx,
		`
		INSERT INTO media(
			id,
			alt,
			type,
			mime_type,
			file_size,
			checksum,
			width,
			height,
			created_by,
			updated_by,
			created_at,
			updated_at,
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12,
			$13
		)
		`,
		media.ID,
		media.Alt,
		media.Type,
		media.MimeType,
		media.FileSize,
		media.Checksum,
		media.Width,
		media.Height,
		media.CreatedBy,
		media.UpdatedBy,
		media.CreatedAt,
		media.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create a new media: %w", err)
	}
	return nil
}

func (r *MediaRepository) GetByID(
	ctx context.Context,
	qe database.QueryExecutor,
	mediaID uuid.UUID,
) ([]*model.Media, error) {
	rows, err := qe.Query(ctx,
		`
		SELECT
			id,
			alt,
			type,
			mime_type,
			width,
			height,
			file_size,
			created_by,
			updated_by,
			created_at,
			updated_at
		FROM media
		WHERE id = $1
		`,
		mediaID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get product media: %w", err)
	}
	defer rows.Close()

	media := make([]*model.Media, 0)
	for rows.Next() {
		var m model.Media
		if err := rows.Scan(
			&m.ID,
			&m.Alt,
			&m.Type,
			&m.MimeType,
			&m.Width,
			&m.Height,
			&m.FileSize,
			&m.CreatedBy,
			&m.UpdatedBy,
			&m.CreatedAt,
			&m.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan product media: %w", err)
		}
		media = append(media, &m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate product media: %w", err)
	}
	return media, nil
}

func (r *MediaRepository) GetByChecksum(
	ctx context.Context,
	qe database.QueryExecutor,
	checksum string,
) (*model.Media, error) {
	media := &model.Media{}
	err := qe.QueryRow(ctx,
		`
		SELECT
			id,
			alt,
			type,
			mime_type,
			width,
			height,
			file_size,
			created_by,
			updated_by,
			created_at,
			updated_at
		FROM media
		WHERE checksum = $1
		`,
		checksum,
	).Scan(
		&media.ID,
		&media.Alt,
		&media.Type,
		&media.MimeType,
		&media.Width,
		&media.Height,
		&media.FileSize,
		&media.CreatedBy,
		&media.UpdatedBy,
		&media.CreatedAt,
		&media.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get product media: %w", err)
	}
	return media, nil
}

type UpdateMediaFields struct {
	Alt       *string
	Type      *string
	MimeType  *string
	Checksum  *string
	Width     *int
	Height    *int
	FileSize  *int64
	UpdatedBy uuid.UUID
	UpdatedAt time.Time
}

func (r *MediaRepository) Update(
	ctx context.Context,
	qe database.QueryExecutor,
	mediaID uuid.UUID,
	fields UpdateMediaFields,
) error {
	_, err := qe.Exec(ctx,
		`
		UPDATE media
		SET
			alt = COALESCE($1, alt),
			type = COALESCE($2, type),
			mime_type = COALESCE($3, mime_type),
			width = COALESCE($4, width),
			height = COALESCE($5, height),
			file_size = COALESCE($6, file_size),
			updated_by = COALESCE($7, updated_by),
			updated_at = COALESCE($8, updated_at)
		WHERE id = $9
		`,
		fields.Alt,
		fields.Type,
		fields.MimeType,
		fields.Width,
		fields.Height,
		fields.FileSize,
		fields.UpdatedBy,
		fields.UpdatedAt,
		mediaID,
	)
	if err != nil {
		return fmt.Errorf("failed to update product media: %w", err)
	}
	return nil
}

func (r *MediaRepository) Delete(
	ctx context.Context,
	qe database.QueryExecutor,
	mediaID uuid.UUID,
) error {
	_, err := qe.Exec(ctx,
		`
		UPDATE media
		SET deleted_at = NOW()
		WHERE id = $1
		`,
		mediaID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete product media: %w", err)
	}
	return nil
}
