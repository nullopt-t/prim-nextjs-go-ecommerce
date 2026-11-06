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
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/nullopt-t/prim-backend/pkg/storage"
	"golang.org/x/text/unicode/norm"
)

// SettlementReader is the vendor domain's read-only window into order data
// (implemented by the order repository). Vendor makes no SQL against order
// tables.
type SettlementReader interface {
	// GetVendorSettledEarnings sums the delivered-order earnings of one vendor.
	GetVendorSettledEarnings(ctx context.Context, vendorID uuid.UUID) (int64, error)
	// GetVendorSalesStats projects revenue, earnings, and order counts.
	GetVendorSalesStats(ctx context.Context, vendorID uuid.UUID) (*model.VendorSalesStats, error)
}

// VendorOrderReader is the vendor portal's read-only window into orders
// containing the vendor's lines (implemented by the order repository).
// The reader owns the join; the vendor package never issues SQL against
// order tables. Only the vendor's own lines are ever returned.
type VendorOrderReader interface {
	// ListVendorOrders returns orders containing the vendor's lines, newest
	// first, with offset pagination and a matching total.
	ListVendorOrders(ctx context.Context, vendorID uuid.UUID, status *model.OrderStatus, q *pagination.ListQuery) ([]model.Order, int64, error)
	// GetVendorOrder loads one order only when it contains a line owned by
	// the vendor; foreign and platform orders answer "not found".
	GetVendorOrder(ctx context.Context, vendorID, orderID uuid.UUID) (*model.Order, error)
}

