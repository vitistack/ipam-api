package mongodbservice

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"
	"github.com/vitistack/ipam-api/internal/logger"
	"github.com/vitistack/ipam-api/internal/responses"
	"github.com/vitistack/ipam-api/internal/utils"
	"github.com/vitistack/ipam-api/pkg/clients/mongodb"
	"github.com/vitistack/ipam-api/pkg/models/apicontracts"
	"github.com/vitistack/ipam-api/pkg/models/mongodbtypes"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// RegisterAddress creates a new address document in the MongoDB collection using the provided
// IpamApiRequest and NetboxPrefix. It encrypts the secret from the request, constructs the
// address document, and inserts it into the database. The function returns the newly created
// mongodbtypes.Address or an error if the operation fails.
//
// Parameters:
//   - request: apicontracts.IpamApiRequest containing the address and service details.
//   - nextPrefix: responses.NetboxPrefix representing the prefix to assign.
//
// Returns:
//   - mongodbtypes.Address: The newly created address document.
//   - error: An error if the operation fails, otherwise nil.
func RegisterAddress(request apicontracts.IpamApiRequest, nextPrefix responses.NetboxPrefix) (mongodbtypes.Address, error) {

	encryptedSecret, err := utils.DeterministicEncrypt(request.Secret)

	if err != nil {
		return mongodbtypes.Address{}, fmt.Errorf("failed to encrypt secret: %w", err)
	}

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

// UpdateAddressDocument updates an address document in the MongoDB collection based on the provided IpamApiRequest.
// It performs the following operations:
//   - Encrypts the provided secret deterministically.
//   - Searches for an address document matching the encrypted secret, zone, address, and IP family.
//   - If a new secret is provided in the request, it validates that only one service is registered and that the service matches,
//     then updates the secret and services array accordingly.
//   - If no new secret is provided, it updates the services array by removing the matching service (if present) and adding the updated service.
//   - Returns an error if the document is not found, if there are mismatches, or if any MongoDB operation fails.
//
// Parameters:
//   - request: apicontracts.IpamApiRequest containing the details for the update operation.
//
// Returns:
//   - error: An error if the update fails or validation does not pass; otherwise, nil.
func UpdateAddressDocument(request apicontracts.IpamApiRequest) (mongodbtypes.Address, error) {
	client := mongodb.GetClient()
	collection := client.Database(viper.GetString("mongodb.database")).Collection(viper.GetString("mongodb.collection"))

	encryptedRequestSecret, err := utils.DeterministicEncrypt(request.Secret)
	encryptedNewSecret, err := utils.DeterministicEncrypt(request.NewSecret)

	if err != nil {
		return mongodbtypes.Address{}, fmt.Errorf("failed to encrypted secret: %w", err)
	}

	filter := bson.M{
		"secret":    encryptedRequestSecret,
		"zone":      request.Zone,
		"address":   request.Address,
		"ip_family": request.IpFamily,
	}

	if request.Address != "" {
		filter["address"] = request.Address
	}

	var registeredAddress mongodbtypes.Address

	err = collection.FindOne(context.Background(), filter).Decode(&registeredAddress)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return mongodbtypes.Address{}, errors.New("no matching address found with the provided secret, zone and address")
		}
		return mongodbtypes.Address{}, fmt.Errorf("failed to read address document: %w", err)
	}

	if request.NewSecret != "" && encryptedRequestSecret != encryptedNewSecret {
		if registeredAddress.Secret == encryptedRequestSecret && len(registeredAddress.Services) > 1 {
			return mongodbtypes.Address{}, errors.New("multiple services registered. unable to change secret")
		}
		if registeredAddress.Secret != encryptedRequestSecret {
			return mongodbtypes.Address{}, errors.New("secret mismatch. unable to change secret")
		}
		encryptedNewSecret, err := utils.DeterministicEncrypt(request.NewSecret)
		if err != nil {
			return mongodbtypes.Address{}, fmt.Errorf("failed to encrypt new secret: %w", err)
		}

		if registeredAddress.Services[0].ServiceName != request.Service.ServiceName ||
			registeredAddress.Services[0].NamespaceId != request.Service.NamespaceId ||
			registeredAddress.Services[0].ClusterId != request.Service.ClusterId {
			return mongodbtypes.Address{}, errors.New("service mismatch. unable to change secret")
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
			return mongodbtypes.Address{}, fmt.Errorf("failed to update address: %w", err)
		}

		return registeredAddress, nil
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
		return mongodbtypes.Address{}, fmt.Errorf("failed to update address: %w", err)
	}

	return registeredAddress, nil
}

// SetServiceExpirationOnAddress sets an expiration date for a specific service associated with an address in MongoDB.
// It performs the following steps:
//  1. Encrypts the provided secret deterministically.
//  2. Finds the address document in MongoDB matching the encrypted secret, zone, address, and IP family.
//  3. Checks if the specified service exists for the address.
//  4. Removes the existing instance of the service from the services array.
//  5. Adds the service back with an updated expiration date based on the retention period.
//  6. Updates the address document in MongoDB with the modified services array.
//
// Parameters:
//   - request: apicontracts.IpamApiRequest containing the secret, zone, address, IP family, and service details.
//
// Returns:
//   - error: An error if the operation fails at any step, or nil if successful.
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
	logger.Log.Infof("Service expiration set for service %s successfully in MongoDB", request.Service.ServiceName)
	return nil
}

// ServiceExists checks if a target Service exists within a slice of Service objects.
// It returns true if there is a Service in the slice that matches the NamespaceId,
// ServiceName, and ClusterId of the target Service; otherwise, it returns false.
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

// ServiceAlreadyRegistered checks if a service, identified by the provided IpamApiRequest,
// is already registered in the MongoDB collection. It encrypts the secret from the request
// and queries the collection for an address document matching the encrypted secret, zone,
// and IP family. If a matching address is found and contains a service with the same
// ServiceName, NamespaceId, and ClusterId as in the request, it returns the registered
// address and a nil error. If no such service is found, it returns an empty Address and nil error.
// Returns an error if encryption, querying, or decoding fails.
func ServiceAlreadyRegistered(request apicontracts.IpamApiRequest) (mongodbtypes.Address, error) {
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
