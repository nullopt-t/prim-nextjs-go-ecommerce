package product

import (
	"context"
	"fmt"
	"time"

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
		)
		`,
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
	ID               uuid.UUID
	Title            string
	ShortDescription string
	Slug             string
	Status           model.ProductStatus

	BrandName    string
	ThumbnailURL string

	Price int64

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (r *ProductRepository) List(
	ctx context.Context,
	qe database.QueryExecutor,
	q *api.ListQuery,
) ([]*ProductListItem, *api.Page, error) {
	query := `
		SELECT
			p.id,
			p.title,
			p.short_description,
			p.slug,
			p.status,

			b.name,

			MIN(pv.price) AS starting_price,

			(
				SELECT pm.url
				FROM product_media pm
				WHERE pm.product_id = p.id
				ORDER BY pm.sort_order
				LIMIT 1
			) AS thumbnail_url,

			p.created_at,
			p.updated_at
		FROM products p
		JOIN brands b
			ON b.id = p.brand_id
		LEFT JOIN product_variants pv
			ON pv.product_id = p.id
		WHERE p.deleted_at IS NULL
		GROUP BY
			p.id,
			b.name
		ORDER BY p.created_at DESC
		LIMIT $1 OFFSET $2;
	`

	rows, err := qe.Query(
		ctx,
		query,
		q.PageSize,
		(q.Page-1)*q.PageSize,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()

	var products []*ProductListItem

	for rows.Next() {
		var p ProductListItem

		if err := rows.Scan(
			&p.ID,
			&p.Title,
			&p.ShortDescription,
			&p.Slug,
			&p.Status,
			&p.BrandName,
			&p.Price,
			&p.ThumbnailURL,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, nil, fmt.Errorf("scan product: %w", err)
		}

		products = append(products, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("iterate products: %w", err)
	}

	var totalItems int
	err = qe.QueryRow(
		ctx,
		`
		SELECT COUNT(id)
		FROM products
		WHERE deleted_at IS NULL
		`).Scan(&totalItems)
	if err != nil {
		return nil, nil, fmt.Errorf("count products: %w", err)
	}

	totalPages := (len(products) + q.PageSize - 1) / q.PageSize
	page := &api.Page{
		Page:        q.Page,
		PageSize:    q.PageSize,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		HasPrevious: q.Page > 1,
		HasNext:     q.Page < totalPages,
	}

	return products, page, nil
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
