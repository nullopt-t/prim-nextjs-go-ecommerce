package category

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/middleware"
)

type CategoryRouter struct {
	chandler     *CategoryHandler
	sessionStore middleware.SessionStore
}

func NewRouter(
	h *CategoryHandler,
	sessionStore middleware.SessionStore,
) *CategoryRouter {
	return &CategoryRouter{
		chandler:     h,
		sessionStore: sessionStore,
	}
}

func (cr *CategoryRouter) MapRoutes(
	vgroup *gin.RouterGroup,
) {
	categories := vgroup.Group("/categories")
	categories.Use(middleware.PublicCache(300))
	{
		categories.GET("", cr.chandler.ListCategories)
	}

	admin := vgroup.Group("/admin/categories")
	admin.Use(
		middleware.Authenticate(cr.sessionStore),
	)
	{
		admin.GET("", cr.chandler.ListAdminCategories)
		admin.POST("", cr.chandler.CreateCategory)
		admin.PATCH("/:id", cr.chandler.UpdateCategory)
		admin.GET("/:id", cr.chandler.GetCategoryByID)
	}
}
