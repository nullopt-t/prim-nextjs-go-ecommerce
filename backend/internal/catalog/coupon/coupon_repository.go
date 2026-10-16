package coupon

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

type ListCouponFilter struct {
	Status *model.CouponStatus
}

type CouponRepository struct{}

func NewRepository() *CouponRepository {
	return &CouponRepository{}
}

// GetByID finds a coupon by its UUID identifier.
func (r *CouponRepository) GetByID(ctx context.Context, qe database.QueryExecutor, id uuid.UUID) (*model.Coupon, error) {
	query := `
		SELECT id, code, discount_type, discount_value, currency, min_cart_amount,
		       usage_limit, per_user_limit, starts_at, expires_at, status, created_at, updated_at
		FROM coupons
		WHERE id = $1
	`
	var (
		c             model.Coupon
		minCartAmount sql.NullInt64
		usageLimit    sql.NullInt32
		perUserLimit  sql.NullInt32
	)

	err := qe.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.Code, &c.DiscountType, &c.DiscountValue, &c.Currency,
		&minCartAmount, &usageLimit, &perUserLimit, &c.StartsAt, &c.ExpiresAt,
		&c.Status, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return nil, database.ErrNotFound
		}
		return nil, fmt.Errorf("get coupon by id: %w", err)
	}

	if minCartAmount.Valid {
		c.MinCartAmount = &minCartAmount.Int64
	}
	if usageLimit.Valid {
		v := int(usageLimit.Int32)
		c.UsageLimit = &v
	}
	if perUserLimit.Valid {
		v := int(perUserLimit.Int32)
		c.PerUserLimit = &v
	}

	return &c, nil
}

// GetByCode finds a coupon by its case-insensitive code.
func (r *CouponRepository) GetByCode(ctx context.Context, qe database.QueryExecutor, code string) (*model.Coupon, error) {
	query := `
		SELECT id, code, discount_type, discount_value, currency, min_cart_amount,
		       usage_limit, per_user_limit, starts_at, expires_at, status, created_at, updated_at
		FROM coupons
		WHERE LOWER(code) = LOWER($1)
	`
	var (
		c             model.Coupon
		minCartAmount sql.NullInt64
		usageLimit    sql.NullInt32
		perUserLimit  sql.NullInt32
	)

	err := qe.QueryRow(ctx, query, strings.TrimSpace(code)).Scan(
		&c.ID, &c.Code, &c.DiscountType, &c.DiscountValue, &c.Currency,
		&minCartAmount, &usageLimit, &perUserLimit, &c.StartsAt, &c.ExpiresAt,
		&c.Status, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return nil, database.ErrNotFound
		}
		return nil, fmt.Errorf("get coupon by code: %w", err)
	}

	if minCartAmount.Valid {
		c.MinCartAmount = &minCartAmount.Int64
	}
	if usageLimit.Valid {
		v := int(usageLimit.Int32)
		c.UsageLimit = &v
	}
	if perUserLimit.Valid {
		v := int(perUserLimit.Int32)
		c.PerUserLimit = &v
	}

	return &c, nil
}

