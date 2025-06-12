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

var showSecret = &cobra.Command{
	Use:   "show-secret",
	Short: "Show address secret",
	Run: func(cmd *cobra.Command, args []string) {
		if err := displaySecret(address, zone); err != nil {
			fmt.Println("Error:", err)
		}
	},
}

func init() {
	showSecret.Flags().StringVar(&address, "address", "", "Address (required)")
	showSecret.Flags().StringVar(&zone, "zone", "", "Zone (required)")
	showSecret.MarkFlagRequired("address")
	showSecret.MarkFlagRequired("zone")
	RootCmd.AddCommand(showSecret)
}

// displaySecret retrieves and displays the decrypted secret associated with a given address and zone
// from a MongoDB collection. It initializes the MongoDB client using configuration values from Viper,
// queries the collection for the specified address and zone, decrypts the stored secret, and prints it.
// Returns an error if the address is not found or if decryption fails.
//
// Parameters:
//   - address: The address to look up in the database.
//   - zone: The zone associated with the address.
//
// Returns:
//   - error: An error if the address is not found, decryption fails, or any database operation fails.
func displaySecret(address, zone string) error {
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
