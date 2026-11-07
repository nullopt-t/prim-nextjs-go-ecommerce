// Package seller owns merchant identity, marketplace standing, and the
// payout ledger (docs/adr-seller-domain.md). Catalog, orders, and stock stay
// in their own domains; this package reads them only through
// consumer-defined interfaces.
package sellers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

// CreateSellerInput carries the columns written on registration.
type CreateSellerInput struct {
	UserID            uuid.UUID
	StoreName         string
	Slug              string
	Description       string
	LogoObjectID      *uuid.UUID
	BannerObjectID    *uuid.UUID
	CommissionRateBps int
}

// UpdateSellerFields carries optional profile mutations.
type UpdateSellerFields struct {
	StoreName       *string
	Slug            *string
	Description     *string
	LogoObjectID    *uuid.UUID
	BannerObjectID  *uuid.UUID
	BankAccountInfo map[string]any
}

// SellerRepository is the seller domain's own persistence (SQL + row mapping).
type SellerRepository struct{}

func NewRepository() *SellerRepository {
	return &SellerRepository{}
}

func scanSeller(row pgx.Row) (*model.Seller, error) {
	v := new(model.Seller)
	err := row.Scan(
		&v.ID,
		&v.UserID,
		&v.StoreName,
		&v.Slug,
		&v.Description,
		&v.LogoObjectID,
		&v.BannerObjectID,
		&v.CommissionRateBps,
		&v.Status,
		&v.StatusReason,
		&v.ReviewedBy,
		&v.ReviewedAt,
		&v.BankAccountInfo,
		&v.CreatedAt,
		&v.UpdatedAt,
		&v.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan seller: %w", err)
	}
	return v, nil
}

const sellerColumns = `
	id,
	user_id,
	store_name,
	slug,
	COALESCE(description, ''),
	logo_object_id,
	banner_object_id,
	commission_rate_bps,
	status,
	COALESCE(status_reason, ''),
	reviewed_by,
	reviewed_at,
	bank_account_info,
	created_at,
	updated_at,
	deleted_at
`

func (r *SellerRepository) Create(ctx context.Context, in *CreateSellerInput) (*model.Seller, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

	v := &model.Seller{
		ID:                uuid.New(),
		UserID:            in.UserID,
		StoreName:         in.StoreName,
		Slug:              in.Slug,
		Description:       in.Description,
		LogoObjectID:      in.LogoObjectID,
		BannerObjectID:    in.BannerObjectID,
		CommissionRateBps: in.CommissionRateBps,
		Status:            model.SellerStatusPending,
		BankAccountInfo:   map[string]any{},
	}

	err = executor.QueryRowContext(ctx, `
		INSERT INTO sellers (
			id, user_id, store_name, slug, description,
			logo_object_id, banner_object_id, commission_rate_bps,
			status, bank_account_info, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, now(), now())
		RETURNING `+sellerColumns,
		v.ID, v.UserID, v.StoreName, v.Slug, v.Description,
		v.LogoObjectID, v.BannerObjectID, v.CommissionRateBps,
		v.Status, []byte(`{}`),
	).Scan(
		&v.ID, &v.UserID, &v.StoreName, &v.Slug, &v.Description,
		&v.LogoObjectID, &v.BannerObjectID, &v.CommissionRateBps,
		&v.Status, &v.StatusReason, &v.ReviewedBy, &v.ReviewedAt,
		&v.BankAccountInfo, &v.CreatedAt, &v.UpdatedAt, &v.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create seller: %w", err)
	}
	return v, nil
}

func (r *SellerRepository) getBy(ctx context.Context, where string, args ...any) (*model.Seller, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

	row := executor.QueryRowContext(
		ctx,
		`SELECT `+sellerColumns+` FROM sellers WHERE `+where,
		args...,
	)
	return scanSeller(row)
}

func (r *SellerRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Seller, error) {
	return r.getBy(ctx, "id = $1 AND deleted_at IS NULL", id)
}

func (r *SellerRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*model.Seller, error) {
	return r.getBy(ctx, "user_id = $1 AND deleted_at IS NULL", userID)
}

