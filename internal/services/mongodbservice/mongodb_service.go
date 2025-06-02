package mongodbservice

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"
	"github.com/vitistack/ipam-api/internal/responses"
	"github.com/vitistack/ipam-api/pkg/clients/mongodb"
	"github.com/vitistack/ipam-api/pkg/models/apicontracts"
	"github.com/vitistack/ipam-api/pkg/models/mongodbtypes"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func RegisterAddress(request apicontracts.IpamApiRequest, nextPrefix responses.NetboxPrefix) (mongodbtypes.Address, error) {
	service := apicontracts.Service{
		ServiceName:         request.Service.ServiceName,
		NamespaceId:         request.Service.NamespaceId,
		ClusterId:           request.Service.ClusterId,
		RetentionPeriodDays: request.Service.RetentionPeriodDays,
		ExpiresAt:           nil,
	}

	newAddressDocument := bson.M{
		"secret":    request.Secret,
		"zone":      request.Zone,
		"address":   nextPrefix.Prefix,
		"netbox_id": nextPrefix.ID,
		"services":  []apicontracts.Service{service},
	}

	client := mongodb.GetClient()
	collection := client.Database(viper.GetString("mongodb.database")).Collection(viper.GetString("mongodb.collection"))

	result, err := collection.InsertOne(context.Background(), newAddressDocument)
	if err != nil {
		return mongodbtypes.Address{}, errors.New("failed to save address: " + err.Error())
	}

	var address mongodbtypes.Address
	err = collection.FindOne(context.Background(), bson.M{"_id": result.InsertedID.(bson.ObjectID)}).Decode(&address)

	if err != nil {
		return mongodbtypes.Address{}, err
	}

	return address, nil

}

func UpdateAddressDocument(request apicontracts.IpamApiRequest) error {
	client := mongodb.GetClient()
	collection := client.Database(viper.GetString("mongodb.database")).Collection(viper.GetString("mongodb.collection"))

	filter := bson.M{
		"secret":  request.Secret,
		"zone":    request.Zone,
		"address": request.Address,
	}

	var registeredAddress mongodbtypes.Address
	err := collection.FindOne(context.Background(), filter).Decode(&registeredAddress)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("address not found: %w", err)
		}
		return fmt.Errorf("failed to read address document: %w", err)
	}

	if request.NewSecret != "" {
		if registeredAddress.Secret == request.Secret && len(registeredAddress.Services) > 1 {
			return fmt.Errorf("multiple services registered. unable to change secret")
		} else if registeredAddress.Secret != request.Secret {
			return errors.New("secret mismatch. unable to change secret")
		} else {
			update := bson.M{
				"$set": bson.M{
					"secret": request.NewSecret,
				},
			}

			_, err = collection.UpdateOne(context.Background(), filter, update)
			if err != nil {
				return fmt.Errorf("failed to update secret: %w", err)
			}
		}

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

	// Add the service to be updated
	newServices = append(newServices, mongodbtypes.Service(apicontracts.Service{
		ServiceName:         request.Service.ServiceName,
		NamespaceId:         request.Service.NamespaceId,
		ClusterId:           request.Service.ClusterId,
		RetentionPeriodDays: request.Service.RetentionPeriodDays,
		ExpiresAt:           nil,
	}))

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

	return nil
}

func SetServiceExpirationOnAddress(request apicontracts.IpamApiRequest) error {
	client := mongodb.GetClient()
	collection := client.Database(viper.GetString("mongodb.database")).Collection(viper.GetString("mongodb.collection"))

	filter := bson.M{
		"secret":  request.Secret,
		"zone":    request.Zone,
		"address": request.Address,
	}

	var registeredAddress mongodbtypes.Address
	err := collection.FindOne(context.Background(), filter).Decode(&registeredAddress)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("address not found: %w", err)
		}
		return fmt.Errorf("failed to read address document: %w", err)
	}

	if !ServiceExists(registeredAddress.Services, mongodbtypes.Service(request.Service)) {
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
	exp := time.Now().Add(time.Minute * time.Duration(request.Service.RetentionPeriodDays)) //! Rembember to change from minutes to days .AddDate(0, 0, retention) for dager
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

	return nil
}

func ServiceExists(services []mongodbtypes.Service, target mongodbtypes.Service) bool {
	for _, s := range services {
		if s.NamespaceId == target.NamespaceId &&
			s.ServiceName == target.ServiceName &&
			s.ClusterId == target.ClusterId {
			return true
		}
	}
	return false
}
