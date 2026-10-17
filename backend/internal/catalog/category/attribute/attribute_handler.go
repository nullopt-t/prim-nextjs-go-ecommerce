package attribute

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
)

type AttributeHandler struct {
	attributeService *AttributeService
}

func NewHandler(s *AttributeService) *AttributeHandler {
	return &AttributeHandler{
		attributeService: s,
	}
}

func (h *AttributeHandler) ListAttributes(c *gin.Context) {
	list, err := h.attributeService.ListAttributes(c.Request.Context())
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

func (h *AttributeHandler) CreateAttribute(c *gin.Context) {
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

	if err := h.attributeService.CreateAttribute(c.Request.Context(), item); err != nil {
		_ = c.Error(apierr.InternalServerError("Failed to create attribute").Wrap(err))
		return
	}
	c.JSON(http.StatusCreated, api.DataResponse{Data: item})
}

func (h *AttributeHandler) DeleteAttribute(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid attribute ID"))
		return
	}

	if err := h.attributeService.DeleteAttribute(c.Request.Context(), id); err != nil {
		_ = c.Error(apierr.InternalServerError("Failed to delete attribute").Wrap(err))
		return
	}
	c.JSON(http.StatusOK, api.MessageResponse{Message: "Attribute deleted"})
}
