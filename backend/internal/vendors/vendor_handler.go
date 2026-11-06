package vendors

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/catalog/inventory"
	"github.com/nullopt-t/prim-backend/internal/catalog/product"
	"github.com/nullopt-t/prim-backend/internal/catalog/product/variant"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/object"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
	"github.com/nullopt-t/prim-backend/internal/shared/validation"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
)

// Handler exposes vendor portal and admin endpoints over HTTP.
type Handler struct {
	service *Service
	media   MediaUploader
}

// MediaUploader is the narrow slice of the object domain the vendor portal
// needs: merchant image uploads. Storage object ids stay internal; the
// vendor surface only ever returns public URLs.
type MediaUploader interface {
	Upload(c *gin.Context) (*object.MediaItem, error)
}

func NewHandler(service *Service, media MediaUploader) *Handler {
	if service == nil {
		panic("vendor: service is required")
	}
	if media == nil {
		panic("vendor: media uploader is required")
	}
	return &Handler{service: service, media: media}
}

// ---- Responses ----------------------------------------------------------

// VendorResponse is the merchant profile payload.
type VendorResponse struct {
	// Vendor profile id.
	ID uuid.UUID `json:"id" example:"50000000-0000-0000-0000-000000000001"`
	// Owning account id.
	UserID uuid.UUID `json:"userId" example:"10000000-0000-0000-0000-000000000001"`
	// Customer-facing store display name.
	StoreName string `json:"storeName" example:"Amr's Gadgets"`
	// URL-safe unique store identifier.
	Slug string `json:"slug" example:"amrs-gadgets"`
	// Optional storefront description.
	Description string `json:"description,omitempty" example:"Electronics and accessories"`
	// Public logo URL, when set.
	LogoURL string `json:"logoUrl,omitempty"`
	// Public banner URL, when set.
	BannerURL string `json:"bannerUrl,omitempty"`
	// Platform commission in basis points.
	CommissionRateBps int `json:"commissionRateBps" example:"1000"`
	// Lifecycle status.
	Status model.VendorStatus `json:"status" enums:"pending,active,suspended,rejected"`
	// Merchant-visible reason for the latest rejection or suspension.
	StatusReason string `json:"statusReason,omitempty"`
	// Administrator who made the latest lifecycle decision.
	ReviewedBy *uuid.UUID `json:"reviewedBy,omitempty"`
	// When the latest lifecycle decision was made (RFC3339).
	ReviewedAt string `json:"reviewedAt,omitempty"`
	// Created timestamp (RFC3339).
	CreatedAt string `json:"createdAt" example:"2026-11-02T10:00:00Z"`
	// Updated timestamp (RFC3339).
	UpdatedAt string `json:"updatedAt" example:"2026-11-02T10:00:00Z"`
}

// PayoutResponse is one ledger row.
type PayoutResponse struct {
	// Payout record id.
	ID uuid.UUID `json:"id" example:"60000000-0000-0000-0000-000000000001"`
	// Vendor id.
	VendorID uuid.UUID `json:"vendorId" example:"50000000-0000-0000-0000-000000000001"`
	// Amount in minor units.
	Amount int64 `json:"amount" example:"125000"`
	// ISO currency.
	Currency string `json:"currency" example:"EGP"`
	// Lifecycle status.
	Status model.PayoutStatus `json:"status" enums:"requested,approved,processing,paid,rejected"`
	// External transfer reference when available.
	ReferenceNumber string `json:"referenceNumber,omitempty"`
	// Optional admin notes (rejection reasons, transfer remarks).
	Notes string `json:"notes,omitempty"`
	// Available balance stamped at request time.
	SnapshotAvailableBalance int64 `json:"snapshotAvailableBalance" example:"300000"`
	// Created timestamp (RFC3339).
	CreatedAt string `json:"createdAt" example:"2026-11-02T10:00:00Z"`
	// Terminal timestamp (RFC3339), when reached.
	ProcessedAt string `json:"processedAt,omitempty" example:"2026-11-03T10:00:00Z"`
}

func toVendorResponse(v *model.Vendor) VendorResponse {
	res := VendorResponse{
		ID:                v.ID,
		UserID:            v.UserID,
		StoreName:         v.StoreName,
		Slug:              v.Slug,
		Description:       v.Description,
		LogoURL:           v.LogoURL,
		BannerURL:         v.BannerURL,
		CommissionRateBps: v.CommissionRateBps,
		Status:            v.Status,
		StatusReason:      v.StatusReason,
		ReviewedBy:        v.ReviewedBy,
		CreatedAt:         v.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         v.UpdatedAt.Format(time.RFC3339),
	}
	if v.ReviewedAt != nil {
		res.ReviewedAt = v.ReviewedAt.Format(time.RFC3339)
	}
	return res
}

func toPayoutResponse(p *model.VendorPayout) PayoutResponse {
	res := PayoutResponse{
		ID:                       p.ID,
		VendorID:                 p.VendorID,
		Amount:                   p.Amount,
		Currency:                 p.Currency,
		Status:                   p.Status,
		Notes:                    p.Notes,
		SnapshotAvailableBalance: p.SnapshotAvailableBalance,
		CreatedAt:                p.CreatedAt.Format(time.RFC3339),
	}
	if p.ReferenceNumber != nil {
		res.ReferenceNumber = *p.ReferenceNumber
	}
	if p.ProcessedAt != nil {
		res.ProcessedAt = p.ProcessedAt.Format(time.RFC3339)
	}
	return res
}

