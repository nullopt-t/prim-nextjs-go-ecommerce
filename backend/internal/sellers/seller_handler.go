package sellers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/catalog/coupon"
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

// Handler exposes seller portal and admin endpoints over HTTP.
type Handler struct {
	service *Service
	media   MediaUploader
}

// MediaUploader is the narrow slice of the object domain the seller portal
// needs: merchant image uploads. Storage object ids stay internal; the
// seller surface only ever returns public URLs.
type MediaUploader interface {
	Upload(c *gin.Context) (*object.MediaItem, error)
}

func NewHandler(service *Service, media MediaUploader) *Handler {
	if service == nil {
		panic("seller: service is required")
	}
	if media == nil {
		panic("seller: media uploader is required")
	}
	return &Handler{service: service, media: media}
}

// ---- Responses ----------------------------------------------------------

// SellerResponse is the merchant profile payload.
type SellerResponse struct {
	// Seller profile id.
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
	Status model.SellerStatus `json:"status" enums:"pending,active,suspended,rejected"`
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
	// Seller id.
	SellerID uuid.UUID `json:"sellerId" example:"50000000-0000-0000-0000-000000000001"`
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

func toSellerResponse(v *model.Seller) SellerResponse {
	res := SellerResponse{
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

func toPayoutResponse(p *model.SellerPayout) PayoutResponse {
	res := PayoutResponse{
		ID:                       p.ID,
		SellerID:                 p.SellerID,
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

// RegisterSellerRequest is the onboarding payload.
type RegisterSellerRequest struct {
	StoreName      string `json:"storeName" binding:"required,min=2,max=150" example:"Amr's Gadgets"`
	Slug           string `json:"slug" binding:"required,max=150" example:"amrs-gadgets"`
	Description    string `json:"description" binding:"max=2000" example:"Electronics and accessories"`
	LogoObjectID   string `json:"logoObjectId" binding:"omitempty,uuid" example:"10000000-0000-0000-0000-00000000000f"`
	BannerObjectID string `json:"bannerObjectId" binding:"omitempty,uuid" example:"10000000-0000-0000-0000-00000000000e"`
}

// RegisterSeller godoc
// @Summary      Register as seller
// @Description  Creates a pending seller application for any authenticated account. A rejected profile may resubmit, which updates the application and returns it to pending review.
// @Tags         Seller
// @Accept       json
// @Produce      json
// @Param        request body RegisterSellerRequest true "Seller registration payload"
// @Success      201 {object} SellerResponse
// @Failure      400 {object} api.ErrorResponse "Validation failed"
// @Failure      401 {object} api.ErrorResponse "Authentication required"
// @Failure      409 {object} api.ErrorResponse "Already registered or slug taken"
// @Router       /seller/register [post]
// @Security     SessionAuth
func (h *Handler) RegisterSeller(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	var req RegisterSellerRequest
	if !bindSellerJSON(c, &req) {
		return
	}

	seller, err := h.service.RegisterSeller(c.Request.Context(), userID, RegisterSellerInput{
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

	c.JSON(http.StatusCreated, api.DataResponse{Data: toSellerResponse(seller)})
}

// UploadMedia godoc
// @Summary      Upload a merchant product image
// @Description  Uploads a raster image to the catalog bucket and returns its public URL, which callers pass as thumbnailUrl when creating variants. Only registered sellers may upload.
// @Tags         Seller
// @Accept       multipart/form-data
// @Param        file formData string true "Image file (JPEG, PNG, GIF, WebP or AVIF, up to 10 MB)"
// @Success      201 {object} api.DataResponse{data=SellerMediaUploadResponse}
// @Failure      401 {object} api.ErrorResponse "Authentication required"
// @Failure      404 {object} api.ErrorResponse "Seller not found"
// @Router       /seller/media [post]
// @Security     SessionAuth
func (h *Handler) UploadMedia(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	if _, err := h.service.GetSellerByUser(c.Request.Context(), userID); err != nil {
		_ = c.Error(err)
		return
	}

	item, err := h.media.Upload(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, api.DataResponse{Data: SellerMediaUploadResponse{URL: item.URL}})
}

// SellerMediaUploadResponse is the merchant-facing upload result. It exposes
// only the public URL; storage object ids never leave the backend.
type SellerMediaUploadResponse struct {
	// Public URL of the uploaded image.
	URL string `json:"url" example:"https://cdn.example.com/catalog/8f14e45f-ea0a-4a7a-9c9a-123456789abc"`
}

// GetMySeller godoc
// @Summary      Get my seller profile
// @Description  Returns the seller profile belonging to the authenticated user.
// @Tags         Seller
// @Produce      json
// @Success      200 {object} SellerResponse
// @Failure      401 {object} api.ErrorResponse "Authentication required"
// @Failure      404 {object} api.ErrorResponse "Seller not found"
// @Router       /seller/profile [get]
// @Security     SessionAuth
func (h *Handler) GetMySeller(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	seller, err := h.service.GetSellerByUser(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: toSellerResponse(seller)})
}

// CheckSlugAvailability godoc
// @Summary      Check store slug availability
// @Description  Validates slug syntax and reserved words, then checks the authoritative uniqueness set. A merchant's own current slug is reported available for re-application.
// @Tags         Seller
// @Produce      json
// @Param        slug query string true "Candidate store slug"
// @Success      200 {object} api.DataResponse{data=SlugAvailability}
// @Failure      401 {object} api.UnauthorizedErrorResponse
// @Router       /seller/slug-availability [get]
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

// UpdateMySellerRequest is the profile edit payload; nil fields are unchanged.
type UpdateMySellerRequest struct {
	StoreName       string         `json:"storeName" binding:"omitempty,min=2,max=150" example:"Amr's Gadgets"`
	Slug            string         `json:"slug" binding:"omitempty,max=63" example:"amrs-gadgets"`
	Description     string         `json:"description" binding:"max=2000" example:"Electronics and accessories"`
	LogoObjectID    string         `json:"logoObjectId" binding:"omitempty,uuid"`
	BannerObjectID  string         `json:"bannerObjectId" binding:"omitempty,uuid"`
	BankAccountInfo map[string]any `json:"bankAccountInfo"`
}

// UpdateMySeller godoc
// @Summary      Update my seller profile
// @Description  Edits the merchant profile for pending, active, or rejected sellers. Rejected edits return the application to pending; suspended sellers are blocked. The slug is immutable once the seller is active. Commission and status remain admin-controlled.
// @Tags         Seller
// @Accept       json
// @Produce      json
// @Param        request body UpdateMySellerRequest true "Profile update payload"
// @Success      200 {object} SellerResponse
// @Failure      400 {object} api.ErrorResponse "Validation failed"
// @Failure      403 {object} api.ErrorResponse "Seller not authorized"
// @Failure      404 {object} api.ErrorResponse "Seller not found"
// @Router       /seller/profile [patch]
// @Security     SessionAuth
func (h *Handler) UpdateMySeller(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	var req UpdateMySellerRequest
	if !bindSellerJSON(c, &req) {
		return
	}

	in := UpdateSellerInput{
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

	seller, err := h.service.UpdateMyProfile(c.Request.Context(), userID, in)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: toSellerResponse(seller)})
}

// bindSellerJSON binds the request body and renders validation errors via the
// shared middleware.
func bindSellerJSON(c *gin.Context, req any) bool {
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

// AdminChangeSellerStatusRequest is the admin lifecycle payload.
type AdminChangeSellerStatusRequest struct {
	// New seller lifecycle state.
	Status model.SellerStatus `json:"status" binding:"required,oneof=pending active suspended rejected" enums:"pending,active,suspended,rejected"`
	// Merchant-visible reason; required for rejection and suspension.
	Reason string `json:"reason" binding:"omitempty,max=2000"`
}

// AdminChangeSellerStatus godoc
// @Summary      Change seller status (admin)
// @Description  Applies a guarded lifecycle transition (pending->active/rejected, active<->suspended) and records reviewer metadata. Rejection and suspension require a merchant-facing reason.
// @Tags         Admin Sellers
// @Accept       json
// @Produce      json
// @Param        id      path string                       true "Seller UUID" format(uuid)
// @Param        request body AdminChangeSellerStatusRequest true "New status"
// @Success      200 {object} SellerResponse
// @Failure      400 {object} api.BadReqResponse "Validation failed"
// @Failure      404 {object} api.NotFoundErrorResponse "Seller not found"
// @Failure      409 {object} api.ConflictErrorResponse "Transition not permitted"
// @Router       /admin/sellers/{id}/status [patch]
// @Security     SessionAuth
func (h *Handler) AdminChangeSellerStatus(c *gin.Context) {
	sellerID, ok := bindSellerID(c)
	if !ok {
		return
	}

	var req AdminChangeSellerStatusRequest
	if !bindSellerJSON(c, &req) {
		return
	}

	reviewedBy, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	seller, err := h.service.AdminChangeSellerStatus(c.Request.Context(), sellerID, req.Status, req.Reason, reviewedBy)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: toSellerResponse(seller)})
}

// AdminSetCommissionRequest is the admin commission payload.
type AdminSetCommissionRequest struct {
	// Commission as basis points (0..10000).
	RateBps int `json:"rateBps" binding:"required,min=0,max=10000" example:"1000"`
}

// AdminSetCommission godoc
// @Summary      Set seller commission rate (admin)
// @Description  Overrides the per-seller commission in basis points. Future orders use the new rate; already-placed orders keep their frozen earning.
// @Tags         Admin Sellers
// @Accept       json
// @Produce      json
// @Param        id      path string true "Seller UUID" format(uuid)
// @Param        request body AdminSetCommissionRequest true "Rate in bps"
// @Success      200 {object} SellerResponse
// @Failure      400 {object} api.BadReqResponse "Validation failed"
// @Failure      404 {object} api.NotFoundErrorResponse "Seller not found"
// @Router       /admin/sellers/{id}/commission [patch]
// @Security     SessionAuth
func (h *Handler) AdminSetCommission(c *gin.Context) {
	sellerID, ok := bindSellerID(c)
	if !ok {
		return
	}

	var req AdminSetCommissionRequest
	if !bindSellerJSON(c, &req) {
		return
	}

	seller, err := h.service.AdminSetCommissionRate(c.Request.Context(), sellerID, req.RateBps)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: toSellerResponse(seller)})
}

// AdminListSellers godoc
// @Summary      List sellers (admin)
// @Description  Returns the seller directory, newest first, optionally filtered by lifecycle status.
// @Tags         Admin Sellers
// @Produce      json
// @Param        status    query string false "Lifecycle filter" Enums(pending, active, suspended, rejected)
// @Param        page      query int    false "Page number (1-based)"
// @Param        pageSize  query int    false "Page size (max 100)"
// @Success      200 {object} api.PaginatedResponse{data=[]SellerResponse}
// @Failure      400 {object} api.BadReqResponse "Invalid status filter"
// @Router       /admin/sellers [get]
// @Security     SessionAuth
func (h *Handler) AdminListSellers(c *gin.Context) {
	var status *model.SellerStatus
	if raw := c.Query("status"); raw != "" {
		s := model.SellerStatus(raw)
		switch s {
		case model.SellerStatusPending, model.SellerStatusActive, model.SellerStatusSuspended, model.SellerStatusRejected:
		default:
			_ = c.Error(apierr.BadRequestError("Invalid seller status filter"))
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

	result, err := h.service.AdminListSellers(c.Request.Context(), status, &q)
	if err != nil {
		_ = c.Error(err)
		return
	}

	responses := make([]SellerResponse, 0, len(result.Items))
	for _, v := range result.Items {
		responses = append(responses, toSellerResponse(v))
	}

	c.JSON(http.StatusOK, api.PaginatedResponse{
		Data: responses,
		Meta: result.Meta(),
	})
}

// AdminGetSeller godoc
// @Summary      Get seller by ID (admin)
// @Description  Returns any seller profile for administration.
// @Tags         Admin Sellers
// @Produce      json
// @Param        id path string true "Seller UUID" format(uuid)
// @Success      200 {object} SellerResponse
// @Failure      404 {object} api.NotFoundErrorResponse "Seller not found"
// @Router       /admin/sellers/{id} [get]
// @Security     SessionAuth
func (h *Handler) AdminGetSeller(c *gin.Context) {
	sellerID, ok := bindSellerID(c)
	if !ok {
		return
	}

	seller, err := h.service.AdminGetSellerByID(c.Request.Context(), sellerID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: toSellerResponse(seller)})
}

// ---- Payouts ------------------------------------------------------------

// RequestPayoutRequest is the seller payout request payload.
type RequestPayoutRequest struct {
	// Amount in minor currency units; must not exceed the available balance.
	Amount int64 `json:"amount" binding:"required,min=1" example:"125000"`
}

// RequestPayout godoc
// @Summary      Request a payout
// @Description  Registers a payout request against settled earnings, validating the available balance under the seller row lock. Active sellers only. Reusing an Idempotency-Key with the same amount replays the original request; a different amount returns 409.
// @Tags         Seller
// @Accept       json
// @Produce      json
// @Param        request body RequestPayoutRequest true "Payout request"
// @Param        Idempotency-Key header string false "Client retry key (max 100 chars)"
// @Success      201 {object} PayoutResponse
// @Failure      400 {object} api.BadReqResponse "Validation failed or insufficient balance"
// @Failure      403 {object} api.ForbiddenErrorResponse "Seller not active"
// @Failure      404 {object} api.NotFoundErrorResponse "Seller not found"
// @Router       /seller/payouts [post]
// @Security     SessionAuth
func (h *Handler) RequestPayout(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	var req RequestPayoutRequest
	if !bindSellerJSON(c, &req) {
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
// @Description  Returns the caller's payout ledger including rejected records (full audit trail). Available to active and suspended sellers.
// @Tags         Seller
// @Produce      json
// @Success      200 {object} api.DataResponse{data=[]PayoutResponse}
// @Failure      401 {object} api.UnauthorizedErrorResponse
// @Failure      404 {object} api.NotFoundErrorResponse "Seller not found"
// @Router       /seller/payouts [get]
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
// @Description  Returns every payout across all sellers for administration.
// @Tags         Admin Sellers
// @Produce      json
// @Success      200 {object} api.DataResponse{data=[]PayoutResponse}
// @Router       /admin/sellers/payouts [get]
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
// @Tags         Admin Sellers
// @Accept       json
// @Produce      json
// @Param        id path string true "Payout UUID" format(uuid)
// @Param        request body AdminTransitionPayoutRequest true "Transition payload"
// @Success      200 {object} PayoutResponse
// @Failure      404 {object} api.NotFoundErrorResponse "Payout not found"
// @Failure      409 {object} api.ConflictErrorResponse "Transition not permitted"
// @Router       /admin/sellers/payouts/{id}/transition [patch]
// @Security     SessionAuth
func (h *Handler) AdminTransitionPayout(c *gin.Context) {
	payoutID, ok := bindSellerID(c)
	if !ok {
		return
	}

	var req AdminTransitionPayoutRequest
	if !bindSellerJSON(c, &req) {
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

// bindSellerID parses the route ":id" path parameter.
func bindSellerID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.InvalidUUIDError().WithFields(
			api.FieldError{Field: "id", Constraint: "uuid", Message: "invalid seller id"},
		))
		return uuid.Nil, false
	}
	return id, true
}

func toPayoutList(payouts []*model.SellerPayout) []PayoutResponse {
	list := make([]PayoutResponse, 0, len(payouts))
	for _, p := range payouts {
		list = append(list, toPayoutResponse(p))
	}
	return list
}

// GetDashboard godoc
// @Summary      Seller dashboard
// @Description  Returns the merchant's money position (settled, pending, outstanding) and catalog health counts. Active sellers only; other lifecycle states receive 403.
// @Tags         Seller
// @Produce      json
// @Success      200 {object} api.DataResponse{data=Dashboard}
// @Failure      401 {object} api.UnauthorizedErrorResponse
// @Failure      404 {object} api.NotFoundErrorResponse "Seller not found"
// @Router       /seller/dashboard [get]
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

// ---- Seller product catalog ---------------------------------------------

// SellerCreateProductRequest is the merchant product creation payload.
type SellerCreateProductRequest struct {
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
	// Product classification: simple, variable, or downloadable.
	ProductType string `json:"productType" binding:"omitempty,oneof=simple variable downloadable" example:"simple"`
	// Ordered variation axis keys for variable products.
	VariationAxes []string `json:"variationAxes,omitempty" example:"color,storage"`
	// Public file URL for downloadable products.
	DownloadURL *string `json:"downloadUrl,omitempty"`
	// Delivery note shown to buyers of a downloadable product.
	DeliveryInstructions *string `json:"deliveryInstructions,omitempty"`
	// Optional public product image URL, as returned by POST /seller/media.
	// Required before publishing; drafts may omit it.
	ThumbnailURL string `json:"thumbnailUrl,omitempty" example:"https://cdn.example.com/catalog/8f14e45f-ea0a-4a7a-9c9a-123456789abc"`
}

// SellerUpdateProductRequest is the merchant product edit payload; omitted
// fields are unchanged.
type SellerUpdateProductRequest struct {
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
	// Product classification: simple, variable, or downloadable.
	ProductType *string `json:"productType,omitempty" binding:"omitempty,oneof=simple variable downloadable"`
	// Variation axes; omit to leave unchanged, [] to make the product simple.
	VariationAxes *[]string `json:"variationAxes,omitempty"`
}

// SellerProductResponse is the merchant-facing product payload.
type SellerProductResponse struct {
	// Product UUID.
	ID uuid.UUID `json:"id" example:"70000000-0000-0000-0000-000000000001"`
	// Owning seller UUID.
	SellerID uuid.UUID `json:"sellerId" example:"50000000-0000-0000-0000-000000000001"`
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

func toSellerProductResponse(p *model.Product) SellerProductResponse {
	res := SellerProductResponse{
		ID:          p.ID,
		Slug:        p.Slug,
		Title:       p.Title,
		Status:      p.Status,
		ProductType: p.ProductType,
		CreatedAt:   p.CreatedAt.Format(timeLayout),
		UpdatedAt:   p.UpdatedAt.Format(timeLayout),
	}
	if p.SellerID != nil {
		res.SellerID = *p.SellerID
	}
	return res
}

// ListMyProducts godoc
// @Summary      List my products
// @Description  Returns the authenticated seller's catalog, drafts included. Ownership is enforced server-side and the seller must be active.
// @Tags         Seller
// @Produce      json
// @Param        page      query int    false "Page number (1-based)"
// @Param        pageSize  query int    false "Page size (max 100)"
// @Param        status    query string false "Publication status filter" Enums(draft, published, archived)
// @Param        category  query string false "Category name(s) or UUID(s), comma-separated"
// @Param        brand     query string false "Brand name(s) or UUID(s), comma-separated"
// @Success      200 {object} api.PaginatedResponse{data=[]SellerProductSummary}
// @Failure      401 {object} api.UnauthorizedErrorResponse
// @Failure      404 {object} api.NotFoundErrorResponse "Seller not found"
// @Router       /seller/products [get]
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
// @Description  Creates a draft product owned by the authenticated seller. Requires an active seller profile.
// @Tags         Seller
// @Accept       json
// @Produce      json
// @Param        request body SellerCreateProductRequest true "Product draft payload"
// @Success      201 {object} api.DataResponse{data=SellerProductResponse}
// @Failure      400 {object} api.BadReqResponse "Validation failed"
// @Failure      403 {object} api.ForbiddenErrorResponse "Seller not active"
// @Failure      404 {object} api.NotFoundErrorResponse "Seller not found"
// @Router       /seller/products [post]
// @Security     SessionAuth
func (h *Handler) CreateProduct(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	var req SellerCreateProductRequest
	if !bindSellerJSON(c, &req) {
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

	created, err := h.service.CreateProduct(c.Request.Context(), userID, req.ThumbnailURL, product.CreateProductInput{
		BrandID:              brandID,
		CategoryID:           categoryID,
		Slug:                 &req.Slug,
		Title:                req.Title,
		Description:          req.Description,
		Highlights:           req.Highlights,
		ProductType:          req.ProductType,
		VariationAxes:        req.VariationAxes,
		DownloadURL:          req.DownloadURL,
		DeliveryInstructions: req.DeliveryInstructions,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, api.DataResponse{Data: toSellerProductResponse(created)})
}

// UpdateProduct godoc
// @Summary      Update a product
// @Description  Edits a product owned by the authenticated seller. Foreign or platform products return 404.
// @Tags         Seller
// @Accept       json
// @Produce      json
// @Param        id      path string                     true "Product UUID" format(uuid)
// @Param        request body SellerUpdateProductRequest true "Product edit payload"
// @Success      200 {object} api.MessageResponse
// @Failure      400 {object} api.BadReqResponse "Validation failed"
// @Failure      403 {object} api.ForbiddenErrorResponse "Seller not active"
// @Failure      404 {object} api.NotFoundErrorResponse "Product or seller not found"
// @Router       /seller/products/{id} [patch]
// @Security     SessionAuth
func (h *Handler) UpdateProduct(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	productID, ok := bindSellerID(c)
	if !ok {
		return
	}

	var req SellerUpdateProductRequest
	if !bindSellerJSON(c, &req) {
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
// @Description  Soft-deletes a product owned by the authenticated seller.
// @Tags         Seller
// @Produce      json
// @Param        id path string true "Product UUID" format(uuid)
// @Success      200 {object} api.MessageResponse
// @Failure      403 {object} api.ForbiddenErrorResponse "Seller not active"
// @Failure      404 {object} api.NotFoundErrorResponse "Product or seller not found"
// @Router       /seller/products/{id} [delete]
// @Security     SessionAuth
func (h *Handler) DeleteProduct(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	productID, ok := bindSellerID(c)
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
// @Description  Publishes a product owned by the authenticated seller. Product-domain preconditions (e.g. at least one variant) still apply.
// @Tags         Seller
// @Produce      json
// @Param        id path string true "Product UUID" format(uuid)
// @Success      200 {object} api.MessageResponse
// @Failure      400 {object} api.BadReqResponse "Publishing preconditions not met"
// @Failure      403 {object} api.ForbiddenErrorResponse "Seller not active"
// @Failure      404 {object} api.NotFoundErrorResponse "Product or seller not found"
// @Router       /seller/products/{id}/publish [post]
// @Security     SessionAuth
func (h *Handler) PublishProduct(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	productID, ok := bindSellerID(c)
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
// @Description  Archives a product owned by the authenticated seller, removing it from the storefront.
// @Tags         Seller
// @Produce      json
// @Param        id path string true "Product UUID" format(uuid)
// @Success      200 {object} api.MessageResponse
// @Failure      403 {object} api.ForbiddenErrorResponse "Seller not active"
// @Failure      404 {object} api.NotFoundErrorResponse "Product or seller not found"
// @Router       /seller/products/{id}/archive [post]
// @Security     SessionAuth
func (h *Handler) ArchiveProduct(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	productID, ok := bindSellerID(c)
	if !ok {
		return
	}

	if err := h.service.ArchiveProduct(c.Request.Context(), userID, productID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "Product archived"})
}

// ---- Seller promotions --------------------------------------------------

// SellerCreatePromotionRequest is the merchant coupon creation payload.
type SellerCreatePromotionRequest struct {
	// Promo code; stored uppercased.
	Code string `json:"code" binding:"required" example:"SUMMER20"`
	// Discount kind.
	DiscountType string `json:"discountType" binding:"required,oneof=percentage fixed_amount" example:"percentage"`
	// Percentage (1-100) or fixed amount in minor units.
	DiscountValue int64 `json:"discountValue" binding:"required" example:"20"`
	// ISO currency code; defaults to EGP.
	Currency *string `json:"currency,omitempty" example:"EGP"`
	// Minimum eligible subtotal in minor units before the coupon applies.
	MinCartAmount *int64 `json:"minCartAmount,omitempty"`
	// Optional total redemption cap.
	UsageLimit *int `json:"usageLimit,omitempty"`
	// Optional per-customer redemption cap.
	PerUserLimit *int `json:"perUserLimit,omitempty"`
	// Optional activation window.
	StartsAt  *time.Time `json:"startsAt,omitempty"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

// SellerPromotionStatusRequest is the merchant coupon lifecycle payload.
type SellerPromotionStatusRequest struct {
	// New lifecycle state.
	Status string `json:"status" binding:"required,oneof=active disabled archived" example:"disabled"`
}

// ListPromotions godoc
// @Summary      List my promotions
// @Description  Lists coupons owned by the authenticated seller.
// @Tags         Seller
// @Produce      json
// @Param        page     query int    false "Page number (1-based)"
// @Param        pageSize query int    false "Page size (max 100)"
// @Success      200 {object} api.PaginatedResponse{data=[]model.Coupon}
// @Failure      403 {object} api.ForbiddenErrorResponse "Seller not active"
// @Router       /seller/promotions [get]
// @Security     SessionAuth
func (h *Handler) ListPromotions(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	var q pagination.ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		validation.ValidationError(c, err)
		return
	}
	q.Process(pagination.QueryOptions{DefaultPageSize: 20, MaxPageSize: 100})

	cursor, err := q.Keyset()
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid pagination cursor").Wrap(err))
		return
	}

	result, page, err := h.service.ListPromotions(c.Request.Context(), userID, q.PageSize, cursor)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.SuccessResponse{Data: result, Meta: page})
}

// CreatePromotion godoc
// @Summary      Create a promotion
// @Description  Creates a coupon scoped to the authenticated seller's products. The discount applies only to that seller's eligible subtotal at checkout.
// @Tags         Seller
// @Accept       json
// @Produce      json
// @Param        request body SellerCreatePromotionRequest true "Promotion payload"
// @Success      201 {object} api.DataResponse{data=model.Coupon}
// @Failure      400 {object} api.BadReqResponse "Validation failed"
// @Failure      403 {object} api.ForbiddenErrorResponse "Seller not active"
// @Router       /seller/promotions [post]
// @Security     SessionAuth
func (h *Handler) CreatePromotion(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	var req SellerCreatePromotionRequest
	if !bindSellerJSON(c, &req) {
		return
	}

	created, err := h.service.CreatePromotion(c.Request.Context(), userID, coupon.CreateCouponInput{
		Code:          req.Code,
		DiscountType:  model.CouponDiscountType(req.DiscountType),
		DiscountValue: req.DiscountValue,
		Currency:      derefString(req.Currency),
		MinCartAmount: req.MinCartAmount,
		UsageLimit:    req.UsageLimit,
		PerUserLimit:  req.PerUserLimit,
		StartsAt:      req.StartsAt,
		ExpiresAt:     req.ExpiresAt,
		Status:        model.CouponStatusActive,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, api.DataResponse{Data: created})
}

// SetPromotionStatus godoc
// @Summary      Enable or disable a promotion
// @Description  Changes the lifecycle state of a coupon owned by the authenticated seller.
// @Tags         Seller
// @Accept       json
// @Produce      json
// @Param        id      path string                      true "Coupon UUID" format(uuid)
// @Param        request body SellerPromotionStatusRequest true "Status payload"
// @Success      200 {object} api.MessageResponse
// @Failure      403 {object} api.ForbiddenErrorResponse "Seller not active"
// @Failure      404 {object} api.NotFoundErrorResponse "Promotion not found"
// @Router       /seller/promotions/{id}/status [patch]
// @Security     SessionAuth
func (h *Handler) SetPromotionStatus(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	couponID, ok := bindSellerID(c)
	if !ok {
		return
	}

	var req SellerPromotionStatusRequest
	if !bindSellerJSON(c, &req) {
		return
	}

	if err := h.service.SetPromotionStatus(c.Request.Context(), userID, couponID, model.CouponStatus(req.Status)); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "Promotion updated"})
}

// DeletePromotion godoc
// @Summary      Delete a promotion
// @Description  Permanently removes a coupon owned by the authenticated seller.
// @Tags         Seller
// @Produce      json
// @Param        id path string true "Coupon UUID" format(uuid)
// @Success      200 {object} api.MessageResponse
// @Failure      403 {object} api.ForbiddenErrorResponse "Seller not active"
// @Failure      404 {object} api.NotFoundErrorResponse "Promotion not found"
// @Router       /seller/promotions/{id} [delete]
// @Security     SessionAuth
func (h *Handler) DeletePromotion(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	couponID, ok := bindSellerID(c)
	if !ok {
		return
	}

	if err := h.service.DeletePromotion(c.Request.Context(), userID, couponID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "Promotion deleted"})
}

// ---- Seller variants & stock --------------------------------------------

// SellerCreateVariantRequest is the merchant variant creation payload.
type SellerCreateVariantRequest struct {
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
	// Optional public thumbnail image URL, as returned by POST /seller/media.
	ThumbnailURL *string `json:"thumbnailUrl,omitempty"`
}

// SellerUpdateVariantRequest is the merchant variant edit payload.
type SellerUpdateVariantRequest struct {
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
	// Updated public thumbnail image URL, as returned by POST /seller/media.
	ThumbnailURL *string `json:"thumbnailUrl,omitempty"`
}

// SellerStockAdjustRequest is a merchant stock movement payload.
type SellerStockAdjustRequest struct {
	// Signed delta: positive restocks, negative reduces on-hand.
	Quantity int `json:"quantity" binding:"required" example:"50"`
	// Business reason for the movement.
	Reason string `json:"reason" binding:"required,oneof=restock return adjustment" example:"restock"`
}

// SellerStockResponse is the resulting stock state after an adjustment.
type SellerStockResponse struct {
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
// @Description  Returns the variants of a product owned by the authenticated seller.
// @Tags         Seller
// @Produce      json
// @Param        id       path  string true  "Product UUID" format(uuid)
// @Param        page     query int    false "Page number (1-based)"
// @Param        pageSize query int    false "Page size (max 100)"
// @Success      200 {object} api.PaginatedResponse{data=[]SellerVariantResponse}
// @Failure      403 {object} api.ForbiddenErrorResponse "Seller not active"
// @Failure      404 {object} api.NotFoundErrorResponse "Product or seller not found"
// @Router       /seller/products/{id}/variants [get]
// @Security     SessionAuth
func (h *Handler) ListProductVariants(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	productID, ok := bindSellerID(c)
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
// @Description  Creates a purchasable SKU under a product owned by the authenticated seller.
// @Tags         Seller
// @Accept       json
// @Produce      json
// @Param        id      path string                     true "Product UUID" format(uuid)
// @Param        request body SellerCreateVariantRequest true "Variant payload"
// @Success      201 {object} api.DataResponse{data=SellerVariantResponse}
// @Failure      400 {object} api.BadReqResponse "Validation failed"
// @Failure      403 {object} api.ForbiddenErrorResponse "Seller not active"
// @Failure      404 {object} api.NotFoundErrorResponse "Product or seller not found"
// @Router       /seller/products/{id}/variants [post]
// @Security     SessionAuth
func (h *Handler) CreateVariant(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	productID, ok := bindSellerID(c)
	if !ok {
		return
	}

	var req SellerCreateVariantRequest
	if !bindSellerJSON(c, &req) {
		return
	}

	created, err := h.service.CreateVariant(c.Request.Context(), userID, productID, derefString(req.ThumbnailURL), &variant.CreateVariantInput{
		SKU:             req.SKU,
		Title:           req.Title,
		Price:           req.Price,
		CrossedOutPrice: req.CrossedOutPrice,
		Currency:        req.Currency,
		Attributes:      req.Attributes,
		IsDefault:       req.IsDefault,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, api.DataResponse{Data: toSellerVariantResponse(created)})
}

// UpdateVariant godoc
// @Summary      Update my variant
// @Description  Edits a variant whose parent product belongs to the authenticated seller.
// @Tags         Seller
// @Accept       json
// @Produce      json
// @Param        id      path string                     true "Variant UUID" format(uuid)
// @Param        request body SellerUpdateVariantRequest true "Variant edit payload"
// @Success      200 {object} api.MessageResponse
// @Failure      400 {object} api.BadReqResponse "Validation failed"
// @Failure      403 {object} api.ForbiddenErrorResponse "Seller not active"
// @Failure      404 {object} api.NotFoundErrorResponse "Variant or seller not found"
// @Router       /seller/variants/{id} [patch]
// @Security     SessionAuth
func (h *Handler) UpdateVariant(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	variantID, ok := bindSellerID(c)
	if !ok {
		return
	}

	var req SellerUpdateVariantRequest
	if !bindSellerJSON(c, &req) {
		return
	}

	if err := h.service.UpdateVariant(c.Request.Context(), userID, variantID, req.ThumbnailURL, variant.UpdateVariantInput{
		SKU:             req.SKU,
		Title:           req.Title,
		Price:           req.Price,
		CrossedOutPrice: req.CrossedOutPrice,
		Currency:        req.Currency,
		Attributes:      req.Attributes,
		IsDefault:       req.IsDefault,
	}); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "Variant updated"})
}

// DeleteVariant godoc
// @Summary      Delete my variant
// @Description  Soft-deletes a variant whose parent product belongs to the authenticated seller.
// @Tags         Seller
// @Produce      json
// @Param        id path string true "Variant UUID" format(uuid)
// @Success      200 {object} api.MessageResponse
// @Failure      403 {object} api.ForbiddenErrorResponse "Seller not active"
// @Failure      404 {object} api.NotFoundErrorResponse "Variant or seller not found"
// @Router       /seller/variants/{id} [delete]
// @Security     SessionAuth
func (h *Handler) DeleteVariant(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	variantID, ok := bindSellerID(c)
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
// @Description  Records a merchant stock movement (restock, return, correction) for a variant of the authenticated seller.
// @Tags         Seller
// @Accept       json
// @Produce      json
// @Param        id      path string                  true "Variant UUID" format(uuid)
// @Param        request body SellerStockAdjustRequest true "Stock movement payload"
// @Success      200 {object} api.DataResponse{data=SellerStockResponse}
// @Failure      400 {object} api.BadReqResponse "Validation failed"
// @Failure      403 {object} api.ForbiddenErrorResponse "Seller not active"
// @Failure      404 {object} api.NotFoundErrorResponse "Variant or seller not found"
// @Router       /seller/variants/{id}/stock [post]
// @Security     SessionAuth
func (h *Handler) AdjustStock(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	variantID, ok := bindSellerID(c)
	if !ok {
		return
	}

	var req SellerStockAdjustRequest
	if !bindSellerJSON(c, &req) {
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

	c.JSON(http.StatusOK, api.DataResponse{Data: SellerStockResponse{
		VariantID:         stock.VariantID,
		OnHandQuantity:    stock.OnHandQuantity,
		ReservedQuantity:  stock.ReservedQuantity,
		AvailableQuantity: stock.AvailableQuantity,
		IsInStock:         stock.IsInStock,
	}})
}
