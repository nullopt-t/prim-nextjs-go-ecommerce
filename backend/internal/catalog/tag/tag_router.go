package tag

import (
	"github.com/gin-gonic/gin"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/middleware"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
)

type TagRouter struct {
	th      *TagHandler
	secrets *config.Secrets
}

func NewRouter(h *TagHandler, secrets *config.Secrets) *TagRouter {
	return &TagRouter{
		th:      h,
		secrets: secrets,
	}
}

func (tr *TagRouter) MapRoutes(vgroup *gin.RouterGroup) {
	admin := vgroup.Group("/admin/tags")
	admin.Use(
		middleware.Authenticate(tr.secrets, true),
	)
	{
		admin.GET("", tr.th.AdminListTags)
		admin.GET("/:id", tr.th.GetTagByID)
		admin.POST("", tr.th.CreateTag)
		admin.PATCH("/:id", tr.th.UpdateTagByID)
		admin.DELETE("/:id", tr.th.DeleteTagByID)
	}

	public := vgroup.Group("/tags")
	{
		public.GET("", tr.th.ListTags)
	}
}
