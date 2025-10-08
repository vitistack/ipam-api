package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vitistack/ipam-api/internal/utils"
	"github.com/vitistack/ipam-api/pkg/clients/mongodb"
	"github.com/vitistack/ipam-api/pkg/models/apicontracts"
	"github.com/vitistack/ipam-api/pkg/models/mongodbtypes"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	deleteServiceAddress string
	deleteServiceZone    string
	deleteServiceSecret  string
	deleteServiceName    string
	deleteNamespaceID    string
	deleteClusterID      string
)

var deleteService = &cobra.Command{
	Use:   "delete-service",
	Short: "Delete service from address",
	Run: func(cmd *cobra.Command, args []string) {
		expireAddress := apicontracts.IpamAPIRequest{
			Address: deleteServiceAddress,
			Zone:    deleteServiceZone,
			Secret:  deleteServiceSecret,
			Service: apicontracts.Service{
				ServiceName: deleteServiceName,
				NamespaceID: deleteNamespaceID,
				ClusterID:   deleteClusterID,
			},
		}

		err := setServiceExpirationOnAddress(expireAddress)

		if err != nil {
			fmt.Println("Error:", err)
		}
	},
}

func init() {
	deleteService.Flags().StringVar(&deleteServiceAddress, "address", "", "Address (required)")
	deleteService.Flags().StringVar(&deleteServiceZone, "zone", "", "Zone (required)")
	deleteService.Flags().StringVar(&deleteServiceSecret, "secret", "", "Secret (required)")
	deleteService.Flags().StringVar(&deleteServiceName, "service-name", "", "Service Name (required)")
	deleteService.Flags().StringVar(&deleteNamespaceID, "namespace-id", "", "Namespace ID (required)")
	deleteService.Flags().StringVar(&deleteClusterID, "cluster-id", "", "Cluster ID (required)")
	if err := deleteService.MarkFlagRequired("address"); err != nil {
		fmt.Printf("Error marking 'address' flag as required: %v\n", err)
	}
	if err := deleteService.MarkFlagRequired("zone"); err != nil {
		fmt.Printf("Error marking 'zone' flag as required: %v\n", err)
	}
	if err := deleteService.MarkFlagRequired("secret"); err != nil {
		fmt.Printf("Error marking 'secret' flag as required: %v\n", err)
	}
	if err := deleteService.MarkFlagRequired("service-name"); err != nil {
		fmt.Printf("Error marking 'service-name' flag as required: %v\n", err)
	}
	if err := deleteService.MarkFlagRequired("namespace-id"); err != nil {
		fmt.Printf("Error marking 'namespace-id' flag as required: %v\n", err)
	}
	if err := deleteService.MarkFlagRequired("cluster-id"); err != nil {
		fmt.Printf("Error marking 'cluster-id' flag as required: %v\n", err)
	}

	RootCmd.AddCommand(deleteService)
}

// setServiceExpirationOnAddress sets an expiration date for a specific service associated with an address in the MongoDB database.
// It performs the following steps:
//  1. Initializes the MongoDB client using configuration from viper.
//  2. Encrypts the provided secret deterministically.
//  3. Finds the address document matching the encrypted secret, zone, and address.
//  4. Checks if the specified service exists for the address.
//  5. Removes the existing instance of the service from the address's services array.
//  6. Adds the service back with an updated expiration date based on the retention period.
//  7. Updates the address document in MongoDB with the modified services array.
//
// Parameters:
//   - request: apicontracts.IpamApiRequest containing the secret, zone, address, and service details.
//
// Returns:
//   - error: An error if any step fails, or nil on success.
func setServiceExpirationOnAddress(request apicontracts.IpamAPIRequest) error {

	mongoConfig := mongodb.MongoConfig{
		Host:     viper.GetString("mongodb.host"),
		Username: viper.GetString("mongodb.username"),
		Password: viper.GetString("mongodb.password"),
	}
	mongodb.InitClient(mongoConfig)

	client := mongodb.GetClient()
	collection := client.Database(viper.GetString("mongodb.database")).Collection(viper.GetString("mongodb.collection"))

	encryptedSecret, err := utils.DeterministicEncrypt(request.Secret)

	if err != nil {
		return fmt.Errorf("failed to encrypt secret: %w", err)
	}

	// address := utils.NormalizeCIDR(request.Address)

	filter := bson.M{
		"secret":  encryptedSecret,
		"zone":    request.Zone,
		"address": utils.NormalizeCIDR(request.Address),
	}

	var registeredAddress mongodbtypes.Address
	err = collection.FindOne(context.Background(), filter).Decode(&registeredAddress)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("no matching address found with the provided secret, zone and address")
		}
		return fmt.Errorf("failed to read address document: %w", err)
	}

	if !serviceExists(registeredAddress.Services, mongodbtypes.Service(request.Service)) {
		return errors.New("service does not exist for this address")
	}

	// Loop through the services array and remove the service that matches the request
	var newServices []mongodbtypes.Service
	for _, service := range registeredAddress.Services {
		if !(service.NamespaceID == request.Service.NamespaceID &&
			service.ServiceName == request.Service.ServiceName &&
			service.ClusterID == request.Service.ClusterID) {
			newServices = append(newServices, mongodbtypes.Service{
				ServiceName:         service.ServiceName,
				NamespaceID:         service.NamespaceID,
				ClusterID:           service.ClusterID,
				RetentionPeriodDays: service.RetentionPeriodDays,
				ExpiresAt:           service.ExpiresAt,
			})
		}
	}

	// Add the service to be expired with an expiration date
	var expiresAt *time.Time
	exp := time.Now().AddDate(0, 0, request.Service.RetentionPeriodDays)
	expiresAt = &exp
	newServices = append(newServices, mongodbtypes.Service{
		ServiceName:         request.Service.ServiceName,
		NamespaceID:         request.Service.NamespaceID,
		ClusterID:           request.Service.ClusterID,
		RetentionPeriodDays: request.Service.RetentionPeriodDays,
		ExpiresAt:           expiresAt})

	// Update mongodb
	update := bson.M{
		"$set": bson.M{
			"services": newServices,
		},
	}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to update services array: %w", err)
	}
	fmt.Printf("Expiration set for service '%s' cluster id '%s' 'namespace id '%s' on address '%s'\n",
		request.Service.ServiceName, request.Service.ClusterID, request.Service.NamespaceID, request.Address)
	return nil
}

// serviceExists checks if a target Service exists within a slice of Service objects.
// It returns true if a Service with matching NamespaceId, ServiceName, and ClusterId is found,
// otherwise it returns false.
func serviceExists(services []mongodbtypes.Service, target mongodbtypes.Service) bool {
	for _, s := range services {
		if s.NamespaceID == target.NamespaceID &&
			s.ServiceName == target.ServiceName &&
			s.ClusterID == target.ClusterID {
			return true
		}
	}
	return false
}
