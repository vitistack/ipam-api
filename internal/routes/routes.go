package routes

import (
	docs "github.com/NorskHelsenett/oss-ipam-api/docs"
	"github.com/NorskHelsenett/oss-ipam-api/internal/handlers/prefixeshandler"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"     // swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware
)

func SetupRoutes(server *gin.Engine) {

	docs.SwaggerInfo.Title = "oss-ipam-api API"
	docs.SwaggerInfo.Description = "This the oss-ipam-api API server."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:3000"
	docs.SwaggerInfo.BasePath = "/v2"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	server.POST("/prefixes", prefixeshandler.RegisterPrefix)
	server.DELETE("/prefixes", prefixeshandler.DeregisterPrefix)

	server.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
