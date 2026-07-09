package category

import (
	"github.com/gin-gonic/gin"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/middleware"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
)

type CategoryRouter struct {
	chandler *CategoryHandler
	secrets  *config.Secrets
}

func NewRouter(
	h *CategoryHandler,
	secrets *config.Secrets,
) *CategoryRouter {
	return &CategoryRouter{
		chandler: h,
		secrets:  secrets,
	}
}

func (cr *CategoryRouter) MapRoutes(
	vgroup *gin.RouterGroup,
) {
	categories := vgroup.Group("/categories")
	categories.GET("", cr.chandler.ListCategories)
	categories.GET("/:id", cr.chandler.GetCategoryByID)

	admin := vgroup.Group("/admin/categories")
	admin.Use(
		middleware.Authanticate(cr.secrets),
		middleware.Authorize(model.AdminRole),
	)
	admin.GET("", cr.chandler.ListAdminCategories)
	admin.POST("", cr.chandler.CreateCategory)
	admin.GET("/:id", cr.chandler.GetCategoryByID)
	admin.PATCH("/:id", cr.chandler.UpdateCategory)
}
