// Package vendor owns merchant identity, marketplace standing, and the
// payout ledger (docs/adr-vendor-domain.md). Catalog, orders, and stock stay
// in their own domains; this package reads them only through
// consumer-defined interfaces.
package vendors

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

// CreateVendorInput carries the columns written on registration.
type CreateVendorInput struct {
	UserID            uuid.UUID
	StoreName         string
	Slug              string
	Description       string
	LogoObjectID      *uuid.UUID
	BannerObjectID    *uuid.UUID
	CommissionRateBps int
}

// UpdateVendorFields carries optional profile mutations.
type UpdateVendorFields struct {
	StoreName       *string
	Description     *string
	LogoObjectID    *uuid.UUID
	BannerObjectID  *uuid.UUID
	BankAccountInfo map[string]any
}

// VendorRepository is the vendor domain's own persistence (SQL + row mapping).
type VendorRepository struct{}

func NewRepository() *VendorRepository {
	return &VendorRepository{}
}

func scanVendor(row pgx.Row) (*model.Vendor, error) {
	v := new(model.Vendor)
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
		&v.BankAccountInfo,
		&v.CreatedAt,
		&v.UpdatedAt,
		&v.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan vendor: %w", err)
	}
	return v, nil
}

const vendorColumns = `
	id,
	user_id,
	store_name,
	slug,
	description,
	logo_object_id,
	banner_object_id,
	commission_rate_bps,
	status,
	bank_account_info,
	created_at,
	updated_at,
	deleted_at
`

func (r *VendorRepository) Create(ctx context.Context, in *CreateVendorInput) (*model.Vendor, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

	v := &model.Vendor{
		ID:                uuid.New(),
		UserID:            in.UserID,
		StoreName:         in.StoreName,
		Slug:              in.Slug,
		Description:       in.Description,
		LogoObjectID:      in.LogoObjectID,
		BannerObjectID:    in.BannerObjectID,
		CommissionRateBps: in.CommissionRateBps,
		Status:            model.VendorStatusPending,
		BankAccountInfo:   map[string]any{},
	}

	err = executor.QueryRow(ctx, `
		INSERT INTO vendors (
			id, user_id, store_name, slug, description,
			logo_object_id, banner_object_id, commission_rate_bps,
			status, bank_account_info, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, now(), now())
		RETURNING `+vendorColumns,
		v.ID, v.UserID, v.StoreName, v.Slug, v.Description,
		v.LogoObjectID, v.BannerObjectID, v.CommissionRateBps,
		v.Status, []byte(`{}`),
	).Scan(
		&v.ID, &v.UserID, &v.StoreName, &v.Slug, &v.Description,
		&v.LogoObjectID, &v.BannerObjectID, &v.CommissionRateBps,
		&v.Status, &v.BankAccountInfo, &v.CreatedAt, &v.UpdatedAt, &v.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create vendor: %w", err)
	}
	return v, nil
}

func (r *VendorRepository) getBy(ctx context.Context, where string, args ...any) (*model.Vendor, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

	row := executor.QueryRow(
		ctx,
		`SELECT `+vendorColumns+` FROM vendors WHERE `+where,
		args...,
	)
	return scanVendor(row)
}

func (r *VendorRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Vendor, error) {
	return r.getBy(ctx, "id = $1 AND deleted_at IS NULL", id)
}

func (r *VendorRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*model.Vendor, error) {
	return r.getBy(ctx, "user_id = $1 AND deleted_at IS NULL", userID)
}

func (r *VendorRepository) GetBySlug(ctx context.Context, slug string) (*model.Vendor, error) {
	return r.getBy(ctx, "slug = $1 AND deleted_at IS NULL", slug)
}

// ListAll returns vendors ordered newest-first, optionally narrowed to one
// lifecycle status, with classic offset pagination and a matching total.
func (r *VendorRepository) ListAll(ctx context.Context, status *model.VendorStatus, q *pagination.ListQuery) ([]*model.Vendor, int64, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, 0, err
	}

	where := "WHERE deleted_at IS NULL"
	args := []any{}
	if status != nil && *status != "" {
		where += " AND status = $1::vendor_status"
		args = append(args, *status)
	}

	var total int64
	countQuery := `SELECT COUNT(*) FROM vendors ` + where
	if err := executor.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count vendors: %w", err)
	}

	offsetIdx := len(args) + 1
	limitIdx := len(args) + 2
	rows, err := executor.Query(ctx, `
		SELECT `+vendorColumns+` FROM vendors `+where+`
		ORDER BY created_at DESC, id DESC
		LIMIT $`+fmt.Sprint(limitIdx)+` OFFSET $`+fmt.Sprint(offsetIdx),
		append(args, q.PageSize, q.Offset)...,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list vendors: %w", err)
	}
	defer rows.Close()

	vendors := make([]*model.Vendor, 0)
	for rows.Next() {
		v, err := scanVendor(rows)
		if err != nil {
			return nil, 0, err
		}
		vendors = append(vendors, v)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list vendors: %w", err)
	}
	return vendors, total, nil
}

// GetByIDForUpdate locks the vendor row for the remainder of the caller's
// transaction. It requires an active transaction (money decisions depend on
// the lock serializing concurrent balance claims).
func (r *VendorRepository) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.Vendor, error) {
	if _, err := database.RequireTx(ctx); err != nil {
		return nil, err
	}
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

	row := executor.QueryRow(
		ctx,
		`SELECT `+vendorColumns+` FROM vendors WHERE id = $1 AND deleted_at IS NULL FOR UPDATE`,
		id,
	)
	return scanVendor(row)
}

