package user

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
)

// UserHandler defines the HTTP handler interface for administrative user management endpoints.
type UserHandler interface {
	GetUserByID(c *gin.Context)
	DeleteUser(c *gin.Context)
	GetAllUsers(c *gin.Context)
	UpdateUserRole(c *gin.Context)
	UpdateAccountStatus(c *gin.Context)
	RevokeUserSessions(c *gin.Context)
}

// Router configures routing and administrative authorization hierarchies for /admin/users endpoints.
type Router struct {
	handler      UserHandler
	sessionStore middleware.SessionStore
}

// NewRouter constructs a new user management Router instance.
func NewRouter(handler UserHandler, sessionStore middleware.SessionStore) *Router {
	return &Router{
		handler:      handler,
		sessionStore: sessionStore,
	}
}

// MapRoutes registers administrative user management routes under /admin/users.
// Group-level middleware requires authentication and at least model.AdminRole clearance.
// Sensitive mutations (role changes, status changes, session revocations) require model.SuperRole.
func (r *Router) MapRoutes(vgroup *gin.RouterGroup) {
	adminUsers := vgroup.Group("/admin/users")
	adminUsers.Use(
		middleware.Authenticate(r.sessionStore),
		middleware.Authorize(model.AdminRole),
	)
	{
		// GET /api/v1/admin/users - List paginated users with role/status filters (Admin or Superadmin)
		adminUsers.GET("", r.handler.GetAllUsers)
		// GET /api/v1/admin/users/:id - View user details by UUID (Admin or Superadmin)
		adminUsers.GET("/:id", r.handler.GetUserByID)

		// Account management endpoints: Admin and Superadmin (enforces strict role rank hierarchy in service)
		// DELETE /api/v1/admin/users/:id - Delete account
		adminUsers.DELETE("/:id", middleware.Authorize(model.AdminRole), r.handler.DeleteUser)
		// PATCH /api/v1/admin/users/:id/role - Alter user administrative role tier
		adminUsers.PATCH("/:id/role", middleware.Authorize(model.AdminRole), r.handler.UpdateUserRole)
		// PATCH /api/v1/admin/users/:id/status - Alter account standing (suspend, activate)
		adminUsers.PATCH("/:id/status", middleware.Authorize(model.AdminRole), r.handler.UpdateAccountStatus)
		// POST /api/v1/admin/users/:id/revoke-sessions - Invalidate all active user login sessions
		adminUsers.POST("/:id/revoke-sessions", middleware.Authorize(model.AdminRole), r.handler.RevokeUserSessions)
	}
}
