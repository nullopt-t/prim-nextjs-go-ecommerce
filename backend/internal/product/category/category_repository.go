package category

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

type Filter struct {
	ID *uuid.UUID
}

type CategoryRepository struct {
}

func NewRepository() *CategoryRepository {
	return &CategoryRepository{}
}

func (cr *CategoryRepository) Create(
	ctx context.Context,
	qe database.QueryExecutor,
	category *model.ProductCategory,
) error {
	_, err := qe.Exec(
		ctx,
		`
		INSERT INTO
		categories (
			id,
			name,
			slug,
			parent_id,
			publication_status,
			created_at,
			updated_at
		)
		VALUES(
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7
		)
		`,
		category.ID,
		category.Name,
		category.Slug,
		category.ParentID,
		model.PublicationStatusDraft,
		category.CreatedAt,
		category.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create category: %w", err)
	}
	return err
}

func (cr *CategoryRepository) Get(
	ctx context.Context,
	qe database.QueryExecutor,
	filter Filter,
) (*model.ProductCategory, error) {
	var category model.ProductCategory

	query :=
		`
			SELECT
				id,
				name,
				slug,
				parent_id,
				publication_status,
				created_at,
				updated_at
			FROM
				categories
			WHERE
			deleted_at IS NULL
		`

	args := []any{}
	argID := 1
	if filter.ID != nil {
		query += fmt.Sprintf(" AND id = $%d", argID)
		args = append(args, *filter.ID)
		argID++
	}

	err := qe.QueryRow(
		ctx,
		query,
		args...,
	).Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
		&category.ParentID,
		&category.PublicationStatus,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("get category: %w", err)
	}

	return &category, nil
}

type UpdateCategoryFields struct {
	Name              *string
	Slug              *string
	ParentID          *uuid.UUID
	PublicationStatus *model.PublicationStatus
}

func (cr *CategoryRepository) Update(
	ctx context.Context,
	qe database.QueryExecutor,
	categoryID uuid.UUID,
	fields UpdateCategoryFields,
) error {
	query := `
	UPDATE
		categories
	SET
		name = COALESCE($1, name),
		slug = COALESCE($2, slug),
		parent_id = COALESCE($3, parent_id),
		publication_status = COALESCE($4, publication_status)
	WHERE
		id = $5
	`
	_, err := qe.Exec(
		ctx,
		query,
		fields.Name,
		fields.Slug,
		fields.ParentID,
		fields.PublicationStatus,
		categoryID,
	)
	if err != nil {
		return fmt.Errorf("update category: %w", err)
	}

	return nil
}

func (cr *CategoryRepository) ListProductCategories(
	ctx context.Context,
	qe database.QueryExecutor,
	productID uuid.UUID,
) ([]*model.ProductCategory, error) {
	query := `
	SELECT
		c.id,
		c.name,
		c.slug,
		c.parent_id,
		c.publication_status,
		c.created_at,
		c.updated_at
	FROM
		product_categories pc
	JOIN
		categories c ON pc.category_id = c.id
	WHERE
		deleted_at IS NULL AND pc.product_id = $1
	`
	rows, err := qe.Query(
		ctx,
		query,
		productID,
	)
	if err != nil {
		return nil, fmt.Errorf("list product categories: %w", err)
	}
	defer rows.Close()

	var result = make([]*model.ProductCategory, 0)
	for rows.Next() {
		var category model.ProductCategory
		if err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Slug,
			&category.ParentID,
			&category.PublicationStatus,
			&category.CreatedAt,
			&category.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("list product categories: %w", err)
		}
		result = append(result, &category)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list product categories: %w", err)
	}
	return result, nil
}

func (cr *CategoryRepository) PutProductCategories(
	ctx context.Context,
	tx database.QueryExecutor,
	productID uuid.UUID,
	categoryIDs []uuid.UUID,
) error {
	query := `
	DELETE FROM
	product_categories
	WHERE product_id = $1
	`
	if _, err := tx.Exec(
		ctx,
		query,
		productID,
	); err != nil {
		return fmt.Errorf(
			"put product categories: %w",
			err,
		)
	}

	for _, categoryID := range categoryIDs {
		query := `
		INSERT INTO
		product_categories (
			product_id,
			category_id
		)
		VALUES (
			$1,
			$2
		)
		`
		if _, err := tx.Exec(
			ctx,
			query,
			productID,
			categoryID,
		); err != nil {
			return fmt.Errorf(
				"put product categories: %w",
				err,
			)
		}
	}
	return nil
}

