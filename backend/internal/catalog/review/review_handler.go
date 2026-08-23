package review

import (
	"net/http"
	"time"

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
	ProductID   string  `json:"productId" binding:"required,uuid" example:"356cbaee-4700-4af5-ac9c-61aeeafd541c"`
	OrderItemID string  `json:"orderItemId" binding:"required,uuid" example:"c8ccec1c-ded5-4380-9f78-a1d4eb3d4f28"`
	Rating      int16   `json:"rating" binding:"required,min=1,max=5" example:"5"`
	Title       *string `json:"title,omitempty" example:"Amazing headphones!"`
	Body        *string `json:"body,omitempty" example:"The sound quality and active noise cancellation are top notch."`
}

type UpdateReviewRequest struct {
	Rating *int16  `json:"rating,omitempty" binding:"omitempty,min=1,max=5" example:"4"`
	Title  *string `json:"title,omitempty" example:"Updated title"`
	Body   *string `json:"body,omitempty" example:"Updated review body."`
}

type UpdateReviewStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending approved rejected" example:"approved"`
}

type ReviewResponse struct {
	ID          string  `json:"id" example:"8f123456-e89b-12d3-a456-426614174000"`
	ProductID   string  `json:"productId" example:"356cbaee-4700-4af5-ac9c-61aeeafd541c"`
	UserID      string  `json:"userId" example:"a1b2c3d4-e5f6-7890-1234-56789abcdef0"`
	OrderItemID string  `json:"orderItemId" example:"c8ccec1c-ded5-4380-9f78-a1d4eb3d4f28"`
	Rating      int16   `json:"rating" example:"5"`
	Title       *string `json:"title,omitempty" example:"Amazing headphones!"`
	Body        *string `json:"body,omitempty" example:"The sound quality is top notch."`
	Status      string  `json:"status" example:"pending"`
	CreatedAt   string  `json:"createdAt" example:"2026-08-10T15:00:00Z"`
	UpdatedAt   string  `json:"updatedAt" example:"2026-08-10T15:00:00Z"`
}

type RatingSummaryResponse struct {
	AverageRating float64       `json:"averageRating" example:"4.5"`
	ReviewCount   int           `json:"reviewCount" example:"24"`
	Distribution  map[int16]int `json:"distribution"`
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
		CreatedAt:   r.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   r.UpdatedAt.Format(time.RFC3339),
	}
}

func mapRatingSummaryResponse(s *model.RatingSummary) RatingSummaryResponse {
	if s == nil {
		return RatingSummaryResponse{
			Distribution: map[int16]int{1: 0, 2: 0, 3: 0, 4: 0, 5: 0},
		}
	}
	return RatingSummaryResponse{
		AverageRating: s.AverageRating,
		ReviewCount:   s.ReviewCount,
		Distribution:  s.Distribution,
	}
}

func getUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, apierr.ErrUnauthorized("unauthorized")
	}
	if id, ok := userIDVal.(uuid.UUID); ok {
		return id, nil
	}
	if idStr, ok := userIDVal.(string); ok {
		id, err := uuid.Parse(idStr)
		if err == nil {
			return id, nil
		}
	}
	return uuid.Nil, apierr.ErrUnauthorized("invalid user identifier in context")
}

// CreateReview godoc
//
//	@Summary		Submit a product review
//	@Description	Allows a customer to review a product they purchased via a specific order item.
//	@Tags			Reviews
//	@Accept			json
//	@Produce		json
//	@Param			request	body		CreateReviewRequest				true	"Review creation payload"
//	@Success		201		{object}	api.DataResponse{data=ReviewResponse}	"Review submitted successfully"
//	@Failure		400		{object}	api.BadRequestErrorResponse		"Invalid input or unverified purchase"
//	@Failure		401		{object}	api.UnauthorizedErrorResponse	"Unauthorized"
//	@Failure		403		{object}	api.ForbiddenErrorResponse		"Forbidden - not your purchase"
//	@Failure		409		{object}	api.ConflictErrorResponse		"Item already reviewed"
//	@Failure		500		{object}	api.InternalServerErrorResponse	"Internal server error"
//	@Router			/reviews [post]
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

	userID, err := getUserIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	rv, err := h.service.CreateReview(c.Request.Context(), CreateReviewInput{
		ProductID:   productID,
		UserID:      userID,
		OrderItemID: orderItemID,
		Rating:      req.Rating,
		Title:       req.Title,
		Body:        req.Body,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, api.DataResponse{Data: mapReviewToResponse(rv)})
}

