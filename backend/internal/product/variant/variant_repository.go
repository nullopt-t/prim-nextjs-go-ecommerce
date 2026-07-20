package variant

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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

func (vr *VariantRepository) GetVariant(
	ctx context.Context,
	db database.QueryExecutor,
	variantID uuid.UUID,
) (*model.ProductVariant, error) {
	variant := &model.ProductVariant{}
	err := db.QueryRow(
		ctx,
		`
		SELECT
			id,
			product_id,
			sku,
			price,
			currency,
			created_at,
			updated_at
		FROM product_variants
		WHERE id = $1
		`,
		variantID,
	).Scan(
		&variant.ID,
		&variant.ProductID,
		&variant.SKU,
		&variant.Price,
		&variant.Currency,
		&variant.CreatedAt,
		&variant.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get variant by id: %w", err)
	}
	return variant, nil
}
