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
	admin := vgroup.Group("/admin/products")
	{
		admin.POST("", r.ph.CreateProductAsDraft)
		admin.GET("", r.ph.AdminGetAllProducts)
		admin.GET("/:id", r.ph.GetProductByID)
		admin.PATCH("/:id", r.ph.UpdateProduct)
		admin.DELETE("/:id", r.ph.SoftDeleteProduct)

		// Lifecycle / Publication
		admin.POST("/:id/publish", r.ph.PublishProduct)
		admin.POST("/:id/archive", r.ph.ArchiveProduct)

		// Variants Management
		admin.POST("/:id/variants", r.ph.CreateProductVariant)
		admin.POST("/:id/variants/default", r.ph.SetDefaultVariant)

		// Media Operations
		admin.POST("/:id/media", r.ph.UploadProductMedia)
		admin.DELETE("/:id/media/:media_id", r.ph.DetachMedia)
		admin.PATCH("/:id/media/reorder", r.ph.ReorderMedia)

		// Categories & Tags Linkage
		admin.PUT("/:id/categories", r.ph.PutProductCategories)
		admin.PUT("/:id/tags", r.ph.PutProductTags)
	}
}