// ObjectReader is the subset of object persistence needed to resolve vendor
// asset URLs. Storage object UUIDs never cross the vendor API boundary:
// clients send public URLs and the object domain resolves them.
type ObjectReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.Object, error)
	ResolvePublicURL(ctx context.Context, rawURL string) (*model.Object, error)
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
	orders     VendorOrderReader
	products   ProductCatalog
	variants   VariantCatalog
	stock      StockAdjuster
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
	orders VendorOrderReader,
	products ProductCatalog,
	variants VariantCatalog,
	stock StockAdjuster,
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
	if orders == nil {
		panic("vendor: order reader is required")
	}
	if products == nil {
		panic("vendor: product catalog is required")
	}
	if variants == nil {
		panic("vendor: variant catalog is required")
	}
	if stock == nil {
		panic("vendor: stock adjuster is required")
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
		orders:     orders,
		products:   products,
		variants:   variants,
		stock:      stock,
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

// Slug policy shared by registration, profile edits, and the live
// availability endpoint. The DB CHECK constraint mirrors the length/charset.
const (
	MinSlugLength = 3
	MaxSlugLength = 63
)

var (
	slugPattern   = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	reservedSlugs = map[string]struct{}{
		"admin": {}, "api": {}, "auth": {}, "vendor": {}, "sell": {},
		"store": {}, "stores": {}, "products": {}, "categories": {}, "brands": {},
		"cart": {}, "checkout": {}, "account": {}, "user": {}, "orders": {},
		"payments": {}, "settings": {}, "support": {}, "help": {}, "about": {},
		"terms": {}, "privacy": {}, "login": {}, "register": {}, "search": {},
		"wishlist": {}, "en": {}, "ar": {},
	}
)

// NormalizeSlug applies the canonical NFKC + trim + lowercase policy.
func NormalizeSlug(raw string) string {
	return strings.ToLower(strings.TrimSpace(norm.NFKC.String(raw)))
}

// validateSlug returns the normalized slug and a user-facing rejection reason.
// An empty reason means the slug is valid.
func validateSlug(raw string) (string, string) {
	slug := NormalizeSlug(raw)
	if slug == "" {
		return slug, "Slug is required"
	}
	if len(slug) < MinSlugLength || len(slug) > MaxSlugLength {
		return slug, "Slug must be between 3 and 63 characters"
	}
	if !slugPattern.MatchString(slug) {
		return slug, "Slug must contain only lowercase letters, numbers, and single hyphens"
	}
	if _, reserved := reservedSlugs[slug]; reserved {
		return slug, "This slug is a reserved platform keyword"
	}
	return slug, ""
}

func validationError(field, constraint, message string) *apierr.APIError {
	return apierr.BadRequestError(message).
		WithCode(apierr.CodeValidationFailed).
		WithFields(api.FieldError{Field: field, Constraint: constraint, Message: message})
}

func slugTakenError() *apierr.APIError {
	return apierr.ConflictError("This store slug is already taken").
		WithCode(errcode.CodeVendorSlugTaken).
		WithFields(api.FieldError{
			Field: "slug", Constraint: "unique", Message: "this store slug is already taken",
		})
}

// SlugAvailability is the live availability response for onboarding.
type SlugAvailability struct {
	Available bool   `json:"available"`
	Slug      string `json:"slug"`
	Reason    string `json:"reason,omitempty"`
}

// CheckSlugAvailability validates syntax/reserved words and checks the
// authoritative uniqueness set. A merchant's own current slug is available
// so a rejected application can be resubmitted unchanged.
func (s *Service) CheckSlugAvailability(ctx context.Context, userID uuid.UUID, raw string) (*SlugAvailability, error) {
	slug, reason := validateSlug(raw)
	if reason != "" {
		return &SlugAvailability{Available: false, Slug: slug, Reason: reason}, nil
	}

	result := &SlugAvailability{Available: true, Slug: slug}
	err := s.runner.WithExecutor(ctx, func(ctx context.Context) error {
		vendor, err := s.repo.GetBySlug(ctx, slug)
		if err == nil {
			if vendor.UserID != userID {
				result.Available = false
				result.Reason = "This store slug is already taken"
			}
			return nil
		}
		if !errors.Is(database.MapError(err), database.ErrNotFound) {
			return s.handleError(err, "Failed to check vendor slug")
		}

		exists, err := s.repo.SlugExists(ctx, slug)
		if err != nil {
			return s.handleError(err, "Failed to check vendor slug")
		}
		if exists {
			result.Available = false
			result.Reason = "This store slug is already taken"
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// RegisterVendor creates a pending vendor application for an existing user.
// A rejected application is resubmitted (and returned to pending) instead of
// being rejected as a duplicate.
func (s *Service) RegisterVendor(ctx context.Context, userID uuid.UUID, in RegisterVendorInput) (*model.Vendor, error) {
	storeName := strings.TrimSpace(in.StoreName)
	if len(storeName) < 2 {
		return nil, validationError("storeName", "min", "Store name must be at least 2 characters")
	}
	if len(storeName) > 150 {
		return nil, validationError("storeName", "max", "Store name must be at most 150 characters")
	}
	description := strings.TrimSpace(in.Description)
	if len(description) > 2000 {
		return nil, validationError("description", "max", "Description must be at most 2000 characters")
	}

	slug, reason := validateSlug(in.Slug)
	if reason != "" {
		return nil, validationError("slug", "slug", reason)
	}

	var saved *model.Vendor
	slugConflict := false
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		if _, err := s.userReader.GetByID(ctx, userID); err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.NotFoundError("User not found").WithCode(apierr.CodeNotFound)
			}
			return s.handleError(err, "Failed to resolve user")
		}

		existing, err := s.repo.GetByUserID(ctx, userID)
		if err == nil {
			if existing.Status != model.VendorStatusRejected {
				return apierr.ConflictError("This account already has a vendor profile").
					WithCode(errcode.CodeVendorAlreadyRegistered)
			}

			updated, updateErr := s.repo.Update(ctx, existing.ID, UpdateVendorFields{
				StoreName:      &storeName,
				Slug:           &slug,
				Description:    &description,
				LogoObjectID:   in.LogoObjectID,
				BannerObjectID: in.BannerObjectID,
			})
			if updateErr != nil {
				if errors.Is(database.MapError(updateErr), database.ErrConflict) {
					return slugTakenError()
				}
				return updateErr
			}
			if updateErr := s.repo.MarkResubmitted(ctx, updated.ID); updateErr != nil {
				return updateErr
			}
			saved, err = s.repo.GetByID(ctx, updated.ID)
			return err
		}
		if !errors.Is(database.MapError(err), database.ErrNotFound) {
			return s.handleError(err, "Failed to check existing vendor profile")
		}

		slugTaken, err := s.repo.SlugExists(ctx, slug)
		if err != nil {
			return err
		}
		if slugTaken {
			return slugTakenError()
		}

		created, createErr := s.repo.Create(ctx, &CreateVendorInput{
			UserID:            userID,
			StoreName:         storeName,
			Slug:              slug,
			Description:       description,
			LogoObjectID:      in.LogoObjectID,
			BannerObjectID:    in.BannerObjectID,
			CommissionRateBps: s.settings.VendorCommissionBps(),
		})
		if createErr != nil {
			if errors.Is(database.MapError(createErr), database.ErrConflict) {
				// The transaction is aborted after a unique violation, so the
				// constraint lookup must happen after rollback below.
				slugConflict = true
			}
			return createErr
		}
		saved = created
		return nil
	})
	if err != nil {
		if slugConflict {
			var taken bool
			lookupErr := s.runner.WithExecutor(ctx, func(ctx context.Context) error {
				var e error
				taken, e = s.repo.SlugExists(ctx, slug)
				return e
			})
			if lookupErr != nil {
				return nil, s.handleError(lookupErr, "Failed to check vendor slug")
			}
			if taken {
				return nil, slugTakenError()
			}
			return nil, apierr.ConflictError("This account already has a vendor profile").
				WithCode(errcode.CodeVendorAlreadyRegistered)
		}
		return nil, s.handleError(err, "Failed to register vendor")
	}

	s.attachAssetURLs(ctx, saved)
	return saved, nil
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
	Slug            *string
	Description     *string
	LogoObjectID    *uuid.UUID
	BannerObjectID  *uuid.UUID
	BankAccountInfo map[string]any
}

// UpdateMyProfile edits the caller's own profile. Pending, active, and
// rejected applications may edit; a rejected edit re-enters review. Suspended
// vendors are frozen. Slugs are immutable once active.
func (s *Service) UpdateMyProfile(ctx context.Context, userID uuid.UUID, in UpdateVendorInput) (*model.Vendor, error) {
	if in.StoreName != nil {
		name := strings.TrimSpace(*in.StoreName)
		if len(name) < 2 || len(name) > 150 {
			return nil, validationError("storeName", "length", "Store name must be between 2 and 150 characters")
		}
		in.StoreName = &name
	}
	if in.Description != nil {
		description := strings.TrimSpace(*in.Description)
		if len(description) > 2000 {
			return nil, validationError("description", "max", "Description must be at most 2000 characters")
		}
		in.Description = &description
	}

	var normalizedSlug *string
	if in.Slug != nil {
		slug, reason := validateSlug(*in.Slug)
		if reason != "" {
			return nil, validationError("slug", "slug", reason)
		}
		normalizedSlug = &slug
	}

	var updated *model.Vendor
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		existing, err := s.repo.GetByUserID(ctx, userID)
		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.NotFoundError("Vendor not found").WithCode(errcode.CodeVendorNotFound)
			}
			return s.handleError(err, "Failed to load vendor")
		}

		switch existing.Status {
		case model.VendorStatusPending, model.VendorStatusActive, model.VendorStatusRejected:
		default:
			return apierr.ForbiddenError("Vendor is not authorized for this action").
				WithCode(errcode.CodeVendorNotAuthorizedForAction)
		}

		if existing.Status == model.VendorStatusActive && normalizedSlug != nil && *normalizedSlug != existing.Slug {
			return validationError("slug", "immutable", "Store slug cannot be changed after approval")
		}

		updated, err = s.repo.Update(ctx, existing.ID, UpdateVendorFields{
			StoreName:       in.StoreName,
			Slug:            normalizedSlug,
			Description:     in.Description,
			LogoObjectID:    in.LogoObjectID,
			BannerObjectID:  in.BannerObjectID,
			BankAccountInfo: in.BankAccountInfo,
		})
		if err != nil {
			if errors.Is(database.MapError(err), database.ErrConflict) {
				return slugTakenError()
			}
			return err
		}

		if existing.Status == model.VendorStatusRejected {
			if err := s.repo.MarkResubmitted(ctx, updated.ID); err != nil {
				return err
			}
			updated, err = s.repo.GetByID(ctx, updated.ID)
		}
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
	case from == model.VendorStatusRejected && to == model.VendorStatusPending:
		return true
	default:
		return false
	}
}

