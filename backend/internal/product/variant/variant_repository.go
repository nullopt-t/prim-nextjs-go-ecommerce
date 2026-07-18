package variant

import (
	"context"
	"fmt"

	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

type VariantRepository struct{}

func NewRepository() *VariantRepository {
	return &VariantRepository{}
}

func (vr *VariantRepository) Create(
	ctx context.Context,
	db database.QueryExecutor,
	variant *model.ProductVariant,
) error {
	_, err := db.Exec(
		ctx,
		`
		INSERT INTO product_variants(
			id,
			product_id,
			sku,
			price,
			currency,
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
		variant.ID,
		variant.ProductID,
		variant.SKU,
		variant.Price,
		variant.Currency,
		variant.CreatedAt,
		variant.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create a new variant :%w", err)
	}
	return nil
}
