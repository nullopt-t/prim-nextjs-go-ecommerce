package product

import (
	"context"
	"fmt"

	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"

	"github.com/google/uuid"
)

type Filter struct {
	ID   *uuid.UUID
	Slug *string
}

type ProductRepository struct {
}

func NewProductRepository() *ProductRepository {
	return &ProductRepository{}
}

func (r *ProductRepository) Create(
	ctx context.Context,
	qe database.QueryExecutor,
	p *model.Product,
) error {
	_, err := qe.Exec(ctx,
		`INSERT INTO products (
			id,
			title,
	 		short_description,
			description,
			slug,
			status,
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
		p.ID,
		p.Title,
		p.ShortDescription,
		p.Description,
		p.Slug,
		p.Status,
		p.CreatedAt,
		p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf(
			"create a product :%w",
			err,
		)
	}
	return nil
}

func (r *ProductRepository) Get(
	ctx context.Context,
	qe database.QueryExecutor,
	filter Filter,
) (*model.Product, error) {

	query := `
	SELECT
		id,
		brand_id,
		title,
		short_description,
		description,
		slug,
		status,
		created_at,
		updated_at
	FROM products
	WHERE deleted_at IS NULL`

	args := []any{}
	argID := 1
	if filter.ID != nil {
		query += fmt.Sprintf(" AND id=$%d", argID)
		args = append(args, *filter.ID)
		argID++
	}

	if filter.Slug != nil {
		query += fmt.Sprintf(" AND slug=$%d", argID)
		args = append(args, *filter.Slug)
	}

	product := &model.Product{}
	err := qe.QueryRow(ctx,
		query,
		args...,
	).Scan(
		&product.ID,
		&product.BrandID,
		&product.Title,
		&product.ShortDescription,
		&product.Description,
		&product.Slug,
		&product.Status,
		&product.CreatedAt,
		&product.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf(
			"get a product:%w",
			err,
		)
	}

	return product, nil
}

func (r *ProductRepository) Update(
	ctx context.Context,
	qe database.QueryExecutor,
	product *model.Product,
) error {
	query := `
		UPDATE products
		SET
			title = $1,
			short_description = $2,
			description = $3,
			slug = $4,
			status = $5,
			updated_at = NOW()
		WHERE id = $6
	`
	args := []any{
		product.Title,
		product.ShortDescription,
		product.Description,
		product.Slug,
		product.Status,
		product.ID,
	}
	_, err := qe.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update product: %w", err)
	}
	return nil
}

type ProductListItem struct {
	ID           uuid.UUID `json:"id"`
	Title        string    `json:"title"`
	Slug         string    `json:"slug"`
	ThumbnailURL *string   `json:"thumbnailUrl"`
	Price        int64     `json:"price"`
	Currency     string    `json:"currency"`
}

type ProductList struct {
	Items []*ProductListItem `json:"items"`
	Page  *api.Page          `json:"page"`
}

func (r *ProductRepository) List(
	ctx context.Context,
	qe database.QueryExecutor,
	q *api.ListQuery,
) (*ProductList, error) {
	query := `
	SELECT
    p.id,
    p.title,
    p.slug,
    dv.price,
		dv.currency,
    thumb.url
FROM products p
LEFT JOIN product_variants dv
    ON dv.id = p.default_variant_id
    AND dv.deleted_at IS NULL
LEFT JOIN product_media thumb
    ON thumb.variant_id = dv.id
    AND thumb.is_primary = true
WHERE p.deleted_at IS NULL
ORDER BY p.created_at DESC
LIMIT $1 OFFSET $2	`

	offset := (q.Page - 1) * q.PageSize

	rows, err := qe.Query(
		ctx,
		query,
		q.PageSize,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list products: %w",
			err,
		)
	}
	defer rows.Close()

	var products []*ProductListItem

	for rows.Next() {
		var p ProductListItem

		err = rows.Scan(
			&p.ID,
			&p.Title,
			&p.Slug,
			&p.Price,
			&p.Currency,
			&p.ThumbnailURL,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"scan product: %w",
				err,
			)
		}

		products = append(products, &p)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate products: %w",
			err,
		)
	}

	var totalItems int

	err = qe.QueryRow(
		ctx,
		`
		SELECT COUNT(*)
		FROM products
		WHERE deleted_at IS NULL
		`,
	).Scan(&totalItems)
	if err != nil {
		return nil, fmt.Errorf(
			"count products: %w",
			err,
		)
	}

	totalPages := 0
	if totalItems > 0 {
		totalPages =
			(totalItems + q.PageSize - 1) /
				q.PageSize
	}

	page := &api.Page{
		Page:        q.Page,
		PageSize:    q.PageSize,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		HasPrevious: q.Page > 1,
		HasNext:     q.Page < totalPages,
	}

	return &ProductList{
		Items: products,
		Page:  page,
	}, nil
}

func (r *ProductRepository) SoftDelete(
	ctx context.Context,
	qe database.QueryExecutor,
	filter Filter,
) error {
	query := `
	UPDATE products
	SET deleted_at = NOW()
	WHERE deleted_at IS NULL`

	args := []any{}
	argID := 1
	if filter.ID != nil {
		query += fmt.Sprintf(" AND id=$%d", argID)
		args = append(args, *filter.ID)
		argID++
	}

	if filter.Slug != nil {
		query += fmt.Sprintf(" AND slug=$%d", argID)
		args = append(args, *filter.Slug)
	}

	_, err := qe.Exec(ctx,
		query,
		args...,
	)
	if err != nil {
		return fmt.Errorf(
			"soft delete a product:%w",
			err,
		)
	}
	return nil
}

func (r *ProductRepository) GetVariants(
	ctx context.Context,
	qe database.QueryExecutor,
	productID uuid.UUID,
) ([]*model.ProductVariant, error) {
	var variants []*model.ProductVariant
	rows, err := qe.Query(
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
		WHERE product_id=$1
		`,
		productID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get variants: %w",
			err,
		)
	}
	for rows.Next() {
		var variant model.ProductVariant
		err = rows.Scan(
			&variant.ID,
			&variant.ProductID,
			&variant.SKU,
			&variant.Price,
			&variant.Currency,
			&variant.CreatedAt,
			&variant.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"get variants: %w",
				err,
			)
		}
		variants = append(variants, &variant)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf(
			"get variants: %w",
			err,
		)
	}
	return variants, nil
}

