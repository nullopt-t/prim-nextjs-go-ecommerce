package history

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
)

// Handler translates HTTP requests into browsing-history service calls.
type Handler struct {
	service *Service
}

// NewHandler constructs a browsing-history handler.
func NewHandler(service *Service) *Handler {
	if service == nil {
		panic("history: service is required")
	}
	return &Handler{service: service}
}

// ListHistoryQuery bounds the recent-products projection.
type ListHistoryQuery struct {
	// Maximum number of products to return (1-50, default 10).
	Limit int `form:"limit" example:"10"`
}

// SyncHistoryRequest carries guest identifiers collected before sign-in.
// At least one of productIds or slugs is required.
type SyncHistoryRequest struct {
	// Product UUIDs viewed by the guest.
	ProductIDs []string `json:"productIds,omitempty" example:"[\"60000000-0000-0000-0000-000000000001\"]"`
	// Product slugs viewed by the guest, newest first.
	Slugs []string `json:"slugs,omitempty" example:"[\"macbook-pro-16\",\"wireless-headphones\"]"`
}

// SyncHistoryResponse reports how many products were accepted into the account.
type SyncHistoryResponse struct {
	// Number of products merged into the authenticated account's history.
	Synced int `json:"synced" example:"2"`
}

func userIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return uuid.Nil, apierr.UnauthorizedError("unauthorized")
	}
	return userID, nil
}

// GetHistory godoc
//
//	@Summary		List recently viewed products
//	@Description	Returns the authenticated customer's recently viewed products, newest first. Unpublished, deleted, and inactive-vendor products are omitted.
//	@Tags			History
//	@Produce		json
//	@Security		SessionAuth
//	@Param			limit	query		int	false	"Maximum products to return (1-50, default 10)"
//	@Success		200		{object}	api.DataResponse{data=[]ViewedProductItem}	"Recently viewed products"
//	@Failure		400		{object}	api.BadRequestErrorResponse					"Invalid limit"
//	@Failure		401		{object}	api.UnauthorizedErrorResponse				"Unauthorized"
//	@Failure		500		{object}	api.InternalServerErrorResponse				"Internal server error"
//	@Router			/user/history/products [get]
func (h *Handler) GetHistory(c *gin.Context) {
	var query ListHistoryQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		_ = c.Error(apierr.InvalidQueryParameterError("invalid history query"))
		return
	}

	userID, err := userIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	items, err := h.service.List(c.Request.Context(), userID, query.Limit)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: items})
}

// RecordView godoc
//
//	@Summary		Record a product view
//	@Description	Records or refreshes a product view for the authenticated customer. Repeated views update viewed_at without creating duplicates.
//	@Tags			History
//	@Produce		json
//	@Security		SessionAuth
//	@Param			id	path		string	true	"Product UUID"
//	@Success		204					"View recorded"
//	@Failure		400	{object}	api.BadRequestErrorResponse		"Invalid product id"
//	@Failure		401	{object}	api.UnauthorizedErrorResponse	"Unauthorized"
//	@Failure		404	{object}	api.NotFoundErrorResponse		"Product not found or not visible"
//	@Failure		429	{object}	api.TooManyRequestsErrorResponse	"Rate limit exceeded"
//	@Failure		500	{object}	api.InternalServerErrorResponse	"Internal server error"
//	@Router			/user/history/products/{id} [post]
func (h *Handler) RecordView(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil || productID == uuid.Nil {
		_ = c.Error(apierr.InvalidUUIDError("invalid product id"))
		return
	}

	userID, err := userIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	if err := h.service.RecordView(c.Request.Context(), userID, productID); err != nil {
		_ = c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}

// SyncHistory godoc
//
//	@Summary		Merge guest browsing history
//	@Description	Merges guest product IDs or slugs into the authenticated account after sign-in. Unknown or no-longer-visible products are ignored.
//	@Tags			History
//	@Accept			json
//	@Produce		json
//	@Security		SessionAuth
//	@Param			request	body		SyncHistoryRequest	true	"Guest product identifiers, newest first"
//	@Success		200		{object}	api.DataResponse{data=SyncHistoryResponse}	"History merged"
//	@Failure		400		{object}	api.BadRequestErrorResponse					"Invalid or oversized payload"
//	@Failure		401		{object}	api.UnauthorizedErrorResponse				"Unauthorized"
//	@Failure		429		{object}	api.TooManyRequestsErrorResponse			"Rate limit exceeded"
//	@Failure		500		{object}	api.InternalServerErrorResponse				"Internal server error"
//	@Router			/user/history/products/sync [post]
func (h *Handler) SyncHistory(c *gin.Context) {
	var req SyncHistoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apierr.InvalidPayloadError("invalid request payload"))
		return
	}

	productIDs := make([]uuid.UUID, 0, len(req.ProductIDs))
	for _, raw := range req.ProductIDs {
		id, err := uuid.Parse(raw)
		if err != nil || id == uuid.Nil {
			_ = c.Error(apierr.ValidationFailedError("invalid product id").
				WithFields(api.FieldError{
					Field:      "productIds",
					Constraint: "uuid",
					Message:    "productIds must contain valid UUIDs",
				}))
			return
		}
		productIDs = append(productIDs, id)
	}

	userID, err := userIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	synced, err := h.service.SyncGuestHistory(c.Request.Context(), userID, productIDs, req.Slugs)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: SyncHistoryResponse{Synced: synced}})
}

// DeleteHistoryItem godoc
//
//	@Summary		Delete one recently viewed product
//	@Description	Removes a single product from the authenticated customer's browsing history.
//	@Tags			History
//	@Produce		json
//	@Security		SessionAuth
//	@Param			id	path		string	true	"Product UUID"
//	@Success		204					"History item deleted"
//	@Failure		400	{object}	api.BadRequestErrorResponse		"Invalid product id"
//	@Failure		401	{object}	api.UnauthorizedErrorResponse	"Unauthorized"
//	@Failure		404	{object}	api.NotFoundErrorResponse		"History item not found"
//	@Failure		500	{object}	api.InternalServerErrorResponse	"Internal server error"
//	@Router			/user/history/products/{id} [delete]
func (h *Handler) DeleteHistoryItem(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil || productID == uuid.Nil {
		_ = c.Error(apierr.InvalidUUIDError("invalid product id"))
		return
	}

	userID, err := userIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	if err := h.service.DeleteView(c.Request.Context(), userID, productID); err != nil {
		_ = c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ClearHistory godoc
//
//	@Summary		Clear browsing history
//	@Description	Deletes every recently viewed product for the authenticated customer.
//	@Tags			History
//	@Produce		json
//	@Security		SessionAuth
//	@Success		204					"History cleared"
//	@Failure		401	{object}	api.UnauthorizedErrorResponse	"Unauthorized"
//	@Failure		500	{object}	api.InternalServerErrorResponse	"Internal server error"
//	@Router			/user/history/products [delete]
func (h *Handler) ClearHistory(c *gin.Context) {
	userID, err := userIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	if err := h.service.Clear(c.Request.Context(), userID); err != nil {
		_ = c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}
