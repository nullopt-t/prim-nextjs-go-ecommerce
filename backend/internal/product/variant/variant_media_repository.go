package variant

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

type MediaRepository struct{}

func NewMediaRepository() *MediaRepository {
	return &MediaRepository{}
}

func (mr *MediaRepository) CreateVariantMedia(
	ctx context.Context,
	qe database.QueryExecutor,
	media *model.VariantMedia,
) error {
	_, err := qe.Exec(
		ctx,
		`
		INSERT INTO variant_media(
			id,
			variant_id,
			object_id,
			sort_order,
			type,
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
			$7
		)
		`,
		media.ID,
		media.VariantID,
		media.ObjectID,
		media.SortOrder,
		media.Type,
		media.CreatedAt,
		media.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf(
			"create a new product media :%w",
			err,
		)
	}
	return nil
}

func (mr *MediaRepository) ReorderVariantMedia(
	ctx context.Context,
	qe database.QueryExecutor,
	variantID uuid.UUID,
	mediaID uuid.UUID,
	idx int,
) error {
	_, err := qe.Exec(
		ctx,
		`
		UPDATE variant_media
		SET sort_order = $1
		WHERE variant_id = $2 AND id = $3
		`,
		idx,
		variantID,
		mediaID,
	)
	if err != nil {
		return fmt.Errorf(
			"reorder variant media :%w",
			err,
		)
	}
	return nil
}

func (mr *MediaRepository) GetVariantMedia(
	ctx context.Context,
	qe database.QueryExecutor,
	variantID uuid.UUID,
) ([]*model.VariantMedia, error) {

	rows, err := qe.Query(
		ctx,
		`
		SELECT
			vm.id,
			vm.type,
			vm.sort_order,
			vm.created_at,
			vm.updated_at,
			os.id,
			os.bucket,
			os.key,
			os.content_type,
			os.size,
			os.status,
			os.created_at,
			os.updated_at
		FROM
			variant_media vm
		JOIN objects os
			ON vm.object_id = os.id
		WHERE
			vm.variant_id = $1
		ORDER BY
			vm.sort_order
		`,
		variantID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get variant media :%w",
			err,
		)
	}

	var media = make([]*model.VariantMedia, 0)
	for rows.Next() {
		m := &model.VariantMedia{}
		o := &model.Object{}
		err := rows.Scan(
			&m.ID,
			&m.Type,
			&m.SortOrder,
			&m.CreatedAt,
			&m.UpdatedAt,
			&o.ID,
			&o.Bucket,
			&o.Key,
			&o.ContentType,
			&o.Size,
			&o.Status,
			&o.CreatedAt,
			&o.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"get variant media :%w",
				err,
			)
		}
		m.Object = o
		media = append(media, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"get variant media :%w",
			err,
		)
	}
	return media, nil
}
