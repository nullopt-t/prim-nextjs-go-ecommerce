package vendors

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/errcode"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/nullopt-t/prim-backend/pkg/storage"
)

// SettlementReader is the vendor domain's read-only window into order data
// (implemented by the order repository). Vendor makes no SQL against order
// tables.
type SettlementReader interface {
	// GetVendorSettledEarnings sums the delivered-order earnings of one vendor.
	GetVendorSettledEarnings(ctx context.Context, vendorID uuid.UUID) (int64, error)
}

// ObjectReader is the subset of object persistence needed to resolve vendor
// asset URLs.
type ObjectReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.Object, error)
}

// CommissionDefaults supplies the platform-wide default commission rate used
// when a new vendor is registered without an explicit override.
type CommissionDefaults interface {
	VendorCommissionBps() int
}

// Service implements vendor use cases. Cross-domain data access goes through
// the interfaces above; transactions compose via the shared runner.
type Service struct {
	repo       *VendorRepository
	userReader UserReader
	objects    ObjectReader
	settlement SettlementReader
	products   ProductCatalog
	settings   CommissionDefaults
	storage    storage.StorageProvider
	runner     database.Runner
	logger     log.Logger
}

// UserReader is the subset of user persistence vendor needs to validate that
// a vendor registration maps to a real account.
type UserReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
}

func NewService(
	repo *VendorRepository,
	userReader UserReader,
	objects ObjectReader,
	settlement SettlementReader,
	products ProductCatalog,
	settings CommissionDefaults,
	storageProvider storage.StorageProvider,
	runner database.Runner,
	logger log.Logger,
) *Service {
	if repo == nil {
		panic("vendor: repository is required")
	}
	if userReader == nil {
		panic("vendor: user reader is required")
	}
	if objects == nil {
		panic("vendor: object reader is required")
	}
	if settlement == nil {
		panic("vendor: settlement reader is required")
	}
	if products == nil {
		panic("vendor: product catalog is required")
	}
	if settings == nil {
		panic("vendor: commission defaults provider is required")
	}
	if storageProvider == nil {
		panic("vendor: storage provider is required")
	}
	if runner == nil {
		panic("vendor: database runner is required")
	}
	if logger == nil {
		panic("vendor: logger is required")
	}
	return &Service{
		repo:       repo,
		userReader: userReader,
		objects:    objects,
		settlement: settlement,
		products:   products,
		settings:   settings,
		storage:    storageProvider,
		runner:     runner,
		logger:     logger,
	}
}

// handleError translates repository and driver errors into typed APIErrors so
// only apierr.APIError values leave the vendor service.
func (s *Service) handleError(err error, fallbackMsg string) error {
	if err == nil {
		return nil
	}

	var apiErr *apierr.APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}

	mappedErr := database.MapError(err)
	switch {
	case errors.Is(mappedErr, database.ErrNotFound):
		return apierr.NotFoundError(fallbackMsg + ": not found").
			WithCode(apierr.CodeNotFound).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrConflict):
		return apierr.ConflictError(fallbackMsg + ": conflict").
			WithCode(apierr.CodeResourceConflict).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrForeignKeyViolation):
		return apierr.BadRequestError(fallbackMsg + ": referenced entity does not exist").
			WithCode(apierr.CodeInvalidReference).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrCheckViolation),
		errors.Is(mappedErr, database.ErrNotNullViolation),
		errors.Is(mappedErr, database.ErrInvalidInput):
		return apierr.BadRequestError(fallbackMsg + ": invalid input").
			WithCode(apierr.CodeInvalidInput).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrTimeout), errors.Is(mappedErr, database.ErrCanceled):
		return apierr.New(http.StatusGatewayTimeout, "database operation timed out or canceled").
			WithCode(apierr.CodeGatewayTimeout).
			Wrap(err)
	default:
		return apierr.InternalServerError(fallbackMsg).
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}
}

// attachAssetURLs hydrates the vendor's logo/banner public URLs. Missing
// objects are treated as absent (never fail the profile read).
func (s *Service) attachAssetURLs(ctx context.Context, v *model.Vendor) {
	resolve := func(objectID *uuid.UUID) (string, error) {
		if objectID == nil {
			return "", nil
		}
		obj, err := s.objects.GetByID(ctx, *objectID)
		if err != nil {
			return "", err
		}
		return s.storage.GetPublicURL(obj.Bucket, obj.Key), nil
	}

	v.LogoURL, _ = resolve(v.LogoObjectID)
	v.BannerURL, _ = resolve(v.BannerObjectID)
}

// ---- Vendor lifecycle ---------------------------------------------------

// RegisterVendorInput carries a new vendor application.
type RegisterVendorInput struct {
	StoreName      string
	Slug           string
	Description    string
	LogoObjectID   *uuid.UUID
	BannerObjectID *uuid.UUID
}

