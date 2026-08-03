package tag

import "github.com/gin-gonic/gin"

type TagRouter struct {
	th *TagHandler
}

func NewRouter(h *TagHandler) *TagRouter {
	return &TagRouter{th: h}
}

func (tr *TagRouter) MapRoutes(vgroup *gin.RouterGroup) {
	admin := vgroup.Group("/admin/tags")
	{
		admin.GET("", tr.th.AdminListTags)
		admin.GET("/:id", tr.th.GetTagByID)
		admin.POST("", tr.th.CreateTag)
		admin.PATCH("/:id", tr.th.UpdateTagByID)
		admin.DELETE("/:id", tr.th.DeleteTagByID)
	}
}
