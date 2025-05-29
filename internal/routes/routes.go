package routes

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"     // swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware
	docs "github.com/vitistack/ipam-api/docs"
	"github.com/vitistack/ipam-api/internal/handlers/addresseshandler"
)

func SetupRoutes(server *gin.Engine) {

	docs.SwaggerInfo.Title = "oss-ipam-api API"
	docs.SwaggerInfo.Description = "This the oss-ipam-api API server."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:3000"
	docs.SwaggerInfo.BasePath = "/v2"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	server.POST("/", addresseshandler.RegisterAddress)
	server.DELETE("/", addresseshandler.ExpireAddress)

	server.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
