package inventory

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/shared/validation"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
)

type InventoryHandler struct {
	service *InventoryService
}

func NewHandler(service *InventoryService) *InventoryHandler {
	return &InventoryHandler{service: service}
}

func (h *InventoryHandler) ListInventoryOverview(c *gin.Context) {
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

func (h *InventoryHandler) ListRecentLedgers(c *gin.Context) {
	q := &pagination.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		validation.ValidationError(c, err)
		return
	}
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

func (h *InventoryHandler) AdjustInventory(c *gin.Context) {
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

func (h *InventoryHandler) AdjustProductInventory(c *gin.Context) {
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
