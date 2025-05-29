package mongodbservice

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/viper"
	"github.com/vitistack/ipam-api/internal/responses"
	"github.com/vitistack/ipam-api/internal/services/netboxservice"
	"github.com/vitistack/ipam-api/pkg/clients/mongodb"
	"github.com/vitistack/ipam-api/pkg/models/apicontracts"
	"github.com/vitistack/ipam-api/pkg/models/mongodbtypes"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func InsertNewPrefixDocument(request apicontracts.K8sRequestBody, nextPrefix responses.NetboxPrefix) (mongodbtypes.Prefix, error) {

	retention := request.Service.RetentionPeriodDays
	// var expiresAt *time.Time
	// if retention > 0 {
	// 	exp := time.Now().Add(time.Minute * time.Duration(retention)) // eller .AddDate(0, 0, retention) for dager
	// 	expiresAt = &exp
	// }

	service := apicontracts.Service{
		ServiceName:         request.Service.ServiceName,
		ServiceId:           request.Service.ServiceId,
		ClusterId:           request.Service.ClusterId,
		RetentionPeriodDays: retention,
		ExpiresAt:           nil,
	}

	newDoc := bson.M{
		"secret":   request.Secret,
		"zone":     request.Zone,
		"prefix":   nextPrefix.Prefix,
		"id":       nextPrefix.ID,
		"services": []apicontracts.Service{service},
	}

	client := mongodb.GetClient()
	collection := client.Database(viper.GetString("mongodb.database")).Collection(viper.GetString("mongodb.collection"))

	result, err := collection.InsertOne(context.Background(), newDoc)
	if err != nil {
		return mongodbtypes.Prefix{}, errors.New("failed to save prefix: " + err.Error())
	}

	var prefix mongodbtypes.Prefix
	err = collection.FindOne(context.Background(), bson.M{"_id": result.InsertedID.(bson.ObjectID)}).Decode(&prefix)

	if err != nil {
		return mongodbtypes.Prefix{}, err
	}

	return prefix, nil

}

func UpdatePrefixDocument(request apicontracts.K8sRequestBody) error {
	// retention := request.Service.RetentionPeriodDays
	// var expiresAt *time.Time
	// if retention > 0 {
	// 	exp := time.Now().AddDate(0, 0, retention)
	// 	expiresAt = &exp
	// }
	// request.Service.ExpiresAt = expiresAt

	client := mongodb.GetClient()
	collection := client.Database(viper.GetString("mongodb.database")).Collection(viper.GetString("mongodb.collection"))

	filter := bson.M{
		"secret": request.Secret,
		"zone":   request.Zone,
		"prefix": request.Address,
	}

	var doc mongodbtypes.Prefix
	err := collection.FindOne(context.Background(), filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("prefix not found: %w", err)
		}
		return fmt.Errorf("failed to read prefix document: %w", err)
	}

	// 1. Filtrer ut eventuell eksisterende service med samme ID + navn + cluster
	var newServices []mongodbtypes.Service
	for _, s := range doc.Services {
		if !(s.ServiceId == request.Service.ServiceId &&
			s.ServiceName == request.Service.ServiceName &&
			s.ClusterId == request.Service.ClusterId) {
			newServices = append(newServices, mongodbtypes.Service{
				ServiceName:         s.ServiceName,
				ServiceId:           s.ServiceId,
				ClusterId:           s.ClusterId,
				RetentionPeriodDays: s.RetentionPeriodDays,
				ExpiresAt:           s.ExpiresAt,
			})
		}
	}

	// 2. Legg til ny/oppdatert service
	newServices = append(newServices, mongodbtypes.Service(request.Service))

	// 3. Oppdater dokumentet med ny services-array
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

// func UpdatePrefixDocument(request apicontracts.K8sRequestBody) error {
// 	retention := request.Service.RetentionPeriodDays
// 	var expiresAt *time.Time
// 	if retention > 0 {
// 		exp := time.Now().Add(time.Minute * time.Duration(retention)) // eller .AddDate(0, 0, retention) for dager
// 		expiresAt = &exp
// 	}
// 	request.Service.ExpiresAt = expiresAt

// 	client := mongodb.GetClient()
// 	collection := client.Database(viper.GetString("mongodb.database")).Collection(viper.GetString("mongodb.collection"))

// 	filter := bson.M{
// 		"secret": request.Secret,
// 		"zone":   request.Zone,
// 		"prefix": request.Address,
// 	}

// 	// 1. Fjern eksisterende service med samme id + navn + cluster
// 	update := bson.M{
// 		"$pull": bson.M{
// 			"services": bson.M{
// 				"service_id":   request.Service.ServiceId,
// 				"service_name": request.Service.ServiceName,
// 				"cluster_id":   request.Service.ClusterId,
// 			},
// 		},
// 		"$addToSet": bson.M{
// 			"services": request.Service,
// 		},
// 	}
// 	_, err := collection.UpdateOne(context.Background(), filter, update)
// 	if err != nil {
// 		return fmt.Errorf("failed to pull existing service: %w", err)
// 	}

// 	// 2. Legg til oppdatert service
// 	// fmt.Println("request.service", request.Service)
// 	// push := bson.M{
// 	// 	"$addToSet": bson.M{
// 	// 		"services": request.Service,
// 	// 	},
// 	// }
// 	// _, err = collection.UpdateOne(context.Background(), filter, push)
// 	if err != nil {
// 		return fmt.Errorf("failed to add updated service: %w", err)
// 	}

// 	return nil
// }

// DeleteServiceFromPrefix removes a service from the services array in a prefix document.
func DeleteServiceFromPrefix(request apicontracts.K8sRequestBody) error {

	client := mongodb.GetClient()
	collection := client.Database(viper.GetString("mongodb.database")).Collection(viper.GetString("mongodb.collection"))
	filter := bson.M{"secret": request.Secret, "zone": request.Zone, "prefix": request.Address}

	var result mongodbtypes.Prefix
	err := collection.FindOne(context.Background(), filter).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("no matching prefix found")
		}
		return errors.New(err.Error())
	}
	// Check if the service exists in the prefix document
	serviceFound := false
	for _, service := range result.Services {
		if request.Service.ServiceName == service.ServiceName && request.Service.ServiceId == service.ServiceId {
			serviceFound = true
			break
		}
	}

	if !serviceFound {
		return errors.New("service not found in prefix document")
	}

	update := bson.M{
		"$pull": bson.M{
			"services": request.Service,
		},
	}

	_, err = collection.UpdateOne(context.Background(), filter, update)

	if err != nil {
		return errors.New("failed to deregister service from prefix: " + err.Error())
	}

	err = collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		return errors.New("could not find prefix: " + err.Error())
	}

	if len(result.Services) == 0 {
		_, err = collection.DeleteOne(context.Background(), filter)
		if err != nil {
			return errors.New("failed to delete prefix from mongodb: " + err.Error())
		}
		err = netboxservice.DeleteNetboxPrefix(strconv.Itoa(result.NetboxID))
		if err != nil {
			return errors.New("failed to delete prefix from Netbox: " + err.Error())
		}
	}

	return nil
}