// ---- Portal: registration & profile -------------------------------------

// RegisterVendorRequest is the onboarding payload.
type RegisterVendorRequest struct {
	StoreName      string `json:"storeName" binding:"required,min=2,max=150" example:"Amr's Gadgets"`
	Slug           string `json:"slug" binding:"required,max=150" example:"amrs-gadgets"`
	Description    string `json:"description" binding:"max=2000" example:"Electronics and accessories"`
	LogoObjectID   string `json:"logoObjectId" binding:"omitempty,uuid" example:"10000000-0000-0000-0000-00000000000f"`
	BannerObjectID string `json:"bannerObjectId" binding:"omitempty,uuid" example:"10000000-0000-0000-0000-00000000000e"`
}

// RegisterVendor godoc
// @Summary      Register as vendor
// @Description  Creates a pending vendor application for any authenticated account. A rejected profile may resubmit, which updates the application and returns it to pending review.
// @Tags         Vendor
// @Accept       json
// @Produce      json
// @Param        request body RegisterVendorRequest true "Vendor registration payload"
// @Success      201 {object} VendorResponse
// @Failure      400 {object} api.ErrorResponse "Validation failed"
// @Failure      401 {object} api.ErrorResponse "Authentication required"
// @Failure      409 {object} api.ErrorResponse "Already registered or slug taken"
// @Router       /vendor/register [post]
// @Security     SessionAuth
func (h *Handler) RegisterVendor(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	var req RegisterVendorRequest
	if !bindVendorJSON(c, &req) {
		return
	}

	vendor, err := h.service.RegisterVendor(c.Request.Context(), userID, RegisterVendorInput{
		StoreName:      req.StoreName,
		Slug:           req.Slug,
		Description:    req.Description,
		LogoObjectID:   parseOptionalUUID(req.LogoObjectID),
		BannerObjectID: parseOptionalUUID(req.BannerObjectID),
	})
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, api.DataResponse{Data: toVendorResponse(vendor)})
}

// UploadMedia godoc
// @Summary      Upload a merchant product image
// @Description  Uploads a raster image to the catalog bucket and returns its public URL, which callers pass as thumbnailUrl when creating variants. Only registered vendors may upload.
// @Tags         Vendor
// @Accept       multipart/form-data
// @Param        file formData string true "Image file (JPEG, PNG, GIF, WebP or AVIF, up to 10 MB)"
// @Success      201 {object} api.DataResponse{data=VendorMediaUploadResponse}
// @Failure      401 {object} api.ErrorResponse "Authentication required"
// @Failure      404 {object} api.ErrorResponse "Vendor not found"
// @Router       /vendor/media [post]
// @Security     SessionAuth
func (h *Handler) UploadMedia(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	if _, err := h.service.GetVendorByUser(c.Request.Context(), userID); err != nil {
		_ = c.Error(err)
		return
	}

	item, err := h.media.Upload(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, api.DataResponse{Data: VendorMediaUploadResponse{URL: item.URL}})
}

// VendorMediaUploadResponse is the merchant-facing upload result. It exposes
// only the public URL; storage object ids never leave the backend.
type VendorMediaUploadResponse struct {
	// Public URL of the uploaded image.
	URL string `json:"url" example:"https://cdn.example.com/catalog/8f14e45f-ea0a-4a7a-9c9a-123456789abc"`
}

// GetMyVendor godoc
// @Summary      Get my vendor profile
// @Description  Returns the vendor profile belonging to the authenticated user.
// @Tags         Vendor
// @Produce      json
// @Success      200 {object} VendorResponse
// @Failure      401 {object} api.ErrorResponse "Authentication required"
// @Failure      404 {object} api.ErrorResponse "Vendor not found"
// @Router       /vendor/profile [get]
// @Security     SessionAuth
func (h *Handler) GetMyVendor(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	vendor, err := h.service.GetVendorByUser(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: toVendorResponse(vendor)})
}

// CheckSlugAvailability godoc
// @Summary      Check store slug availability
// @Description  Validates slug syntax and reserved words, then checks the authoritative uniqueness set. A merchant's own current slug is reported available for re-application.
// @Tags         Vendor
// @Produce      json
// @Param        slug query string true "Candidate store slug"
// @Success      200 {object} api.DataResponse{data=SlugAvailability}
// @Failure      401 {object} api.UnauthorizedErrorResponse
// @Router       /vendor/slug-availability [get]
// @Security     SessionAuth
func (h *Handler) CheckSlugAvailability(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	availability, err := h.service.CheckSlugAvailability(c.Request.Context(), userID, c.Query("slug"))
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: availability})
}

// UpdateMyVendorRequest is the profile edit payload; nil fields are unchanged.
type UpdateMyVendorRequest struct {
	StoreName       string         `json:"storeName" binding:"omitempty,min=2,max=150" example:"Amr's Gadgets"`
	Slug            string         `json:"slug" binding:"omitempty,max=63" example:"amrs-gadgets"`
	Description     string         `json:"description" binding:"max=2000" example:"Electronics and accessories"`
	LogoObjectID    string         `json:"logoObjectId" binding:"omitempty,uuid"`
	BannerObjectID  string         `json:"bannerObjectId" binding:"omitempty,uuid"`
	BankAccountInfo map[string]any `json:"bankAccountInfo"`
}