// Update writes the optional profile fields on the vendor row. Bank account
// info may be nil (no change).
func (r *VendorRepository) Update(ctx context.Context, vendorID uuid.UUID, fields UpdateVendorFields) (*model.Vendor, error) {
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

	row := executor.QueryRow(ctx, `
		UPDATE vendors SET
			store_name         = COALESCE($2, store_name),
			description        = COALESCE($3, description),
			logo_object_id     = COALESCE($4, logo_object_id),
			banner_object_id   = COALESCE($5, banner_object_id),
			bank_account_info  = COALESCE($6, bank_account_info),
			updated_at         = now()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING `+vendorColumns,
		vendorID, fields.StoreName, fields.Description, fields.LogoObjectID,
		fields.BannerObjectID, bankJSON,
	)
	return scanVendor(row)
}

// SetStatus transitions the lifecycle state (admin action).
func (r *VendorRepository) SetStatus(ctx context.Context, vendorID uuid.UUID, status model.VendorStatus) error {
	executor, err := database.Executor(ctx)
	if err != nil {
		return err
	}

	cmd, err := executor.Exec(ctx, `
		UPDATE vendors SET status = $2, updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`, vendorID, status)
	if err != nil {
		return fmt.Errorf("set vendor status: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// SetCommissionRate assigns the per-vendor commission override (admin action).
func (r *VendorRepository) SetCommissionRate(ctx context.Context, vendorID uuid.UUID, rateBps int) error {
	executor, err := database.Executor(ctx)
	if err != nil {
		return err
	}

	cmd, err := executor.Exec(ctx, `
		UPDATE vendors SET commission_rate_bps = $2, updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`, vendorID, rateBps)
	if err != nil {
		return fmt.Errorf("set vendor commission: %w", err)
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
func (r *VendorRepository) GetOutstandingPayoutTotal(ctx context.Context, vendorID uuid.UUID) (int64, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return 0, err
	}

	var total int64
	err = executor.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0)
		FROM vendor_payouts
		WHERE vendor_id = $1
		  AND status = ANY($2::payout_status[])
	`, vendorID, []string{
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

func scanPayout(row pgx.Row) (*model.VendorPayout, error) {
	p := new(model.VendorPayout)
	err := row.Scan(
		&p.ID,
		&p.VendorID,
		&p.Amount,
		&p.Currency,
		&p.Status,
		&p.ReferenceNumber,
		&p.Notes,
		&p.SnapshotAvailableBalance,
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
// balance under the vendor row lock.
func (r *VendorRepository) CreatePayout(ctx context.Context, vendorID uuid.UUID, amount int64, snapshotBalance int64) (*model.VendorPayout, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

	row := executor.QueryRow(ctx, `
		INSERT INTO vendor_payouts (
			id, vendor_id, amount, currency, status,
			snapshot_available_balance, created_at, updated_at
		)
		VALUES ($1, $2, $3, 'EGP', $4, $5, now(), now())
		RETURNING
			id, vendor_id, amount, currency, status,
			reference_number, COALESCE(notes, ''), snapshot_available_balance,
			created_at, updated_at, processed_at
	`,
		uuid.New(), vendorID, amount, model.PayoutStatusRequested, snapshotBalance,
	)
	return scanPayout(row)
}

// ListPayoutsByVendor returns the vendor's full ledger (audit trail retained,
// rejected rows included).
func (r *VendorRepository) ListPayoutsByVendor(ctx context.Context, vendorID uuid.UUID) ([]*model.VendorPayout, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := executor.Query(ctx, `
		SELECT
			id, vendor_id, amount, currency, status,
			reference_number, COALESCE(notes, ''), snapshot_available_balance,
			created_at, updated_at, processed_at
		FROM vendor_payouts
		WHERE vendor_id = $1
		ORDER BY created_at DESC
	`, vendorID)
	if err != nil {
		return nil, fmt.Errorf("list payouts: %w", err)
	}
	defer rows.Close()

	payouts := make([]*model.VendorPayout, 0)
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
func (r *VendorRepository) UpdatePayoutStatus(
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

	cmd, err := executor.Exec(ctx, `
		UPDATE vendor_payouts SET
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
func (r *VendorRepository) GetPayoutByID(ctx context.Context, payoutID uuid.UUID) (*model.VendorPayout, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

	row := executor.QueryRow(ctx, `
		SELECT
			id, vendor_id, amount, currency, status,
			reference_number, COALESCE(notes, ''), snapshot_available_balance,
			created_at, updated_at, processed_at
		FROM vendor_payouts
		WHERE id = $1
	`, payoutID)
	return scanPayout(row)
}

// ListPayouts returns every payout across all vendors (admin view).
func (r *VendorRepository) ListPayouts(ctx context.Context) ([]*model.VendorPayout, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := executor.Query(ctx, `
		SELECT
			id, vendor_id, amount, currency, status,
			reference_number, COALESCE(notes, ''), snapshot_available_balance,
			created_at, updated_at, processed_at
		FROM vendor_payouts
		ORDER BY created_at DESC, id DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list payouts: %w", err)
	}
	defer rows.Close()

	payouts := make([]*model.VendorPayout, 0)
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
