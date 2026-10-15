package user

import (
	"net/http"
	"time"

	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/validation"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/api/security"
	"github.com/nullopt-t/prim-backend/pkg/config"
	"github.com/nullopt-t/prim-backend/pkg/log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UsersResponse struct {
	Users []UserResponse `json:"users"`
}

type UserByEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type UserURIParam struct {
	UserID uuid.UUID `uri:"id" json:"id" binding:"required"`
}

type UpdateUserRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

type UpdateUserStatusRequest struct {
	Status         string     `json:"status" binding:"required"`
	SuspendedUntil *time.Time `json:"suspendedUntil,omitempty"`
}

type UserResponse struct {
	ID              uuid.UUID  `json:"id"`
	Email           string     `json:"email"`
	FullName        *string    `json:"fullName,omitempty"`
	Role            string     `json:"role"`
	IsEmailVerified bool       `json:"isEmailVerified"`
	Status          string     `json:"status"`
	SuspendedUntil  *time.Time `json:"suspendedUntil,omitempty"`
	LockedUntil     *time.Time `json:"lockedUntil,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

func ToUserResponse(user model.User) UserResponse {
	roleStr := "customer"
	if user.Role != nil {
		roleStr = string(*user.Role)
	}
	return UserResponse{
		ID:              user.ID,
		Email:           user.Identifier,
		FullName:        user.FullName,
		Role:            roleStr,
		IsEmailVerified: user.IsEmailVerified,
		Status:          string(user.Status),
		SuspendedUntil:  user.SuspendedUntil,
		LockedUntil:     user.LockedUntil,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}
}

func ToUserListResponse(users []model.User) UsersResponse {
	var list UsersResponse
	list.Users = make([]UserResponse, 0, len(users))
	for _, user := range users {
		list.Users = append(list.Users, ToUserResponse(user))
	}
	return list
}

type Handler struct {
	userService *UserService
	limiter     *security.RateLimiter
	secrets     config.Secrets
	logger      log.Logger
}

func NewHandler(
	userService *UserService,
	limiter *security.RateLimiter,
	secrets config.Secrets,
	logger log.Logger,
) *Handler {
	return &Handler{
		userService: userService,
		limiter:     limiter,
		secrets:     secrets,
		logger:      logger,
	}
}

func (h *Handler) GetUserByID(c *gin.Context) {
	var uri UserURIParam
	if err := c.ShouldBindUri(&uri); err != nil {
		validation.ValidationError(c, err)
		return
	}

	user, err := h.userService.GetUserByID(
		c.Request.Context(),
		uri.UserID,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(
		http.StatusOK,
		api.SuccessResponse{
			Data: ToUserResponse(*user),
		},
	)
}

func (h *Handler) GetAllUsers(c *gin.Context) {
	var q pagination.ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		validation.ValidationError(c, err)
		return
	}

	if q.PageSize <= 0 {
		q.PageSize = 10
	}

	if q.Page <= 0 {
		q.Page = 1
	}

	var rolePtr *string
	if r := c.Query("role"); r != "" {
		rolePtr = &r
	}

	var statusPtr *string
	if s := c.Query("status"); s != "" {
		statusPtr = &s
	}

	users, page, err := h.userService.GetAllUsers(
		c.Request.Context(),
		q,
		rolePtr,
		statusPtr,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(
		http.StatusOK,
		api.SuccessResponse{
			Data: ToUserListResponse(users).Users,
			Meta: page,
		},
	)
}

func (h *Handler) DeleteUser(c *gin.Context) {
	var uri UserURIParam
	if err := c.ShouldBindUri(&uri); err != nil {
		validation.ValidationError(c, err)
		return
	}

	err := h.userService.DeleteUserByID(
		c.Request.Context(),
		uri.UserID,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(
		http.StatusOK,
		api.SuccessResponse{
			Message: "User deleted successfully",
		},
	)
}

func (h *Handler) UpdateUserRole(c *gin.Context) {
	var uri UserURIParam
	if err := c.ShouldBindUri(&uri); err != nil {
		validation.ValidationError(c, err)
		return
	}

	var req UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	callerIDVal, exists := c.Get("userID")
	if !exists {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}
	callerID := callerIDVal.(uuid.UUID)

	if err := h.userService.UpdateUserRole(
		c.Request.Context(),
		uri.UserID,
		model.UserRole(req.Role),
		callerID,
	); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{
		Message: "User role updated successfully",
	})
}

func (h *Handler) UpdateAccountStatus(c *gin.Context) {
	var uri UserURIParam
	if err := c.ShouldBindUri(&uri); err != nil {
		validation.ValidationError(c, err)
		return
	}

	var req UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	callerIDVal, exists := c.Get("userID")
	if !exists {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}
	callerID := callerIDVal.(uuid.UUID)

	if err := h.userService.UpdateAccountStatus(
		c.Request.Context(),
		uri.UserID,
		model.AccountStatus(req.Status),
		req.SuspendedUntil,
		callerID,
	); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{
		Message: "User status updated successfully",
	})
}

func (h *Handler) RevokeUserSessions(c *gin.Context) {
	var uri UserURIParam
	if err := c.ShouldBindUri(&uri); err != nil {
		validation.ValidationError(c, err)
		return
	}

	callerIDVal, exists := c.Get("userID")
	if !exists {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}
	callerID := callerIDVal.(uuid.UUID)

	if err := h.userService.RevokeAllUserSessions(
		c.Request.Context(),
		uri.UserID,
		callerID,
	); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{
		Message: "User sessions revoked successfully",
	})
}