// AdminListVendors returns the marketplace's vendor directory, optionally
// narrowed to one lifecycle status, newest first.
func (s *Service) AdminListVendors(ctx context.Context, status *model.VendorStatus, q *pagination.ListQuery) (*pagination.PagedResult[model.Vendor], error) {
	var result *pagination.PagedResult[model.Vendor]
	err := s.runner.WithExecutor(ctx, func(ctx context.Context) error {
		vendors, total, err := s.repo.ListAll(ctx, status, q)
		if err != nil {
			return s.handleError(err, "Failed to list vendors")
		}

		result = pagination.NewPagedResult(vendors, pagination.NewPage(q.Page, q.PageSize, int(total)))
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Hydrate asset URLs outside the executor; the attach helper only
	// resolves object lookups and is safe to run without a DB context.
	for _, v := range result.Items {
		s.attachAssetURLs(ctx, v)
	}
	return result, nil
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

// AdminChangeVendorStatus applies a lifecycle transition guarded by the state
// machine and records the reviewing administrator plus a merchant-visible
// reason for rejection/suspension.
func (s *Service) AdminChangeVendorStatus(
	ctx context.Context,
	vendorID uuid.UUID,
	to model.VendorStatus,
	reason string,
	reviewedBy uuid.UUID,
) (*model.Vendor, error) {
	reason = strings.TrimSpace(reason)
	if len(reason) > 2000 {
		return nil, validationError("reason", "max", "Status reason must be at most 2000 characters")
	}
	if (to == model.VendorStatusRejected || to == model.VendorStatusSuspended) && reason == "" {
		return nil, validationError("reason", "required", "A reason is required for rejection or suspension")
	}

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

		var reasonPtr *string
		if to == model.VendorStatusRejected || to == model.VendorStatusSuspended {
			reasonPtr = &reason
		}
		if err := s.repo.SetStatus(ctx, vendorID, to, reasonPtr, &reviewedBy); err != nil {
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
func (s *Service) RequestPayout(ctx context.Context, userID uuid.UUID, amount int64, idempotencyKey string) (*model.VendorPayout, error) {
	if amount <= 0 {
		return nil, apierr.BadRequestError("Payout amount must be positive").
			WithCode(errcode.CodePayoutInvalidAmount)
	}
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if len(idempotencyKey) > 100 {
		return nil, validationError("Idempotency-Key", "max", "Idempotency key must be at most 100 characters")
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

		if idempotencyKey != "" {
			replayed, lookupErr := s.repo.GetPayoutByIdempotencyKey(ctx, existing.ID, idempotencyKey)
			if lookupErr == nil {
				if replayed.Amount != amount {
					return apierr.ConflictError("Idempotency key was already used for a different amount").
						WithCode(errcode.CodePayoutIdempotencyConflict)
				}
				payout = replayed
				return nil
			}
			if !errors.Is(database.MapError(lookupErr), database.ErrNotFound) {
				return s.handleError(lookupErr, "Failed to check payout idempotency key")
			}
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

		var keyPtr *string
		if idempotencyKey != "" {
			keyPtr = &idempotencyKey
		}
		payout, err = s.repo.CreatePayout(ctx, existing.ID, amount, available, keyPtr)
		if err != nil && keyPtr != nil && errors.Is(database.MapError(err), database.ErrConflict) {
			// A concurrent retry may have won the insert. Return that request
			// when it represents the same withdrawal intent.
			replayed, lookupErr := s.repo.GetPayoutByIdempotencyKey(ctx, existing.ID, idempotencyKey)
			if lookupErr == nil {
				if replayed.Amount != amount {
					return apierr.ConflictError("Idempotency key was already used for a different amount").
						WithCode(errcode.CodePayoutIdempotencyConflict)
				}
				payout = replayed
				return nil
			}
		}
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
		existing, err := s.requireVendorWithStatus(ctx, userID, model.VendorStatusActive, model.VendorStatusSuspended)
		if err != nil {
			return err
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

		// Suspended/rejected stores may not receive money. Rejection remains
		// legal so the funds can be released back to the balance.
		if to == model.PayoutStatusApproved || to == model.PayoutStatusProcessing || to == model.PayoutStatusPaid {
			vendor, err := s.repo.GetByIDForUpdate(ctx, existing.VendorID)
			if err != nil {
				return s.handleError(err, "Failed to load vendor")
			}
			if vendor.Status != model.VendorStatusActive {
				return apierr.ConflictError("Payouts cannot be processed while vendor is not active").
					WithCode(errcode.CodeVendorNotAuthorizedForAction)
			}
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
