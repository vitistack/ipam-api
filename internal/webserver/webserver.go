package webserver

import (
	"github.com/gin-gonic/gin"
	"github.com/vitistack/ipam-api/internal/logger"
	"github.com/vitistack/ipam-api/internal/middleware"
	"github.com/vitistack/ipam-api/internal/routes"
)

func InitHTTPServer() {
	gin.SetMode(gin.ReleaseMode)
	server := gin.New() // or gin.Default()

	server.Use(gin.Recovery())
	server.Use(middleware.ZapLogger())
	server.Use(middleware.ZapErrorLogger())

	routes.SetupRoutes(server)

	logger.Log.Info("Vitistack IPAM API server starting on port 3000")
	err := server.Run(":3000")

	if err != nil {
		panic(err)
	}
}
