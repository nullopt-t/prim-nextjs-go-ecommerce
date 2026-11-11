package review

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
)

type ReviewHandler struct {
	service *ReviewService
}

func NewHandler(s *ReviewService) *ReviewHandler {
	if s == nil {
		panic("review: service is required")
	}
	return &ReviewHandler{service: s}
}

type CreateReviewRequest struct {
	// Product UUID to review
	ProductID string `json:"productId" binding:"required,uuid" example:"60000000-0000-0000-0000-000000000001"`
	// Optional purchased Order Item UUID (if omitted, server automatically verifies buyer eligibility from delivered orders)
	OrderItemID *string `json:"orderItemId,omitempty" binding:"omitempty,uuid" example:"91000000-0000-0000-0000-000000000001"`
	// Star score between 1 and 5
	Rating int16 `json:"rating" binding:"required,min=1,max=5" example:"5"`
	// Optional review headline title
	Title *string `json:"title,omitempty" example:"Unbelievable performance!"`
	// Review detailed feedback narrative
	Body *string `json:"body,omitempty" example:"The M3 Pro chip handles heavy 4K rendering and compiles code effortlessly."`
}

type UpdateReviewRequest struct {
	// Updated star score between 1 and 5
	Rating *int16 `json:"rating,omitempty" binding:"omitempty,min=1,max=5" example:"5"`
	// Updated review headline title
	Title *string `json:"title,omitempty" example:"Updated review headline"`
	// Updated review feedback narrative
	Body *string `json:"body,omitempty" example:"Updated feedback after 3 months of continuous daily use."`
}

type UpdateReviewStatusRequest struct {
	// Moderation state: pending, approved, rejected
	Status string `json:"status" binding:"required,oneof=pending approved rejected" example:"approved"`
}

// ReviewUserResponse identifies the customer who wrote a review.
type ReviewUserResponse struct {
	// Display name of the reviewer
	Name string `json:"name" example:"Sarah Connor"`
}

type ReviewResponse struct {
	// Review UUID
	ID string `json:"id" example:"92000000-0000-0000-0000-000000000001"`
	// Reviewed Product UUID
	ProductID string `json:"productId" example:"60000000-0000-0000-0000-000000000001"`
	// Author Customer UUID
	UserID string `json:"userId" example:"10000000-0000-0000-0000-000000000002"`
	// Reviewer display name, omitted when the review carries none
	User *ReviewUserResponse `json:"user,omitempty"`
	// Title of the reviewed product, resolved at read time
	ProductTitle string `json:"productTitle,omitempty" example:"Aurora 27 4K Monitor"`
	// Public URL of the reviewed product's thumbnail, resolved at read time
	ProductImage string `json:"productImage,omitempty" example:"http://localhost:9000/catalog/products/abc.jpg"`
	// Verified Order Item purchase UUID
	OrderItemID string `json:"orderItemId" example:"91000000-0000-0000-0000-000000000001"`
	// Star score (1-5)
	Rating int16 `json:"rating" example:"5"`
	// Optional review headline
	Title *string `json:"title,omitempty" example:"Unbelievable performance!"`
	// Review detailed text
	Body *string `json:"body,omitempty" example:"The M3 Pro chip handles heavy 4K rendering and compiles code effortlessly."`
	// Moderation status: pending, approved, rejected
	Status string `json:"status" example:"approved"`
	// Number of helpful votes
	HelpfulCount int `json:"helpfulCount" example:"12"`
	// Whether requesting user voted this review helpful
	UserVoted bool `json:"userVoted" example:"false"`
	// Submission timestamp (RFC3339)
	CreatedAt string `json:"createdAt" example:"2026-08-13T07:56:17Z"`
	// Last updated timestamp (RFC3339)
	UpdatedAt string `json:"updatedAt" example:"2026-08-13T07:56:17Z"`
}

type HelpfulVoteResponse struct {
	Voted        bool `json:"voted" example:"true"`
	HelpfulCount int  `json:"helpfulCount" example:"13"`
}

type RatingSummaryResponse struct {
	// Average review score (1.0 to 5.0)
	AverageRating float64 `json:"averageRating" example:"4.5"`
	// Total number of approved reviews
	ReviewCount int `json:"reviewCount" example:"24"`
	// Star distribution mapping score (1-5) to count
	Distribution map[int16]int `json:"distribution"`
}

