package review

import (
	"github.com/gin-gonic/gin"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/middleware"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
)

type ReviewRouter struct {
	rh      *ReviewHandler
	secrets config.Secrets
}

func NewRouter(
	rh *ReviewHandler,
	secrets config.Secrets,
) *ReviewRouter {
	return &ReviewRouter{
		rh:      rh,
		secrets: secrets,
	}
}

func (r *ReviewRouter) MapRoutes(vgroup *gin.RouterGroup) {
	// User endpoints (Authenticated customer)
	user := vgroup.Group("/reviews")
	user.Use(middleware.Authenticate(r.secrets, false))
	{
		user.POST("", r.rh.CreateReview)
		user.GET("/me", r.rh.GetMyReviews)
		user.GET("/:id", r.rh.GetReviewByID)
		user.PATCH("/:id", r.rh.UpdateReview)
		user.DELETE("/:id", r.rh.DeleteReview)
	}

	// Admin endpoints (Moderation)
	admin := vgroup.Group("/admin/reviews")
	admin.Use(middleware.Authenticate(r.secrets, false))
	{
		admin.GET("", r.rh.AdminListReviews)
		admin.GET("/:id", r.rh.AdminGetReviewByID)
		admin.PATCH("/:id/status", r.rh.UpdateReviewStatus)
		admin.DELETE("/:id", r.rh.AdminDeleteReview)
	}
}
