package media

import (
	"context"

	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

type ProductMediaRepository struct {
}

func (pm *ProductMediaRepository) Create(
	ctx context.Context,
	qe database.QueryExecutor,
	model *model.ProductMedia,
) error {
	_, err := qe.Exec(ctx, `
		INSERT INTO product_media (
			product_id,
			media_id,
			sort_order,
			is_primary
		)
		VALUES (
			$1,
			$2,
			$3,
			$4
		)
	`,
		model.ProductID,
		model.MediaID,
		model.SortOrder,
		model.IsPrimary,
	)
	return err
}
