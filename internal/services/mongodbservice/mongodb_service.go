package mongodbservice

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"
	"github.com/vitistack/ipam-api/internal/responses"
	"github.com/vitistack/ipam-api/internal/utils"
	"github.com/vitistack/ipam-api/pkg/clients/mongodb"
	"github.com/vitistack/ipam-api/pkg/models/apicontracts"
	"github.com/vitistack/ipam-api/pkg/models/mongodbtypes"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func RegisterAddress(request apicontracts.IpamApiRequest, nextPrefix responses.NetboxPrefix) (mongodbtypes.Address, error) {

	encryptedSecret, err := utils.DeterministicEncrypt(request.Secret)

	if err != nil {
		return mongodbtypes.Address{}, fmt.Errorf("failed to encrypt secret: %w", err)
	}

	// denyCleanup := request.Service.DenyExternalCleanup

	service := apicontracts.Service{
		ServiceName:         request.Service.ServiceName,
		NamespaceId:         request.Service.NamespaceId,
		ClusterId:           request.Service.ClusterId,
		RetentionPeriodDays: request.Service.RetentionPeriodDays,
		DenyExternalCleanup: request.Service.DenyExternalCleanup,
		ExpiresAt:           nil,
	}

	newAddressDocument := bson.M{
		"secret":    encryptedSecret,
		"zone":      request.Zone,
		"address":   nextPrefix.Prefix,
		"ip_family": request.IpFamily,
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

	encryptedRequestSecret, err := utils.DeterministicEncrypt(request.Secret)

	if err != nil {
		return fmt.Errorf("failed to encrypted secret: %w", err)
	}

	filter := bson.M{
		"secret":    encryptedRequestSecret,
		"zone":      request.Zone,
		"address":   request.Address,
		"ip_family": request.IpFamily,
	}

	var registeredAddress mongodbtypes.Address
	err = collection.FindOne(context.Background(), filter).Decode(&registeredAddress)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("no matching address found with the provided secret, zone and address")
		}
		return fmt.Errorf("failed to read address document: %w", err)
	}

	if request.NewSecret != "" {
		if registeredAddress.Secret == encryptedRequestSecret && len(registeredAddress.Services) > 1 {
			return errors.New("multiple services registered. unable to change secret")
		}
		if registeredAddress.Secret != encryptedRequestSecret {
			return errors.New("secret mismatch. unable to change secret")
		}
		encryptedNewSecret, err := utils.DeterministicEncrypt(request.NewSecret)
		if err != nil {
			return fmt.Errorf("failed to encrypt new secret: %w", err)
		}

		if registeredAddress.Services[0].ServiceName != request.Service.ServiceName ||
			registeredAddress.Services[0].NamespaceId != request.Service.NamespaceId ||
			registeredAddress.Services[0].ClusterId != request.Service.ClusterId {
			return errors.New("service mismatch. unable to change secret")
		}

		var currentServices []mongodbtypes.Service
		currentServices = append(currentServices, mongodbtypes.Service{
			ServiceName:         request.Service.ServiceName,
			NamespaceId:         request.Service.NamespaceId,
			ClusterId:           request.Service.ClusterId,
			RetentionPeriodDays: request.Service.RetentionPeriodDays,
			DenyExternalCleanup: request.Service.DenyExternalCleanup})

		update := bson.M{
			"$set": bson.M{
				"secret":   encryptedNewSecret,
				"services": currentServices,
			},
		}

		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			return fmt.Errorf("failed to update secret: %w", err)
		}
		return nil
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

	encryptedSecret, err := utils.DeterministicEncrypt(request.Secret)

	if err != nil {
		return fmt.Errorf("failed to encrypt secret: %w", err)
	}

	filter := bson.M{
		"secret":    encryptedSecret,
		"zone":      request.Zone,
		"address":   request.Address,
		"ip_family": request.IpFamily,
	}

	var registeredAddress mongodbtypes.Address
	err = collection.FindOne(context.Background(), filter).Decode(&registeredAddress)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("no matching address found with the provivded secret, zone and address")
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
			newServices = append(newServices, service)
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
		DenyExternalCleanup: request.Service.DenyExternalCleanup,
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

func AddressAlreadyRegistered(request apicontracts.IpamApiRequest) (mongodbtypes.Address, error) {
	client := mongodb.GetClient()
	collection := client.Database(viper.GetString("mongodb.database")).Collection(viper.GetString("mongodb.collection"))

	encryptedSecret, err := utils.DeterministicEncrypt(request.Secret)
	if err != nil {
		return mongodbtypes.Address{}, fmt.Errorf("failed to encrypt secret: %w", err)
	}

	filter := bson.M{
		"secret":    encryptedSecret,
		"zone":      request.Zone,
		"ip_family": request.IpFamily,
	}

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return mongodbtypes.Address{}, fmt.Errorf("failed to query address documents: %w", err)
	}

	for cursor.Next(context.Background()) {
		var registeredAddress mongodbtypes.Address
		if err := cursor.Decode(&registeredAddress); err != nil {
			return mongodbtypes.Address{}, fmt.Errorf("failed to decode address document: %w", err)
		}
		for _, service := range registeredAddress.Services {
			if service.ServiceName == request.Service.ServiceName &&
				service.NamespaceId == request.Service.NamespaceId &&
				service.ClusterId == request.Service.ClusterId {
				return registeredAddress, nil
			}
		}
	}

	if err := cursor.Err(); err != nil {
		return mongodbtypes.Address{}, fmt.Errorf("cursor error: %w", err)
	}

	return mongodbtypes.Address{}, nil
}
