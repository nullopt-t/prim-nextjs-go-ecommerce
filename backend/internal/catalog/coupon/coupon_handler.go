package coupon

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/validation"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
)

type CouponHandler struct {
	couponService *CouponService
}

func NewHandler(couponService *CouponService) *CouponHandler {
	return &CouponHandler{
		couponService: couponService,
	}
}

type CreateCouponRequest struct {
	Code          string  `json:"code" binding:"required" example:"SUMMER20"`
	DiscountType  string  `json:"discountType" binding:"required" example:"percentage"`
	DiscountValue int64   `json:"discountValue" binding:"required" example:"20"`
	Currency      *string `json:"currency,omitempty" example:"EGP"`
	MinCartAmount *int64  `json:"minCartAmount,omitempty" example:"50000"`
	UsageLimit    *int    `json:"usageLimit,omitempty" example:"1000"`
	PerUserLimit  *int    `json:"perUserLimit,omitempty" example:"1"`
	StartsAt      *string `json:"startsAt,omitempty" example:"2026-06-01T00:00:00Z"`
	ExpiresAt     *string `json:"expiresAt,omitempty" example:"2026-08-31T23:59:59Z"`
	Status        string  `json:"status,omitempty" example:"active"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required" example:"disabled"`
}

type CouponResponse struct {
	ID            string  `json:"id" example:"60000000-0000-0000-0000-000000000001"`
	Code          string  `json:"code" example:"SUMMER20"`
	DiscountType  string  `json:"discountType" example:"percentage"`
	DiscountValue int64   `json:"discountValue" example:"20"`
	Currency      string  `json:"currency" example:"EGP"`
	MinCartAmount *int64  `json:"minCartAmount,omitempty" example:"50000"`
	UsageLimit    *int    `json:"usageLimit,omitempty" example:"1000"`
	PerUserLimit  *int    `json:"perUserLimit,omitempty" example:"1"`
	StartsAt      *string `json:"startsAt,omitempty" example:"2026-06-01T00:00:00Z"`
	ExpiresAt     *string `json:"expiresAt,omitempty" example:"2026-08-31T23:59:59Z"`
	Status        string  `json:"status" example:"active"`
	UsageCount    int     `json:"usageCount" example:"42"`
	CreatedAt     string  `json:"createdAt" example:"2026-06-01T00:00:00Z"`
	UpdatedAt     string  `json:"updatedAt" example:"2026-06-01T00:00:00Z"`
}

func toCouponResponse(c *model.Coupon) CouponResponse {
	resp := CouponResponse{
		ID:            c.ID.String(),
		Code:          c.Code,
		DiscountType:  string(c.DiscountType),
		DiscountValue: c.DiscountValue,
		Currency:      c.Currency,
		MinCartAmount: c.MinCartAmount,
		UsageLimit:    c.UsageLimit,
		PerUserLimit:  c.PerUserLimit,
		Status:        string(c.Status),
		UsageCount:    c.UsageCount,
		CreatedAt:     c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     c.UpdatedAt.Format(time.RFC3339),
	}
	if c.StartsAt != nil {
		s := c.StartsAt.Format(time.RFC3339)
		resp.StartsAt = &s
	}
	if c.ExpiresAt != nil {
		e := c.ExpiresAt.Format(time.RFC3339)
		resp.ExpiresAt = &e
	}
	return resp
}

// ListCoupons godoc
//
//	@Summary		List promotions and coupons
//	@Description	Returns all promotional coupons (Admin only)
//	@Tags			Coupons
//	@Produce		json
//	@Success		200	{object}	api.DataResponse
//	@Router			/admin/promotions [get]
func (h *CouponHandler) ListCoupons(c *gin.Context) {
	var q pagination.ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		validation.ValidationError(c, err)
		return
	}
	q.Process(pagination.QueryOptions{DefaultPageSize: 20, MaxPageSize: 100})

	var filter ListCouponFilter
	if st := c.Query("status"); st != "" {
		s := model.CouponStatus(st)
		filter.Status = &s
	}

	cursor, err := q.Keyset()
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid pagination cursor").Wrap(err))
		return
	}

	coupons, page, err := h.couponService.ListCoupons(c.Request.Context(), filter, q.PageSize, cursor)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := make([]CouponResponse, len(coupons))
	for i, cp := range coupons {
		res[i] = toCouponResponse(cp)
	}

	c.JSON(http.StatusOK, api.SuccessResponse{Data: res, Meta: page})
}

// GetCouponByID godoc
//
//	@Summary		Get coupon details
//	@Description	Returns coupon details by UUID (Admin only)
//	@Tags			Coupons
//	@Produce		json
//	@Param			id	path		string	true	"Coupon UUID"
//	@Success		200	{object}	api.DataResponse
//	@Router			/admin/promotions/{id} [get]
func (h *CouponHandler) GetCouponByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid coupon ID"))
		return
	}

	cp, err := h.couponService.GetByID(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: toCouponResponse(cp)})
}

