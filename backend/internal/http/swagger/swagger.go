// Package swagger configures and exposes HTTP endpoints for API documentation
// generated via Swagger / OpenAPI annotations in the codebase.
package swagger

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/middleware"
	"github.com/nullopt-t/prim-backend/internal/model"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetUpDocs binds Swagger UI documentation handlers to the provided Gin router group.
// It establishes separate documentation entry points for customer-facing public APIs
// and privileged administrative/operational APIs secured with auth middleware.
//
// Query Parameters & Routing:
// - /docs/public/*any: Serves UI assets and the swagger.json spec for customer storefront APIs.
// - /docs/admin/*any: Serves UI assets and the swagger.json spec for backoffice administration APIs.
func SetUpDocs(rg *gin.RouterGroup, sessionStore ...middleware.SessionStore) {
	// HTTP redirect from bare "/docs/public" to "/docs/public/index.html"
	rg.GET("/docs/public", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/api/v1/docs/public/index.html")
	})

	// Public Customer API Documentation
	rg.GET("/docs/public/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.InstanceName("public"),
	))

	// Private Admin / Ops API Documentation
	adminDocs := rg.Group("/docs/admin")
	if len(sessionStore) > 0 && sessionStore[0] != nil {
		adminDocs.Use(
			middleware.Authenticate(sessionStore[0]),
			middleware.Authorize(model.AdminRole),
		)
	}

	adminDocs.GET("", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/api/v1/docs/admin/index.html")
	})
	adminDocs.GET("/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.InstanceName("admin"),
	))
}
