package routes

import (
	"github.com/NorskHelsenett/oss-ipam-api/internal/controllers/prefixescontroller"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(server *gin.Engine) {

	server.GET("/prefixes", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "Yay! You are using the OSS IPAM API"})
	})

	server.POST("/prefixes", prefixescontroller.RegisterPrefix)
	server.DELETE("/prefixes", prefixescontroller.DeregisterPrefix)
}
