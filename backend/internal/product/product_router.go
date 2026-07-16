package product

import "github.com/gin-gonic/gin"

type ProductRouter struct {
	ph *ProductHandler
}

func NewRouter(
	ph *ProductHandler,
) *ProductRouter {
	return &ProductRouter{
		ph: ph,
	}
}

func (r *ProductRouter) MapRoutes(vgroup *gin.RouterGroup) {
	products := vgroup.Group("/products")
	{
		products.GET("", r.ph.GetAllProducts)
		products.GET("/:slug", r.ph.GetProductBySlug)
	}

	admin := vgroup.Group("/admin/products")
	{
		admin.POST("", r.ph.CreateProductAsDraft)
		admin.GET("/:id", r.ph.GetProductByID)
		admin.POST("/:id/variants/default", r.ph.SetDefaultVariant)
		admin.POST("/:id/publish", r.ph.PublishProduct)
		admin.POST("/:id/archive", r.ph.ArchiveProduct)

		admin.GET("/:id/variants", r.ph.GetProductVariants)
		// admin.POST("/:id/variants", r.ph.CreateProductVariant)

		admin.GET("/:id/categories", r.ph.GetProductCategories)
		// admin.PUT("/:id/categories", r.ph.UpdateProductCategories)

		admin.GET("/:id/tags", r.ph.GetProductTags)
		// admin.PUT("/:id/tags", r.ph.UpdateProductTags)
	}
}
