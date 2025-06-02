package mongodb

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	clientInstance *mongo.Client
	clientOnce     sync.Once
)

type MongoConfig struct {
	Host     string `json:"host"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// InitClient connects to MongoDB and initializes the shared client
func InitClient(config MongoConfig) *mongo.Client {
	uri := fmt.Sprintf("mongodb://%v:%v@%v:27017/?authSource=admin&readPreference=primary&ssl=false", config.Username, config.Password, config.Host)
	clientOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var err error
		clientInstance, err = mongo.Connect(options.Client().ApplyURI(uri))
		if err != nil {
			log.Fatalf("MongoDB connection error: %v", err)
		}

		if err := clientInstance.Ping(ctx, nil); err != nil {
			log.Fatalf("MongoDB ping failed: %v", err)

		}
	})

	return clientInstance
}

// GetClient returns the initialized MongoDB client
func GetClient() *mongo.Client {
	return clientInstance
}
