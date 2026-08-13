package review

import (
	"github.com/gin-gonic/gin"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/middleware"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
)

type ReviewRouter struct {
	rh      *ReviewHandler
	secrets *config.Secrets
}

func NewRouter(
	rh *ReviewHandler,
	secrets *config.Secrets,
) *ReviewRouter {
	return &ReviewRouter{
		rh:      rh,
		secrets: secrets,
	}
}

func (r *ReviewRouter) MapRoutes(vgroup *gin.RouterGroup) {
	// User endpoints (Create review)
	user := vgroup.Group("/reviews")
	user.Use(middleware.Authenticate(r.secrets, false))
	{
		user.POST("", r.rh.CreateReview)
	}

	// Admin endpoints
	admin := vgroup.Group("/admin/reviews")
	admin.Use(middleware.Authenticate(r.secrets, true))
	{
		admin.GET("", r.rh.AdminListReviews)
		admin.PATCH("/:id/status", r.rh.UpdateReviewStatus)
	}
}
