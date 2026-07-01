package brand

import "github.com/gin-gonic/gin"

type Router struct {
	bh *BrandHandler
}

func NewRouter(
	h *BrandHandler,
) *Router {
	return &Router{
		bh: h,
	}
}

func (r *Router) MapRoutes(
	vgroup *gin.RouterGroup,
) {
	brands := vgroup.Group("/brands")
	brands.POST("", r.bh.CreateBrand)
	brands.GET("", r.bh.ListBrands)
	brands.GET("/:id", r.bh.GetBrandByID)
}