// CreateCoupon godoc
//
//	@Summary		Create promotional coupon
//	@Description	Creates a new coupon code with usage rules (Admin only)
//	@Tags			Coupons
//	@Accept			json
//	@Produce		json
//	@Param			request	body		CreateCouponRequest	true	"Create coupon request"
//	@Success		201		{object}	api.DataResponse
//	@Router			/admin/promotions [post]
func (h *CouponHandler) CreateCoupon(c *gin.Context) {
	var req CreateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid request body").Wrap(err))
		return
	}

	in := CreateCouponInput{
		Code:          req.Code,
		DiscountType:  model.CouponDiscountType(req.DiscountType),
		DiscountValue: req.DiscountValue,
		MinCartAmount: req.MinCartAmount,
		UsageLimit:    req.UsageLimit,
		PerUserLimit:  req.PerUserLimit,
		Status:        model.CouponStatus(req.Status),
	}
	if req.Currency != nil {
		in.Currency = *req.Currency
	}
	if req.StartsAt != nil && *req.StartsAt != "" {
		if t, err := time.Parse(time.RFC3339, *req.StartsAt); err == nil {
			in.StartsAt = &t
		}
	}
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		if t, err := time.Parse(time.RFC3339, *req.ExpiresAt); err == nil {
			in.ExpiresAt = &t
		}
	}

	cp, err := h.couponService.CreateCoupon(c.Request.Context(), in)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, api.DataResponse{Data: toCouponResponse(cp)})
}

// UpdateStatus godoc
//
//	@Summary		Update coupon status
//	@Description	Activates, disables, or archives a coupon (Admin only)
//	@Tags			Coupons
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string				true	"Coupon UUID"
//	@Param			request	body		UpdateStatusRequest	true	"Update status request"
//	@Success		200		{object}	api.MessageResponse
//	@Router			/admin/promotions/{id}/status [patch]
func (h *CouponHandler) UpdateStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid coupon ID"))
		return
	}

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid request body").Wrap(err))
		return
	}

	if err := h.couponService.UpdateStatus(c.Request.Context(), id, model.CouponStatus(req.Status)); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "Promotion status updated"})
}

// DeleteCoupon godoc
//
//	@Summary		Delete coupon
//	@Description	Deletes a promotion coupon by ID (Admin only)
//	@Tags			Coupons
//	@Produce		json
//	@Param			id	path		string	true	"Coupon UUID"
//	@Success		200	{object}	api.MessageResponse
//	@Router			/admin/promotions/{id} [delete]
func (h *CouponHandler) DeleteCoupon(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid coupon ID"))
		return
	}

	if err := h.couponService.Delete(c.Request.Context(), id); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "Promotion deleted"})
}

type CouponRedemptionResponse struct {
	ID             string  `json:"id" example:"c1000000-0000-0000-0000-000000000001"`
	UserID         *string `json:"userId,omitempty" example:"10000000-0000-0000-0000-000000000003"`
	UserName       *string `json:"userName,omitempty" example:"Jane Doe"`
	UserEmail      *string `json:"userEmail,omitempty" example:"jane@example.com"`
	OrderID        string  `json:"orderId" example:"90000000-0000-0000-0000-000000000002"`
	DiscountAmount int64   `json:"discountAmount" example:"13000"`
	CreatedAt      string  `json:"createdAt" example:"2026-06-01T00:00:00Z"`
}

func toRedemptionResponse(e *RedemptionEntry) CouponRedemptionResponse {
	resp := CouponRedemptionResponse{
		ID:             e.ID.String(),
		OrderID:        e.OrderID.String(),
		DiscountAmount: e.DiscountAmount,
		CreatedAt:      e.CreatedAt.Format(time.RFC3339),
		UserName:       e.UserName,
		UserEmail:      e.UserEmail,
	}
	if e.UserID != nil {
		v := e.UserID.String()
		resp.UserID = &v
	}
	return resp
}

// ListRedemptions godoc
//
//	@Summary		List coupon redemptions (usage log)
//	@Description	Returns every user and order that redeemed a coupon code (Admin only)
//	@Tags			Coupons
//	@Produce		json
//	@Param			id	path		string	true	"Coupon UUID"
//	@Success		200	{object}	api.DataResponse
//	@Router			/admin/promotions/{id}/redemptions [get]
func (h *CouponHandler) ListRedemptions(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid coupon ID"))
		return
	}

	var q pagination.ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		validation.ValidationError(c, err)
		return
	}
	q.Process(pagination.QueryOptions{DefaultPageSize: 50, MaxPageSize: 100})

	cursor, err := q.Keyset()
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid pagination cursor").Wrap(err))
		return
	}

	entries, page, err := h.couponService.ListRedemptions(c.Request.Context(), id, q.PageSize, cursor)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := make([]CouponRedemptionResponse, len(entries))
	for i, e := range entries {
		res[i] = toRedemptionResponse(e)
	}

	c.JSON(http.StatusOK, api.SuccessResponse{Data: res, Meta: page})
}
