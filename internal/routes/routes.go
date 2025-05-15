package routes

import (
	"github.com/NorskHelsenett/oss-ipam-api/internal/controllers/prefixescontroller"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func SetupRoutes(server *gin.Engine) {

	server.GET("/prefixes", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": viper.GetString("mongodb.username"),
		})
	})

	server.POST("/prefixes", prefixescontroller.RegisterPrefix)
}
