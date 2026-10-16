// Package variant provides routing, HTTP handlers, domain services, and database persistence
// for product variants (SKUs), variant-specific media assets, and variant stock integration.
package variant

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/middleware"
)

// VariantRouter registers HTTP routes for storefront variant lookups and administrative
// variant lifecycle management (creation, update, deletion, media management, and stock adjustments).
type VariantRouter struct {
	vh           *VariantHandler
	sessionStore middleware.SessionStore
}

// NewRouter constructs a new VariantRouter with the provided VariantHandler and session store.
func NewRouter(vh *VariantHandler, sessionStore middleware.SessionStore) *VariantRouter {
	return &VariantRouter{
		vh:           vh,
		sessionStore: sessionStore,
	}
}

// MapRoutes mounts public and administrative variant endpoints onto the provided Gin router group.
// Public routes provide cached read-only access to variants and inventory availability.
// Admin routes require session authentication and administrator role privileges.
func (vr *VariantRouter) MapRoutes(vgroup *gin.RouterGroup) {
	// Public routes for storefront customer browsing (cached for 300s)
	public := vgroup.Group("")
	public.Use(middleware.PublicCache(300))
	{
		// GET /variants/:id - Retrieve variant by UUID
		public.GET("/variants/:id", vr.vh.GetVariantByID)
		// GET /variants/sku/:sku - Retrieve variant by unique SKU string
		public.GET("/variants/sku/:sku", vr.vh.GetVariantBySKU)
		// GET /variants/:id/media - List media gallery attachments for a variant
		public.GET("/variants/:id/media", vr.vh.ListVariantMedia)
	}

	// Public inventory availability endpoint (cached for 60s)
	publicInventory := vgroup.Group("/variants/:id/inventory")
	publicInventory.Use(middleware.PublicCache(60))
	{
		// GET /variants/:id/inventory - Customer-facing available stock check
		publicInventory.GET("", vr.vh.GetVariantStockPublic)
	}

	// Admin routes requiring authentication and AdminRole authorization
	admin := vgroup.Group("/admin")
	admin.Use(
		middleware.Authenticate(vr.sessionStore),
		middleware.Authorize(model.AdminRole),
	)
	{
		// Variant CRUD lifecycle
		admin.PATCH("/variants/:id", vr.vh.UpdateVariantByID)
		admin.DELETE("/variants/:id", vr.vh.DeleteVariantByID)
		admin.POST("/variants/:id/restore", vr.vh.RestoreVariantByID)

		// Media attachment management
		admin.POST("/variants/:id/media", vr.vh.AttachMedia)
		admin.DELETE("/variants/:id/media/:media_id", vr.vh.DetachMedia)
		admin.PATCH("/variants/:id/media/reorder", vr.vh.ReorderMedia)

		// Inventory stock management and audit ledger
		admin.POST("/variants/:id/inventory/adjust", vr.vh.AdjustStock)
		admin.GET("/variants/:id/inventory", vr.vh.GetVariantStockAdmin)
		admin.GET("/variants/:id/inventory/ledgers", vr.vh.ListVariantLedgers)
	}
}