// UpdateMyVendor godoc
// @Summary      Update my vendor profile
// @Description  Edits the merchant profile for pending, active, or rejected vendors. Rejected edits return the application to pending; suspended vendors are blocked. The slug is immutable once the vendor is active. Commission and status remain admin-controlled.
// @Tags         Vendor
// @Accept       json
// @Produce      json
// @Param        request body UpdateMyVendorRequest true "Profile update payload"
// @Success      200 {object} VendorResponse
// @Failure      400 {object} api.ErrorResponse "Validation failed"
// @Failure      403 {object} api.ErrorResponse "Vendor not authorized"
// @Failure      404 {object} api.ErrorResponse "Vendor not found"
// @Router       /vendor/profile [patch]
// @Security     SessionAuth
func (h *Handler) UpdateMyVendor(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	var req UpdateMyVendorRequest
	if !bindVendorJSON(c, &req) {
		return
	}

	in := UpdateVendorInput{
		BankAccountInfo: req.BankAccountInfo,
	}
	if req.StoreName != "" {
		in.StoreName = &req.StoreName
	}
	if req.Slug != "" {
		in.Slug = &req.Slug
	}
	if req.Description != "" {
		in.Description = &req.Description
	}
	if req.LogoObjectID != "" {
		in.LogoObjectID = parseOptionalUUID(req.LogoObjectID)
	}
	if req.BannerObjectID != "" {
		in.BannerObjectID = parseOptionalUUID(req.BannerObjectID)
	}

	vendor, err := h.service.UpdateMyProfile(c.Request.Context(), userID, in)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: toVendorResponse(vendor)})
}

// bindVendorJSON binds the request body and renders validation errors via the
// shared middleware.
func bindVendorJSON(c *gin.Context, req any) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		validation.ValidationError(c, err)
		return false
	}
	return true
}

// Helper: parse an optional UUID string; empty means absent.
func parseOptionalUUID(s string) *uuid.UUID {
	if s == "" {
		return nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return nil
	}
	return &id
}

// Helper: dereference an optional string; nil means absent.
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ---- Admin lifecycle ----------------------------------------------------

// AdminChangeVendorStatusRequest is the admin lifecycle payload.
type AdminChangeVendorStatusRequest struct {
	// New vendor lifecycle state.
	Status model.VendorStatus `json:"status" binding:"required,oneof=pending active suspended rejected" enums:"pending,active,suspended,rejected"`
	// Merchant-visible reason; required for rejection and suspension.
	Reason string `json:"reason" binding:"omitempty,max=2000"`
}

// AdminChangeVendorStatus godoc
// @Summary      Change vendor status (admin)
// @Description  Applies a guarded lifecycle transition (pending->active/rejected, active<->suspended) and records reviewer metadata. Rejection and suspension require a merchant-facing reason.
// @Tags         Admin Vendors
// @Accept       json
// @Produce      json
// @Param        id      path string                       true "Vendor UUID" format(uuid)
// @Param        request body AdminChangeVendorStatusRequest true "New status"
// @Success      200 {object} VendorResponse
// @Failure      400 {object} api.BadReqResponse "Validation failed"
// @Failure      404 {object} api.NotFoundErrorResponse "Vendor not found"
// @Failure      409 {object} api.ConflictErrorResponse "Transition not permitted"
// @Router       /admin/vendors/{id}/status [patch]
// @Security     SessionAuth
func (h *Handler) AdminChangeVendorStatus(c *gin.Context) {
	vendorID, ok := bindVendorID(c)
	if !ok {
		return
	}

	var req AdminChangeVendorStatusRequest
	if !bindVendorJSON(c, &req) {
		return
	}

	reviewedBy, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	vendor, err := h.service.AdminChangeVendorStatus(c.Request.Context(), vendorID, req.Status, req.Reason, reviewedBy)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: toVendorResponse(vendor)})
}

// AdminSetCommissionRequest is the admin commission payload.
type AdminSetCommissionRequest struct {
	// Commission as basis points (0..10000).
	RateBps int `json:"rateBps" binding:"required,min=0,max=10000" example:"1000"`
}

// AdminSetCommission godoc
// @Summary      Set vendor commission rate (admin)
// @Description  Overrides the per-vendor commission in basis points. Future orders use the new rate; already-placed orders keep their frozen earning.
// @Tags         Admin Vendors
// @Accept       json
// @Produce      json
// @Param        id      path string true "Vendor UUID" format(uuid)
// @Param        request body AdminSetCommissionRequest true "Rate in bps"
// @Success      200 {object} VendorResponse
// @Failure      400 {object} api.BadReqResponse "Validation failed"
// @Failure      404 {object} api.NotFoundErrorResponse "Vendor not found"
// @Router       /admin/vendors/{id}/commission [patch]
// @Security     SessionAuth
func (h *Handler) AdminSetCommission(c *gin.Context) {
	vendorID, ok := bindVendorID(c)
	if !ok {
		return
	}

	var req AdminSetCommissionRequest
	if !bindVendorJSON(c, &req) {
		return
	}

	vendor, err := h.service.AdminSetCommissionRate(c.Request.Context(), vendorID, req.RateBps)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: toVendorResponse(vendor)})
}

