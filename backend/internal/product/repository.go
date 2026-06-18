package product

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"

	"github.com/google/uuid"
)

type Filter struct {
	ID uuid.UUID
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
) (uuid.UUID, error) {
	var id uuid.UUID
	err := qe.QueryRow(ctx,
		`INSERT INTO products (
			title,
	 		short_description,
			description,
			sku,
			slug,
			status,
			price,
			currency)
		VALUES (
			$1, $2, $3, $4,
		 	$5, $6, $7, $8
		)
		RETURNING id`,
		p.Title,
		p.ShortDescription,
		p.Description,
		p.SKU,
		p.Slug,
		p.Status,
		p.Price,
		p.Currency,
	).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf(
			"create product :%w",
			err,
		)
	}
	return id, nil
}

func (r *ProductRepository) Get(
	ctx context.Context,
	qe database.QueryExecutor,
	filter Filter,
) (*model.Product, error) {
	var p model.Product
	query := `SELECT id,
					title,
				 	price,
					created_at,
				 	updated_at
					FROM products
					WHERE deleted_at IS NULL`

	args := []any{}
	argID := 1
	if filter.ID != uuid.Nil {
		query += fmt.Sprintf(" AND id=$%d", argID)
		args = append(args, filter.ID)
	}

	err := qe.QueryRow(ctx,
		query,
		args...,
	).Scan(&p.ID,
		&p.Title,
		&p.Price,
		&p.CreatedAt,
		&p.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *ProductRepository) GetAll(
	ctx context.Context,
	qe database.QueryExecutor,
	q api.PageQuery,
) ([]*model.Product, *api.Page, error) {
	offset := (q.Page - 1) * q.PageSize

	query := `
		SELECT
			id,
			title,
			short_description,
			description,
			sku,
			slug,
			status,
			price,
			currency,
			created_at,
			updated_at
		FROM products
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := qe.Query(ctx, query, q.PageSize, offset)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var total int
	err = qe.QueryRow(
		ctx,
		`SELECT COUNT(*) FROM products WHERE deleted_at IS NULL`,
	).Scan(&total)
	if err != nil {
		return nil, nil, err
	}

	len := min(total, q.PageSize)
	products := make([]*model.Product, 0, len)
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(
			&p.ID,
			&p.Title,
			&p.ShortDescription,
			&p.Description,
			&p.SKU,
			&p.Slug,
			&p.Status,
			&p.Price,
			&p.Currency,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, nil, err
		}

		products = append(products, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	page := &api.Page{
		Page:     q.Page,
		PageSize: q.PageSize,
		Total:    total,
	}

	return products, page, nil
}

func (r *ProductRepository) Delete(
	ctx context.Context,
	qe database.QueryExecutor,
	filter Filter,
) error {
	query := `UPDATE products
					SET deleted_at = NOW()
				 	WHERE deleted_at IS NULL`

	args := []any{}

	if filter.ID != uuid.Nil {
		query += " AND id = $1"
		args = append(args, filter.ID)
	}

	res, err := qe.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return sql.ErrNoRows
	}

	return nil
}