// GetMyReviews godoc
//
//	@Summary		List current user's reviews
//	@Description	Returns a paginated list of all reviews written by the authenticated user.
//	@Tags			Reviews
//	@Produce		json
//	@Param			q	query		pagination.ListQuery												true	"Pagination and sorting parameters"
//	@Success		200	{object}	api.PaginatedResponse{data=[]ReviewResponse,meta=pagination.Page}	"User's reviews"
//	@Failure		401	{object}	api.UnauthorizedErrorResponse										"Unauthorized"
//	@Failure		500	{object}	api.InternalServerErrorResponse										"Internal server error"
//	@Router			/reviews/me [get]
func (h *ReviewHandler) GetMyReviews(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	q := &pagination.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		_ = c.Error(apierr.ErrValidationFailed("invalid query parameters"))
		return
	}
	q.Process(pagination.QueryOptions{DefaultPageSize: 10, MaxPageSize: 100})

	result, err := h.service.ListReviews(c.Request.Context(), q, nil, &userID, nil)
	if err != nil {
		_ = c.Error(err)
		return
	}

	responses := make([]ReviewResponse, 0, len(result.Items))
	for _, rv := range result.Items {
		responses = append(responses, mapReviewToResponse(rv))
	}

	c.JSON(http.StatusOK, api.PaginatedResponse{
		Data: responses,
		Meta: result.Page,
	})
}

// GetReviewByID godoc
//
//	@Summary		Get review by ID
//	@Description	Retrieves a review by its unique identifier.
//	@Tags			Reviews
//	@Produce		json
//	@Param			id	path		string									true	"Review ID"
//	@Success		200	{object}	api.DataResponse{data=ReviewResponse}	"Review details"
//	@Failure		400	{object}	api.BadRequestErrorResponse				"Invalid review ID format"
//	@Failure		404	{object}	api.NotFoundErrorResponse				"Review not found"
//	@Failure		500	{object}	api.InternalServerErrorResponse			"Internal server error"
//	@Router			/reviews/{id} [get]
func (h *ReviewHandler) GetReviewByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		_ = c.Error(apierr.ErrValidationFailed("invalid review id"))
		return
	}

	rv, err := h.service.GetReviewByID(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: mapReviewToResponse(rv)})
}

// UpdateReview godoc
//
//	@Summary		Update user's review
//	@Description	Updates rating, title, or body of a review written by the authenticated user. Resets status to pending for re-moderation.
//	@Tags			Reviews
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string									true	"Review ID"
//	@Param			request	body		UpdateReviewRequest						true	"Review update payload"
//	@Success		200		{object}	api.DataResponse{data=ReviewResponse}	"Review updated successfully"
//	@Failure		400		{object}	api.BadRequestErrorResponse				"Invalid input"
//	@Failure		401		{object}	api.UnauthorizedErrorResponse			"Unauthorized"
//	@Failure		403		{object}	api.ForbiddenErrorResponse				"Forbidden"
//	@Failure		404		{object}	api.NotFoundErrorResponse				"Review not found"
//	@Failure		500		{object}	api.InternalServerErrorResponse			"Internal server error"
//	@Router			/reviews/{id} [patch]
func (h *ReviewHandler) UpdateReview(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		_ = c.Error(apierr.ErrValidationFailed("invalid review id"))
		return
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var req UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apierr.ErrInvalidPayload("invalid request payload"))
		return
	}

	rv, err := h.service.UpdateUserReview(c.Request.Context(), id, userID, UpdateReviewInput{
		Rating: req.Rating,
		Title:  req.Title,
		Body:   req.Body,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: mapReviewToResponse(rv)})
}

// DeleteReview godoc
//
//	@Summary		Delete user's review
//	@Description	Deletes a review written by the authenticated user.
//	@Tags			Reviews
//	@Produce		json
//	@Param			id	path		string							true	"Review ID"
//	@Success		200	{object}	api.MessageResponse				"Review deleted successfully"
//	@Failure		400	{object}	api.BadRequestErrorResponse		"Invalid review ID"
//	@Failure		401	{object}	api.UnauthorizedErrorResponse	"Unauthorized"
//	@Failure		403	{object}	api.ForbiddenErrorResponse		"Forbidden"
//	@Failure		404	{object}	api.NotFoundErrorResponse		"Review not found"
//	@Failure		500	{object}	api.InternalServerErrorResponse	"Internal server error"
//	@Router			/reviews/{id} [delete]
func (h *ReviewHandler) DeleteReview(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		_ = c.Error(apierr.ErrValidationFailed("invalid review id"))
		return
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	err = h.service.DeleteReview(c.Request.Context(), id, userID, false)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "review deleted successfully"})
}