func mapReviewToResponse(r *model.Review) ReviewResponse {
	resp := ReviewResponse{
		ID:           r.ID.String(),
		ProductID:    r.ProductID.String(),
		UserID:       r.UserID.String(),
		OrderItemID:  r.OrderItemID.String(),
		Rating:       r.Rating,
		Title:        r.Title,
		Body:         r.Body,
		Status:       r.Status.String(),
		HelpfulCount: r.HelpfulCount,
		UserVoted:    r.UserVoted,
		ProductTitle: r.ProductTitle,
		ProductImage: r.ProductImageURL,
		CreatedAt:    r.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    r.UpdatedAt.Format(time.RFC3339),
	}
	if r.AuthorName != "" {
		resp.User = &ReviewUserResponse{Name: r.AuthorName}
	}
	return resp
}

func getUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return uuid.Nil, apierr.UnauthorizedError("unauthorized")
	}
	return userID, nil
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
		_ = c.Error(apierr.InvalidPayloadError("invalid request payload"))
		return
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		_ = c.Error(apierr.ValidationFailedError("invalid product id"))
		return
	}

	var orderItemID *uuid.UUID
	if req.OrderItemID != nil && *req.OrderItemID != "" {
		parsedID, err := uuid.Parse(*req.OrderItemID)
		if err != nil {
			_ = c.Error(apierr.ValidationFailedError("invalid order item id"))
			return
		}
		orderItemID = &parsedID
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
//	@Description	Returns a cursor-paginated list of all reviews written by the authenticated user.
//	@Tags			Reviews
//	@Produce		json
//	@Param			cursor	query		string																false	"Cursor for the next page"
//	@Param			pageSize	query		int																		false	"Reviews per page (max 100)"
//	@Success		200	{object}	api.PaginatedResponse{data=[]ReviewResponse,meta=pagination.CursorPage}	"User's reviews"
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
		_ = c.Error(apierr.ValidationFailedError("invalid query parameters"))
		return
	}
	q.Process(pagination.QueryOptions{DefaultPageSize: 10, MaxPageSize: 100})

	cursor, err := q.Keyset()
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid pagination cursor").Wrap(err))
		return
	}

	reviews, page, err := h.service.ListReviews(c.Request.Context(), q.PageSize, cursor, model.ReviewFilter{
		UserID:  &userID,
		VoterID: &userID,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}

	responses := make([]ReviewResponse, 0, len(reviews))
	for _, rv := range reviews {
		responses = append(responses, mapReviewToResponse(&rv))
	}

	c.JSON(http.StatusOK, api.PaginatedResponse{
		Data: responses,
		Meta: page,
	})
}

// ReviewEligibilityResponse defines the review eligibility status
type ReviewEligibilityResponse struct {
	CanReview bool `json:"canReview"`
}

// CheckEligibility godoc
//
//	@Summary		Check review eligibility
//	@Description	Checks whether the authenticated user has purchased and received the product and can review it.
//	@Tags			Reviews
//	@Produce		json
//	@Param			productId	query		string												true	"Product UUID"
//	@Success		200			{object}	api.DataResponse{data=ReviewEligibilityResponse}	"Eligibility status"
//	@Failure		400			{object}	api.BadRequestErrorResponse							"Invalid product ID"
//	@Failure		401			{object}	api.UnauthorizedErrorResponse						"Unauthorized"
//	@Failure		500			{object}	api.InternalServerErrorResponse						"Internal server error"
//	@Router			/reviews/eligibility [get]
func (h *ReviewHandler) CheckEligibility(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	productIDStr := c.Query("productId")
	if productIDStr == "" {
		_ = c.Error(apierr.ValidationFailedError("productId is required"))
		return
	}

	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		_ = c.Error(apierr.ValidationFailedError("invalid productId format"))
		return
	}

	canReview, err := h.service.CheckReviewEligibility(c.Request.Context(), userID, productID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{
		Data: ReviewEligibilityResponse{
			CanReview: canReview,
		},
	})
}

