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
	"github.com/nullopt-t/prim-backend/pkg/log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UsersResponse wraps a slice of UserResponse objects for batch administrative user listings.
type UsersResponse struct {
	Users []UserResponse `json:"users"`
}

// UserByEmailRequest represents the request payload when querying or referencing a user by their primary email.
type UserByEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// UserURIParam captures and validates UUID route parameters from path bindings.
type UserURIParam struct {
	UserID uuid.UUID `uri:"id" json:"id" binding:"required"`
}

// UpdateUserRoleRequest represents the request body for modifying a user's administrative role tier.
type UpdateUserRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

// UpdateUserStatusRequest represents the payload for changing account standing (active, suspended, deleted).
type UpdateUserStatusRequest struct {
	Status         string     `json:"status" binding:"required"`
	SuspendedUntil *time.Time `json:"suspendedUntil,omitempty"`
}

// UserResponse represents the public administrative view of a user account entity,
// including administrative roles, email verification status, account status, and timestamps.
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

// ToUserResponse converts an internal User model entity to the public UserResponse DTO,
// defaulting a nil role pointer to the standard "customer" role.
func ToUserResponse(user model.User) UserResponse {
	roleStr := "customer"
	if user.Role != "" {
		roleStr = string(user.Role)
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

// ToUserListResponse converts a slice of domain User entities into the UsersResponse container.
func ToUserListResponse(users []model.User) UsersResponse {
	var list UsersResponse
	list.Users = make([]UserResponse, 0, len(users))
	for _, user := range users {
		list.Users = append(list.Users, ToUserResponse(user))
	}
	return list
}

// Handler orchestrates HTTP transport operations for administrative user management:
// querying users, listing paginated accounts, deleting users, altering roles, updating status,
// and revoking user sessions.
type Handler struct {
	userService *UserService
	limiter     *security.RateLimiter
	logger      log.Logger
}

// NewHandler constructs a new administrative user Handler.
func NewHandler(
	userService *UserService,
	limiter *security.RateLimiter,
	logger log.Logger,
) *Handler {
	return &Handler{
		userService: userService,
		limiter:     limiter,
		logger:      logger,
	}
}

// GetUserByID retrieves full profile and status details for a user by their UUID.
func (h *Handler) GetUserByID(c *gin.Context) {
	// Parse target user UUID from path
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid user ID"))
		return
	}

	// Fetch user record from service
	user, err := h.userService.GetUserByID(
		c.Request.Context(),
		userID,
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

// GetAllUsers retrieves a paginated, filterable list of user accounts for staff administration.
func (h *Handler) GetAllUsers(c *gin.Context) {
	// Bind pagination query parameters
	var q pagination.ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		validation.ValidationError(c, err)
		return
	}

	// Apply pagination bounds: cap page size and compute an overflow-safe offset
	q.Process(pagination.QueryOptions{DefaultPageSize: 10, MaxPageSize: 100})

	// Optional administrative filters
	var rolePtr *string
	if r := c.Query("role"); r != "" {
		rolePtr = &r
	}

	var statusPtr *string
	if s := c.Query("status"); s != "" {
		statusPtr = &s
	}

	// Query database via user service
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

// DeleteUser deletes an account by its UUID.
func (h *Handler) DeleteUser(c *gin.Context) {
	// Parse target user UUID
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid user ID"))
		return
	}

	// Extract caller's authenticated user ID
	callerID, err := getUserIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	if userID == callerID {
		_ = c.Error(apierr.BadRequestError("Superadmins cannot delete their own account").WithCode("CANNOT_DELETE_SELF"))
		return
	}

	err = h.userService.DeleteUserByID(
		c.Request.Context(),
		userID,
		callerID,
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

// UpdateUserRole modifies the administrative role of a target user (Superadmin only).
func (h *Handler) UpdateUserRole(c *gin.Context) {
	// Parse target user UUID
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid user ID"))
		return
	}

	// Parse JSON request body
	var req UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	// Extract caller's authenticated user ID
	callerID, err := getUserIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Update user role through service (prevents modifying own role)
	if err := h.userService.UpdateUserRole(
		c.Request.Context(),
		userID,
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

// UpdateAccountStatus modifies a user's account standing (active, suspended, deleted) (Superadmin only).
func (h *Handler) UpdateAccountStatus(c *gin.Context) {
	// Parse target user UUID
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid user ID"))
		return
	}

	// Parse JSON request body
	var req UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	// Extract caller's user UUID from context
	callerID, err := getUserIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Update account status in database
	if err := h.userService.UpdateAccountStatus(
		c.Request.Context(),
		userID,
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

// RevokeUserSessions forcefully revokes all active Redis sessions for the specified user (Superadmin only).
func (h *Handler) RevokeUserSessions(c *gin.Context) {
	// Parse target user UUID
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid user ID"))
		return
	}

	// Extract caller's authenticated user ID
	callerID, err := getUserIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Invalidate all sessions in Redis
	if err := h.userService.RevokeAllUserSessions(
		c.Request.Context(),
		userID,
		callerID,
	); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{
		Message: "User sessions revoked successfully",
	})
}




