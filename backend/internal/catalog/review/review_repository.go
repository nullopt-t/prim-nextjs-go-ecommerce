package review

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

type ReviewRepository struct{}

func NewReviewRepository() *ReviewRepository {
	return &ReviewRepository{}
}

// PurchaseVerification contains customer ownership, product association, and order lifecycle status
// used to validate whether an order item can be reviewed.
type PurchaseVerification struct {
	CustomerID  *uuid.UUID
	ProductID   uuid.UUID
	OrderStatus model.OrderStatus
}

// VerifyPurchase checks the database for an order item, resolving its buyer customer ID,
// parent product ID, and parent order status.
func (r *ReviewRepository) VerifyPurchase(
	ctx context.Context,
	orderItemID uuid.UUID,
) (*PurchaseVerification, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

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
	err = executor.QueryRowContext(ctx, query, orderItemID).Scan(&pv.CustomerID, &pv.ProductID, &pv.OrderStatus)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &pv, nil
}

// FindEligibleOrderItem finds a delivered order item for the given customer and product that hasn't been reviewed yet,
// locking the order item row (FOR UPDATE) to prevent concurrency races on review creation.
func (r *ReviewRepository) FindEligibleOrderItem(
	ctx context.Context,
	userID uuid.UUID,
	productID uuid.UUID,
) (uuid.UUID, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return uuid.Nil, err
	}

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
		FOR UPDATE OF oi
	`
	var orderItemID uuid.UUID
	err = executor.QueryRowContext(ctx, query, userID, productID).Scan(&orderItemID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return uuid.Nil, nil
		}
		return uuid.Nil, err
	}
	return orderItemID, nil
}

func (r *ReviewRepository) Create(ctx context.Context, rv *model.Review) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO reviews (
			id,
			product_id,
			user_id,
			user_name,
			order_item_id,
			rating,
			title,
			body,
			status,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now(), now())
		RETURNING created_at, updated_at
	`

	return tx.QueryRowContext(
		ctx,
		query,
		rv.ID,
		rv.ProductID,
		rv.UserID,
		rv.AuthorName,
		rv.OrderItemID,
		rv.Rating,
		rv.Title,
		rv.Body,
		rv.Status,
	).Scan(&rv.CreatedAt, &rv.UpdatedAt)
}

func (r *ReviewRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Review, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT
			id,
			product_id,
			user_id,
			user_name,
			order_item_id,
			rating,
			title,
			body,
			status,
			helpful_count,
			created_at,
			updated_at
		FROM reviews
		WHERE id = $1
	`
	rv := &model.Review{}
	err = executor.QueryRowContext(ctx, query, id).Scan(
		&rv.ID,
		&rv.ProductID,
		&rv.UserID,
		&rv.AuthorName,
		&rv.OrderItemID,
		&rv.Rating,
		&rv.Title,
		&rv.Body,
		&rv.Status,
		&rv.HelpfulCount,
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

func (r *ReviewRepository) GetByOrderItemID(ctx context.Context, orderItemID uuid.UUID) (*model.Review, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

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
			helpful_count,
			created_at,
			updated_at
		FROM reviews
		WHERE order_item_id = $1
	`
	rv := &model.Review{}
	err = executor.QueryRowContext(ctx, query, orderItemID).Scan(
		&rv.ID,
		&rv.ProductID,
		&rv.UserID,
		&rv.OrderItemID,
		&rv.Rating,
		&rv.Title,
		&rv.Body,
		&rv.Status,
		&rv.HelpfulCount,
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

func (r *ReviewRepository) Update(ctx context.Context, rv *model.Review) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}

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
	return tx.QueryRowContext(ctx, query, rv.Rating, rv.Title, rv.Body, rv.Status, rv.ID).Scan(&rv.UpdatedAt)
}

func (r *ReviewRepository) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	status model.ReviewStatus,
) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}

	query := `
		UPDATE reviews
		SET
			status = $1,
			updated_at = now()
		WHERE id = $2
	`
	_, err = tx.ExecContext(ctx, query, status, id)
	return err
}

func (r *ReviewRepository) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}

	query := `DELETE FROM reviews WHERE id = $1`
	_, err = tx.ExecContext(ctx, query, id)
	return err
}

