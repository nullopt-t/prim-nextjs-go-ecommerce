package user

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/middleware"
	"github.com/nullopt-t/prim-backend/internal/model"
)

type UserHandler interface {
	GetUserByID(c *gin.Context)
	DeleteUser(c *gin.Context)
	GetAllUsers(c *gin.Context)
	UpdateUserRole(c *gin.Context)
	UpdateAccountStatus(c *gin.Context)
	RevokeUserSessions(c *gin.Context)
}

type Router struct {
	handler      UserHandler
	sessionStore middleware.SessionStore
}

func NewRouter(handler UserHandler, sessionStore middleware.SessionStore) *Router {
	return &Router{
		handler:      handler,
		sessionStore: sessionStore,
	}
}

func (r *Router) MapRoutes(vgroup *gin.RouterGroup) {
	adminUsers := vgroup.Group("/admin/users")
	adminUsers.Use(
		middleware.Authenticate(r.sessionStore),
		middleware.Authorize(model.AdminRole),
	)
	{
		adminUsers.GET("", r.handler.GetAllUsers)
		adminUsers.GET("/:id", r.handler.GetUserByID)
		adminUsers.DELETE("/:id", r.handler.DeleteUser)

		// Superadmin-only mutation endpoints
		adminUsers.PATCH("/:id/role", middleware.Authorize(model.SuperRole), r.handler.UpdateUserRole)
		adminUsers.PATCH("/:id/status", middleware.Authorize(model.SuperRole), r.handler.UpdateAccountStatus)
		adminUsers.POST("/:id/revoke-sessions", middleware.Authorize(model.SuperRole), r.handler.RevokeUserSessions)
	}
}
