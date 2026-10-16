package object

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/middleware"
	"github.com/nullopt-t/prim-backend/internal/model"
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
	}
}
