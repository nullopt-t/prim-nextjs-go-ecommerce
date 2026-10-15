package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
)

type AdminHandler struct {
	service *AdminService
}

func NewHandler(service *AdminService) *AdminHandler {
	return &AdminHandler{service: service}
}

func (h *AdminHandler) GetDashboard(c *gin.Context) {
	stats, err := h.service.GetDashboardStats(c.Request.Context())
	if err != nil {
		_ = c.Error(apierr.InternalServerError("Failed to fetch dashboard stats").Wrap(err))
		return
	}
	c.JSON(http.StatusOK, api.DataResponse{Data: stats})
}

// Coupons / Promotions
func (h *AdminHandler) ListPromotions(c *gin.Context) {
	list, err := h.service.ListCoupons(c.Request.Context())
	if err != nil {
		_ = c.Error(apierr.InternalServerError("Failed to fetch promotions").Wrap(err))
		return
	}
	c.JSON(http.StatusOK, api.DataResponse{Data: list})
}

type CreatePromotionRequest struct {
	Code          string  `json:"code" binding:"required"`
	DiscountType  string  `json:"discountType" binding:"required"`
	DiscountValue int64   `json:"discountValue" binding:"required"`
	Currency      *string `json:"currency"`
	MinCartAmount *int64  `json:"minCartAmount"`
	UsageLimit    *int    `json:"usageLimit"`
	PerUserLimit  *int    `json:"perUserLimit"`
	StartsAt      *string `json:"startsAt"`
	ExpiresAt     *string `json:"expiresAt"`
	Status        string  `json:"status"`
}

func (h *AdminHandler) CreatePromotion(c *gin.Context) {
	var req CreatePromotionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid request body").Wrap(err))
		return
	}

	item := &PromotionItem{
		Code:          req.Code,
		DiscountType:  req.DiscountType,
		DiscountValue: req.DiscountValue,
		Currency:      req.Currency,
		MinCartAmount: req.MinCartAmount,
		UsageLimit:    req.UsageLimit,
		PerUserLimit:  req.PerUserLimit,
		Status:        req.Status,
	}

	if req.StartsAt != nil && *req.StartsAt != "" {
		if t, err := time.Parse(time.RFC3339, *req.StartsAt); err == nil {
			item.StartsAt = &t
		}
	}
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		if t, err := time.Parse(time.RFC3339, *req.ExpiresAt); err == nil {
			item.ExpiresAt = &t
		}
	}

	if err := h.service.CreateCoupon(c.Request.Context(), item); err != nil {
		_ = c.Error(apierr.InternalServerError("Failed to create promotion").Wrap(err))
		return
	}

	c.JSON(http.StatusCreated, api.DataResponse{Data: item})
}

func (h *AdminHandler) UpdatePromotionStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid coupon ID"))
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid request body"))
		return
	}

	if err := h.service.UpdateCouponStatus(c.Request.Context(), id, req.Status); err != nil {
		_ = c.Error(apierr.InternalServerError("Failed to update status").Wrap(err))
		return
	}
	c.JSON(http.StatusOK, api.MessageResponse{Message: "Promotion status updated"})
}

func (h *AdminHandler) DeletePromotion(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid coupon ID"))
		return
	}

	if err := h.service.DeleteCoupon(c.Request.Context(), id); err != nil {
		_ = c.Error(apierr.InternalServerError("Failed to delete promotion").Wrap(err))
		return
	}
	c.JSON(http.StatusOK, api.MessageResponse{Message: "Promotion deleted"})
}

// Payments
func (h *AdminHandler) ListPayments(c *gin.Context) {
	limit := 50
	offset := 0
	if l := c.Query("pageSize"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}
	if p := c.Query("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			offset = (val - 1) * limit
		}
	}

	payments, total, err := h.service.ListPayments(c.Request.Context(), limit, offset)
	if err != nil {
		_ = c.Error(apierr.InternalServerError("Failed to fetch payments").Wrap(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": payments,
		"meta": gin.H{
			"totalItems": total,
			"pageSize":   limit,
		},
	})
}

func (h *AdminHandler) RefundPayment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid payment ID"))
		return
	}

	if err := h.service.RefundPayment(c.Request.Context(), id); err != nil {
		_ = c.Error(apierr.InternalServerError("Failed to refund payment").Wrap(err))
		return
	}
	c.JSON(http.StatusOK, api.MessageResponse{Message: "Payment marked as refunded"})
}