// List returns coupons matching optional filters, ordered chronologically descending.
func (r *CouponRepository) List(ctx context.Context, qe database.QueryExecutor, filter ListCouponFilter) ([]*model.Coupon, error) {
	var (
		sb   strings.Builder
		args []interface{}
		idx  = 1
	)

	sb.WriteString(`
		SELECT id, code, discount_type, discount_value, currency, min_cart_amount,
		       usage_limit, per_user_limit, starts_at, expires_at, status, created_at, updated_at
		FROM coupons
	`)

	if filter.Status != nil {
		sb.WriteString(fmt.Sprintf(" WHERE status = $%d::coupon_status", idx))
		args = append(args, string(*filter.Status))
		idx++
	}

	sb.WriteString(" ORDER BY created_at DESC")

	rows, err := qe.Query(ctx, sb.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("list coupons: %w", err)
	}
	defer rows.Close()

	var list []*model.Coupon
	for rows.Next() {
		var (
			c             model.Coupon
			minCartAmount sql.NullInt64
			usageLimit    sql.NullInt32
			perUserLimit  sql.NullInt32
		)

		if err := rows.Scan(
			&c.ID, &c.Code, &c.DiscountType, &c.DiscountValue, &c.Currency,
			&minCartAmount, &usageLimit, &perUserLimit, &c.StartsAt, &c.ExpiresAt,
			&c.Status, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan coupon: %w", err)
		}

		if minCartAmount.Valid {
			c.MinCartAmount = &minCartAmount.Int64
		}
		if usageLimit.Valid {
			v := int(usageLimit.Int32)
			c.UsageLimit = &v
		}
		if perUserLimit.Valid {
			v := int(perUserLimit.Int32)
			c.PerUserLimit = &v
		}

		list = append(list, &c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate coupons: %w", err)
	}

	return list, nil
}

// Create inserts a new coupon into the database.
func (r *CouponRepository) Create(ctx context.Context, qe database.QueryExecutor, c *model.Coupon) error {
	query := `
		INSERT INTO coupons (
			id, code, discount_type, discount_value, currency, min_cart_amount,
			usage_limit, per_user_limit, starts_at, expires_at, status, created_at, updated_at
		)
		VALUES ($1, $2, $3::discount_type, $4, $5, $6, $7, $8, $9, $10, $11::coupon_status, NOW(), NOW())
		RETURNING created_at, updated_at
	`
	currency := c.Currency
	if currency == "" {
		currency = "EGP"
	}
	status := c.Status
	if status == "" {
		status = model.CouponStatusActive
	}

	return qe.QueryRow(ctx, query,
		c.ID, c.Code, c.DiscountType, c.DiscountValue, currency, c.MinCartAmount,
		c.UsageLimit, c.PerUserLimit, c.StartsAt, c.ExpiresAt, status,
	).Scan(&c.CreatedAt, &c.UpdatedAt)
}

// UpdateStatus changes the coupon status.
func (r *CouponRepository) UpdateStatus(ctx context.Context, qe database.QueryExecutor, id uuid.UUID, status model.CouponStatus) error {
	query := `UPDATE coupons SET status = $1::coupon_status, updated_at = NOW() WHERE id = $2`
	ct, err := qe.Exec(ctx, query, string(status), id)
	if err != nil {
		return fmt.Errorf("update coupon status: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return database.ErrNotFound
	}
	return nil
}

// Delete permanently removes a coupon by its UUID.
func (r *CouponRepository) Delete(ctx context.Context, qe database.QueryExecutor, id uuid.UUID) error {
	query := `DELETE FROM coupons WHERE id = $1`
	ct, err := qe.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete coupon: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return database.ErrNotFound
	}
	return nil
}

// GetRedemptionCount returns total times a coupon has been redeemed globally.
func (r *CouponRepository) GetRedemptionCount(ctx context.Context, qe database.QueryExecutor, couponID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM coupon_redemptions WHERE coupon_id = $1`
	var count int
	err := qe.QueryRow(ctx, query, couponID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get redemption count: %w", err)
	}
	return count, nil
}

// GetUserRedemptionCount returns times a coupon has been redeemed by a specific user.
func (r *CouponRepository) GetUserRedemptionCount(ctx context.Context, qe database.QueryExecutor, couponID uuid.UUID, userID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM coupon_redemptions WHERE coupon_id = $1 AND user_id = $2`
	var count int
	err := qe.QueryRow(ctx, query, couponID, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get user redemption count: %w", err)
	}
	return count, nil
}

// RecordRedemption logs a coupon usage against an order.
func (r *CouponRepository) RecordRedemption(ctx context.Context, qe database.QueryExecutor, red *model.CouponRedemption) error {
	query := `
		INSERT INTO coupon_redemptions (id, coupon_id, user_id, order_id, discount_amount, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING created_at
	`
	if red.ID == uuid.Nil {
		red.ID = uuid.New()
	}
	return qe.QueryRow(ctx, query,
		red.ID, red.CouponID, red.UserID, red.OrderID, red.DiscountAmount,
	).Scan(&red.CreatedAt)
}