// ListPublicReviews godoc
//
//	@Summary		List public approved reviews
//	@Description	Returns a cursor-paginated list of approved reviews for storefront display.
//	@Tags			Reviews
//	@Produce		json
//	@Param			productId	query		string																false	"Filter by Product UUID"
//	@Param			userId		query		string																false	"Filter by author Customer UUID, e.g. for a reviewer's public profile"
//	@Param			rating		query		int																		false	"Filter by star rating (1-5)"
//	@Param			cursor		query		string																false	"Cursor for the next page"
//	@Param			pageSize	query		int																		false	"Reviews per page (max 100)"
//	@Success		200			{object}	api.PaginatedResponse{data=[]ReviewResponse,meta=pagination.CursorPage}	"Paginated list of approved reviews"
//	@Failure		400			{object}	api.BadRequestErrorResponse											"Invalid input"
//	@Failure		500			{object}	api.InternalServerErrorResponse										"Internal server error"
//	@Router			/reviews [get]
func (h *ReviewHandler) ListPublicReviews(c *gin.Context) {
	q := &pagination.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		_ = c.Error(apierr.ValidationFailedError("invalid query parameters"))
		return
	}
	q.Process(pagination.QueryOptions{DefaultPageSize: 10, MaxPageSize: 100})

	var productIDPtr *uuid.UUID
	if prodStr := c.Query("productId"); prodStr != "" {
		pid, err := uuid.Parse(prodStr)
		if err != nil {
			_ = c.Error(apierr.ValidationFailedError("invalid product id"))
			return
		}
		productIDPtr = &pid
	}

	var userIDPtr *uuid.UUID
	if uidStr := c.Query("userId"); uidStr != "" {
		uid, err := uuid.Parse(uidStr)
		if err != nil {
			_ = c.Error(apierr.ValidationFailedError("invalid user id"))
			return
		}
		userIDPtr = &uid
	}

	var ratingPtr *int16
	if ratingStr := c.Query("rating"); ratingStr != "" {
		rating, err := strconv.ParseInt(ratingStr, 10, 16)
		if err != nil || rating < 1 || rating > 5 {
			_ = c.Error(apierr.ValidationFailedError("rating must be between 1 and 5"))
			return
		}
		parsed := int16(rating)
		ratingPtr = &parsed
	}

	cursor, err := q.Keyset()
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid pagination cursor").Wrap(err))
		return
	}

	var voterID *uuid.UUID
	if uid, ok := middleware.GetUserID(c); ok {
		voterID = &uid
	}

	approvedStatus := model.ReviewStatusApproved
	reviews, page, err := h.service.ListReviews(c.Request.Context(), q.PageSize, cursor, model.ReviewFilter{
		ProductID: productIDPtr,
		UserID:    userIDPtr,
		Status:    &approvedStatus,
		Rating:    ratingPtr,
		VoterID:   voterID,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}

	responses := make([]ReviewResponse, 0, len(reviews))
	for _, rv := range reviews {
		responses = append(responses, mapReviewToResponse(&rv))
	}

	c.JSON(http.StatusOK, api.PaginatedResponse{
		Data: responses,
		Meta: page,
	})
}

