package system

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
)

// Router registers the administrative system health endpoint.
type Router struct {
	handler      *Handler
	sessionStore middleware.SessionStore
}

// NewRouter constructs a system router.
func NewRouter(h *Handler, sessionStore middleware.SessionStore) *Router {
	if h == nil {
		panic("system: handler is required")
	}
	if sessionStore == nil {
		panic("system: session store is required")
	}
	return &Router{
		handler:      h,
		sessionStore: sessionStore,
	}
}

// MapRoutes binds /admin/system routes behind authentication and AdminRole RBAC.
func (r *Router) MapRoutes(vgroup *gin.RouterGroup) {
	admin := vgroup.Group("/admin/system")
	admin.Use(
		middleware.Authenticate(r.sessionStore),
		middleware.Authorize(model.AdminRole, model.SuperRole),
	)
	{
		admin.GET("/health", r.handler.GetHealth)
	}
}
