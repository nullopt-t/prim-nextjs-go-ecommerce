package variant

import (
	"github.com/gin-gonic/gin"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/middleware"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
)

type VariantRouter struct {
	vh      *VariantHandler
	secrets *config.Secrets
}

func NewRouter(vh *VariantHandler, secrets *config.Secrets) *VariantRouter {
	return &VariantRouter{
		vh:      vh,
		secrets: secrets,
	}
}

func (vr *VariantRouter) MapRoutes(vgroup *gin.RouterGroup) {
	// Public routes
	public := vgroup.Group("")
	public.Use(middleware.PublicCache(300))
	{
		public.GET("/variants/:id", vr.vh.GetVariantByID)
		public.GET("/variants/:id/media", vr.vh.ListVariantMedia)
	}

	// Admin routes
	admin := vgroup.Group("/admin")
	admin.Use(
		middleware.Authenticate(vr.secrets),
	)
	{
		admin.PATCH("/variants/:id", vr.vh.UpdateVariantByID)
		admin.DELETE("/variants/:id", vr.vh.DeleteVariantByID)

		admin.POST("/variants/:id/media", vr.vh.AttachMedia)
		admin.DELETE("/variants/:id/media/:media_id", vr.vh.DetachMedia)
		admin.PATCH("/variants/:id/media/reorder", vr.vh.ReorderMedia)
	}
}
