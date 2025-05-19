package routes

import (
	"github.com/NorskHelsenett/oss-ipam-api/internal/handlers/prefixeshandler"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(server *gin.Engine) {
	server.POST("/prefixes", prefixeshandler.RegisterPrefix)
	server.DELETE("/prefixes", prefixeshandler.DeregisterPrefix)
}