// slugPattern mirrors the DB CHECK constraint on vendors.slug.
var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// RegisterVendor creates a pending vendor application for an existing user.
// Rules:
//   - one profile per user, enforced by the unique user_id index
//   - slug uniqueness is backed by the unique index (races map to 409)
//   - commission starts at the platform default configured in settings
func (s *Service) RegisterVendor(ctx context.Context, userID uuid.UUID, in RegisterVendorInput) (*model.Vendor, error) {
	// 1. Validate shape and refer to the canonical account.
	if strings.TrimSpace(in.StoreName) == "" {
		return nil, apierr.BadRequestError("Store name is required").
			WithCode(apierr.CodeValidationFailed).
			WithFields(api.FieldError{Field: "storeName", Constraint: "required", Message: "store name is required"})
	}
	if len(strings.TrimSpace(in.StoreName)) > 150 {
		return nil, apierr.BadRequestError("Store name must be at most 150 characters").
			WithCode(apierr.CodeValidationFailed).
			WithFields(api.FieldError{Field: "storeName", Constraint: "max", Message: "store name must be at most 150 characters"})
	}

	slug := strings.ToLower(strings.TrimSpace(in.Slug))
	if slug == "" {
		return nil, apierr.BadRequestError("Slug is required").
			WithCode(apierr.CodeValidationFailed).
			WithFields(api.FieldError{Field: "slug", Constraint: "required", Message: "slug is required"})
	}
	if !slugPattern.MatchString(slug) || len(slug) > 150 {
		return nil, apierr.BadRequestError("Slug must be lowercase letters, numbers, and hyphens").
			WithCode(apierr.CodeValidationFailed).
			WithFields(api.FieldError{Field: "slug", Constraint: "slug", Message: "slug must be lowercase letters, numbers, and hyphens"})
	}

	var created *model.Vendor
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		if _, err := s.userReader.GetByID(ctx, userID); err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.NotFoundError("User not found").WithCode(apierr.CodeNotFound)
			}
			return s.handleError(err, "Failed to resolve user")
		}

		if _, err := s.repo.GetByUserID(ctx, userID); err == nil {
			return apierr.ConflictError("This account already has a vendor profile").
				WithCode(errcode.CodeVendorAlreadyRegistered)
		} else if !errors.Is(database.MapError(err), database.ErrNotFound) {
			return s.handleError(err, "Failed to check existing vendor profile")
		}

		var createErr error
		created, createErr = s.repo.Create(ctx, &CreateVendorInput{
			UserID:            userID,
			StoreName:         strings.TrimSpace(in.StoreName),
			Slug:              slug,
			Description:       strings.TrimSpace(in.Description),
			LogoObjectID:      in.LogoObjectID,
			BannerObjectID:    in.BannerObjectID,
			CommissionRateBps: s.settings.VendorCommissionBps(),
		})
		return createErr
	})
	if err != nil {
		return nil, s.handleError(err, "Failed to register vendor")
	}

	s.attachAssetURLs(ctx, created)
	return created, nil
}

// GetMyVendor returns the caller's vendor profile.
func (s *Service) GetVendorByUser(ctx context.Context, userID uuid.UUID) (*model.Vendor, error) {
	var vendor *model.Vendor
	err := s.runner.WithExecutor(ctx, func(ctx context.Context) error {
		v, err := s.repo.GetByUserID(ctx, userID)
		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.NotFoundError("Vendor not found").WithCode(errcode.CodeVendorNotFound)
			}
			return s.handleError(err, "Failed to load vendor")
		}
		vendor = v
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.attachAssetURLs(ctx, vendor)
	return vendor, nil
}

// UpdateVendorInput is the vendor-editable profile payload.
type UpdateVendorInput struct {
	StoreName       *string
	Description     *string
	LogoObjectID    *uuid.UUID
	BannerObjectID  *uuid.UUID
	BankAccountInfo map[string]any
}

// UpdateMyProfile edits the caller's own profile. Suspended/rejected vendors
// may not edit.
func (s *Service) UpdateMyProfile(ctx context.Context, userID uuid.UUID, in UpdateVendorInput) (*model.Vendor, error) {
	var updated *model.Vendor
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		existing, err := s.repo.GetByUserID(ctx, userID)
		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.NotFoundError("Vendor not found").WithCode(errcode.CodeVendorNotFound)
			}
			return s.handleError(err, "Failed to load vendor")
		}

		if existing.Status != model.VendorStatusActive && existing.Status != model.VendorStatusPending {
			return apierr.ForbiddenError("Vendor is not authorized for this action").
				WithCode(errcode.CodeVendorNotAuthorizedForAction)
		}

		fields := UpdateVendorFields{
			StoreName:       in.StoreName,
			Description:     in.Description,
			LogoObjectID:    in.LogoObjectID,
			BannerObjectID:  in.BannerObjectID,
			BankAccountInfo: in.BankAccountInfo,
		}

		updated, err = s.repo.Update(ctx, existing.ID, fields)
		return err
	})
	if err != nil {
		return nil, s.handleError(err, "Failed to update vendor profile")
	}

	s.attachAssetURLs(ctx, updated)
	return updated, nil
}

