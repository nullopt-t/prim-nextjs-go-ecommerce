package tag

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

type TagRepository struct {
}

func NewRepository() *TagRepository {
	return &TagRepository{}
}

func (tr *TagRepository) Create(
	ctx context.Context,
	qe database.QueryExecutor,
	tag *model.ProductTag,
) error {

	now := time.Now().UTC()
	if tag.CreatedAt.IsZero() {
		tag.CreatedAt = now
	}

	if tag.UpdatedAt.IsZero() {
		tag.UpdatedAt = now
	}

	query := `
	INSERT INTO tags(
		id,
		name,
		publication_status,
		created_at,
		updated_at
	)
	VALUES (
		$1,
		$2,
		$3,
		$4,
		$5
	)
	`
	_, err := qe.Exec(
		ctx,
		query,
		tag.ID,
		tag.Name,
		model.PublicationStatusDraft,
		tag.CreatedAt,
		tag.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create tag:%w", err)
	}

	return nil
}

type Filter struct {
	ID   *uuid.UUID
	Name *string
}

func (tr *TagRepository) Get(
	ctx context.Context,
	qe database.QueryExecutor,
	filter *Filter,
) (*model.ProductTag, error) {
	query := `
	SELECT
		id,
		name,
		created_at,
		updated_at
	FROM
		tags
	WHERE deleted_at IS NULL
	`

	args := []any{}
	argID := 1
	if filter.ID != nil {
		query += fmt.Sprintf(" AND id = $%d", argID)
		args = append(args, *filter.ID)
		argID++
	}

	if filter.Name != nil {
		query += fmt.Sprintf(" AND name = $%d", argID)
		args = append(args, *filter.Name)
	}

	var tag model.ProductTag
	err := qe.QueryRow(
		ctx,
		query,
		args...,
	).Scan(
		&tag.ID,
		&tag.Name,
		&tag.CreatedAt,
		&tag.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get a product by id: %w", err)
	}

	return &tag, nil
}

func (tr *TagRepository) AdminList(
	ctx context.Context,
	qe database.QueryExecutor,
	q *api.ListQuery,
) ([]*model.ProductTag, *api.Page, error) {
	var query strings.Builder
	var countQuery strings.Builder

	query.WriteString(`
		SELECT
			id,
			name,
			publication_status,
			created_at,
			updated_at,
			deleted_at
		FROM
			tags
		WHERE 1 = 1
	`)

	countQuery.WriteString(`
			SELECT
				COUNT(*)
			FROM
				tags
			WHERE 1 = 1
		`)

	args := []any{}
	argID := 1

	if q.Search != "" {
		condition := fmt.Sprintf(" AND name ILIKE $%d", argID)
		query.WriteString(condition)
		countQuery.WriteString(condition)
		args = append(args, fmt.Sprintf("%%%s%%", q.Search))
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

	queryArgs := append(slices.Clone(args), q.Page, q.Offset)

	var total int
	err := qe.QueryRow(
		ctx,
		countQuery.String(),
		args...,
	).Scan(&total)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"list tags: %w",
			err,
		)
	}

	rows, err := qe.Query(
		ctx,
		query.String(),
		queryArgs...,
	)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"list tags: %w",
			err,
		)
	}
	defer rows.Close()

	var tags []*model.ProductTag
	for rows.Next() {
		var tag model.ProductTag
		err := rows.Scan(
			&tag.ID,
			&tag.Name,
			&tag.PublicationStatus,
			&tag.CreatedAt,
			&tag.UpdatedAt,
			&tag.DeletedAt,
		)
		if err != nil {
			return nil, nil, fmt.Errorf(
				"list tags: %w",
				err,
			)
		}
		tags = append(tags, &tag)
	}

	return tags, &api.Page{
		Page:        q.Page,
		PageSize:    q.PageSize,
		TotalItems:  total,
		TotalPages:  (total + q.PageSize - 1) / q.PageSize,
		HasPrevious: q.Page > 1,
		HasNext:     q.Page*q.PageSize < total,
	}, nil
}

func (tr *TagRepository) List(
	ctx context.Context,
	qe database.QueryExecutor,
	q *api.ListQuery,
) ([]*model.ProductTag, *api.Page, error) {
	var query strings.Builder
	var countQuery strings.Builder

	query.WriteString(`
		SELECT
			id,
			name
		FROM
			tags
		WHERE deleted_at IS NULL
			AND publication_status = 'published'
	`)

	countQuery.WriteString(`
			SELECT
				COUNT(*)
			FROM
				tags
			WHERE deleted_at IS NULL
				AND publication_status = 'published'
		`)

	args := []any{}
	argID := 1

	if q.Search != "" {
		condition := fmt.Sprintf(" AND name ILIKE $%d", argID)
		query.WriteString(condition)
		countQuery.WriteString(condition)
		args = append(args, fmt.Sprintf("%%%s%%", q.Search))
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
			"list tags: %w",
			err,
		)
	}

	queryArgs := append(args, q.Page, q.PageSize)

	rows, err := qe.Query(
		ctx,
		query.String(),
		queryArgs...,
	)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"list tags: %w",
			err,
		)
	}
	defer rows.Close()

	var tags []*model.ProductTag
	for rows.Next() {
		var tag model.ProductTag
		err := rows.Scan(&tag.ID, &tag.Name, &tag.CreatedAt, &tag.UpdatedAt, &tag.DeletedAt)
		if err != nil {
			return nil, nil, fmt.Errorf(
				"list tags: %w",
				err,
			)
		}
		tags = append(tags, &tag)
	}

	return tags, &api.Page{
		Page:        q.Page,
		PageSize:    q.PageSize,
		TotalItems:  total,
		TotalPages:  (total + q.PageSize - 1) / q.PageSize,
		HasPrevious: q.Page > 1,
		HasNext:     q.Page*q.PageSize < total,
	}, nil
}

func (tr *TagRepository) PutProductTags(
	ctx context.Context,
	tx database.QueryExecutor,
	productID uuid.UUID,
	tagsIDs []uuid.UUID,
) error {
	_, err := tx.Exec(
		ctx,
		`DELETE FROM
			product_tags
		WHERE product_id = $1`,
		productID,
	)
	if err != nil {
		return fmt.Errorf(
			"put product tags: %w",
			err,
		)
	}

	for _, tagID := range tagsIDs {
		_, err := tx.Exec(
			ctx,
			`INSERT INTO product_tags (
				product_id,
				tag_id
			) VALUES (
				$1,
				$2
			)`,
			productID,
			tagID,
		)
		if err != nil {
			return fmt.Errorf(
				"put product tags: %w",
				err,
			)
		}
	}
	return nil
}

func (tr *TagRepository) ListProductTags(
	ctx context.Context,
	qe database.QueryExecutor,
	productID uuid.UUID,
) ([]*model.ProductTag, error) {
	query := `
	SELECT
		id,
		name,
		created_at,
		updated_at
	FROM
		product_tags pt
	JOIN
		tags ts ON pt.tag_id = ts.id
	WHERE
		deleted_at IS NULL
	AND
		pt.product_id = $1
	`
	rows, err := qe.Query(
		ctx,
		query,
		productID,
	)
	if err != nil {
		return nil, fmt.Errorf("list product tags: %w", err)
	}

	tags := make([]*model.ProductTag, 0)
	for rows.Next() {
		var tag model.ProductTag
		err := rows.Scan(
			&tag.ID,
			&tag.Name,
			&tag.CreatedAt,
			&tag.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("list product tags: %w", err)
		}

		tags = append(
			tags,
			&tag,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list product tags: %w", err)
	}

	return tags, nil
}