// Media
func (h *AdminHandler) ListMedia(c *gin.Context) {
	limit := 50
	offset := 0
	if l := c.Query("pageSize"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}
	if p := c.Query("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			offset = (val - 1) * limit
		}
	}

	media, total, err := h.service.ListMedia(c.Request.Context(), limit, offset)
	if err != nil {
		_ = c.Error(apierr.InternalServerError("Failed to fetch media").Wrap(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": media,
		"meta": gin.H{
			"totalItems": total,
			"pageSize":   limit,
		},
	})
}

// Attributes
func (h *AdminHandler) ListAttributes(c *gin.Context) {
	list, err := h.service.ListAttributes(c.Request.Context())
	if err != nil {
		_ = c.Error(apierr.InternalServerError("Failed to fetch attributes").Wrap(err))
		return
	}
	c.JSON(http.StatusOK, api.DataResponse{Data: list})
}

type CreateAttributeRequest struct {
	CategoryID    *string  `json:"categoryId"`
	Key           string   `json:"key" binding:"required"`
	Label         string   `json:"label" binding:"required"`
	ValueType     string   `json:"valueType"`
	AllowedValues []string `json:"allowedValues"`
	IsFilterable  bool     `json:"isFilterable"`
	SortOrder     int      `json:"sortOrder"`
}

func (h *AdminHandler) CreateAttribute(c *gin.Context) {
	var req CreateAttributeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid request body").Wrap(err))
		return
	}

	var catUUID *uuid.UUID
	if req.CategoryID != nil && *req.CategoryID != "" {
		if parsed, err := uuid.Parse(*req.CategoryID); err == nil {
			catUUID = &parsed
		}
	}

	allowedBytes, _ := json.Marshal(req.AllowedValues)
	if len(req.AllowedValues) == 0 {
		allowedBytes = []byte("[]")
	}

	item := &AttributeItem{
		CategoryID:    catUUID,
		Key:           req.Key,
		Label:         req.Label,
		ValueType:     req.ValueType,
		AllowedValues: allowedBytes,
		IsFilterable:  req.IsFilterable,
		SortOrder:     req.SortOrder,
	}

	if err := h.service.CreateAttribute(c.Request.Context(), item); err != nil {
		_ = c.Error(apierr.InternalServerError("Failed to create attribute").Wrap(err))
		return
	}
	c.JSON(http.StatusCreated, api.DataResponse{Data: item})
}

func (h *AdminHandler) DeleteAttribute(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid attribute ID"))
		return
	}

	if err := h.service.DeleteAttribute(c.Request.Context(), id); err != nil {
		_ = c.Error(apierr.InternalServerError("Failed to delete attribute").Wrap(err))
		return
	}
	c.JSON(http.StatusOK, api.MessageResponse{Message: "Attribute deleted"})
}

// Inventory
func (h *AdminHandler) ListInventoryOverview(c *gin.Context) {
	limit := 50
	offset := 0
	if l := c.Query("pageSize"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}
	if p := c.Query("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			offset = (val - 1) * limit
		}
	}

	items, total, err := h.service.ListInventoryOverview(c.Request.Context(), limit, offset)
	if err != nil {
		_ = c.Error(apierr.InternalServerError("Failed to fetch inventory overview").Wrap(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": items,
		"meta": gin.H{
			"totalItems": total,
			"pageSize":   limit,
		},
	})
}

func (h *AdminHandler) ListRecentLedgers(c *gin.Context) {
	limit := 30
	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	ledgers, err := h.service.ListRecentLedgers(c.Request.Context(), limit)
	if err != nil {
		_ = c.Error(apierr.InternalServerError("Failed to fetch ledgers").Wrap(err))
		return
	}
	c.JSON(http.StatusOK, api.DataResponse{Data: ledgers})
}

type AdjustInventoryRequest struct {
	Quantity int    `json:"quantity" binding:"required"`
	Reason   string `json:"reason" binding:"required"`
}

func (h *AdminHandler) AdjustInventory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("variantId"))
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid variant ID"))
		return
	}

	var req AdjustInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid request body").Wrap(err))
		return
	}

	if err := h.service.AdjustInventory(c.Request.Context(), id, req.Quantity, req.Reason); err != nil {
		_ = c.Error(apierr.InternalServerError("Failed to adjust inventory").Wrap(err))
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "Inventory adjusted successfully"})
}
