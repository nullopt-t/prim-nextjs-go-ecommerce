package review

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/pagination"
)

type ReviewHandler struct {
	service *ReviewService
}

func NewHandler(s *ReviewService) *ReviewHandler {
	return &ReviewHandler{service: s}
}

type CreateReviewRequest struct {
	ProductID   string  `json:"product_id" binding:"required,uuid"`
	OrderItemID string  `json:"order_item_id" binding:"required,uuid"`
	Rating      int16   `json:"rating" binding:"required,min=1,max=5"`
	Title       *string `json:"title,omitempty"`
	Body        *string `json:"body,omitempty"`
}

type UpdateReviewStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type ReviewResponse struct {
	ID          string    `json:"id"`
	ProductID   string    `json:"product_id"`
	UserID      string    `json:"user_id"`
	OrderItemID string    `json:"order_item_id"`
	Rating      int16     `json:"rating"`
	Title       *string   `json:"title,omitempty"`
	Body        *string   `json:"body,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
}

func mapReviewToResponse(r *model.Review) ReviewResponse {
	return ReviewResponse{
		ID:          r.ID.String(),
		ProductID:   r.ProductID.String(),
		UserID:      r.UserID.String(),
		OrderItemID: r.OrderItemID.String(),
		Rating:      r.Rating,
		Title:       r.Title,
		Body:        r.Body,
		Status:      r.Status.String(),
		CreatedAt:   r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   r.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *ReviewHandler) CreateReview(c *gin.Context) {
	var req CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apierr.ErrInvalidPayload("invalid request payload"))
		return
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		_ = c.Error(apierr.ErrValidationFailed("invalid product id"))
		return
	}

	orderItemID, err := uuid.Parse(req.OrderItemID)
	if err != nil {
		_ = c.Error(apierr.ErrValidationFailed("invalid order item id"))
		return
	}

	userIDVal, exists := c.Get("userID")
	if !exists {
		_ = c.Error(apierr.ErrUnauthorized("unauthorized"))
		return
	}
	userIDStr, ok := userIDVal.(string)
	if !ok {
		_ = c.Error(apierr.ErrUnauthorized("invalid user id in context"))
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		_ = c.Error(apierr.ErrUnauthorized("invalid user id format"))
		return
	}

	rv, err := h.service.CreateReview(c.Request.Context(), productID, userID, orderItemID, req.Rating, req.Title, req.Body)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, api.DataResponse{Data: mapReviewToResponse(rv)})
}

func (h *ReviewHandler) AdminListReviews(c *gin.Context) {
	q := &pagination.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		_ = c.Error(apierr.ErrValidationFailed("invalid query parameters"))
		return
	}
	q.Process(pagination.QueryOptions{DefaultPageSize: 10, MaxPageSize: 100})
	
	var productIDPtr *uuid.UUID
	if productIDStr := c.Query("product_id"); productIDStr != "" {
		id, err := uuid.Parse(productIDStr)
		if err == nil {
			productIDPtr = &id
		}
	}

	var statusPtr *model.ReviewStatus
	if statusStr := c.Query("status"); statusStr != "" {
		s, err := model.ParseReviewStatus(statusStr)
		if err == nil {
			statusPtr = &s
		}
	}

	result, err := h.service.ListReviews(c.Request.Context(), q, productIDPtr, nil, statusPtr)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var responses []ReviewResponse
	for _, rv := range result.Items {
		responses = append(responses, mapReviewToResponse(rv))
	}

	c.JSON(http.StatusOK, api.PaginatedResponse{
		Data: responses,
		Meta: result.Page,
	})
}

func (h *ReviewHandler) UpdateReviewStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		_ = c.Error(apierr.ErrValidationFailed("invalid review id"))
		return
	}

	var req UpdateReviewStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apierr.ErrInvalidPayload("invalid request payload"))
		return
	}

	status, err := model.ParseReviewStatus(req.Status)
	if err != nil {
		_ = c.Error(apierr.ErrValidationFailed("invalid status"))
		return
	}

	err = h.service.UpdateReviewStatus(c.Request.Context(), id, status)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "status updated successfully"})
}