func (r *SellerRepository) GetBySlug(ctx context.Context, slug string) (*model.Seller, error) {
	return r.getBy(ctx, "slug = $1 AND deleted_at IS NULL", slug)
}

// ListAll returns sellers ordered newest-first, optionally narrowed to one
// lifecycle status, with classic offset pagination and a matching total.
func (r *SellerRepository) ListAll(ctx context.Context, status *model.SellerStatus, q *pagination.ListQuery) ([]*model.Seller, int64, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, 0, err
	}

	where := "WHERE deleted_at IS NULL"
	args := []any{}
	if status != nil && *status != "" {
		where += " AND status = $1::seller_status"
		args = append(args, *status)
	}

	var total int64
	countQuery := `SELECT COUNT(*) FROM sellers ` + where
	if err := executor.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count sellers: %w", err)
	}

	offsetIdx := len(args) + 1
	limitIdx := len(args) + 2
	rows, err := executor.QueryContext(ctx, `
		SELECT `+sellerColumns+` FROM sellers `+where+`
		ORDER BY created_at DESC, id DESC
		LIMIT $`+fmt.Sprint(limitIdx)+` OFFSET $`+fmt.Sprint(offsetIdx),
		append(args, q.Offset, q.PageSize)...,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list sellers: %w", err)
	}
	defer rows.Close()

	sellers := make([]*model.Seller, 0)
	for rows.Next() {
		v, err := scanSeller(rows)
		if err != nil {
			return nil, 0, err
		}
		sellers = append(sellers, v)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list sellers: %w", err)
	}
	return sellers, total, nil
}

// GetByIDForUpdate locks the seller row for the remainder of the caller's
// transaction. It requires an active transaction (money decisions depend on
// the lock serializing concurrent balance claims).
func (r *SellerRepository) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.Seller, error) {
	if _, err := database.RequireTx(ctx); err != nil {
		return nil, err
	}
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

	row := executor.QueryRowContext(
		ctx,
		`SELECT `+sellerColumns+` FROM sellers WHERE id = $1 AND deleted_at IS NULL FOR UPDATE`,
		id,
	)
	return scanSeller(row)
}

// Update writes the optional profile fields on the seller row. Bank account
// info may be nil (no change).
func (r *SellerRepository) Update(ctx context.Context, sellerID uuid.UUID, fields UpdateSellerFields) (*model.Seller, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

	bankJSON := []byte(nil)
	if fields.BankAccountInfo != nil {
		var marshalErr error
		bankJSON, marshalErr = marshalJSONB(fields.BankAccountInfo)
		if marshalErr != nil {
			return nil, fmt.Errorf("marshal bank info: %w", marshalErr)
		}
	}

	row := executor.QueryRowContext(ctx, `
		UPDATE sellers SET
			store_name         = COALESCE($2, store_name),
			slug               = COALESCE($3, slug),
			description        = COALESCE($4, description),
			logo_object_id     = COALESCE($5, logo_object_id),
			banner_object_id   = COALESCE($6, banner_object_id),
			bank_account_info  = COALESCE($7, bank_account_info),
			updated_at         = now()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING `+sellerColumns,
		sellerID, fields.StoreName, fields.Slug, fields.Description, fields.LogoObjectID,
		fields.BannerObjectID, bankJSON,
	)
	return scanSeller(row)
}

