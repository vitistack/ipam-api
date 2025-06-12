package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/viper"
	"github.com/vitistack/ipam-api/cmd/ipam-api/settings"

	"github.com/vitistack/ipam-api/internal/logger"
	"github.com/vitistack/ipam-api/internal/services/netboxservice"
	"github.com/vitistack/ipam-api/internal/utils"
	"github.com/vitistack/ipam-api/internal/webserver"
	"github.com/vitistack/ipam-api/pkg/clients/mongodb"
)

func main() {
	// read config.json file
	err := settings.InitConfig()

	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	if err := logger.InitLogger("./logs"); err != nil {
		log.Fatalf("logger init failed: %v", err)
	}

	defer logger.Sync()

	// Initialize MongoDB client
	mongoConfig := mongodb.MongoConfig{
		Host:     viper.GetString("mongodb.host"),
		Username: viper.GetString("mongodb.username"),
		Password: viper.GetString("mongodb.password"),
	}

	mongodb.InitClient(mongoConfig) // check if running before starting webserver

	logger.Log.Info("Waiting for Netbox to become available...")
	if err := netboxservice.WaitForNetbox(); err != nil {
		logger.Log.Fatalf("Netbox is not available: %v", err)
	}

	logger.Log.Info("Netbox is available. Caching prefix containers...")
	err = netboxservice.Cache.FetchPrefixContainers()

	if err != nil {
		logger.Log.Fatalf("Failed to fetch prefix containers: %v", err)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start web server in a goroutine
	_, serverCancel := context.WithCancel(context.Background())
	go func() {
		webserver.InitHttpServer()
	}()

	// Start cleanup worker
	go func() {
		utils.StartCleanupWorker()
	}()

	// Wait for termination signal
	sig := <-sigChan
	logger.Log.Infof("Received signal: %s. IPAM-API shutting down...", sig)
	serverCancel()

}
