package review

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/middleware"
	"github.com/nullopt-t/prim-backend/internal/model"
)

// ReviewRouter configures the HTTP routes and middleware for customer review interactions
// and administrator review moderation workflows.
type ReviewRouter struct {
	rh           *ReviewHandler
	sessionStore middleware.SessionStore
}

// NewRouter constructs a ReviewRouter with its handler and sessionStore dependencies.
func NewRouter(
	rh *ReviewHandler,
	sessionStore middleware.SessionStore,
) *ReviewRouter {
	return &ReviewRouter{
		rh:           rh,
		sessionStore: sessionStore,
	}
}

// MapRoutes registers customer and administrative review routes on the provided Gin router group.
func (r *ReviewRouter) MapRoutes(vgroup *gin.RouterGroup) {
	// Public review listing and detail routes (allow public storefront display of approved reviews)
	public := vgroup.Group("/reviews")
	public.Use(middleware.OptionalAuthenticate(r.sessionStore))
	{
		public.GET("", r.rh.ListPublicReviews)
		public.GET("/:id", r.rh.GetReviewByID)
	}

	// User endpoints (Authenticated customer)
	user := vgroup.Group("/reviews")
	user.Use(middleware.Authenticate(r.sessionStore))
	{
		user.POST("", r.rh.CreateReview)
		user.GET("/me", r.rh.GetMyReviews)
		user.GET("/eligibility", r.rh.CheckEligibility)
		user.PATCH("/:id", r.rh.UpdateReview)
		user.DELETE("/:id", r.rh.DeleteReview)
	}

	// Admin endpoints (Moderation) - protected with Administrator role check
	admin := vgroup.Group("/admin/reviews")
	admin.Use(middleware.Authenticate(r.sessionStore), middleware.Authorize(model.AdminRole))
	{
		admin.GET("", r.rh.AdminListReviews)
		admin.GET("/:id", r.rh.AdminGetReviewByID)
		admin.PATCH("/:id/status", r.rh.UpdateReviewStatus)
		admin.DELETE("/:id", r.rh.AdminDeleteReview)
	}
}


