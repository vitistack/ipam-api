package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/NorskHelsenett/oss-ipam-api/cmd/oss-ipam-api/settings"
	"github.com/NorskHelsenett/oss-ipam-api/internal/webserver"
	"github.com/NorskHelsenett/oss-ipam-api/pkg/clients/mongodb"
	"github.com/spf13/viper"
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

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start web server in a goroutine
	_, serverCancel := context.WithCancel(context.Background())
	go func() {
		webserver.InitHttpServer()
	}()

	// Wait for termination signal
	sig := <-sigChan
	log.Printf("Received signal: %s. Shutting down...", sig)
	serverCancel()

}
