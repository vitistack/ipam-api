package webserver

import (
	"github.com/NorskHelsenett/oss-ipam-api/internal/routes"
	"github.com/gin-gonic/gin"
)

func InitHttpServer() {
	server := gin.Default()
	routes.SetupRoutes(server)
	server.Run(":3000")

}
