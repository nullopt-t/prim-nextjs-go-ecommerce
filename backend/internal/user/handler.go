package user

import (
	"net/http"
	"time"

	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/validation"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/security"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"

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
	UserID uuid.UUID `json:"id" validate:"required,uuid"`
}

type UserResponse struct {
	ID             uuid.UUID  `json:"id,omitempty"`
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

func (h *Handler) GetUserByID(c *gin.Context) {
	var uri UserURIParam
	if err := c.ShouldBindJSON(&uri); err != nil {
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
			Data: UserResponse{
				ID:          user.ID,
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
		validation.ValidationError(c, err)
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
		_ = c.Error(err)
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
