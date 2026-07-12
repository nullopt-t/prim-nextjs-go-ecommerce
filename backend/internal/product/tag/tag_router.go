package tag

import "github.com/gin-gonic/gin"

type TagRouter struct {
	th *TagHandler
}

func NewRouter(h *TagHandler) *TagRouter {
	return &TagRouter{th: h}
}

func (tr *TagRouter) MapRoutes(vgroup *gin.RouterGroup) {
	tags := vgroup.Group("/tags")
	tags.GET("", tr.th.ListTags)
	tags.GET("/:id", tr.th.GetTagByID)
	tags.POST("", tr.th.CreateTag)
}