// AdminListVendors godoc
// @Summary      List vendors (admin)
// @Description  Returns the vendor directory, newest first, optionally filtered by lifecycle status.
// @Tags         Admin Vendors
// @Produce      json
// @Param        status    query string false "Lifecycle filter" Enums(pending, active, suspended, rejected)
// @Param        page      query int    false "Page number (1-based)"
// @Param        pageSize  query int    false "Page size (max 100)"
// @Success      200 {object} api.PaginatedResponse{data=[]VendorResponse}
// @Failure      400 {object} api.BadReqResponse "Invalid status filter"
// @Router       /admin/vendors [get]
// @Security     SessionAuth
func (h *Handler) AdminListVendors(c *gin.Context) {
	var status *model.VendorStatus
	if raw := c.Query("status"); raw != "" {
		s := model.VendorStatus(raw)
		switch s {
		case model.VendorStatusPending, model.VendorStatusActive, model.VendorStatusSuspended, model.VendorStatusRejected:
		default:
			_ = c.Error(apierr.BadRequestError("Invalid vendor status filter"))
			return
		}
		status = &s
	}

	var q pagination.ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		validation.ValidationError(c, err)
		return
	}
	q.Process(pagination.QueryOptions{DefaultPageSize: 20, MaxPageSize: 100})

	result, err := h.service.AdminListVendors(c.Request.Context(), status, &q)
	if err != nil {
		_ = c.Error(err)
		return
	}

	responses := make([]VendorResponse, 0, len(result.Items))
	for _, v := range result.Items {
		responses = append(responses, toVendorResponse(v))
	}

	c.JSON(http.StatusOK, api.PaginatedResponse{
		Data: responses,
		Meta: result.Meta(),
	})
}

// AdminGetVendor godoc
// @Summary      Get vendor by ID (admin)
// @Description  Returns any vendor profile for administration.
// @Tags         Admin Vendors
// @Produce      json
// @Param        id path string true "Vendor UUID" format(uuid)
// @Success      200 {object} VendorResponse
// @Failure      404 {object} api.NotFoundErrorResponse "Vendor not found"
// @Router       /admin/vendors/{id} [get]
// @Security     SessionAuth
func (h *Handler) AdminGetVendor(c *gin.Context) {
	vendorID, ok := bindVendorID(c)
	if !ok {
		return
	}

	vendor, err := h.service.AdminGetVendorByID(c.Request.Context(), vendorID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: toVendorResponse(vendor)})
}

// ---- Payouts ------------------------------------------------------------

// RequestPayoutRequest is the vendor payout request payload.
type RequestPayoutRequest struct {
	// Amount in minor currency units; must not exceed the available balance.
	Amount int64 `json:"amount" binding:"required,min=1" example:"125000"`
}

// RequestPayout godoc
// @Summary      Request a payout
// @Description  Registers a payout request against settled earnings, validating the available balance under the vendor row lock. Active vendors only. Reusing an Idempotency-Key with the same amount replays the original request; a different amount returns 409.
// @Tags         Vendor
// @Accept       json
// @Produce      json
// @Param        request body RequestPayoutRequest true "Payout request"
// @Param        Idempotency-Key header string false "Client retry key (max 100 chars)"
// @Success      201 {object} PayoutResponse
// @Failure      400 {object} api.BadReqResponse "Validation failed or insufficient balance"
// @Failure      403 {object} api.ForbiddenErrorResponse "Vendor not active"
// @Failure      404 {object} api.NotFoundErrorResponse "Vendor not found"
// @Router       /vendor/payouts [post]
// @Security     SessionAuth
func (h *Handler) RequestPayout(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	var req RequestPayoutRequest
	if !bindVendorJSON(c, &req) {
		return
	}

	payout, err := h.service.RequestPayout(c.Request.Context(), userID, req.Amount, c.GetHeader("Idempotency-Key"))
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, api.DataResponse{Data: toPayoutResponse(payout)})
}

// ListMyPayouts godoc
// @Summary      List my payouts
// @Description  Returns the caller's payout ledger including rejected records (full audit trail). Available to active and suspended vendors.
// @Tags         Vendor
// @Produce      json
// @Success      200 {object} api.DataResponse{data=[]PayoutResponse}
// @Failure      401 {object} api.UnauthorizedErrorResponse
// @Failure      404 {object} api.NotFoundErrorResponse "Vendor not found"
// @Router       /vendor/payouts [get]
// @Security     SessionAuth
func (h *Handler) ListMyPayouts(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	payouts, err := h.service.ListMyPayouts(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: toPayoutList(payouts)})
}

// AdminListPayouts godoc
// @Summary      List all payouts (admin)
// @Description  Returns every payout across all vendors for administration.
// @Tags         Admin Vendors
// @Produce      json
// @Success      200 {object} api.DataResponse{data=[]PayoutResponse}
// @Router       /admin/vendors/payouts [get]
// @Security     SessionAuth
func (h *Handler) AdminListPayouts(c *gin.Context) {
	payouts, err := h.service.AdminListPayouts(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: toPayoutList(payouts)})
}

// AdminTransitionPayoutRequest is the admin payout action payload.
type AdminTransitionPayoutRequest struct {
	// Target payout state; validated against the lifecycle by the service.
	Status model.PayoutStatus `json:"status" binding:"required,oneof=approved processing paid rejected" enums:"approved,processing,paid,rejected"`
	// External transfer reference (for processing/paid).
	ReferenceNumber string `json:"referenceNumber" binding:"omitempty,max=100" example:"PAY-2026-0001"`
	// Admin notes, required for rejection reasons.
	Notes string `json:"notes" binding:"max=2000"`
}

