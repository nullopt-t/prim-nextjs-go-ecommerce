package category

import (
	"github.com/gin-gonic/gin"
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
	{
		categories.GET("", cr.chandler.ListCategories)
	}

	admin := vgroup.Group("/admin/categories")
	{
		admin.GET("", cr.chandler.ListAdminCategories)
		admin.POST("", cr.chandler.CreateCategory)
		admin.PATCH("/:id", cr.chandler.UpdateCategory)
		admin.GET("/:id", cr.chandler.GetCategoryByID)
	}
}
