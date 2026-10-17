package audit

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
)

// Router registers the superadmin-only audit trail endpoint.
type Router struct {
	handler      *Handler
	sessionStore middleware.SessionStore
}

// NewRouter constructs an audit router.
func NewRouter(handler *Handler, sessionStore middleware.SessionStore) *Router {
	return &Router{handler: handler, sessionStore: sessionStore}
}

// MapRoutes binds /admin/audit behind authentication and SuperRole RBAC.
func (r *Router) MapRoutes(vgroup *gin.RouterGroup) {
	admin := vgroup.Group("/admin/audit")
	admin.Use(
		middleware.Authenticate(r.sessionStore),
		middleware.Authorize(model.SuperRole),
	)
	{
		admin.GET("", r.handler.ListAuditLogs)
	}
}