// AdminTransitionPayout godoc
// @Summary      Transition a payout (admin)
// @Description  Moves a payout through its lifecycle (conditional on prior state). State conflicts return 409.
// @Tags         Admin Vendors
// @Accept       json
// @Produce      json
// @Param        id path string true "Payout UUID" format(uuid)
// @Param        request body AdminTransitionPayoutRequest true "Transition payload"
// @Success      200 {object} PayoutResponse
// @Failure      404 {object} api.NotFoundErrorResponse "Payout not found"
// @Failure      409 {object} api.ConflictErrorResponse "Transition not permitted"
// @Router       /admin/vendors/payouts/{id}/transition [patch]
// @Security     SessionAuth
func (h *Handler) AdminTransitionPayout(c *gin.Context) {
	payoutID, ok := bindVendorID(c)
	if !ok {
		return
	}

	var req AdminTransitionPayoutRequest
	if !bindVendorJSON(c, &req) {
		return
	}

	var ref *string
	if req.ReferenceNumber != "" {
		ref = &req.ReferenceNumber
	}
	var notes *string
	if req.Notes != "" {
		notes = &req.Notes
	}

	payout, err := h.service.AdminTransitionPayout(c.Request.Context(), payoutID, req.Status, ref, notes)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: toPayoutResponse(payout)})
}

// bindVendorID parses the route ":id" path parameter.
func bindVendorID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.InvalidUUIDError().WithFields(
			api.FieldError{Field: "id", Constraint: "uuid", Message: "invalid vendor id"},
		))
		return uuid.Nil, false
	}
	return id, true
}

func toPayoutList(payouts []*model.VendorPayout) []PayoutResponse {
	list := make([]PayoutResponse, 0, len(payouts))
	for _, p := range payouts {
		list = append(list, toPayoutResponse(p))
	}
	return list
}

// GetDashboard godoc
// @Summary      Vendor dashboard
// @Description  Returns the merchant's money position (settled, pending, outstanding) and catalog health counts. Active vendors only; other lifecycle states receive 403.
// @Tags         Vendor
// @Produce      json
// @Success      200 {object} api.DataResponse{data=Dashboard}
// @Failure      401 {object} api.UnauthorizedErrorResponse
// @Failure      404 {object} api.NotFoundErrorResponse "Vendor not found"
// @Router       /vendor/dashboard [get]
// @Security     SessionAuth
func (h *Handler) GetDashboard(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	dashboard, err := h.service.GetDashboard(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: dashboard})
}

// ---- Vendor product catalog ---------------------------------------------

// VendorCreateProductRequest is the merchant product creation payload.
type VendorCreateProductRequest struct {
	// Optional brand UUID.
	BrandID *string `json:"brandId" binding:"omitempty,uuid" example:"30000000-0000-0000-0000-000000000001"`
	// Required category UUID.
	CategoryID string `json:"categoryId" binding:"required,uuid" example:"40000000-0000-0000-0000-000000000002"`
	// URL-friendly product identifier.
	Slug string `json:"slug" binding:"required,max=200" example:"wireless-headphones"`
	// Product display title.
	Title string `json:"title" binding:"required,max=200" example:"Wireless Headphones"`
	// Full product description.
	Description *string `json:"description,omitempty"`
	// Key bullet points shown on the storefront.
	Highlights []string `json:"highlights,omitempty" example:"Active Noise Cancellation,30h battery"`
	// Product classification: simple or variable.
	ProductType string `json:"productType" binding:"omitempty,oneof=simple variable" example:"simple"`
	// Ordered variation axis keys for variable products.
	VariationAxes []string `json:"variationAxes,omitempty" example:"color,storage"`
}

// VendorUpdateProductRequest is the merchant product edit payload; omitted
// fields are unchanged.
type VendorUpdateProductRequest struct {
	// Optional brand UUID.
	BrandID *string `json:"brandId" binding:"omitempty,uuid"`
	// Optional category UUID.
	CategoryID *string `json:"categoryId" binding:"omitempty,uuid"`
	// Updated product title.
	Title *string `json:"title,omitempty" binding:"omitempty,max=200"`
	// Updated product description.
	Description *string `json:"description,omitempty"`
	// Updated highlight bullets.
	Highlights []string `json:"highlights,omitempty"`
	// Product classification: simple or variable.
	ProductType *string `json:"productType,omitempty" binding:"omitempty,oneof=simple variable"`
	// Variation axes; omit to leave unchanged, [] to make the product simple.
	VariationAxes *[]string `json:"variationAxes,omitempty"`
}

// VendorProductResponse is the merchant-facing product payload.
type VendorProductResponse struct {
	// Product UUID.
	ID uuid.UUID `json:"id" example:"70000000-0000-0000-0000-000000000001"`
	// Owning vendor UUID.
	VendorID uuid.UUID `json:"vendorId" example:"50000000-0000-0000-0000-000000000001"`
	// URL-friendly slug.
	Slug string `json:"slug" example:"wireless-headphones"`
	// Display title.
	Title string `json:"title" example:"Wireless Headphones"`
	// Lifecycle status.
	Status model.PublicationStatus `json:"status" enums:"draft,published,archived"`
	// Product classification.
	ProductType model.ProductType `json:"productType" enums:"simple,variable"`
	// Created timestamp (RFC3339).
	CreatedAt string `json:"createdAt"`
	// Updated timestamp (RFC3339).
	UpdatedAt string `json:"updatedAt"`
}

func toVendorProductResponse(p *model.Product) VendorProductResponse {
	res := VendorProductResponse{
		ID:          p.ID,
		Slug:        p.Slug,
		Title:       p.Title,
		Status:      p.Status,
		ProductType: p.ProductType,
		CreatedAt:   p.CreatedAt.Format(timeLayout),
		UpdatedAt:   p.UpdatedAt.Format(timeLayout),
	}
	if p.VendorID != nil {
		res.VendorID = *p.VendorID
	}
	return res
}