// SetStatus transitions the lifecycle state and records the reviewer
// metadata for admin actions.
func (r *SellerRepository) SetStatus(
	ctx context.Context,
	sellerID uuid.UUID,
	status model.SellerStatus,
	reason *string,
	reviewedBy *uuid.UUID,
) error {
	executor, err := database.Executor(ctx)
	if err != nil {
		return err
	}

	cmd, err := executor.ExecContext(ctx, `
		UPDATE sellers SET
			status = $2,
			status_reason = $3,
			reviewed_by = $4,
			reviewed_at = now(),
			updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`, sellerID, status, reason, reviewedBy)
	if err != nil {
		return fmt.Errorf("set seller status: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// MarkResubmitted returns a rejected application to the review queue and
// clears the prior admin decision metadata.
func (r *SellerRepository) MarkResubmitted(ctx context.Context, sellerID uuid.UUID) error {
	executor, err := database.Executor(ctx)
	if err != nil {
		return err
	}

	cmd, err := executor.ExecContext(ctx, `
		UPDATE sellers SET
			status = 'pending',
			status_reason = NULL,
			reviewed_by = NULL,
			reviewed_at = NULL,
			updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL AND status = 'rejected'
	`, sellerID)
	if err != nil {
		return fmt.Errorf("mark seller resubmitted: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// SlugExists reports whether any seller row, including soft-deleted ones,
// owns the candidate slug. The unique index is the final authority; this is
// the live availability lookup used by the onboarding UI.
func (r *SellerRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return false, err
	}

	var exists bool
	if err := executor.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM sellers WHERE slug = $1)`, slug).Scan(&exists); err != nil {
		return false, fmt.Errorf("check seller slug: %w", err)
	}
	return exists, nil
}

// SetCommissionRate assigns the per-seller commission override (admin action).
func (r *SellerRepository) SetCommissionRate(ctx context.Context, sellerID uuid.UUID, rateBps int) error {
	executor, err := database.Executor(ctx)
	if err != nil {
		return err
	}

	cmd, err := executor.ExecContext(ctx, `
		UPDATE sellers SET commission_rate_bps = $2, updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`, sellerID, rateBps)
	if err != nil {
		return fmt.Errorf("set seller commission: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// ---- Payouts ------------------------------------------------------------

// GetOutstandingPayoutTotal sums payouts whose status (requested, approved,
// processing, paid) has claimed funds from the balance; rejected rows do not
// hold money, so they are excluded.
func (r *SellerRepository) GetOutstandingPayoutTotal(ctx context.Context, sellerID uuid.UUID) (int64, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return 0, err
	}

	var total int64
	err = executor.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(amount), 0)
		FROM seller_payouts
		WHERE seller_id = $1
		  AND status = ANY($2::payout_status[])
	`, sellerID, []string{
		string(model.PayoutStatusRequested),
		string(model.PayoutStatusApproved),
		string(model.PayoutStatusProcessing),
		string(model.PayoutStatusPaid),
	}).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("sum outstanding payouts: %w", err)
	}
	return total, nil
}

func scanPayout(row pgx.Row) (*model.SellerPayout, error) {
	p := new(model.SellerPayout)
	err := row.Scan(
		&p.ID,
		&p.SellerID,
		&p.Amount,
		&p.Currency,
		&p.Status,
		&p.ReferenceNumber,
		&p.Notes,
		&p.SnapshotAvailableBalance,
		&p.IdempotencyKey,
		&p.CreatedAt,
		&p.UpdatedAt,
		&p.ProcessedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan payout: %w", err)
	}
	return p, nil
}

// CreatePayout inserts a payout request after the caller has validated the
// balance under the seller row lock. idempotencyKey deduplicates retries when
// supplied.
func (r *SellerRepository) CreatePayout(ctx context.Context, sellerID uuid.UUID, amount int64, snapshotBalance int64, idempotencyKey *string) (*model.SellerPayout, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

	row := executor.QueryRowContext(ctx, `
		INSERT INTO seller_payouts (
			id, seller_id, amount, currency, status,
			snapshot_available_balance, idempotency_key, created_at, updated_at
		)
		VALUES ($1, $2, $3, 'EGP', $4, $5, $6, now(), now())
		RETURNING
			id, seller_id, amount, currency, status,
			reference_number, COALESCE(notes, ''), snapshot_available_balance,
			idempotency_key, created_at, updated_at, processed_at
	`,
		uuid.New(), sellerID, amount, model.PayoutStatusRequested, snapshotBalance, idempotencyKey,
	)
	return scanPayout(row)
}

// GetPayoutByIdempotencyKey returns a previously accepted request for the
// same seller/key pair, if one exists.
func (r *SellerRepository) GetPayoutByIdempotencyKey(ctx context.Context, sellerID uuid.UUID, key string) (*model.SellerPayout, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

	row := executor.QueryRowContext(ctx, `
		SELECT
			id, seller_id, amount, currency, status,
			reference_number, COALESCE(notes, ''), snapshot_available_balance,
			idempotency_key, created_at, updated_at, processed_at
		FROM seller_payouts
		WHERE seller_id = $1 AND idempotency_key = $2
	`, sellerID, key)
	return scanPayout(row)
}

// ListPayoutsBySeller returns the seller's full ledger (audit trail retained,
// rejected rows included).
func (r *SellerRepository) ListPayoutsBySeller(ctx context.Context, sellerID uuid.UUID) ([]*model.SellerPayout, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := executor.QueryContext(ctx, `
		SELECT
			id, seller_id, amount, currency, status,
			reference_number, COALESCE(notes, ''), snapshot_available_balance,
			idempotency_key, created_at, updated_at, processed_at
		FROM seller_payouts
		WHERE seller_id = $1
		ORDER BY created_at DESC
	`, sellerID)
	if err != nil {
		return nil, fmt.Errorf("list payouts: %w", err)
	}
	defer rows.Close()

	payouts := make([]*model.SellerPayout, 0)
	for rows.Next() {
		p, err := scanPayout(rows)
		if err != nil {
			return nil, err
		}
		payouts = append(payouts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list payouts: %w", err)
	}
	return payouts, nil
}

// UpdatePayoutStatus performs a conditional lifecycle transition. It only
// succeeds when the record is still (or already) in fromStatus; zero rows
// affected maps to ErrNotFound (idempotency guard against double-approve /
// double-pay and concurrent workers).
func (r *SellerRepository) UpdatePayoutStatus(
	ctx context.Context,
	payoutID uuid.UUID,
	fromStatus model.PayoutStatus,
	toStatus model.PayoutStatus,
	referenceNumber *string,
	notes *string,
) error {
	executor, err := database.Executor(ctx)
	if err != nil {
		return err
	}

	cmd, err := executor.ExecContext(ctx, `
		UPDATE seller_payouts SET
			status = $3::payout_status,
			reference_number = COALESCE($4::text, reference_number),
			notes            = COALESCE($5::text, notes),
			processed_at     = CASE WHEN $3::payout_status IN ('paid', 'rejected') THEN now() ELSE processed_at END,
			updated_at       = now()
		WHERE id = $1 AND status = $2::payout_status
	`, payoutID, fromStatus, toStatus, referenceNumber, notes)
	if err != nil {
		return fmt.Errorf("update payout status: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// GetPayoutByID loads a payout by id for admin actions.
func (r *SellerRepository) GetPayoutByID(ctx context.Context, payoutID uuid.UUID) (*model.SellerPayout, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

	row := executor.QueryRowContext(ctx, `
		SELECT
			id, seller_id, amount, currency, status,
			reference_number, COALESCE(notes, ''), snapshot_available_balance,
			idempotency_key, created_at, updated_at, processed_at
		FROM seller_payouts
		WHERE id = $1
	`, payoutID)
	return scanPayout(row)
}

// ListPayouts returns every payout across all sellers (admin view).
func (r *SellerRepository) ListPayouts(ctx context.Context) ([]*model.SellerPayout, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := executor.QueryContext(ctx, `
		SELECT
			id, seller_id, amount, currency, status,
			reference_number, COALESCE(notes, ''), snapshot_available_balance,
			idempotency_key, created_at, updated_at, processed_at
		FROM seller_payouts
		ORDER BY created_at DESC, id DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list payouts: %w", err)
	}
	defer rows.Close()

	payouts := make([]*model.SellerPayout, 0)
	for rows.Next() {
		p, err := scanPayout(rows)
		if err != nil {
			return nil, err
		}
		payouts = append(payouts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list payouts: %w", err)
	}
	return payouts, nil
}

// marshalJSONB serializes a jsonb column payload.
func marshalJSONB(v any) ([]byte, error) {
	return json.Marshal(v)
}
