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
	private := vgroup.Group("/admin/variants")
	{
		private.GET("/:id/media", vr.vh.GetVariantMedia)
	}
}