// ListMyProducts godoc
// @Summary      List my products
// @Description  Returns the authenticated vendor's catalog, drafts included. Ownership is enforced server-side and the vendor must be active.
// @Tags         Vendor
// @Produce      json
// @Param        page      query int    false "Page number (1-based)"
// @Param        pageSize  query int    false "Page size (max 100)"
// @Param        status    query string false "Publication status filter" Enums(draft, published, archived)
// @Param        category  query string false "Category name(s) or UUID(s), comma-separated"
// @Param        brand     query string false "Brand name(s) or UUID(s), comma-separated"
// @Success      200 {object} api.PaginatedResponse{data=[]VendorProductSummary}
// @Failure      401 {object} api.UnauthorizedErrorResponse
// @Failure      404 {object} api.NotFoundErrorResponse "Vendor not found"
// @Router       /vendor/products [get]
// @Security     SessionAuth
func (h *Handler) ListMyProducts(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	var q product.ListProductsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		validation.ValidationError(c, err)
		return
	}
	q.Process(pagination.QueryOptions{DefaultPageSize: 10, MaxPageSize: 100})
	if _, err := q.Keyset(); err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid pagination cursor").Wrap(err))
		return
	}

	result, err := h.service.ListOwnProducts(c.Request.Context(), userID, &q)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.PaginatedResponse{
		Data: result.Items,
		Meta: result.Meta(),
	})
}

// CreateProduct godoc
// @Summary      Create a product
// @Description  Creates a draft product owned by the authenticated vendor. Requires an active vendor profile.
// @Tags         Vendor
// @Accept       json
// @Produce      json
// @Param        request body VendorCreateProductRequest true "Product draft payload"
// @Success      201 {object} api.DataResponse{data=VendorProductResponse}
// @Failure      400 {object} api.BadReqResponse "Validation failed"
// @Failure      403 {object} api.ForbiddenErrorResponse "Vendor not active"
// @Failure      404 {object} api.NotFoundErrorResponse "Vendor not found"
// @Router       /vendor/products [post]
// @Security     SessionAuth
func (h *Handler) CreateProduct(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	var req VendorCreateProductRequest
	if !bindVendorJSON(c, &req) {
		return
	}

	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		_ = c.Error(apierr.InvalidUUIDError().WithFields(api.FieldError{
			Field: "categoryId", Constraint: "uuid", Message: "categoryId must be a UUID",
		}))
		return
	}

	var brandID *uuid.UUID
	if req.BrandID != nil {
		brandID = parseOptionalUUID(*req.BrandID)
	}

	created, err := h.service.CreateProduct(c.Request.Context(), userID, product.CreateProductInput{
		BrandID:       brandID,
		CategoryID:    categoryID,
		Slug:          &req.Slug,
		Title:         req.Title,
		Description:   req.Description,
		Highlights:    req.Highlights,
		ProductType:   req.ProductType,
		VariationAxes: req.VariationAxes,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, api.DataResponse{Data: toVendorProductResponse(created)})
}

// UpdateProduct godoc
// @Summary      Update a product
// @Description  Edits a product owned by the authenticated vendor. Foreign or platform products return 404.
// @Tags         Vendor
// @Accept       json
// @Produce      json
// @Param        id      path string                     true "Product UUID" format(uuid)
// @Param        request body VendorUpdateProductRequest true "Product edit payload"
// @Success      200 {object} api.MessageResponse
// @Failure      400 {object} api.BadReqResponse "Validation failed"
// @Failure      403 {object} api.ForbiddenErrorResponse "Vendor not active"
// @Failure      404 {object} api.NotFoundErrorResponse "Product or vendor not found"
// @Router       /vendor/products/{id} [patch]
// @Security     SessionAuth
func (h *Handler) UpdateProduct(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	productID, ok := bindVendorID(c)
	if !ok {
		return
	}

	var req VendorUpdateProductRequest
	if !bindVendorJSON(c, &req) {
		return
	}

	var brandID *uuid.UUID
	if req.BrandID != nil {
		brandID = parseOptionalUUID(*req.BrandID)
	}
	var categoryID *uuid.UUID
	if req.CategoryID != nil {
		categoryID = parseOptionalUUID(*req.CategoryID)
	}

	if err := h.service.UpdateProduct(c.Request.Context(), userID, productID, product.UpdateProductInput{
		BrandID:       brandID,
		CategoryID:    categoryID,
		Title:         req.Title,
		Description:   req.Description,
		Highlights:    req.Highlights,
		ProductType:   req.ProductType,
		VariationAxes: req.VariationAxes,
	}); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "Product updated"})
}

// DeleteProduct godoc
// @Summary      Delete a product
// @Description  Soft-deletes a product owned by the authenticated vendor.
// @Tags         Vendor
// @Produce      json
// @Param        id path string true "Product UUID" format(uuid)
// @Success      200 {object} api.MessageResponse
// @Failure      403 {object} api.ForbiddenErrorResponse "Vendor not active"
// @Failure      404 {object} api.NotFoundErrorResponse "Product or vendor not found"
// @Router       /vendor/products/{id} [delete]
// @Security     SessionAuth
func (h *Handler) DeleteProduct(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	productID, ok := bindVendorID(c)
	if !ok {
		return
	}

	if err := h.service.DeleteProduct(c.Request.Context(), userID, productID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "Product deleted"})
}

