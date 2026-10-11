package review

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/middleware"
)

type ReviewRouter struct {
	rh           *ReviewHandler
	sessionStore middleware.SessionStore
}

func NewRouter(
	rh *ReviewHandler,
	sessionStore middleware.SessionStore,
) *ReviewRouter {
	return &ReviewRouter{
		rh:           rh,
		sessionStore: sessionStore,
	}
}

func (r *ReviewRouter) MapRoutes(vgroup *gin.RouterGroup) {
	// User endpoints (Authenticated customer)
	user := vgroup.Group("/reviews")
	user.Use(middleware.Authenticate(r.sessionStore))
	{
		user.POST("", r.rh.CreateReview)
		user.GET("/me", r.rh.GetMyReviews)
		user.GET("/eligibility", r.rh.CheckEligibility)
		user.GET("/:id", r.rh.GetReviewByID)
		user.PATCH("/:id", r.rh.UpdateReview)
		user.DELETE("/:id", r.rh.DeleteReview)
	}

	// Admin endpoints (Moderation)
	admin := vgroup.Group("/admin/reviews")
	admin.Use(middleware.Authenticate(r.sessionStore))
	{
		admin.GET("", r.rh.AdminListReviews)
		admin.GET("/:id", r.rh.AdminGetReviewByID)
		admin.PATCH("/:id/status", r.rh.UpdateReviewStatus)
		admin.DELETE("/:id", r.rh.AdminDeleteReview)
	}
}


