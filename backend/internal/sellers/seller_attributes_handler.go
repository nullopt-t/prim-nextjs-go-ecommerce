package sellers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/catalog/category"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
	"github.com/nullopt-t/prim-backend/internal/shared/validation"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
)

// SellerAttributeRequest is the merchant custom attribute payload.
type SellerAttributeRequest struct {
	// Category the attribute applies to.
	CategoryID string `json:"categoryId" binding:"required,uuid"`
	// Machine key, e.g. "fabric".
	Key string `json:"key" binding:"required,max=100"`
	// Display label.
	Label string `json:"label" binding:"required,max=200"`
	// Value kind; defaults to enum.
	ValueType string `json:"valueType" binding:"omitempty,oneof=enum text number boolean"`
	// Option set for enum attributes.
	AllowedValues []string `json:"allowedValues,omitempty"`
}

// SellerAttributeUpdateRequest is the merchant attribute edit payload.
type SellerAttributeUpdateRequest struct {
	// Updated display label.
	Label *string `json:"label,omitempty" binding:"omitempty,max=200"`
	// Updated option set.
	AllowedValues []string `json:"allowedValues,omitempty"`
}

// SellerAttributeResponse is the merchant-facing attribute definition. It
// mirrors the category attribute shape without stdlib-typed fields so the
// public spec can render it.
type SellerAttributeResponse struct {
	// Attribute id.
	ID string `json:"id" format:"uuid"`
	// Category the attribute applies to.
	CategoryID *string `json:"categoryId,omitempty" format:"uuid"`
	// Machine key.
	Key string `json:"key" example:"fabric"`
	// Display label.
	Label string `json:"label" example:"Fabric"`
	// Value kind.
	ValueType string `json:"valueType" example:"enum"`
	// Option set for enum attributes.
	AllowedValues []string `json:"allowedValues,omitempty"`
	// Whether the attribute is filterable on the storefront.
	IsFilterable bool `json:"isFilterable"`
	// Display order.
	SortOrder int `json:"sortOrder"`
}

func toSellerAttributeResponse(ai *category.AttributeItem) SellerAttributeResponse {
	var values []string
	_ = json.Unmarshal(ai.AllowedValues, &values)
	res := SellerAttributeResponse{
		ID:            ai.ID.String(),
		Key:           ai.Key,
		Label:         ai.Label,
		ValueType:     ai.ValueType,
		AllowedValues: values,
		IsFilterable:  ai.IsFilterable,
		SortOrder:     ai.SortOrder,
	}
	if ai.CategoryID != nil {
		id := ai.CategoryID.String()
		res.CategoryID = &id
	}
	return res
}

// ListSellerAttributes godoc
// @Summary      List my custom attributes
// @Description  Returns the authenticated seller's custom attribute definitions for a category.
// @Tags         Seller
// @Produce      json
// @Param        categoryId query string true "Category UUID" format(uuid)
// @Success      200 {object} api.DataResponse{data=[]SellerAttributeResponse}
// @Failure      403 {object} api.ForbiddenErrorResponse "Seller not active"
// @Router       /seller/attributes [get]
// @Security     SessionAuth
func (h *Handler) ListSellerAttributes(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	categoryID, err := uuid.Parse(c.Query("categoryId"))
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid category id").WithCode(apierr.CodeInvalidInput))
		return
	}

	items, err := h.service.ListAttributes(c.Request.Context(), userID, categoryID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := make([]SellerAttributeResponse, len(items))
	for i, item := range items {
		res[i] = toSellerAttributeResponse(&item)
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: res})
}

// CreateSellerAttribute godoc
// @Summary      Define a custom attribute
// @Description  Creates a seller-scoped attribute definition for a category.
// @Tags         Seller
// @Accept       json
// @Produce      json
// @Param        request body SellerAttributeRequest true "Attribute payload"
// @Success      201 {object} api.DataResponse{data=SellerAttributeResponse}
// @Failure      400 {object} api.BadReqResponse "Validation failed"
// @Failure      403 {object} api.ForbiddenErrorResponse "Seller not active"
// @Router       /seller/attributes [post]
// @Security     SessionAuth
func (h *Handler) CreateSellerAttribute(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	var req SellerAttributeRequest
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

	created, err := h.service.CreateAttribute(c.Request.Context(), userID, categoryID, req.Key, req.Label, req.ValueType, req.AllowedValues)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, api.DataResponse{Data: toSellerAttributeResponse(created)})
}

// UpdateSellerAttribute godoc
// @Summary      Edit a custom attribute
// @Description  Edits a seller-scoped attribute definition.
// @Tags         Seller
// @Accept       json
// @Produce      json
// @Param        id      path string                         true "Attribute UUID" format(uuid)
// @Param        request body SellerAttributeUpdateRequest true "Attribute edit payload"
// @Success      200 {object} api.MessageResponse
// @Failure      403 {object} api.ForbiddenErrorResponse "Seller not active"
// @Failure      404 {object} api.NotFoundErrorResponse "Attribute not found"
// @Router       /seller/attributes/{id} [patch]
// @Security     SessionAuth
func (h *Handler) UpdateSellerAttribute(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	attributeID, ok := bindSellerID(c)
	if !ok {
		return
	}

	var req SellerAttributeUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	if err := h.service.UpdateAttribute(c.Request.Context(), userID, attributeID, req.Label, req.AllowedValues); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "Attribute updated"})
}

// DeleteSellerAttribute godoc
// @Summary      Delete a custom attribute
// @Description  Removes a seller-scoped attribute definition.
// @Tags         Seller
// @Produce      json
// @Param        id path string true "Attribute UUID" format(uuid)
// @Success      200 {object} api.MessageResponse
// @Failure      403 {object} api.ForbiddenErrorResponse "Seller not active"
// @Failure      404 {object} api.NotFoundErrorResponse "Attribute not found"
// @Router       /seller/attributes/{id} [delete]
// @Security     SessionAuth
func (h *Handler) DeleteSellerAttribute(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	attributeID, ok := bindSellerID(c)
	if !ok {
		return
	}

	if err := h.service.DeleteAttribute(c.Request.Context(), userID, attributeID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "Attribute deleted"})
}
