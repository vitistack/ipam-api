package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vitistack/ipam-api/internal/utils"
	"github.com/vitistack/ipam-api/pkg/clients/mongodb"
	"github.com/vitistack/ipam-api/pkg/models/mongodbtypes"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	showServicesAddress string
	showServicesZone    string
	showServicesFormat  string
)

var showServices = &cobra.Command{
	Use:   "show-services",
	Short: "Show address services",
	Run: func(cmd *cobra.Command, args []string) {
		if err := displayServices(showServicesAddress, showServicesZone); err != nil {
			fmt.Println("Error:", err)
		}
	},
}

func init() {
	showServices.Flags().StringVar(&showServicesAddress, "address", "", "Address (required)")
	showServices.Flags().StringVar(&showServicesZone, "zone", "", "Zone (required)")
	showServices.Flags().StringVar(&showServicesFormat, "format", "", "Output format (optional, default text. Use 'json' for JSON output)")
	showServices.MarkFlagRequired("address")
	showServices.MarkFlagRequired("zone")

	RootCmd.AddCommand(showServices)
}

// displayServices retrieves and displays information about a specific address and zone from MongoDB.
// It initializes the MongoDB client using configuration from Viper, queries the database for the
// address and zone, decrypts the stored secret, and prints the result in either JSON or a human-readable
// format. If the address is not found or any error occurs during the process (database, decryption, or
// marshaling), an error is returned.
//
// Parameters:
//   - address: the address to look up in the database.
//   - zone: the zone associated with the address.
//
// Returns:
//   - error: if any operation fails, an error describing the failure is returned; otherwise, nil.
func displayServices(address, zone string) error {
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

	savedAddress.Secret = decryptedSecret

	addressJson, err := json.MarshalIndent(savedAddress, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal address to JSON: %w", err)
	}

	if showServicesFormat == "json" {
		fmt.Println(string(addressJson))
	} else {
		fmt.Println("Address\t\t " + savedAddress.Address)
		fmt.Println("Secret\t\t " + savedAddress.Secret)
		fmt.Println("Zone\t\t " + savedAddress.Zone)
		if len(savedAddress.Services) > 0 {
			for index, service := range savedAddress.Services {
				if index == 0 {
					fmt.Println("Services\t- Service Name:\t\t\t" + service.ServiceName)
				} else {
					fmt.Println("\t\t- Service Name:\t\t\t" + service.ServiceName)
				}
				fmt.Println("\t\t  Namespace ID:\t\t\t" + service.NamespaceId)
				fmt.Println("\t\t  Cluster ID:\t\t\t" + service.ClusterId)
				fmt.Println("\t\t  Retention Period Days:\t" + strconv.Itoa(service.RetentionPeriodDays))
				fmt.Println("\t\t  Deny External Cleanup:\t" + strconv.FormatBool(service.DenyExternalCleanup))
				if service.ExpiresAt != nil {
					fmt.Println("\t\t  Expires:\t\t\t" + service.ExpiresAt.String())
				}
			}
		}

	}

	return nil
}
