package webserver

import (
	"github.com/gin-gonic/gin"
	"github.com/vitistack/ipam-api/internal/routes"
)

func InitHttpServer() {
	server := gin.Default()
	routes.SetupRoutes(server)
	err := server.Run(":3000")

	if err != nil {
		panic(err)
	}
}