func (cr *CategoryRepository) AdminList(
	ctx context.Context,
	qe database.QueryExecutor,
	q *api.ListQuery,
) ([]*model.ProductCategory, *api.Page, error) {
	var query strings.Builder
	var countQuery strings.Builder

	query.WriteString(`
		SELECT
			id,
			name,
			slug,
			parent_id,
			publication_status,
			created_at,
			updated_at
		FROM categories
		WHERE 1=1
		`)

	countQuery.WriteString(`
		SELECT COUNT(*)
		FROM categories
		WHERE 1=1
		`)

	args := make([]any, 0)
	argID := 1

	if q.Search != "" {
		condition := fmt.Sprintf(" AND name ILIKE $%d OR slug ILIKE $%d", argID, +1)
		query.WriteString(condition)
		countQuery.WriteString(condition)
		args = append(args, "%"+q.Search+"%")
		argID++
	}

	if len(q.Sort) > 0 {
		query.WriteString(" ORDER BY ")
		for i, sort := range q.Sort {
			field := sort.Field
			order := sort.Order
			fmt.Fprintf(&query, "%s %s", field, order)
			if i < len(q.Sort)-1 {
				query.WriteString(", ")
			}
		}
	} else {
		query.WriteString(" ORDER BY created_at DESC")
	}

	fmt.Fprintf(&query, " LIMIT $%d OFFSET $%d", argID, argID+1)

	var total int
	err := qe.QueryRow(
		ctx,
		countQuery.String(),
		args...,
	).Scan(&total)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"admin list categories: %w",
			err,
		)
	}

	queryArgs := append(slices.Clone(args), q.PageSize, (q.Page-1)*q.PageSize)
	var categories []*model.ProductCategory
	rows, err := qe.Query(
		ctx,
		query.String(),
		queryArgs...,
	)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"admin list categories: %w",
			err,
		)
	}
	defer rows.Close()

	for rows.Next() {
		var category model.ProductCategory
		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Slug,
			&category.ParentID,
			&category.PublicationStatus,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			return nil, nil, fmt.Errorf(
				"admin list categories: %w",
				err,
			)
		}
		categories = append(categories, &category)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf(
			"admin list categories: %w",
			err,
		)
	}

	return categories, &api.Page{
		Page:        q.Page,
		PageSize:    q.PageSize,
		TotalItems:  total,
		TotalPages:  (total + q.PageSize - 1) / q.PageSize,
		HasPrevious: q.Page > 1,
		HasNext:     q.Page < (total+q.PageSize-1)/q.PageSize,
	}, nil
}

func (cr *CategoryRepository) List(
	ctx context.Context,
	qe database.QueryExecutor,
	q *api.ListQuery,
) ([]*model.ProductCategory, *api.Page, error) {
	var query strings.Builder
	var countQuery strings.Builder

	query.WriteString(`
		SELECT
			id,
			name,
			slug
		FROM categories
		WHERE 1=1
		`)

	countQuery.WriteString(`
		SELECT COUNT(*)
		FROM categories
		WHERE 1=1
		`)

	args := make([]any, 0)
	argID := 1

	if q.Search != "" {
		condition := fmt.Sprintf(" AND name ILIKE $%d OR slug ILIKE $%d", argID, +1)
		query.WriteString(condition)
		countQuery.WriteString(condition)
		args = append(args, "%"+q.Search+"%")
		argID++
	}

	if len(q.Sort) > 0 {
		query.WriteString(" ORDER BY ")
		for i, sort := range q.Sort {
			field := sort.Field
			order := sort.Order
			fmt.Fprintf(&query, "%s %s", field, order)
			if i < len(q.Sort)-1 {
				query.WriteString(", ")
			}
		}
	} else {
		query.WriteString(" ORDER BY created_at DESC")
	}

	fmt.Fprintf(&query, " LIMIT $%d OFFSET $%d", argID, argID+1)

	var total int
	err := qe.QueryRow(
		ctx,
		countQuery.String(),
		args...,
	).Scan(&total)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"admin list categories: %w",
			err,
		)
	}

	queryArgs := append(slices.Clone(args), q.PageSize, (q.Page-1)*q.PageSize)
	var categories []*model.ProductCategory
	rows, err := qe.Query(
		ctx,
		query.String(),
		queryArgs...,
	)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"admin list categories: %w",
			err,
		)
	}
	defer rows.Close()

	for rows.Next() {
		var category model.ProductCategory
		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Slug,
		)
		if err != nil {
			return nil, nil, fmt.Errorf(
				"admin list categories: %w",
				err,
			)
		}
		categories = append(categories, &category)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf(
			"admin list categories: %w",
			err,
		)
	}

	return categories, &api.Page{
		Page:        q.Page,
		PageSize:    q.PageSize,
		TotalItems:  total,
		TotalPages:  (total + q.PageSize - 1) / q.PageSize,
		HasPrevious: q.Page > 1,
		HasNext:     q.Page < (total+q.PageSize-1)/q.PageSize,
	}, nil
}
