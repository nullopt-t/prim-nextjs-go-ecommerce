package object

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
)

type Router struct {
	handler      *ObjectHandler
	sessionStore middleware.SessionStore
}

func NewRouter(handler *ObjectHandler, sessionStore middleware.SessionStore) *Router {
	return &Router{
		handler:      handler,
		sessionStore: sessionStore,
	}
}

func (r *Router) MapRoutes(vgroup *gin.RouterGroup) {
	admin := vgroup.Group("/admin/media")
	admin.Use(
		middleware.Authenticate(r.sessionStore),
		middleware.Authorize(model.AdminRole),
	)
	{
		admin.GET("", r.handler.ListMedia)
		admin.POST("", r.handler.UploadMedia)
		// DELETE /admin/media/:id destructively removes stored objects: Superadmin only.
		admin.DELETE("/:id", middleware.Authorize(model.SuperRole), r.handler.DeleteMedia)
	}
}
