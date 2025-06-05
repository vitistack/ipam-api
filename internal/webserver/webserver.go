package webserver

import (
	"github.com/gin-gonic/gin"
	"github.com/vitistack/ipam-api/internal/middleware"
	"github.com/vitistack/ipam-api/internal/routes"
)

func InitHttpServer() {
	server := gin.Default()

	server.Use(middleware.ZapLogger())
	server.Use(middleware.ZapErrorLogger())

	routes.SetupRoutes(server)

	server.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	err := server.Run(":3000")

	if err != nil {
		panic(err)
	}
}
