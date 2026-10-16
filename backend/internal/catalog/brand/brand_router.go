// Package brand provides catalog domain services, HTTP handlers, routing,
// and database persistence for managing product manufacturers and brands.
package brand

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/middleware"
)

// Router orchestrates HTTP routing for public storefront and administrative
// brand operations, binding paths to BrandHandler receiver methods.
type Router struct {
	bh           *BrandHandler
	sessionStore middleware.SessionStore
}

// NewRouter constructs a new Router instance with injected handler and session store dependencies.
func NewRouter(
	h *BrandHandler,
	sessionStore middleware.SessionStore,
) *Router {
	return &Router{
		bh:           h,
		sessionStore: sessionStore,
	}
}

// MapRoutes configures route groups for public customer browsing and protected backoffice administration.
//
// Architectural Intentions & Route Specifications:
// 1. Public Storefront Routes (/brands):
//   - GET /brands: Retrieves active product brands for catalog filtering and storefront display.
//   - Configured with HTTP cache headers (max-age=300s) via middleware.PublicCache.
//
// 2. Backoffice Administration Routes (/admin/brands):
//   - POST /admin/brands: Creates a new brand entity (unique name required).
//   - GET /admin/brands: Lists all brands including soft-deleted entries.
//   - GET /admin/brands/:id: Retrieves a single brand by its UUID identifier.
//   - PATCH /admin/brands/:id: Applies partial updates to brand attributes (e.g. name, external link).
//   - PUT /admin/brands/:id/logo: Uploads a brand logo image to object storage and links it.
//   - DELETE /admin/brands/:id: Soft-deletes a brand if not currently in use by active products.
func (r *Router) MapRoutes(
	vgroup *gin.RouterGroup,
) {
	// Public customer-facing brand endpoints with HTTP caching
	brands := vgroup.Group("/brands")
	brands.Use(middleware.PublicCache(300))
	{
		brands.GET("", r.bh.ListBrands)
	}

	// Backoffice administrative brand management endpoints
	admin := vgroup.Group("/admin/brands")
	admin.Use(
		middleware.Authenticate(r.sessionStore),
		middleware.Authorize(model.AdminRole),
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
