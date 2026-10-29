package coupon

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

type ListCouponFilter struct {
	Status *model.CouponStatus
	// Limit is the page size; repositories over-fetch one row to detect more.
	Limit int
	// Cursor, when set, switches the query to keyset pagination on (created_at, id) DESC.
	Cursor *pagination.Keyset
}

// RedemptionEntry is a single coupon redemption joined with the redeeming
// user's profile details, used by the admin promotion usage log.
type RedemptionEntry struct {
	ID             int64
	UserID         *uuid.UUID
	UserName       *string
	UserEmail      *string
	OrderID        uuid.UUID
	DiscountAmount int64
	Currency       string
	CreatedAt      time.Time
}

type CouponRepository struct{}

func NewRepository() *CouponRepository {
	return &CouponRepository{}
}

// GetByID finds a coupon by its UUID identifier.
func (r *CouponRepository) GetByID(ctx context.Context, qe database.QueryExecutor, id uuid.UUID) (*model.Coupon, error) {
	return r.get(ctx, qe, `WHERE id = $1`, id)
}

// GetByIDAndLock finds a coupon by UUID and takes an exclusive row lock, so
// callers can validate usage limits and redeem atomically within their tx.
func (r *CouponRepository) GetByIDAndLock(ctx context.Context, qe database.QueryExecutor, id uuid.UUID) (*model.Coupon, error) {
	return r.get(ctx, qe, `WHERE id = $1 FOR UPDATE`, id)
}

// GetByCodeAndLock finds a coupon by its case-insensitive code.
func (r *CouponRepository) GetByCodeAndLock(ctx context.Context, qe database.QueryExecutor, code string) (*model.Coupon, error) {
	return r.get(ctx, qe, `WHERE LOWER(code) = LOWER($1) FOR UPDATE`, strings.TrimSpace(code))
}

func (r *CouponRepository) GetByCode(ctx context.Context, qe database.QueryExecutor, code string) (*model.Coupon, error) {
	return r.get(ctx, qe, `WHERE LOWER(code) = LOWER($1)`, strings.TrimSpace(code))
}

