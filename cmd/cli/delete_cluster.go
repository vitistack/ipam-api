package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vitistack/ipam-api/pkg/clients/mongodb"
	"github.com/vitistack/ipam-api/pkg/models/mongodbtypes"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	clusterID string
)

var deleteClusterCmd = &cobra.Command{
	Use:   "delete-cluster",
	Short: "Set expiresAt == time.Now() for services linked to a cluster id",
	Run: func(cmd *cobra.Command, args []string) {
		if err := setExpiresForCluster(clusterID); err != nil {
			fmt.Println("Error:", err)
		}
	},
}

func init() {
	deleteClusterCmd.Flags().StringVar(&clusterID, "cluster", "", "Cluster ID (required)")
	deleteClusterCmd.MarkFlagRequired("cluster")
	RootCmd.AddCommand(deleteClusterCmd)
}

// setExpiresForCluster sets the expiration date for all services associated with the given clusterId
// in the MongoDB collection. It finds all address documents containing services with the specified
// clusterId, updates the ExpiresAt field to the current time, and sets RetentionPeriodDays to 0 for
// those services. Returns an error if no addresses are found, if there are issues querying or updating
// the database, or if decoding fails.
//
// Parameters:
//   - clusterId: The ID of the cluster whose services' expiration should be set.
//
// Returns:
//   - error: An error if the operation fails, or nil on success.
func setExpiresForCluster(clusterId string) error {
	// Initialize MongoDB client
	mongoConfig := mongodb.MongoConfig{
		Host:     viper.GetString("mongodb.host"),
		Username: viper.GetString("mongodb.username"),
		Password: viper.GetString("mongodb.password"),
	}

	mongodb.InitClient(mongoConfig)

	client := mongodb.GetClient()
	collection := client.Database(viper.GetString("mongodb.database")).Collection(viper.GetString("mongodb.collection"))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"services.cluster_id": clusterId,
	}

	var addresses []mongodbtypes.Address
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to find addresses: %w", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var addr mongodbtypes.Address
		if err := cursor.Decode(&addr); err != nil {
			return fmt.Errorf("failed to decode address: %w", err)
		}
		addresses = append(addresses, addr)
	}
	if err := cursor.Err(); err != nil {
		return fmt.Errorf("cursor error: %w", err)
	}

	if len(addresses) == 0 {
		return fmt.Errorf("no address found for cluster_id=%s", clusterId)
	}

	for _, a := range addresses {
		newServices := []mongodbtypes.Service{}
		for _, service := range a.Services {
			if service.ClusterId == clusterId {
				exp := time.Now()
				service.ExpiresAt = &exp
				service.RetentionPeriodDays = 0
			}
			newServices = append(newServices, service)
		}

		update := bson.M{
			"$set": bson.M{
				"services": newServices,
			},
		}
		filter := bson.M{"_id": a.ID}
		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			return fmt.Errorf("failed to update services array: %w", err)
		}
	}
	fmt.Println("Expiration set for addresses with cluster ID:", clusterId)
	return nil
}
