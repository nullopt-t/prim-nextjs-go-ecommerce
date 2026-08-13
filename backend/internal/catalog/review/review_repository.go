package review

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/pagination"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

var allowedReviewSortFields = map[string]string{
	"id":         "r.id",
	"rating":     "r.rating",
	"status":     "r.status",
	"created_at": "r.created_at",
	"updated_at": "r.updated_at",
}

type ReviewRepository struct{}

func NewReviewRepository() *ReviewRepository {
	return &ReviewRepository{}
}

func (r *ReviewRepository) Create(ctx context.Context, qe database.QueryExecutor, rv *model.Review) error {
	query := `
		INSERT INTO reviews (
			id,
			product_id,
			user_id,
			order_item_id,
			rating,
			title,
			body,
			status,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now(), now())
		RETURNING created_at, updated_at
	`

	err := qe.QueryRow(
		ctx,
		query,
		rv.ID,
		rv.ProductID,
		rv.UserID,
		rv.OrderItemID,
		rv.Rating,
		rv.Title,
		rv.Body,
		rv.Status,
	).Scan(&rv.CreatedAt, &rv.UpdatedAt)

	return err
}

func (r *ReviewRepository) GetByID(ctx context.Context, qe database.QueryExecutor, id uuid.UUID) (*model.Review, error) {
	query := `
		SELECT
			id,
			product_id,
			user_id,
			order_item_id,
			rating,
			title,
			body,
			status,
			created_at,
			updated_at
		FROM reviews
		WHERE id = $1
	`
	rv := &model.Review{}
	err := qe.QueryRow(ctx, query, id).Scan(
		&rv.ID,
		&rv.ProductID,
		&rv.UserID,
		&rv.OrderItemID,
		&rv.Rating,
		&rv.Title,
		&rv.Body,
		&rv.Status,
		&rv.CreatedAt,
		&rv.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return rv, nil
}

func (r *ReviewRepository) UpdateStatus(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
	status model.ReviewStatus,
) error {
	query := `
		UPDATE reviews
		SET
			status = $1,
			updated_at = now()
		WHERE id = $2
	`
	_, err := qe.Exec(ctx, query, status, id)
	return err
}

func (r *ReviewRepository) Delete(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
) error {
	query := `DELETE FROM reviews WHERE id = $1`
	_, err := qe.Exec(ctx, query, id)
	return err
}

func (r *ReviewRepository) List(
	ctx context.Context,
	qe database.QueryExecutor,
	q *pagination.ListQuery,
	productID *uuid.UUID,
	userID *uuid.UUID,
	status *model.ReviewStatus,
) ([]model.Review, int, error) {
	var filters []string
	var args []any
	argID := 1

	if productID != nil {
		filters = append(filters, fmt.Sprintf("product_id = $%d", argID))
		args = append(args, *productID)
		argID++
	}

	if userID != nil {
		filters = append(filters, fmt.Sprintf("user_id = $%d", argID))
		args = append(args, *userID)
		argID++
	}

	if status != nil {
		filters = append(filters, fmt.Sprintf("status = $%d", argID))
		args = append(args, *status)
		argID++
	}

	whereClause := ""
	if len(filters) > 0 {
		whereClause = "WHERE " + strings.Join(filters, " AND ")
	}

	// Count query
	countQuery := fmt.Sprintf(`SELECT COUNT(1) FROM reviews %s`, whereClause)
	var total int
	if err := qe.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []model.Review{}, 0, nil
	}

	sortField := "created_at"
	sortOrder := "DESC"
	if len(q.Sort) > 0 {
		sf := q.Sort[0].Field
		if allowed, ok := allowedReviewSortFields[sf]; ok {
			sortField = allowed
		}
		if q.Sort[0].Order == pagination.SortDesc {
			sortOrder = "DESC"
		} else {
			sortOrder = "ASC"
		}
	}

	listQuery := fmt.Sprintf(`
		SELECT
			id,
			product_id,
			user_id,
			order_item_id,
			rating,
			title,
			body,
			status,
			created_at,
			updated_at
		FROM reviews
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortField, sortOrder, argID, argID+1)

	args = append(args, q.PageSize, q.Offset)

	rows, err := qe.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reviews []model.Review
	for rows.Next() {
		var rv model.Review
		if err := rows.Scan(
			&rv.ID,
			&rv.ProductID,
			&rv.UserID,
			&rv.OrderItemID,
			&rv.Rating,
			&rv.Title,
			&rv.Body,
			&rv.Status,
			&rv.CreatedAt,
			&rv.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		reviews = append(reviews, rv)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return reviews, total, nil
}
