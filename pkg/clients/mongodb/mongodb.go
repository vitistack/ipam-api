package mongodb

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/vitistack/ipam-api/internal/logger"
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

// InitClient initializes and returns a singleton MongoDB client instance using the provided MongoConfig.
// It constructs the MongoDB URI from the configuration, establishes a connection, and pings the database
// to ensure connectivity. If the connection or ping fails, the function logs a fatal error and terminates
// the application.
//
// Parameters:
//   - config: MongoConfig containing the MongoDB connection details.
//
// Returns:
//   - *mongo.Client: A pointer to the initialized MongoDB client instance.
func InitClient(config MongoConfig) *mongo.Client {
	uri := fmt.Sprintf("mongodb://%v:%v@%v:27017/?authSource=admin&readPreference=primary&ssl=false", config.Username, config.Password, config.Host)
	clientOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var err error
		clientInstance, err = mongo.Connect(options.Client().ApplyURI(uri))
		if err != nil {
			logger.Log.Fatalf("MongoDB connection error: %v", err)
		}

		if err := clientInstance.Ping(ctx, nil); err != nil {
			logger.Log.Fatalf("MongoDB ping failed: %v", err)

		}
	})

	return clientInstance
}

// GetClient returns the initialized MongoDB client
func GetClient() *mongo.Client {
	return clientInstance
}
