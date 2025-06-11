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
	deleteNamespaceId    string
	deleteClusterId      string
)

var deleteService = &cobra.Command{
	Use:   "delete-service",
	Short: "Delete service from address",
	Run: func(cmd *cobra.Command, args []string) {
		expireAddress := apicontracts.IpamApiRequest{
			Address: deleteServiceAddress,
			Zone:    deleteServiceZone,
			Secret:  deleteServiceSecret,
			Service: apicontracts.Service{
				ServiceName: deleteServiceName,
				NamespaceId: deleteNamespaceId,
				ClusterId:   deleteClusterId,
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
	deleteService.Flags().StringVar(&deleteNamespaceId, "namespace-id", "", "Namespace ID (required)")
	deleteService.Flags().StringVar(&deleteClusterId, "cluster-id", "", "Cluster ID (required)")
	deleteService.MarkFlagRequired("address")
	deleteService.MarkFlagRequired("zone")
	deleteService.MarkFlagRequired("secret")
	deleteService.MarkFlagRequired("service-name")
	deleteService.MarkFlagRequired("namespace-id")
	deleteService.MarkFlagRequired("cluster-id")

	RootCmd.AddCommand(deleteService)
}

func setServiceExpirationOnAddress(request apicontracts.IpamApiRequest) error {

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

	filter := bson.M{
		"secret":  encryptedSecret,
		"zone":    request.Zone,
		"address": request.Address,
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
		if !(service.NamespaceId == request.Service.NamespaceId &&
			service.ServiceName == request.Service.ServiceName &&
			service.ClusterId == request.Service.ClusterId) {
			newServices = append(newServices, mongodbtypes.Service{
				ServiceName:         service.ServiceName,
				NamespaceId:         service.NamespaceId,
				ClusterId:           service.ClusterId,
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
		NamespaceId:         request.Service.NamespaceId,
		ClusterId:           request.Service.ClusterId,
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
		request.Service.ServiceName, request.Service.ClusterId, request.Service.NamespaceId, request.Address)
	return nil
}

func serviceExists(services []mongodbtypes.Service, target mongodbtypes.Service) bool {
	for _, s := range services {
		if s.NamespaceId == target.NamespaceId &&
			s.ServiceName == target.ServiceName &&
			s.ClusterId == target.ClusterId {
			return true
		}
	}
	return false
}