// ---- Admin lifecycle ----------------------------------------------------

// isValidVendorTransition maps the allowed lifecycle arrows of ADR §state machines.
func isValidVendorTransition(from, to model.VendorStatus) bool {
	switch {
	case from == model.VendorStatusPending && (to == model.VendorStatusActive || to == model.VendorStatusRejected):
		return true
	case from == model.VendorStatusActive && to == model.VendorStatusSuspended:
		return true
	case from == model.VendorStatusSuspended && to == model.VendorStatusActive:
		return true
	default:
		return false
	}
}

// AdminGetVendorByID loads any vendor for administration.
func (s *Service) AdminGetVendorByID(ctx context.Context, vendorID uuid.UUID) (*model.Vendor, error) {
	var vendor *model.Vendor
	err := s.runner.WithExecutor(ctx, func(ctx context.Context) error {
		v, err := s.repo.GetByID(ctx, vendorID)
		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.NotFoundError("Vendor not found").WithCode(errcode.CodeVendorNotFound)
			}
			return s.handleError(err, "Failed to load vendor")
		}
		vendor = v
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.attachAssetURLs(ctx, vendor)
	return vendor, nil
}

// AdminChangeVendorStatus applies a lifecycle transition guarded by the state machine.
func (s *Service) AdminChangeVendorStatus(ctx context.Context, vendorID uuid.UUID, to model.VendorStatus) (*model.Vendor, error) {
	var updated *model.Vendor
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		existing, err := s.repo.GetByID(ctx, vendorID)
		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.NotFoundError("Vendor not found").WithCode(errcode.CodeVendorNotFound)
			}
			return s.handleError(err, "Failed to load vendor")
		}

		if !isValidVendorTransition(existing.Status, to) {
			return apierr.ConflictError("Vendor status transition is not permitted").
				WithCode(errcode.CodeVendorNotAuthorizedForAction)
		}

		if err := s.repo.SetStatus(ctx, vendorID, to); err != nil {
			return err
		}

		updated, err = s.repo.GetByID(ctx, vendorID)
		return err
	})
	if err != nil {
		return nil, s.handleError(err, "Failed to change vendor status")
	}

	s.attachAssetURLs(ctx, updated)
	return updated, nil
}

// AdminSetCommissionRate overrides the per-vendor commission (basis points).
func (s *Service) AdminSetCommissionRate(ctx context.Context, vendorID uuid.UUID, rateBps int) (*model.Vendor, error) {
	if rateBps < 0 || rateBps > 10000 {
		return nil, apierr.BadRequestError("Commission rate out of range").
			WithCode(apierr.CodeValidationFailed)
	}

	var updated *model.Vendor
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		if err := s.repo.SetCommissionRate(ctx, vendorID, rateBps); err != nil {
			return err
		}
		var err error
		updated, err = s.repo.GetByID(ctx, vendorID)
		return err
	})
	if err != nil {
		return nil, s.handleError(err, "Failed to set commission rate")
	}

	s.attachAssetURLs(ctx, updated)
	return updated, nil
}

// ---- Payouts ------------------------------------------------------------

// PayoutOutstandingStatuses lists payout states that have claimed funds from
// the balance (rejected does not, paid does until it is out the door).
func PayoutOutstandingStatuses() []model.PayoutStatus {
	return []model.PayoutStatus{
		model.PayoutStatusRequested,
		model.PayoutStatusApproved,
		model.PayoutStatusProcessing,
		model.PayoutStatusPaid,
	}
}

