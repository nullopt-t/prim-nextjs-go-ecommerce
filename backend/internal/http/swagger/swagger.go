package swagger

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetUpDocs(rg *gin.RouterGroup) {
	// Public Customer API Documentation
	rg.GET("/docs/public/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.InstanceName("public"),
	))

	// Private Admin / Ops API Documentation
	rg.GET("/docs/admin/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.InstanceName("admin"),
	))
}
