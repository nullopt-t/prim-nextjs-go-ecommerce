package category

import "github.com/gin-gonic/gin"

type CategoryRouter struct {
	chandler *CategoryHandler
}

func NewRouter(
	h *CategoryHandler,
) *CategoryRouter {
	return &CategoryRouter{
		chandler: h,
	}
}

func (cr *CategoryRouter) MapRoutes(
	vgroup *gin.RouterGroup,
) {
	categories := vgroup.Group("/categories")
	categories.GET("", cr.chandler.ListCategories)
	categories.GET("/:id", cr.chandler.GetCategoryByID)
	categories.GET("/slug/:slug", cr.chandler.GetCategoryBySlug)
	categories.POST("", cr.chandler.CreateCategory)
}
