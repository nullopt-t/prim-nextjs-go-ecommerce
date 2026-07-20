package brand

import (
	"github.com/gin-gonic/gin"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
)

type Router struct {
	bh      *BrandHandler
	secrets *config.Secrets
}

func NewRouter(
	h *BrandHandler,
	secrets *config.Secrets,
) *Router {
	return &Router{
		bh:      h,
		secrets: secrets,
	}
}

func (r *Router) MapRoutes(
	vgroup *gin.RouterGroup,
) {
	brands := vgroup.Group("/brands")
	brands.GET("", r.bh.ListBrands)
	brands.GET("/slug/:slug", r.bh.GetBrandBySlug)

	admin := vgroup.Group("/admin/brands")
	// admin.Use(
	// 	middleware.Authanticate(r.secrets),
	// 	middleware.Authorize(model.AdminRole),
	// )

	admin.GET("", r.bh.ListAdminBrands)
	admin.GET("/:id", r.bh.GetBrandByID)
	admin.POST("", r.bh.CreateBrand)
	admin.PATCH("/:id", r.bh.UpdateBrand)
}
