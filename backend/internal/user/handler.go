package user

import (
	"errors"
	"net/http"
	"time"

	"github.com/m-mahmoud-alsaid/prim-backend/internal/middleware"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/security"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type UsersResponse struct {
	Users []UserResponse `json:"users"`
}

type UserByEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type UserURIParam struct {
	UserID uuid.UUID `json:"id" validate:"required,uuid"`
}

type UserResponse struct {
	ID             uuid.UUID  `json:"id,omitempty"`
	Role           string     `json:"role,omitempty"`
	Email          string     `json:"email,omitempty"`
	Phone          *string    `json:"phone,omitempty"`
	Status         string     `json:"status,omitempty"`
	SuspendedUntil *time.Time `json:"suspended_until,omitempty"`
	LockedUntil    *time.Time `json:"locked_unitl,omitzero"`
	CreatedAt      time.Time  `json:"created_at,omitzero"`
	UpdatedAt      time.Time  `json:"updated_at,omitzero"`
}

func ToUserResponse(user model.User) UserResponse {
	return UserResponse{
		ID:             user.ID,
		Role:           string(user.Role),
		Email:          user.Email,
		Phone:          user.Phone,
		Status:         string(user.Status),
		SuspendedUntil: user.SuspendedUntil,
		LockedUntil:    user.LockedUntil,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
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
	secrets     *config.Secrets
	logger      log.Logger
}

func NewHandler(
	userService *UserService,
	limiter *security.RateLimiter,
	secrets *config.Secrets,
	logger log.Logger,
) *Handler {
	return &Handler{
		userService: userService,
		limiter:     limiter,
		secrets:     secrets,
		logger:      logger,
	}
}

func (h *Handler) handleValidationError(c *gin.Context, err error) {
	if ve, ok := errors.AsType[validator.ValidationErrors](err); ok && ve != nil {
		fieldErrors := make([]api.FieldError, 0, len(ve))
		for _, e := range ve {
			fieldErrors = append(fieldErrors, api.FieldError{
				Field: e.Field(),
				Tags:  e.Tag(),
			})
		}
		c.Error(security.NewSecureError(
			http.StatusBadRequest,
			security.CodeValidation,
			"bad request data",
			err,
		).WithFields(fieldErrors))
		return
	}
	c.Error(
		security.NewSecureError(
			http.StatusBadRequest,
			security.CodeValidation,
			"bad request data",
			err,
		),
	)
}

func (h *Handler) GetUserByID(c *gin.Context) {
	var uri UserURIParam
	if err := c.ShouldBindJSON(&uri); err != nil {
		h.handleValidationError(c, err)
		return
	}

	user, err := h.userService.GetUserByID(
		c.Request.Context(),
		uri.UserID,
	)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(
		http.StatusOK,
		api.SuccessResponse{
			Data: UserResponse{
				ID:          user.ID,
				Email:       user.Email,
				LockedUntil: user.LockedUntil,
				CreatedAt:   user.CreatedAt,
				UpdatedAt:   user.UpdatedAt,
			},
		},
	)
}

func (h *Handler) GetAllUsers(c *gin.Context) {
	var q api.PageQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		h.handleValidationError(c, err)
		return
	}

	if q.PageSize == 0 {
		q.PageSize = 5
	}

	if q.Page == 0 {
		q.Page = 1
	}

	users, page, err := h.userService.GetAllUsers(
		c.Request.Context(),
		q,
	)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(
		http.StatusOK,
		api.SuccessResponse{
			Data: ToUserListResponse(users),
			Meta: page,
		},
	)
}

func (h *Handler) DeleteUser(c *gin.Context) {
	var uri UserURIParam
	if err := c.ShouldBindUri(&uri); err != nil {
		h.handleValidationError(c, err)
		return
	}

	err := h.userService.DeleteUserByID(
		c.Request.Context(),
		uri.UserID,
	)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(
		http.StatusOK,
		api.SuccessResponse{
			Message: "User deleted successfully",
		},
	)
}

func (h *Handler) GetMe(c *gin.Context) {
	claims, err := middleware.ClaimsWithContext(c)
	if err != nil {
		c.Error(err)
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		c.Error(err)
		return
	}

	user, err := h.userService.GetUserByID(
		c.Request.Context(),
		userID,
	)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(
		http.StatusOK,
		api.SuccessResponse{
			Data: UserResponse{
				ID:             user.ID,
				Role:           string(user.Role),
				Email:          user.Email,
				SuspendedUntil: user.SuspendedUntil,
				LockedUntil:    user.LockedUntil,
				CreatedAt:      user.CreatedAt,
				UpdatedAt:      user.UpdatedAt,
			},
		},
	)
}