// GetReviewByID godoc
//
//	@Summary		Get review by ID
//	@Description	Retrieves a review by its unique identifier. Only approved reviews are publicly visible; pending or rejected reviews require the author's identity.
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
		_ = c.Error(apierr.ValidationFailedError("invalid review id"))
		return
	}

	rv, err := h.service.GetReviewByID(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}

	callingUserID, _ := middleware.GetUserID(c)

	// IDOR protection: only return approved reviews, or allow author to view their own review
	if rv.Status != model.ReviewStatusApproved && (callingUserID == uuid.Nil || rv.UserID != callingUserID) {
		_ = c.Error(apierr.NotFoundError("review not found"))
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
		_ = c.Error(apierr.ValidationFailedError("invalid review id"))
		return
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var req UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apierr.InvalidPayloadError("invalid request payload"))
		return
	}

	rv, err := h.service.UpdateUserReview(c.Request.Context(), id, userID, UpdateReviewInput(req))
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
		_ = c.Error(apierr.ValidationFailedError("invalid review id"))
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
//	@Description	Returns a cursor-paginated list of reviews with optional filters for product, user, and status.
//	@Tags			Admin Reviews
//	@Produce		json
//	@Param			cursor		query		string																false	"Cursor for the next page"
//	@Param			pageSize	query		int																		false	"Reviews per page (max 100)"
//	@Param			productId	query		string																false	"Filter by Product ID (UUID)"
//	@Param			userId		query		string																false	"Filter by User ID (UUID)"
//	@Param			status		query		string																false	"Filter by Review Status (pending, approved, rejected)"
//	@Success		200			{object}	api.PaginatedResponse{data=[]ReviewResponse,meta=pagination.CursorPage}	"Paginated list of reviews"
//	@Failure		401			{object}	api.UnauthorizedErrorResponse										"Unauthorized"
//	@Failure		403			{object}	api.ForbiddenErrorResponse											"Forbidden"
//	@Failure		500			{object}	api.InternalServerErrorResponse										"Internal server error"
//	@Router			/admin/reviews [get]
func (h *ReviewHandler) AdminListReviews(c *gin.Context) {
	q := &pagination.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		_ = c.Error(apierr.ValidationFailedError("invalid query parameters"))
		return
	}
	q.Process(pagination.QueryOptions{DefaultPageSize: 10, MaxPageSize: 100})

	var productIDPtr *uuid.UUID
	productIDStr := c.Query("productId")
	if productIDStr == "" {
		productIDStr = c.Query("product_id")
	}
	if productIDStr != "" {
		id, err := uuid.Parse(productIDStr)
		if err != nil {
			_ = c.Error(apierr.ValidationFailedError("invalid product id"))
			return
		}
		productIDPtr = &id
	}

	var userIDPtr *uuid.UUID
	userIDStr := c.Query("userId")
	if userIDStr == "" {
		userIDStr = c.Query("user_id")
	}
	if userIDStr != "" {
		id, err := uuid.Parse(userIDStr)
		if err != nil {
			_ = c.Error(apierr.ValidationFailedError("invalid user id"))
			return
		}
		userIDPtr = &id
	}

	var statusPtr *model.ReviewStatus
	if statusStr := c.Query("status"); statusStr != "" {
		s, err := model.ParseReviewStatus(statusStr)
		if err != nil {
			_ = c.Error(apierr.ValidationFailedError("invalid review status"))
			return
		}
		statusPtr = &s
	}

	cursor, err := q.Keyset()
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid pagination cursor").Wrap(err))
		return
	}

	reviews, page, err := h.service.ListReviews(c.Request.Context(), q.PageSize, cursor, model.ReviewFilter{
		ProductID: productIDPtr,
		UserID:    userIDPtr,
		Status:    statusPtr,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}

	responses := make([]ReviewResponse, 0, len(reviews))
	for _, rv := range reviews {
		responses = append(responses, mapReviewToResponse(&rv))
	}

	c.JSON(http.StatusOK, api.PaginatedResponse{
		Data: responses,
		Meta: page,
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
		_ = c.Error(apierr.ValidationFailedError("invalid review id"))
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
		_ = c.Error(apierr.ValidationFailedError("invalid review id"))
		return
	}

	var req UpdateReviewStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apierr.InvalidPayloadError("invalid request payload"))
		return
	}

	status, err := model.ParseReviewStatus(req.Status)
	if err != nil {
		_ = c.Error(apierr.ValidationFailedError("invalid status"))
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
		_ = c.Error(apierr.ValidationFailedError("invalid review id"))
		return
	}

	err = h.service.DeleteReview(c.Request.Context(), id, uuid.Nil, true)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "review deleted successfully"})
}

// VoteHelpful godoc
//
//	@Summary		Vote review helpful
//	@Description	Toggles a helpful vote on a product review for the authenticated user
//	@Tags			Reviews
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Review UUID"
//	@Success		200	{object}	api.DataResponse{data=HelpfulVoteResponse}
//	@Failure		400	{object}	api.BadRequestErrorResponse
//	@Failure		401	{object}	api.UnauthorizedResponse
//	@Failure		404	{object}	api.NotFoundErrorResponse
//	@Failure		500	{object}	api.InternalServerErrorResponse
//	@Router			/reviews/{id}/helpful [post]
func (h *ReviewHandler) VoteHelpful(c *gin.Context) {
	reviewIDStr := c.Param("id")
	reviewID, err := uuid.Parse(reviewIDStr)
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid review ID"))
		return
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	voted, count, err := h.service.ToggleHelpful(c.Request.Context(), reviewID, userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{
		Data: HelpfulVoteResponse{
			Voted:        voted,
			HelpfulCount: count,
		},
	})
}