// PublishProduct godoc
// @Summary      Publish a product
// @Description  Publishes a product owned by the authenticated vendor. Product-domain preconditions (e.g. at least one variant) still apply.
// @Tags         Vendor
// @Produce      json
// @Param        id path string true "Product UUID" format(uuid)
// @Success      200 {object} api.MessageResponse
// @Failure      400 {object} api.BadReqResponse "Publishing preconditions not met"
// @Failure      403 {object} api.ForbiddenErrorResponse "Vendor not active"
// @Failure      404 {object} api.NotFoundErrorResponse "Product or vendor not found"
// @Router       /vendor/products/{id}/publish [post]
// @Security     SessionAuth
func (h *Handler) PublishProduct(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	productID, ok := bindVendorID(c)
	if !ok {
		return
	}

	if err := h.service.PublishProduct(c.Request.Context(), userID, productID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "Product published"})
}

// ArchiveProduct godoc
// @Summary      Archive a product
// @Description  Archives a product owned by the authenticated vendor, removing it from the storefront.
// @Tags         Vendor
// @Produce      json
// @Param        id path string true "Product UUID" format(uuid)
// @Success      200 {object} api.MessageResponse
// @Failure      403 {object} api.ForbiddenErrorResponse "Vendor not active"
// @Failure      404 {object} api.NotFoundErrorResponse "Product or vendor not found"
// @Router       /vendor/products/{id}/archive [post]
// @Security     SessionAuth
func (h *Handler) ArchiveProduct(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	productID, ok := bindVendorID(c)
	if !ok {
		return
	}

	if err := h.service.ArchiveProduct(c.Request.Context(), userID, productID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "Product archived"})
}

// ---- Vendor variants & stock --------------------------------------------

// VendorCreateVariantRequest is the merchant variant creation payload.
type VendorCreateVariantRequest struct {
	// Optional custom SKU; generated when omitted.
	SKU *string `json:"sku,omitempty" example:"SKU-BLK-01"`
	// Variant display title.
	Title string `json:"title" binding:"required" example:"Black, 128GB"`
	// Price in minor units.
	Price *int64 `json:"price,omitempty" example:"249900"`
	// Original/crossed-out price in minor units.
	CrossedOutPrice *int64 `json:"crossedOutPrice,omitempty" example:"299900"`
	// ISO currency code.
	Currency *string `json:"currency,omitempty" example:"EGP"`
	// Attribute dimensions (e.g. {"color":"Black","storage":"128GB"}).
	Attributes map[string]any `json:"attributes,omitempty"`
	// Marks this variant as the default storefront selection.
	IsDefault bool `json:"isDefault" example:"true"`
	// Optional public thumbnail image URL, as returned by POST /vendor/media.
	ThumbnailURL *string `json:"thumbnailUrl,omitempty"`
}

// VendorUpdateVariantRequest is the merchant variant edit payload.
type VendorUpdateVariantRequest struct {
	// Updated SKU.
	SKU *string `json:"sku,omitempty"`
	// Updated title.
	Title *string `json:"title,omitempty"`
	// Updated price in minor units.
	Price *int64 `json:"price,omitempty"`
	// Updated original price in minor units.
	CrossedOutPrice *int64 `json:"crossedOutPrice,omitempty"`
	// Updated ISO currency code.
	Currency *string `json:"currency,omitempty"`
	// Updated attribute dimensions.
	Attributes map[string]any `json:"attributes,omitempty"`
	// Updated default-selection flag.
	IsDefault *bool `json:"isDefault,omitempty"`
	// Updated public thumbnail image URL, as returned by POST /vendor/media.
	ThumbnailURL *string `json:"thumbnailUrl,omitempty"`
}

// VendorStockAdjustRequest is a merchant stock movement payload.
type VendorStockAdjustRequest struct {
	// Signed delta: positive restocks, negative reduces on-hand.
	Quantity int `json:"quantity" binding:"required" example:"50"`
	// Business reason for the movement.
	Reason string `json:"reason" binding:"required,oneof=restock return adjustment" example:"restock"`
}

// VendorStockResponse is the resulting stock state after an adjustment.
type VendorStockResponse struct {
	// Variant UUID.
	VariantID uuid.UUID `json:"variantId"`
	// Physical units on hand.
	OnHandQuantity int `json:"onHandQuantity" example:"60"`
	// Units reserved by open orders.
	ReservedQuantity int `json:"reservedQuantity" example:"10"`
	// Salable units (on hand minus reserved).
	AvailableQuantity int `json:"availableQuantity" example:"50"`
	// True when available units exceed zero.
	IsInStock bool `json:"isInStock" example:"true"`
}

// ListProductVariants godoc
// @Summary      List variants of my product
// @Description  Returns the variants of a product owned by the authenticated vendor.
// @Tags         Vendor
// @Produce      json
// @Param        id       path  string true  "Product UUID" format(uuid)
// @Param        page     query int    false "Page number (1-based)"
// @Param        pageSize query int    false "Page size (max 100)"
// @Success      200 {object} api.PaginatedResponse{data=[]VendorVariantResponse}
// @Failure      403 {object} api.ForbiddenErrorResponse "Vendor not active"
// @Failure      404 {object} api.NotFoundErrorResponse "Product or vendor not found"
// @Router       /vendor/products/{id}/variants [get]
// @Security     SessionAuth
func (h *Handler) ListProductVariants(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	productID, ok := bindVendorID(c)
	if !ok {
		return
	}

	var q pagination.ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		validation.ValidationError(c, err)
		return
	}
	q.Process(pagination.QueryOptions{DefaultPageSize: 20, MaxPageSize: 100})

	result, err := h.service.ListProductVariants(c.Request.Context(), userID, productID, &q)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.PaginatedResponse{
		Data: result.Items,
		Meta: result.Meta(),
	})
}

