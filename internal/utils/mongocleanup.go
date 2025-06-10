package utils

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/spf13/viper"
	"github.com/vitistack/ipam-api/internal/logger"
	"github.com/vitistack/ipam-api/internal/services/netboxservice"
	"github.com/vitistack/ipam-api/pkg/clients/mongodb"
	"github.com/vitistack/ipam-api/pkg/models/mongodbtypes"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func StartCleanupWorker() {
	logger.Log.Info("Starting cleanup worker...")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	client := mongodb.GetClient()
	collection := client.Database(viper.GetString("mongodb.database")).Collection(viper.GetString("mongodb.collection"))

	for range ticker.C {
		logger.Log.Info("cleanup worker running")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		CleanupExpiredServices(ctx, collection)
		CleanupRegistrationsWithoutServices(ctx, collection)
		cancel()
	}
}

func CleanupExpiredServices(ctx context.Context, collection *mongo.Collection) {
	update := bson.M{
		"$pull": bson.M{
			"services": bson.M{
				"expires_at": bson.M{"$lte": time.Now()},
			},
		},
	}

	_, err := collection.UpdateMany(ctx, bson.M{}, update)
	if err != nil {
		log.Printf("could not delete service: %v", err)
		return
	}

}

func CleanupRegistrationsWithoutServices(ctx context.Context, collection *mongo.Collection) {
	registrations, err := GetPrefixesWithNoServices(ctx, collection)
	if err != nil {
		log.Printf("mongodb query failed: %v", err)
		return
	}

	for _, prefix := range registrations {
		// Delete the prefix in Netbox
		err := netboxservice.DeleteNetboxPrefix(strconv.Itoa(prefix.NetboxID))

		if err != nil {
			log.Printf("could not delete prefix from netbox: %v", err)
		}
		logger.Log.Infof("Released prefix %s from Netbox", prefix.Address)
		// Delete from MongoDB
		_, err = collection.DeleteOne(ctx, bson.M{"_id": prefix.ID})
		if err != nil {
			log.Printf("could not delete registration from mongodb: %v", err)
		}
		logger.Log.Infof("Released prefix %s from mongodb", prefix.Address)

	}
}

func GetPrefixesWithNoServices(ctx context.Context, collection *mongo.Collection) ([]mongodbtypes.Address, error) {
	// Create filter for finding registrations with no services
	filter := bson.M{
		"$or": []bson.M{
			{"services": bson.M{"$exists": false}},
			{"services": bson.A{}},
		},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var prefixes []mongodbtypes.Address
	if err := cursor.All(ctx, &prefixes); err != nil {
		return nil, err
	}

	return prefixes, nil
}
