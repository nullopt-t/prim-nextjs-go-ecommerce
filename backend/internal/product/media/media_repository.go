package media

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

type MediaRepository struct{}

func NewRepository() *MediaRepository {
	return &MediaRepository{}
}

func (mr *MediaRepository) Create(
	ctx context.Context,
	qe database.QueryExecutor,
	media *model.ProductMedia,
) error {
	_, err := qe.Exec(
		ctx,
		`
		INSERT INTO product_media(
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

func (mr *MediaRepository) GetByID(
	ctx context.Context,
	qe database.QueryExecutor,
	mediaID uuid.UUID,
) (*model.ProductMedia, error) {
	media := &model.ProductMedia{}
	err := qe.QueryRow(
		ctx,
		`
		SELECT 
			id,
			variant_id,
			object_id,
			type,
			sort_order,
			created_at,
			updated_at,
			deleted_at
		FROM
			product_media
		WHERE
			id = $1
		`,
		mediaID,
	).Scan(
		&media.ID,
		&media.VariantID,
		&media.ObjectID,
		&media.Type,
		&media.SortOrder,
		&media.CreatedAt,
		&media.UpdatedAt,
		&media.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get media by id :%w",
			err,
		)
	}
	return media, nil
}

func (mr *MediaRepository) GetVariantMedia(
	ctx context.Context,
	qe database.QueryExecutor,
	variantID uuid.UUID,
) ([]*model.ProductMedia, error) {

	rows, err := qe.Query(
		ctx,
		`
		SELECT
			id,
			variant_id,
			object_id,
			type,
			sort_order,
			created_at,
			updated_at,
			deleted_at
		FROM
			product_media
		WHERE
			variant_id = $1
		`,
		variantID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get variant media :%w",
			err,
		)
	}

	var media = make([]*model.ProductMedia, 0)
	for rows.Next() {
		m := &model.ProductMedia{}
		err := rows.Scan(
			&m.ID,
			&m.VariantID,
			&m.ObjectID,
			&m.Type,
			&m.SortOrder,
			&m.CreatedAt,
			&m.UpdatedAt,
			&m.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"get variant media :%w",
				err,
			)
		}
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