// CreateVariant godoc
// @Summary      Add a variant to my product
// @Description  Creates a purchasable SKU under a product owned by the authenticated vendor.
// @Tags         Vendor
// @Accept       json
// @Produce      json
// @Param        id      path string                     true "Product UUID" format(uuid)
// @Param        request body VendorCreateVariantRequest true "Variant payload"
// @Success      201 {object} api.DataResponse{data=VendorVariantResponse}
// @Failure      400 {object} api.BadReqResponse "Validation failed"
// @Failure      403 {object} api.ForbiddenErrorResponse "Vendor not active"
// @Failure      404 {object} api.NotFoundErrorResponse "Product or vendor not found"
// @Router       /vendor/products/{id}/variants [post]
// @Security     SessionAuth
func (h *Handler) CreateVariant(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	productID, ok := bindVendorID(c)
	if !ok {
		return
	}

	var req VendorCreateVariantRequest
	if !bindVendorJSON(c, &req) {
		return
	}

	created, err := h.service.CreateVariant(c.Request.Context(), userID, productID, derefString(req.ThumbnailURL), &variant.CreateVariantInput{
		SKU:               req.SKU,
		Title:             req.Title,
		Price:             req.Price,
		CrossedOutPrice:   req.CrossedOutPrice,
		Currency:          req.Currency,
		Attributes:        req.Attributes,
		IsDefault:         req.IsDefault,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, api.DataResponse{Data: toVendorVariantResponse(created)})
}

// UpdateVariant godoc
// @Summary      Update my variant
// @Description  Edits a variant whose parent product belongs to the authenticated vendor.
// @Tags         Vendor
// @Accept       json
// @Produce      json
// @Param        id      path string                     true "Variant UUID" format(uuid)
// @Param        request body VendorUpdateVariantRequest true "Variant edit payload"
// @Success      200 {object} api.MessageResponse
// @Failure      400 {object} api.BadReqResponse "Validation failed"
// @Failure      403 {object} api.ForbiddenErrorResponse "Vendor not active"
// @Failure      404 {object} api.NotFoundErrorResponse "Variant or vendor not found"
// @Router       /vendor/variants/{id} [patch]
// @Security     SessionAuth
func (h *Handler) UpdateVariant(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	variantID, ok := bindVendorID(c)
	if !ok {
		return
	}

	var req VendorUpdateVariantRequest
	if !bindVendorJSON(c, &req) {
		return
	}

	if err := h.service.UpdateVariant(c.Request.Context(), userID, variantID, derefString(req.ThumbnailURL), variant.UpdateVariantInput{
		SKU:               req.SKU,
		Title:             req.Title,
		Price:             req.Price,
		CrossedOutPrice:   req.CrossedOutPrice,
		Currency:          req.Currency,
		Attributes:        req.Attributes,
		IsDefault:         req.IsDefault,
	}); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "Variant updated"})
}

// DeleteVariant godoc
// @Summary      Delete my variant
// @Description  Soft-deletes a variant whose parent product belongs to the authenticated vendor.
// @Tags         Vendor
// @Produce      json
// @Param        id path string true "Variant UUID" format(uuid)
// @Success      200 {object} api.MessageResponse
// @Failure      403 {object} api.ForbiddenErrorResponse "Vendor not active"
// @Failure      404 {object} api.NotFoundErrorResponse "Variant or vendor not found"
// @Router       /vendor/variants/{id} [delete]
// @Security     SessionAuth
func (h *Handler) DeleteVariant(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	variantID, ok := bindVendorID(c)
	if !ok {
		return
	}

	if err := h.service.DeleteVariant(c.Request.Context(), userID, variantID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "Variant deleted"})
}

// AdjustStock godoc
// @Summary      Adjust my variant stock
// @Description  Records a merchant stock movement (restock, return, correction) for a variant of the authenticated vendor.
// @Tags         Vendor
// @Accept       json
// @Produce      json
// @Param        id      path string                  true "Variant UUID" format(uuid)
// @Param        request body VendorStockAdjustRequest true "Stock movement payload"
// @Success      200 {object} api.DataResponse{data=VendorStockResponse}
// @Failure      400 {object} api.BadReqResponse "Validation failed"
// @Failure      403 {object} api.ForbiddenErrorResponse "Vendor not active"
// @Failure      404 {object} api.NotFoundErrorResponse "Variant or vendor not found"
// @Router       /vendor/variants/{id}/stock [post]
// @Security     SessionAuth
func (h *Handler) AdjustStock(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	variantID, ok := bindVendorID(c)
	if !ok {
		return
	}

	var req VendorStockAdjustRequest
	if !bindVendorJSON(c, &req) {
		return
	}

	stock, err := h.service.AdjustStock(c.Request.Context(), userID, inventory.AdjustStockInput{
		VariantID: variantID,
		Quantity:  req.Quantity,
		Reason:    req.Reason,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: VendorStockResponse{
		VariantID:         stock.VariantID,
		OnHandQuantity:    stock.OnHandQuantity,
		ReservedQuantity:  stock.ReservedQuantity,
		AvailableQuantity: stock.AvailableQuantity,
		IsInStock:         stock.IsInStock,
	}})
}
