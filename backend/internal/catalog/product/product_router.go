package product

import (
	"github.com/gin-gonic/gin"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/middleware"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
)

type ProductRouter struct {
	ph      *ProductHandler
	secrets *config.Secrets
}

func NewRouter(
	ph *ProductHandler,
	secrets *config.Secrets,
) *ProductRouter {
	return &ProductRouter{
		ph:      ph,
		secrets: secrets,
	}
}

func (r *ProductRouter) MapRoutes(vgroup *gin.RouterGroup) {
	admin := vgroup.Group("/admin/products")
	admin.Use(
		middleware.Authenticate(r.secrets, true),
	)
	{
		admin.POST("", r.ph.CreateProductAsDraft)
		admin.GET("", r.ph.AdminGetAllProducts)
		admin.GET("/:id", r.ph.GetProductByID)
		admin.PATCH("/:id", r.ph.UpdateProduct)
		admin.DELETE("/:id", r.ph.SoftDeleteProduct)

		// Lifecycle / Publication
		admin.POST("/:id/publish", r.ph.PublishProduct)
		admin.POST("/:id/archive", r.ph.ArchiveProduct)

		// Variant default assignment (product-scoped action)
		admin.POST("/:id/variants", r.ph.CreateProductVariant)
		admin.POST("/:id/variants/default", r.ph.SetDefaultVariant)

		// Media Operations
		admin.POST("/:id/media", r.ph.UploadProductMedia)
		admin.DELETE("/:id/media/:media_id", r.ph.DetachMedia)
		admin.PATCH("/:id/media/reorder", r.ph.ReorderMedia)

		// Tags & Category
		admin.PUT("/:id/categories", r.ph.PutProductCategories)
		admin.PUT("/:id/tags", r.ph.PutProductTags)
	}

	public := vgroup.Group("/products")
	public.Use(middleware.PublicCache(300))
	{
		public.GET("", r.ph.GetAllProducts)
		public.GET("/:id", r.ph.GetProductByPublicID)
	}
}
