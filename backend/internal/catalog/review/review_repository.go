package review

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/pagination"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)


var allowedReviewSortFields = map[string]string{
	"id":        "r.id",
	"rating":    "r.rating",
	"status":    "r.status",
	"createdAt": "r.created_at",
	"updatedAt": "r.updated_at",
}

type ReviewRepository struct{}

func NewReviewRepository() *ReviewRepository {
	return &ReviewRepository{}
}

type PurchaseVerification struct {
	CustomerID  *uuid.UUID
	ProductID   uuid.UUID
	OrderStatus model.OrderStatus
}

func (r *ReviewRepository) VerifyPurchase(
	ctx context.Context,
	qe database.QueryExecutor,
	orderItemID uuid.UUID,
) (*PurchaseVerification, error) {
	query := `
		SELECT
			o.customer_id,
			pv.product_id,
			o.status
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		JOIN product_variants pv ON pv.id = oi.variant_id
		WHERE oi.id = $1
	`
	var pv PurchaseVerification
	err := qe.QueryRow(ctx, query, orderItemID).Scan(&pv.CustomerID, &pv.ProductID, &pv.OrderStatus)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &pv, nil
}

// FindEligibleOrderItem finds a delivered order item for the given customer and product that hasn't been reviewed yet.
func (r *ReviewRepository) FindEligibleOrderItem(
	ctx context.Context,
	qe database.QueryExecutor,
	userID uuid.UUID,
	productID uuid.UUID,
) (uuid.UUID, error) {
	query := `
		SELECT oi.id
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		JOIN product_variants pv ON pv.id = oi.variant_id
		LEFT JOIN reviews rev ON rev.order_item_id = oi.id
		WHERE o.customer_id = $1
		  AND pv.product_id = $2
		  AND o.status = 'delivered'
		  AND rev.id IS NULL
		ORDER BY o.created_at DESC
		LIMIT 1
	`
	var orderItemID uuid.UUID
	err := qe.QueryRow(ctx, query, userID, productID).Scan(&orderItemID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return uuid.Nil, nil
		}
		return uuid.Nil, err
	}
	return orderItemID, nil
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

func (r *ReviewRepository) GetByOrderItemID(ctx context.Context, qe database.QueryExecutor, orderItemID uuid.UUID) (*model.Review, error) {
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
		WHERE order_item_id = $1
	`
	rv := &model.Review{}
	err := qe.QueryRow(ctx, query, orderItemID).Scan(
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

func (r *ReviewRepository) Update(ctx context.Context, qe database.QueryExecutor, rv *model.Review) error {
	query := `
		UPDATE reviews
		SET
			rating = $1,
			title = $2,
			body = $3,
			status = $4,
			updated_at = now()
		WHERE id = $5
		RETURNING updated_at
	`
	return qe.QueryRow(ctx, query, rv.Rating, rv.Title, rv.Body, rv.Status, rv.ID).Scan(&rv.UpdatedAt)
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

func (r *ReviewRepository) GetRatingSummaryByProductID(
	ctx context.Context,
	qe database.QueryExecutor,
	productID uuid.UUID,
) (*model.RatingSummary, error) {
	summary := &model.RatingSummary{
		Distribution: map[int16]int{
			1: 0,
			2: 0,
			3: 0,
			4: 0,
			5: 0,
		},
	}

	query := `
		SELECT
			COALESCE(AVG(rating), 0)::float8,
			COUNT(1)
		FROM reviews
		WHERE product_id = $1 AND status = 'approved'
	`
	var avg float64
	var count int
	if err := qe.QueryRow(ctx, query, productID).Scan(&avg, &count); err != nil {
		return nil, err
	}
	summary.AverageRating = math.Round(avg*10) / 10
	summary.ReviewCount = count

	if count > 0 {
		distQuery := `
			SELECT rating, COUNT(1)
			FROM reviews
			WHERE product_id = $1 AND status = 'approved'
			GROUP BY rating
		`
		rows, err := qe.Query(ctx, distQuery, productID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var rating int16
			var rCount int
			if err := rows.Scan(&rating, &rCount); err != nil {
				return nil, err
			}
			summary.Distribution[rating] = rCount
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	return summary, nil
}

func (r *ReviewRepository) GetRatingSummariesByProductIDs(
	ctx context.Context,
	qe database.QueryExecutor,
	productIDs []uuid.UUID,
) (map[uuid.UUID]*model.RatingSummary, error) {
	result := make(map[uuid.UUID]*model.RatingSummary, len(productIDs))
	if len(productIDs) == 0 {
		return result, nil
	}

	for _, pid := range productIDs {
		result[pid] = &model.RatingSummary{
			Distribution: map[int16]int{
				1: 0,
				2: 0,
				3: 0,
				4: 0,
				5: 0,
			},
		}
	}

	query := `
		SELECT
			product_id,
			COALESCE(AVG(rating), 0)::float8,
			COUNT(1)
		FROM reviews
		WHERE product_id = ANY($1) AND status = 'approved'
		GROUP BY product_id
	`
	rows, err := qe.Query(ctx, query, productIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var pid uuid.UUID
		var avg float64
		var count int
		if err := rows.Scan(&pid, &avg, &count); err != nil {
			return nil, err
		}
		if s, ok := result[pid]; ok {
			s.AverageRating = math.Round(avg*10) / 10
			s.ReviewCount = count
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	distQuery := `
		SELECT
			product_id,
			rating,
			COUNT(1)
		FROM reviews
		WHERE product_id = ANY($1) AND status = 'approved'
		GROUP BY product_id, rating
	`
	distRows, err := qe.Query(ctx, distQuery, productIDs)
	if err != nil {
		return nil, err
	}
	defer distRows.Close()

	for distRows.Next() {
		var pid uuid.UUID
		var rating int16
		var rCount int
		if err := distRows.Scan(&pid, &rating, &rCount); err != nil {
			return nil, err
		}
		if s, ok := result[pid]; ok {
			s.Distribution[rating] = rCount
		}
	}
	if err := distRows.Err(); err != nil {
		return nil, err
	}

	return result, nil
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
		filters = append(filters, fmt.Sprintf("r.product_id = $%d", argID))
		args = append(args, *productID)
		argID++
	}

	if userID != nil {
		filters = append(filters, fmt.Sprintf("r.user_id = $%d", argID))
		args = append(args, *userID)
		argID++
	}

	if status != nil {
		filters = append(filters, fmt.Sprintf("r.status = $%d", argID))
		args = append(args, *status)
		argID++
	}

	whereClause := ""
	if len(filters) > 0 {
		whereClause = "WHERE " + strings.Join(filters, " AND ")
	}

	// Count query
	countQuery := fmt.Sprintf(`SELECT COUNT(1) FROM reviews r %s`, whereClause)
	var total int
	if err := qe.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []model.Review{}, 0, nil
	}

	sortField := "r.created_at"
	sortOrder := "DESC"
	if len(q.Sort) > 0 {
		sf := q.Sort[0].Field
		if allowed, ok := allowedReviewSortFields[sf]; ok {
			sortField = allowed
		}
		if q.Sort[0].Order == pagination.SortAsc {
			sortOrder = "ASC"
		} else {
			sortOrder = "DESC"
		}
	}

	listQuery := fmt.Sprintf(`
		SELECT
			r.id,
			r.product_id,
			r.user_id,
			r.order_item_id,
			r.rating,
			r.title,
			r.body,
			r.status,
			r.created_at,
			r.updated_at
		FROM reviews r
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