func (r *ReviewRepository) GetRatingSummaryByProductID(
	ctx context.Context,
	productID uuid.UUID,
) (*model.RatingSummary, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

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
	if err := executor.QueryRowContext(ctx, query, productID).Scan(&avg, &count); err != nil {
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
		rows, err := executor.QueryContext(ctx, distQuery, productID)
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
	productIDs []uuid.UUID,
) (map[uuid.UUID]*model.RatingSummary, error) {
	result := make(map[uuid.UUID]*model.RatingSummary, len(productIDs))
	if len(productIDs) == 0 {
		return result, nil
	}

	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
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
	rows, err := executor.QueryContext(ctx, query, productIDs)
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
	distRows, err := executor.QueryContext(ctx, distQuery, productIDs)
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

// List returns the reviews matching the filter, newest first, plus the cursor
// for the next page. Paging is keyset-based: the boundary row's (created_at,
// id) is carried in the cursor, so concurrent inserts cannot shift rows
// between pages the way OFFSET paging does.
func (r *ReviewRepository) List(
	ctx context.Context,
	limit int,
	cursor *pagination.Keyset,
	filter model.ReviewFilter,
) ([]model.Review, *pagination.CursorPage, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, nil, err
	}

	var filters []string
	var args []any
	argID := 1

	if filter.ProductID != nil {
		filters = append(filters, fmt.Sprintf("r.product_id = $%d", argID))
		args = append(args, *filter.ProductID)
		argID++
	}

	if filter.UserID != nil {
		filters = append(filters, fmt.Sprintf("r.user_id = $%d", argID))
		args = append(args, *filter.UserID)
		argID++
	}

	if filter.Status != nil {
		filters = append(filters, fmt.Sprintf("r.status = $%d", argID))
		args = append(args, *filter.Status)
		argID++
	}

	if filter.Rating != nil {
		filters = append(filters, fmt.Sprintf("r.rating = $%d", argID))
		args = append(args, *filter.Rating)
		argID++
	}

	if cursor != nil {
		clause, cursorArgs := pagination.KeysetClause("r", cursor, argID)
		filters = append(filters, strings.TrimPrefix(clause, " AND "))
		args = append(args, cursorArgs...)
		argID += len(cursorArgs)
	}

	whereClause := ""
	if len(filters) > 0 {
		whereClause = "WHERE " + strings.Join(filters, " AND ")
	}

	var voteSelect string
	var voteJoin string
	if filter.VoterID != nil {
		voteSelect = ", COALESCE(rv_vote.user_id IS NOT NULL, false) AS user_voted"
		voteJoin = fmt.Sprintf("LEFT JOIN review_votes rv_vote ON rv_vote.review_id = r.id AND rv_vote.user_id = $%d", argID)
		args = append(args, *filter.VoterID)
		argID++
	} else {
		voteSelect = ", false AS user_voted"
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
			r.helpful_count,
			r.created_at,
			r.updated_at,
			r.user_name,
			p.title%s
		FROM reviews r
		INNER JOIN products p ON p.id = r.product_id
		%s
		%s%s
		LIMIT $%d
	`, voteSelect, voteJoin, whereClause, pagination.KeysetOrderBy("r"), argID)

	// One extra row tells us whether another page exists without a COUNT.
	args = append(args, limit+1)

	rows, err := executor.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, nil, err
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
			&rv.HelpfulCount,
			&rv.CreatedAt,
			&rv.UpdatedAt,
			&rv.AuthorName,
			&rv.ProductTitle,
			&rv.UserVoted,
		); err != nil {
			return nil, nil, err
		}
		reviews = append(reviews, rv)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	trimmed, page := pagination.NewCursorPageFrom(reviews, limit, func(rv model.Review) pagination.Keyset {
		return pagination.Keyset{CreatedAt: rv.CreatedAt, ID: rv.ID}
	})

	return trimmed, page, nil
}

// ToggleVote toggles a helpful vote on a review by the given user.
// Returns whether the user's vote is active (true if voted, false if unvoted),
// the updated helpful count, or an error.
func (r *ReviewRepository) ToggleVote(
	ctx context.Context,
	reviewID, userID uuid.UUID,
) (bool, int, error) {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return false, 0, err
	}

	var currentCount int
	err = tx.QueryRowContext(ctx, "SELECT helpful_count FROM reviews WHERE id = $1 FOR UPDATE", reviewID).Scan(&currentCount)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, 0, database.ErrNotFound
		}
		return false, 0, err
	}

	tag, err := tx.ExecContext(ctx, "DELETE FROM review_votes WHERE review_id = $1 AND user_id = $2", reviewID, userID)
	if err != nil {
		return false, 0, err
	}

	if tag.RowsAffected() > 0 {
		var newCount int
		err := tx.QueryRowContext(ctx, `
			UPDATE reviews
			SET helpful_count = GREATEST(helpful_count - 1, 0), updated_at = now()
			WHERE id = $1
			RETURNING helpful_count
		`, reviewID).Scan(&newCount)
		if err != nil {
			return false, 0, err
		}
		return false, newCount, nil
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO review_votes (review_id, user_id, created_at) VALUES ($1, $2, now())", reviewID, userID)
	if err != nil {
		return false, 0, err
	}

	var newCount int
	err = tx.QueryRowContext(ctx, `
		UPDATE reviews
		SET helpful_count = helpful_count + 1, updated_at = now()
		WHERE id = $1
		RETURNING helpful_count
	`, reviewID).Scan(&newCount)
	if err != nil {
		return false, 0, err
	}

	return true, newCount, nil
}
