package variant

import "github.com/gin-gonic/gin"

type VariantRouter struct {
	vh *VariantHandler
}

func NewRouter(vh *VariantHandler) *VariantRouter {
	return &VariantRouter{
		vh: vh,
	}
}

func (vr *VariantRouter) MapRoutes(vgroup *gin.RouterGroup) {
	// Public routes
	public := vgroup.Group("")
	{
		public.GET("/variants/:id", vr.vh.GetVariantByID)
		public.GET("/variants/:id/media", vr.vh.ListVariantMedia)
	}

	// Admin routes
	admin := vgroup.Group("/admin")
	{
		admin.PATCH("/variants/:id", vr.vh.UpdateVariantByID)
		admin.DELETE("/variants/:id", vr.vh.DeleteVariantByID)

		admin.POST("/variants/:id/media", vr.vh.AttachMedia)
		admin.DELETE("/variants/:id/media/:media_id", vr.vh.DetachMedia)
		admin.PATCH("/variants/:id/media/reorder", vr.vh.ReorderMedia)
	}
}
