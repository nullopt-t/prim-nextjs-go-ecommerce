package product

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/middleware"
)

type ProductRouter struct {
	ph           *ProductHandler
	sessionStore middleware.SessionStore
}

func NewRouter(
	ph *ProductHandler,
	sessionStore middleware.SessionStore,
) *ProductRouter {
	return &ProductRouter{
		ph:           ph,
		sessionStore: sessionStore,
	}
}

func (r *ProductRouter) MapRoutes(vgroup *gin.RouterGroup) {
	admin := vgroup.Group("/admin/products")
	admin.Use(
		middleware.Authenticate(r.sessionStore),
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
		admin.POST("/:id/thumbnail", r.ph.UploadProductThumbnail)
	
		// Tags & Category
		admin.PUT("/:id/categories", r.ph.PutProductCategories)
		admin.PUT("/:id/tags", r.ph.PutProductTags)
	}

	public := vgroup.Group("/products")
	public.Use(middleware.PublicCache(300))
	{
		public.GET("", r.ph.GetAllProducts)
		public.GET("/:slug", r.ph.GetProductBySlug)
	}

	reviewsGroup := vgroup.Group("/products")
	reviewsGroup.Use(middleware.NoCache())
	{
		reviewsGroup.GET("/:slug/reviews", r.ph.ListProductReviews)
		reviewsGroup.GET("/:slug/reviews/summary", r.ph.GetProductRatingSummary)
	}
}
