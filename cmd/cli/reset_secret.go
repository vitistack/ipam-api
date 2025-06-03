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
	resetAddress string
	resetZone    string
	newSecret    string
)

var resetSecret = &cobra.Command{
	Use:   "reset-secret",
	Short: "Reset address secret",
	Run: func(cmd *cobra.Command, args []string) {
		if resetAddress == "" {
			fmt.Println("Missing --address argument")
			return
		}
		if resetZone == "" {
			fmt.Println("Missing --zone argument")
			return
		}
		if newSecret == "" {
			fmt.Println("Missing --new argument")
			return
		}
		if err := resetSecretForAddress(resetAddress, resetZone, newSecret); err != nil {
			fmt.Println("Error:", err)
		}
	},
}

func init() {
	resetSecret.Flags().StringVar(&resetAddress, "address", "", "Address (required)")
	resetSecret.Flags().StringVar(&resetZone, "zone", "", "Zone (required)")
	resetSecret.Flags().StringVar(&newSecret, "new", "", "Zone (required)")
	resetSecret.MarkFlagRequired("address")
	resetSecret.MarkFlagRequired("zone")
	resetSecret.MarkFlagRequired("new")
	RootCmd.AddCommand(resetSecret)
}

func resetSecretForAddress(address, zone, newSecret string) error {
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

	findOneFilter := bson.M{
		"address": address,
		"zone":    zone,
	}

	var savedAddress mongodbtypes.Address
	err := collection.FindOne(ctx, findOneFilter).Decode(&savedAddress)

	if err != nil {
		return fmt.Errorf("failed to find addresses: %w", err)
	}

	encryptedSecret, err := utils.DeterministicEncrypt(newSecret)

	if err != nil {
		return fmt.Errorf("failed to encrypt secret: %w", err)
	}

	update := bson.M{
		"$set": bson.M{
			"secret": encryptedSecret,
		},
	}

	updateOneFilter := bson.M{"_id": savedAddress.ID}

	_, err = collection.UpdateOne(context.Background(), updateOneFilter, update)

	if err != nil {
		return fmt.Errorf("failed to update secret: %w", err)
	}

	fmt.Println("Secret updated for address '" + address + "' in zone '" + zone)
	return nil
}
