package vendors

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
	"github.com/nullopt-t/prim-backend/internal/shared/validation"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
)

// Handler exposes vendor portal and admin endpoints over HTTP.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	if service == nil {
		panic("vendor: service is required")
	}
	return &Handler{service: service}
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
	return VendorResponse{
		ID:                v.ID,
		UserID:            v.UserID,
		StoreName:         v.StoreName,
		Slug:              v.Slug,
		Description:       v.Description,
		LogoURL:           v.LogoURL,
		BannerURL:         v.BannerURL,
		CommissionRateBps: v.CommissionRateBps,
		Status:            v.Status,
		CreatedAt:         v.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         v.UpdatedAt.Format(time.RFC3339),
	}
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
// @Description  Creates a pending vendor application for the authenticated user. One profile per account.
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

// UpdateMyVendorRequest is the profile edit payload; nil fields are unchanged.
type UpdateMyVendorRequest struct {
	StoreName       string         `json:"storeName" binding:"omitempty,min=2,max=150" example:"Amr's Gadgets"`
	Description     string         `json:"description" binding:"max=2000" example:"Electronics and accessories"`
	LogoObjectID    string         `json:"logoObjectId" binding:"omitempty,uuid"`
	BannerObjectID  string         `json:"bannerObjectId" binding:"omitempty,uuid"`
	BankAccountInfo map[string]any `json:"bankAccountInfo"`
}

// UpdateMyVendor godoc
// @Summary      Update my vendor profile
// @Description  Edits the vendor profile. Only store/slug-independent fields; commission and status are admin-controlled.
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

// ---- Admin lifecycle ----------------------------------------------------

// AdminChangeVendorStatusRequest is the admin lifecycle payload.
type AdminChangeVendorStatusRequest struct {
	// New vendor lifecycle state.
	Status model.VendorStatus `json:"status" binding:"required,oneof=pending active suspended rejected" enums:"pending,active,suspended,rejected"`
}

// AdminChangeVendorStatus godoc
// @Summary      Change vendor status (admin)
// @Description  Applies a guarded lifecycle transition (pending->active/rejected, active<->suspended).
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

	vendor, err := h.service.AdminChangeVendorStatus(c.Request.Context(), vendorID, req.Status)
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
// @Description  Registers a payout request against settled earnings, validating the available balance under the vendor row lock.
// @Tags         Vendor
// @Accept       json
// @Produce      json
// @Param        request body RequestPayoutRequest true "Payout request"
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

	payout, err := h.service.RequestPayout(c.Request.Context(), userID, req.Amount)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, api.DataResponse{Data: toPayoutResponse(payout)})
}

// ListMyPayouts godoc
// @Summary      List my payouts
// @Description  Returns the caller's payout ledger including rejected records (full audit trail).
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