func (r *ProductRepository) GetCategories(
	ctx context.Context,
	qe database.QueryExecutor,
	productID uuid.UUID,
) ([]*model.ProductCategory, error) {
	var categories []*model.ProductCategory
	rows, err := qe.Query(
		ctx,
		`
		SELECT
			c.id,
			c.name,
			c.parent_id,
			c.created_at,
			c.updated_at,
			c.deleted_at
		FROM product_categories pc
		JOIN categories c ON pc.category_id = c.id
		JOIN products ps ON pc.product_id = ps.id
		WHERE ps.id=$1 AND c.deleted_at IS NULL
		`,
		productID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get categories: %w",
			err,
		)
	}
	for rows.Next() {
		var category model.ProductCategory
		err = rows.Scan(
			&category.ID,
			&category.Name,
			&category.ParentID,
			&category.CreatedAt,
			&category.UpdatedAt,
			&category.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"get categories: %w",
				err,
			)
		}
		categories = append(categories, &category)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf(
			"get categories: %w",
			err,
		)
	}
	return categories, nil
}

func (r *ProductRepository) GetTags(
	ctx context.Context,
	qe database.QueryExecutor,
	productID uuid.UUID,
) ([]*model.ProductTag, error) {
	var tags []*model.ProductTag
	rows, err := qe.Query(
		ctx,
		`
		SELECT
			t.id,
			t.name,
			t.created_at,
			t.updated_at
		FROM product_tags pt
		JOIN tags t ON pt.tag_id = t.id
		JOIN products ps ON pt.product_id = ps.id
		WHERE ps.id=$1 AND t.deleted_at IS NULL
		`,
		productID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get tags: %w",
			err,
		)
	}
	for rows.Next() {
		var tag model.ProductTag
		err = rows.Scan(
			&tag.ID,
			&tag.Name,
			&tag.CreatedAt,
			&tag.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"get tags: %w",
				err,
			)
		}
		tags = append(tags, &tag)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf(
			"get tags: %w",
			err,
		)
	}
	return tags, nil
}

func (r *ProductRepository) SetDefaultVariant(
	ctx context.Context,
	qe database.QueryExecutor,
	productID uuid.UUID,
	variantID uuid.UUID,
) error {
	_, err := qe.Exec(
		ctx,
		`
		UPDATE products
		SET default_variant_id = $2
		WHERE id = $1
		`,
		productID,
		variantID,
	)
	if err != nil {
		return fmt.Errorf(
			"set default variant: %w",
			err,
		)
	}
	return nil
}

func (r *ProductRepository) PublishProduct(
	ctx context.Context,
	qe database.QueryExecutor,
	productID uuid.UUID,
) error {
	_, err := qe.Exec(
		ctx,
		`
		UPDATE products
		SET status = 'published'
		WHERE id = $1
		`,
		productID,
	)
	if err != nil {
		return fmt.Errorf(
			"publish product: %w",
			err,
		)
	}
	return nil
}

func (r *ProductRepository) ArchiveProduct(
	ctx context.Context,
	qe database.QueryExecutor,
	productID uuid.UUID,
) error {
	_, err := qe.Exec(
		ctx,
		`
		UPDATE products
		SET status = 'archived'
		WHERE id = $1
		`,
		productID,
	)
	if err != nil {
		return fmt.Errorf(
			"archive product: %w",
			err,
		)
	}
	return nil
}
