package tag

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
)

type TagRouter struct {
	th           *TagHandler
	sessionStore middleware.SessionStore
}

func NewRouter(h *TagHandler, sessionStore middleware.SessionStore) *TagRouter {
	return &TagRouter{
		th:           h,
		sessionStore: sessionStore,
	}
}

func (tr *TagRouter) MapRoutes(vgroup *gin.RouterGroup) {
	admin := vgroup.Group("/admin/tags")
	admin.Use(
		middleware.Authenticate(tr.sessionStore),
		middleware.Authorize(model.AdminRole),
	)
	{
		admin.GET("", tr.th.AdminListTags)
		admin.GET("/:id", tr.th.GetTagByID)
		admin.POST("", tr.th.CreateTag)
		admin.PATCH("/:id", tr.th.UpdateTagByID)
		admin.DELETE("/:id", tr.th.DeleteTagByID)
	}

	public := vgroup.Group("/tags")
	public.Use(middleware.PublicCache(300))
	{
		public.GET("", tr.th.ListTags)
	}
}
