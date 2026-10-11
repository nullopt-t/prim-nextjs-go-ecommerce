package brand

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/middleware"
)

type Router struct {
	bh           *BrandHandler
	sessionStore middleware.SessionStore
}

func NewRouter(
	h *BrandHandler,
	sessionStore middleware.SessionStore,
) *Router {
	return &Router{
		bh:           h,
		sessionStore: sessionStore,
	}
}

func (r *Router) MapRoutes(
	vgroup *gin.RouterGroup,
) {
	brands := vgroup.Group("/brands")
	brands.Use(middleware.PublicCache(300))
	{
		brands.GET("", r.bh.ListBrands)
	}

	admin := vgroup.Group("/admin/brands")
	admin.Use(
		middleware.Authenticate(r.sessionStore),
	)
	{
		admin.POST("", r.bh.CreateBrand)
		admin.GET("", r.bh.ListAdminBrands)
		admin.PUT("/:id/logo", r.bh.UploadBrandLogo)

		admin.GET("/:id", r.bh.GetBrandByID)
		admin.PATCH("/:id", r.bh.UpdateBrand)
		admin.DELETE("/:id", r.bh.DeleteBrandByID)
	}
}
