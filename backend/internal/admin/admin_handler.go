package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/internal/shared/validation"
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
	q := &pagination.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		validation.ValidationError(c, err)
		return
	}
	q.Process(pagination.QueryOptions{DefaultPageSize: 50, MaxPageSize: 100})

	payments, total, err := h.service.ListPayments(c.Request.Context(), q.PageSize, q.Offset)
	if err != nil {
		_ = c.Error(apierr.InternalServerError("Failed to fetch payments").Wrap(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": payments,
		"meta": gin.H{
			"totalItems": total,
			"pageSize":   q.PageSize,
		},
	})
}

func (h *AdminHandler) UpdatePaymentStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid payment ID"))
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid status body"))
		return
	}

	if err := h.service.UpdatePaymentStatus(c.Request.Context(), id, req.Status); err != nil {
		_ = c.Error(apierr.InternalServerError("Failed to update payment status").Wrap(err))
		return
	}
	c.JSON(http.StatusOK, api.MessageResponse{Message: "Payment status updated successfully"})
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
	q := &pagination.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		validation.ValidationError(c, err)
		return
	}
	q.Process(pagination.QueryOptions{DefaultPageSize: 50, MaxPageSize: 100})

	media, total, err := h.service.ListMedia(c.Request.Context(), q.PageSize, q.Offset)
	if err != nil {
		_ = c.Error(apierr.InternalServerError("Failed to fetch media").Wrap(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": media,
		"meta": gin.H{
			"totalItems": total,
			"pageSize":   q.PageSize,
		},
	})
}

func (h *AdminHandler) UploadMedia(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		_ = c.Error(apierr.BadRequestError("file is required").WithCode(apierr.CodeInvalidInput))
		return
	}

	src, err := file.Open()
	if err != nil {
		_ = c.Error(apierr.InternalServerError("Failed to read uploaded file").Wrap(err))
		return
	}
	defer src.Close()

	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	item, err := h.service.UploadMedia(c.Request.Context(), contentType, file.Size, src)
	if err != nil {
		_ = c.Error(apierr.InternalServerError("Failed to upload media").Wrap(err))
		return
	}

	c.JSON(http.StatusCreated, api.DataResponse{Data: item})
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
	q := &pagination.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		validation.ValidationError(c, err)
		return
	}
	q.Process(pagination.QueryOptions{DefaultPageSize: 50, MaxPageSize: 100})

	items, total, err := h.service.ListInventoryOverview(c.Request.Context(), q.PageSize, q.Offset)
	if err != nil {
		_ = c.Error(apierr.InternalServerError("Failed to fetch inventory overview").Wrap(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": items,
		"meta": gin.H{
			"totalItems": total,
			"pageSize":   q.PageSize,
		},
	})
}

func (h *AdminHandler) ListRecentLedgers(c *gin.Context) {
	q := &pagination.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		validation.ValidationError(c, err)
		return
	}
	// Limit-only listing: clamp page size, no offset
	q.Process(pagination.QueryOptions{DefaultPageSize: 30, MaxPageSize: 100})

	var variantUUID *uuid.UUID
	if vId := c.Query("variantId"); vId != "" {
		if parsed, err := uuid.Parse(vId); err == nil {
			variantUUID = &parsed
		}
	}

	var productUUID *uuid.UUID
	if pId := c.Query("productId"); pId != "" {
		if parsed, err := uuid.Parse(pId); err == nil {
			productUUID = &parsed
		}
	}

	ledgers, err := h.service.ListRecentLedgers(c.Request.Context(), q.PageSize, variantUUID, productUUID)
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

func (h *AdminHandler) AdjustProductInventory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("productId"))
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid product ID"))
		return
	}

	var req AdjustInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid request body").Wrap(err))
		return
	}

	if err := h.service.AdjustProductInventory(c.Request.Context(), id, req.Quantity, req.Reason); err != nil {
		_ = c.Error(apierr.InternalServerError("Failed to adjust product inventory").Wrap(err))
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "Product inventory adjusted successfully"})
}

// Settings
func (h *AdminHandler) GetSettings(c *gin.Context) {
	settings, err := h.service.GetSettings(c.Request.Context())
	if err != nil {
		_ = c.Error(apierr.InternalServerError("Failed to fetch settings").Wrap(err))
		return
	}
	c.JSON(http.StatusOK, api.DataResponse{Data: settings})
}

type UpdateSettingsRequest struct {
	Items map[string]interface{} `json:"items" binding:"required"`
}

func (h *AdminHandler) UpdateSettings(c *gin.Context) {
	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid request body").Wrap(err))
		return
	}

	if err := h.service.UpdateSettings(c.Request.Context(), req.Items); err != nil {
		_ = c.Error(apierr.InternalServerError("Failed to update settings").Wrap(err))
		return
	}
	c.JSON(http.StatusOK, api.MessageResponse{Message: "Settings updated successfully"})
}