// AdminListReviews godoc
//
//	@Summary		List reviews (Admin)
//	@Description	Returns a paginated list of reviews with optional filters for product, user, and status.
//	@Tags			Admin Reviews
//	@Produce		json
//	@Param			q			query		pagination.ListQuery												true	"Pagination and sorting parameters"
//	@Param			productId	query		string																false	"Filter by Product ID (UUID)"
//	@Param			userId		query		string																false	"Filter by User ID (UUID)"
//	@Param			status		query		string																false	"Filter by Review Status (pending, approved, rejected)"
//	@Success		200			{object}	api.PaginatedResponse{data=[]ReviewResponse,meta=pagination.Page}	"Paginated list of reviews"
//	@Failure		401			{object}	api.UnauthorizedErrorResponse										"Unauthorized"
//	@Failure		403			{object}	api.ForbiddenErrorResponse											"Forbidden"
//	@Failure		500			{object}	api.InternalServerErrorResponse										"Internal server error"
//	@Router			/admin/reviews [get]
func (h *ReviewHandler) AdminListReviews(c *gin.Context) {
	q := &pagination.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		_ = c.Error(apierr.ErrValidationFailed("invalid query parameters"))
		return
	}
	q.Process(pagination.QueryOptions{DefaultPageSize: 10, MaxPageSize: 100})

	var productIDPtr *uuid.UUID
	if productIDStr := c.Query("productId"); productIDStr == "" {
		productIDStr = c.Query("product_id")
		if productIDStr != "" {
			if id, err := uuid.Parse(productIDStr); err == nil {
				productIDPtr = &id
			}
		}
	} else {
		if id, err := uuid.Parse(productIDStr); err == nil {
			productIDPtr = &id
		}
	}

	var userIDPtr *uuid.UUID
	if userIDStr := c.Query("userId"); userIDStr == "" {
		userIDStr = c.Query("user_id")
		if userIDStr != "" {
			if id, err := uuid.Parse(userIDStr); err == nil {
				userIDPtr = &id
			}
		}
	} else {
		if id, err := uuid.Parse(userIDStr); err == nil {
			userIDPtr = &id
		}
	}

	var statusPtr *model.ReviewStatus
	if statusStr := c.Query("status"); statusStr != "" {
		s, err := model.ParseReviewStatus(statusStr)
		if err == nil {
			statusPtr = &s
		}
	}

	result, err := h.service.ListReviews(c.Request.Context(), q, productIDPtr, userIDPtr, statusPtr)
	if err != nil {
		_ = c.Error(err)
		return
	}

	responses := make([]ReviewResponse, 0, len(result.Items))
	for _, rv := range result.Items {
		responses = append(responses, mapReviewToResponse(rv))
	}

	c.JSON(http.StatusOK, api.PaginatedResponse{
		Data: responses,
		Meta: result.Page,
	})
}

// AdminGetReviewByID godoc
//
//	@Summary		Get review details (Admin)
//	@Description	Retrieves full details of a review for moderation.
//	@Tags			Admin Reviews
//	@Produce		json
//	@Param			id	path		string									true	"Review ID"
//	@Success		200	{object}	api.DataResponse{data=ReviewResponse}	"Review details"
//	@Failure		400	{object}	api.BadRequestErrorResponse				"Invalid review ID"
//	@Failure		401	{object}	api.UnauthorizedErrorResponse			"Unauthorized"
//	@Failure		403	{object}	api.ForbiddenErrorResponse				"Forbidden"
//	@Failure		404	{object}	api.NotFoundErrorResponse				"Review not found"
//	@Failure		500	{object}	api.InternalServerErrorResponse			"Internal server error"
//	@Router			/admin/reviews/{id} [get]
func (h *ReviewHandler) AdminGetReviewByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		_ = c.Error(apierr.ErrValidationFailed("invalid review id"))
		return
	}

	rv, err := h.service.GetReviewByID(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: mapReviewToResponse(rv)})
}

// UpdateReviewStatus godoc
//
//	@Summary		Moderate review status (Admin)
//	@Description	Updates a review's moderation status (pending, approved, rejected).
//	@Tags			Admin Reviews
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Review ID"
//	@Param			request	body		UpdateReviewStatusRequest		true	"New status payload"
//	@Success		200		{object}	api.MessageResponse				"Status updated successfully"
//	@Failure		400		{object}	api.BadRequestErrorResponse		"Invalid input or status"
//	@Failure		401		{object}	api.UnauthorizedErrorResponse	"Unauthorized"
//	@Failure		403		{object}	api.ForbiddenErrorResponse		"Forbidden"
//	@Failure		404		{object}	api.NotFoundErrorResponse		"Review not found"
//	@Failure		500		{object}	api.InternalServerErrorResponse	"Internal server error"
//	@Router			/admin/reviews/{id}/status [patch]
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

// AdminDeleteReview godoc
//
//	@Summary		Delete review (Admin)
//	@Description	Permanently deletes a review.
//	@Tags			Admin Reviews
//	@Produce		json
//	@Param			id	path		string							true	"Review ID"
//	@Success		200	{object}	api.MessageResponse				"Review deleted successfully"
//	@Failure		400	{object}	api.BadRequestErrorResponse		"Invalid review ID"
//	@Failure		401	{object}	api.UnauthorizedErrorResponse	"Unauthorized"
//	@Failure		403	{object}	api.ForbiddenErrorResponse		"Forbidden"
//	@Failure		404	{object}	api.NotFoundErrorResponse		"Review not found"
//	@Failure		500	{object}	api.InternalServerErrorResponse	"Internal server error"
//	@Router			/admin/reviews/{id} [delete]
func (h *ReviewHandler) AdminDeleteReview(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		_ = c.Error(apierr.ErrValidationFailed("invalid review id"))
		return
	}

	err = h.service.DeleteReview(c.Request.Context(), id, uuid.Nil, true)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "review deleted successfully"})
}
