package utils

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/NorskHelsenett/oss-ipam-api/internal/services/netboxservice"
	"github.com/NorskHelsenett/oss-ipam-api/pkg/clients/mongodb"
	"github.com/NorskHelsenett/oss-ipam-api/pkg/models/mongodbtypes"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func StartCleanupWorker() {
	log.Println("Starting cleanup worker...")
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		CleanupExpiredServices()
		CleanupRegistrationsWithoutServices()
	}
}

func CleanupExpiredServices() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	coll := mongodb.GetClient().
		Database(viper.GetString("mongodb.database")).
		Collection(viper.GetString("mongodb.collection"))

	now := time.Now()

	update := bson.M{
		"$pull": bson.M{
			"services": bson.M{
				"expires_at": bson.M{"$lte": now},
			},
		},
	}

	_, err := coll.UpdateMany(ctx, bson.M{}, update)
	if err != nil {
		log.Printf("could not delete service: %v", err)
		return
	}

}

func CleanupRegistrationsWithoutServices() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	prefixes, err := GetPrefixesWithNoServices()
	if err != nil {
		log.Printf("mongodb query failed: %v", err)
		return
	}

	for _, prefix := range prefixes {
		// Delete the prefix in Netbox
		err := netboxservice.DeleteNetboxPrefix(strconv.Itoa(prefix.NetboxID))
		if err != nil {
			log.Printf("could not delete prefix from netbox: %v", err)
		}

		// Delete from MongoDB
		collection := mongodb.GetClient().Database(viper.GetString("mongodb.database")).Collection(viper.GetString("mongodb.collection"))
		_, err = collection.DeleteOne(ctx, bson.M{"_id": prefix.ID})
		if err != nil {
			log.Printf("could not delete registration from mongodb: %v", err)
		}

	}
}

func GetPrefixesWithNoServices() ([]mongodbtypes.Prefix, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := mongodb.GetClient().Database(viper.GetString("mongodb.database")).Collection(viper.GetString("mongodb.collection"))

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

	var prefixes []mongodbtypes.Prefix
	if err := cursor.All(ctx, &prefixes); err != nil {
		return nil, err
	}

	return prefixes, nil
}
