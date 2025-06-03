package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/viper"
	"github.com/vitistack/ipam-api/cmd/ipam-api/settings"

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

	// Initialize MongoDB client
	mongoConfig := mongodb.MongoConfig{
		Host:     viper.GetString("mongodb.host"),
		Username: viper.GetString("mongodb.username"),
		Password: viper.GetString("mongodb.password"),
	}

	mongodb.InitClient(mongoConfig) // check if running before starting webserver

	err = netboxservice.Cache.FetchPrefixContainers()

	if err != nil {
		log.Fatalf("%s", err.Error())
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
	log.Printf("Received signal: %s. IPAM-API shutting down...", sig)
	serverCancel()

}
