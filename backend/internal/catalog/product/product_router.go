// Package product provides HTTP routing, handlers, domain services, and database persistence
// for master product records, brand and category relations, publishing lifecycle, and search/filtering.
package product

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
)

// ProductRouter configures and registers HTTP routes for storefront product catalog browsing
// and administrative product lifecycle operations.
type ProductRouter struct {
	ph           *ProductHandler
	sessionStore middleware.SessionStore
}

// NewRouter constructs a new ProductRouter with the provided ProductHandler and session store.
func NewRouter(
	ph *ProductHandler,
	sessionStore middleware.SessionStore,
) *ProductRouter {
	return &ProductRouter{
		ph:           ph,
		sessionStore: sessionStore,
	}
}

// MapRoutes configures and binds public storefront product routes and authenticated administrative
// product catalog management routes onto the provided Gin router group.
func (r *ProductRouter) MapRoutes(vgroup *gin.RouterGroup) {
	// Administrative product catalog management routes
	admin := vgroup.Group("/admin/products")
	admin.Use(
		middleware.Authenticate(r.sessionStore),
		middleware.Authorize(model.AdminRole),
	)
	{
		// Product CRUD
		admin.POST("", r.ph.CreateProductAsDraft)
		admin.GET("", r.ph.AdminGetAllProducts)
		admin.GET("/:id", r.ph.GetProductByID)
		admin.PATCH("/:id", r.ph.UpdateProduct)
		admin.DELETE("/:id", r.ph.SoftDeleteProduct)

		// Lifecycle / Publication status transitions
		admin.POST("/:id/publish", r.ph.PublishProduct)
		admin.POST("/:id/archive", r.ph.ArchiveProduct)

		// Variant default assignment and creation (product-scoped actions)
		admin.POST("/:id/variants", r.ph.CreateProductVariant)
		admin.POST("/:id/variants/default", r.ph.SetDefaultVariant)

		// Media Operations
		admin.POST("/:id/thumbnail", r.ph.UploadProductThumbnail)

		// Taxonomy: Categories & Tags assignments
		admin.PUT("/:id/categories", r.ph.PutProductCategories)
		admin.PUT("/:id/tags", r.ph.PutProductTags)
	}

	// Public storefront catalog routes (cached for 300s)
	public := vgroup.Group("/products")
	public.Use(middleware.PublicCache(300))
	{
		// GET /products - List published products with brand/category/tag filters
		public.GET("", r.ph.GetAllProducts)
		// GET /products/:slug - Retrieve published product details by unique slug
		public.GET("/:slug", r.ph.GetProductBySlug)
	}

	// Uncached product reviews and aggregate ratings
	reviewsGroup := vgroup.Group("/products")
	reviewsGroup.Use(middleware.NoCache())
	{
		// GET /products/:slug/reviews - Paginated approved reviews
		reviewsGroup.GET("/:slug/reviews", r.ph.ListProductReviews)
		// GET /products/:slug/reviews/summary - Rating distribution and average score
		reviewsGroup.GET("/:slug/reviews/summary", r.ph.GetProductRatingSummary)
	}
}

