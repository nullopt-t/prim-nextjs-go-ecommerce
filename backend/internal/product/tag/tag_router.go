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
	tags.POST("", tr.th.CreateTag)
}
