package mongodbservice

import (
	"context"
	"errors"
	"strconv"
	"time"
	"fmt"

	"github.com/NorskHelsenett/oss-ipam-api/internal/responses"
	"github.com/NorskHelsenett/oss-ipam-api/internal/services/netboxservice"
	"github.com/NorskHelsenett/oss-ipam-api/pkg/clients/mongodb"
	"github.com/NorskHelsenett/oss-ipam-api/pkg/models/apicontracts"
	"github.com/NorskHelsenett/oss-ipam-api/pkg/models/mongodbtypes"
	"github.com/spf13/viper"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	// "go.mongodb.org/mongo-driver/v2/mongo/options"
)

func InsertNewPrefixDocument(request apicontracts.K8sRequestBody, nextPrefix responses.NetboxPrefix) (mongodbtypes.Prefix, error) {

	retention := request.Service.RetentionPeriodDays
	var expiresAt *time.Time
	if retention > 0 {
		exp := time.Now().Add(time.Minute * time.Duration(retention)) // eller .AddDate(0, 0, retention) for dager
		expiresAt = &exp
	}

	service := apicontracts.Service{
		ServiceName:         request.Service.ServiceName,
		ServiceId:           request.Service.ServiceId,
		ClusterId:           request.Service.ClusterId,
		RetentionPeriodDays: retention,
		ExpiresAt:           expiresAt,
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
	retention := request.Service.RetentionPeriodDays
	var expiresAt *time.Time
	if retention > 0 {
		exp := time.Now().Add(time.Minute * time.Duration(retention)) // eller .AddDate(0, 0, retention) for dager
		expiresAt = &exp
	}
	request.Service.ExpiresAt = expiresAt

	client := mongodb.GetClient()
	collection := client.Database(viper.GetString("mongodb.database")).Collection(viper.GetString("mongodb.collection"))

	filter := bson.M{
		"secret": request.Secret,
		"zone":   request.Zone,
		"prefix": request.Address,
	}

	// 1. Fjern eksisterende service med samme id + navn + cluster
	pull := bson.M{
		"$pull": bson.M{
			"services": bson.M{
				"service_id":   request.Service.ServiceId,
				"service_name": request.Service.ServiceName,
				"cluster_id":   request.Service.ClusterId,
			},
		},
	}
	_, err := collection.UpdateOne(context.Background(), filter, pull)
	if err != nil {
		return fmt.Errorf("failed to pull existing service: %w", err)
	}

	// 2. Legg til oppdatert service
	fmt.Println("request.service", request.Service)
	push := bson.M{
		"$addToSet": bson.M{
			"services": request.Service,
		},
	}
	_, err = collection.UpdateOne(context.Background(), filter, push)
	if err != nil {
		return fmt.Errorf("failed to add updated service: %w", err)
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

// 	update := bson.M{
// 		"$addToSet": bson.M{
// 			"services": request.Service,
// 		},
// 	}

// 	opts := options.UpdateOne().SetUpsert(true)
// 	client := mongodb.GetClient()
// 	collection := client.Database(viper.GetString("mongodb.database")).Collection(viper.GetString("mongodb.collection"))
// 	filter := bson.M{"secret": request.Secret, "zone": request.Zone, "prefix": request.Address}

// 	var result mongodbtypes.Prefix
// 	err := collection.FindOne(context.Background(), filter).Decode(&result)

// 	if err != nil {
// 		if err == mongo.ErrNoDocuments {
// 			return errors.New("no matching prefix found")
// 		}
// 		return errors.New(err.Error())
// 	}

// 	_, err = collection.UpdateOne(context.Background(), filter, update, opts)

// 	if err != nil {
// 		return errors.New("failed to update prefix: " + err.Error())
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
