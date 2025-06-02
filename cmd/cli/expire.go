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
	days      int
)

var expireCmd = &cobra.Command{
	Use:   "expire-addresses",
	Short: "Set expires_at on addresses linked to a cluster ID",
	Run: func(cmd *cobra.Command, args []string) {
		if clusterID == "" {
			fmt.Println("Missing --clusterId argument")
			return
		}
		if err := setExpiresForCluster(clusterID); err != nil {
			fmt.Println("Error:", err)
		}
	},
}

func init() {
	expireCmd.Flags().StringVar(&clusterID, "clusterId", "", "Cluster ID (required)")
	// expireCmd.Flags().IntVar(&days, "days", 30, "Days until expiration")
	expireCmd.MarkFlagRequired("clusterId")
	RootCmd.AddCommand(expireCmd)
}

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

	// json, _ := json.MarshalIndent(addresses, "", "  ")

	// fmt.Println(string(json))

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
