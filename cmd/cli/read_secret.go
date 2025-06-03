package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vitistack/ipam-api/internal/utils"
	"github.com/vitistack/ipam-api/pkg/clients/mongodb"
	"github.com/vitistack/ipam-api/pkg/models/mongodbtypes"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	address string
	zone    string
)

var readSecret = &cobra.Command{
	Use:   "read-secret",
	Short: "Read address secret",
	Run: func(cmd *cobra.Command, args []string) {
		if address == "" {
			fmt.Println("Missing --address argument")
			return
		}
		if zone == "" {
			fmt.Println("Missing --zone argument")
			return
		}
		if err := showSecret(address, zone); err != nil {
			fmt.Println("Error:", err)
		}
	},
}

func init() {
	readSecret.Flags().StringVar(&address, "address", "", "Address (required)")
	readSecret.Flags().StringVar(&zone, "zone", "", "Zone (required)")
	readSecret.MarkFlagRequired("address")
	readSecret.MarkFlagRequired("zone")
	RootCmd.AddCommand(readSecret)
}

func showSecret(address, zone string) error {
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
		"address": address,
		"zone":    zone,
	}

	var savedAddress mongodbtypes.Address
	err := collection.FindOne(ctx, filter).Decode(&savedAddress)

	if err != nil {
		return fmt.Errorf("failed to find addresses: %w", err)
	}

	decryptedSecret, err := utils.DeterministicDecrypt(savedAddress.Secret)

	if err != nil {
		return fmt.Errorf("failed to decrypt secret: %w", err)
	}

	fmt.Println("Secret for address '" + address + "' in zone '" + zone + "': " + decryptedSecret)
	return nil
}