// RequestPayout registers a payout request after validating that settled
// earnings minus outstanding (non-rejected) payouts cover the amount. The
// vendor row is locked FOR UPDATE under the transaction so two concurrent
// requests cannot both claim the same balance; the audit snapshot is stored.
func (s *Service) RequestPayout(ctx context.Context, userID uuid.UUID, amount int64) (*model.VendorPayout, error) {
	if amount <= 0 {
		return nil, apierr.BadRequestError("Payout amount must be positive").
			WithCode(errcode.CodePayoutInvalidAmount)
	}

	var payout *model.VendorPayout
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		existing, err := s.repo.GetByUserID(ctx, userID)
		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.NotFoundError("Vendor not found").WithCode(errcode.CodeVendorNotFound)
			}
			return s.handleError(err, "Failed to load vendor")
		}

		if existing.Status != model.VendorStatusActive {
			return apierr.ForbiddenError("Vendor is not authorized for this action").
				WithCode(errcode.CodeVendorNotAuthorizedForAction)
		}

		// Lock the vendor row BEFORE reading balances so concurrent requests
		// serialize on this single lock.
		if _, err := s.repo.GetByIDForUpdate(ctx, existing.ID); err != nil {
			return s.handleError(err, "Failed to lock vendor")
		}

		settled, err := s.settlement.GetVendorSettledEarnings(ctx, existing.ID)
		if err != nil {
			return s.handleError(err, "Failed to compute settled earnings")
		}
		outstanding, err := s.repo.GetOutstandingPayoutTotal(ctx, existing.ID)
		if err != nil {
			return s.handleError(err, "Failed to compute outstanding payouts")
		}
		available := settled - outstanding

		if amount > available {
			return apierr.BadRequestError(fmt.Sprintf("Requested payout exceeds available balance of %d", available)).
				WithCode(errcode.CodePayoutInsufficientBalance)
		}

		payout, err = s.repo.CreatePayout(ctx, existing.ID, amount, available)
		return err
	})
	if err != nil {
		return nil, s.handleError(err, "Failed to request payout")
	}

	return payout, nil
}

// ListMyPayouts returns the requester's full payout ledger, including
// rejected rows (audit trail never deleted).
func (s *Service) ListMyPayouts(ctx context.Context, userID uuid.UUID) ([]*model.VendorPayout, error) {
	var payouts []*model.VendorPayout
	err := s.runner.WithExecutor(ctx, func(ctx context.Context) error {
		existing, err := s.repo.GetByUserID(ctx, userID)
		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.NotFoundError("Vendor not found").WithCode(errcode.CodeVendorNotFound)
			}
			return s.handleError(err, "Failed to load vendor")
		}

		payouts, err = s.repo.ListPayoutsByVendor(ctx, existing.ID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return payouts, nil
}

// AdminListPayouts exposes the ledger across all vendors (admin filter surface).
func (s *Service) AdminListPayouts(ctx context.Context) ([]*model.VendorPayout, error) {
	var payouts []*model.VendorPayout
	err := s.runner.WithExecutor(ctx, func(ctx context.Context) error {
		var err error
		payouts, err = s.repo.ListPayouts(ctx)
		return err
	})
	if err != nil {
		return nil, s.handleError(err, "Failed to list payouts")
	}
	return payouts, nil
}

// isValidPayoutTransition guards the payout state machine. Rejection after
// processing is allowed only when the transfer is confirmed-failed; the
// caller supplies the transfer note.
func isValidPayoutTransition(from, to model.PayoutStatus) bool {
	switch {
	case from == model.PayoutStatusRequested && to == model.PayoutStatusApproved:
		return true
	case from == model.PayoutStatusRequested && to == model.PayoutStatusRejected:
		return true
	case from == model.PayoutStatusApproved && to == model.PayoutStatusProcessing:
		return true
	case from == model.PayoutStatusApproved && to == model.PayoutStatusRejected:
		return true
	case from == model.PayoutStatusProcessing && to == model.PayoutStatusPaid:
		return true
	case from == model.PayoutStatusProcessing && to == model.PayoutStatusRejected:
		return true
	default:
		return false
	}
}

// AdminTransitionPayout moves a payout through the lifecycle. Transitions are
// conditional on the prior state: a duplicate call whose state already moved
// forms a conflict, never a double-pay.
func (s *Service) AdminTransitionPayout(
	ctx context.Context,
	payoutID uuid.UUID,
	to model.PayoutStatus,
	referenceNumber *string,
	notes *string,
) (*model.VendorPayout, error) {
	var payout *model.VendorPayout
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		existing, err := s.repo.GetPayoutByID(ctx, payoutID)
		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.NotFoundError("Payout not found").WithCode(errcode.CodePayoutNotFound)
			}
			return s.handleError(err, "Failed to load payout")
		}

		if !isValidPayoutTransition(existing.Status, to) {
			return apierr.ConflictError("Payout state transition is not permitted").
				WithCode(errcode.CodePayoutInvalidStateTransition)
		}

		if err := s.repo.UpdatePayoutStatus(ctx, payoutID, existing.Status, to, referenceNumber, notes); err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				// Between read & transition another actor moved the row.
				return apierr.ConflictError("Payout state transition is not permitted").
					WithCode(errcode.CodePayoutInvalidStateTransition)
			}
			return err
		}

		payout, err = s.repo.GetPayoutByID(ctx, payoutID)
		return err
	})
	if err != nil {
		return nil, s.handleError(err, "Failed to transition payout")
	}

	return payout, nil
}