func (r *CouponRepository) get(ctx context.Context, qe database.QueryExecutor, where string, args ...any) (*model.Coupon, error) {
	query := `
		SELECT 
			id, 
			code,
		 	discount_type,
		 	discount_value,
		 	currency,
		 	min_cart_amount,
		 	usage_limit,
		 	per_user_limit,
		 	starts_at,
		 	expires_at,
		 	status,
		 	created_at,
		 	updated_at
		FROM coupons
		` + where

	c := &model.Coupon{}

	var (
		minCartAmount sql.NullInt64
		usageLimit    sql.NullInt32
		perUserLimit  sql.NullInt32
	)

	err := qe.QueryRow(ctx, query, args...).Scan(
		&c.ID,
		&c.Code,
		&c.DiscountType,
		&c.DiscountValue,
		&c.Currency,
		&minCartAmount,
		&usageLimit,
		&perUserLimit,
		&c.StartsAt,
		&c.ExpiresAt,
		&c.Status,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get coupon: %w", err)
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

	return c, nil
}

// List returns coupons matching optional filters, ordered newest first. The
// secondary id sort breaks ties (e.g. rows seeded with identical created_at)
// so the order stays stable across updates instead of following heap layout.
func (r *CouponRepository) List(ctx context.Context, qe database.QueryExecutor, filter ListCouponFilter) ([]*model.Coupon, *pagination.CursorPage, error) {
	var (
		sb   strings.Builder
		args []interface{}
		idx  = 1
	)

	sb.WriteString(`
		SELECT id,
			code,
			discount_type,
			discount_value,
			currency,
			min_cart_amount,
			usage_limit,
			per_user_limit,
			starts_at,
			expires_at,
			status,
			created_at,
			updated_at,
			(SELECT COUNT(*) FROM coupon_redemptions r WHERE r.coupon_id = coupons.id) AS usage_count
		FROM coupons
		WHERE 1=1
	`)

	if filter.Status != nil {
		fmt.Fprintf(&sb, " AND status = $%d::coupon_status", idx)
		args = append(args, string(*filter.Status))
		idx++
	}

	if filter.Cursor != nil {
		fmt.Fprintf(&sb, " AND (created_at, id) < ($%d, $%d)", idx, idx+1)
		args = append(args, filter.Cursor.CreatedAt, filter.Cursor.ID)
		idx += 2
	}

	sb.WriteString(" ORDER BY created_at DESC, id DESC")
	fmt.Fprintf(&sb, " LIMIT $%d", idx)
	args = append(args, filter.Limit+1)

	rows, err := qe.Query(ctx, sb.String(), args...)
	if err != nil {
		return nil, nil, fmt.Errorf("list coupons: %w", err)
	}
	defer rows.Close()

	var list []*model.Coupon
	for rows.Next() {
		c := &model.Coupon{}
		var (
			minCartAmount sql.NullInt64
			usageLimit    sql.NullInt32
			perUserLimit  sql.NullInt32
		)

		if err := rows.Scan(
			&c.ID,
			&c.Code,
			&c.DiscountType,
			&c.DiscountValue,
			&c.Currency,
			&minCartAmount,
			&usageLimit,
			&perUserLimit,
			&c.StartsAt,
			&c.ExpiresAt,
			&c.Status,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.UsageCount,
		); err != nil {
			return nil, nil, fmt.Errorf("scan coupon: %w", err)
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

		list = append(list, c)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("iterate coupons: %w", err)
	}

	list, page := pagination.NewCursorPageFrom(list, filter.Limit, func(c *model.Coupon) pagination.Keyset {
		return pagination.Keyset{CreatedAt: c.CreatedAt, ID: c.ID}
	})
	return list, page, nil
}

// Create inserts a new coupon into the database.
func (r *CouponRepository) Create(ctx context.Context, tx *database.Tx, c *model.Coupon) error {
	query := `
		INSERT INTO coupons (
			id,
			code,
			discount_type,
			discount_value,
			currency,
			min_cart_amount,
			usage_limit,
			per_user_limit,
			starts_at,
			expires_at,
			status,
			created_at,
			updated_at
		)
		VALUES ($1,
		 $2,
		 $3::discount_type,
		 $4,
		 $5,
		 $6,
		 $7,
		 $8,
		 $9,
		 $10,
		 $11::coupon_status,
		 NOW(),
		 NOW())
		RETURNING created_at, updated_at
	`
	
	status := c.Status
	if status == "" {
		status = model.CouponStatusActive
	}

	return tx.QueryRow(ctx, query,
		c.ID,
		c.Code,
		c.DiscountType,
		c.DiscountValue,
		c.Currency,
		c.MinCartAmount,
		c.UsageLimit,
		c.PerUserLimit,
		c.StartsAt,
		c.ExpiresAt,
		status,
	).Scan(&c.CreatedAt, &c.UpdatedAt)
}

// UpdateStatus changes the coupon status.
func (r *CouponRepository) UpdateStatus(ctx context.Context, tx *database.Tx, id uuid.UUID, status model.CouponStatus) error {
	query := `UPDATE coupons SET status = $1::coupon_status, updated_at = NOW() WHERE id = $2`
	ct, err := tx.Exec(ctx, query, string(status), id)
	if err != nil {
		return fmt.Errorf("update coupon status: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// Delete permanently removes a coupon by its UUID.
func (r *CouponRepository) Delete(ctx context.Context, tx *database.Tx, id uuid.UUID) error {
	query := `DELETE FROM coupons WHERE id = $1`
	ct, err := tx.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete coupon: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
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
func (r *CouponRepository) RecordRedemption(ctx context.Context, tx *database.Tx, red *model.CouponRedemption) error {
	query := `
		INSERT INTO coupon_redemptions (coupon_id, user_id, order_id, discount_amount, currency, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, created_at
	`
	return tx.QueryRow(ctx, query,
		red.CouponID,
		red.UserID,
		red.OrderID,
		red.DiscountAmount,
		red.Currency,
	).Scan(&red.ID, &red.CreatedAt)
}

// ListRedemptions returns every redemption of a coupon, newest first, joined
// with the redeeming user's name and email for the admin usage log.
func (r *CouponRepository) ListRedemptions(ctx context.Context, qe database.QueryExecutor, couponID uuid.UUID, limit int, cursor *pagination.Keyset) ([]*RedemptionEntry, *pagination.CursorPage, error) {
	query := `
		SELECT 
			r.id,
			r.user_id,
			u.full_name,
			u.email,
			r.order_id,
			r.discount_amount,
			r.currency,
			r.created_at
		FROM coupon_redemptions r
		LEFT JOIN users u ON u.id = r.user_id
		WHERE r.coupon_id = $1
	`
	args := []any{couponID}
	i := 2

	keysetClause, keysetArgs := pagination.KeysetClause("r", cursor, i)
	query += keysetClause
	args = append(args, keysetArgs...)
	i += len(keysetArgs)

	query += pagination.KeysetOrderBy("r")
	query += fmt.Sprintf(" LIMIT $%d", i)
	args = append(args, limit+1)

	rows, err := qe.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("list coupon redemptions: %w", err)
	}
	defer rows.Close()

	var list []*RedemptionEntry
	for rows.Next() {
		var (
			e         RedemptionEntry
			userName  sql.NullString
			userEmail sql.NullString
		)

		if err := rows.Scan(
			&e.ID, &e.UserID, &userName, &userEmail, &e.OrderID, &e.DiscountAmount, &e.Currency, &e.CreatedAt,
		); err != nil {
			return nil, nil, fmt.Errorf("scan coupon redemption: %w", err)
		}

		if userName.Valid {
			v := userName.String
			e.UserName = &v
		}
		if userEmail.Valid {
			v := userEmail.String
			e.UserEmail = &v
		}

		list = append(list, &e)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("iterate coupon redemptions: %w", err)
	}

	list, page := pagination.NewCursorPageFrom(list, limit, func(e *RedemptionEntry) pagination.Keyset {
		return pagination.Keyset{CreatedAt: e.CreatedAt, IDInt: e.ID}
	})
	return list, page, nil
}
