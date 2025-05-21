package mongodbservice

import (
	"context"
	"errors"
	"strconv"

	"github.com/NorskHelsenett/oss-ipam-api/internal/responses"
	"github.com/NorskHelsenett/oss-ipam-api/internal/services/netboxservice"
	"github.com/NorskHelsenett/oss-ipam-api/pkg/clients/mongodb"
	"github.com/NorskHelsenett/oss-ipam-api/pkg/models/apicontracts"
	"github.com/NorskHelsenett/oss-ipam-api/pkg/models/mongodbtypes"
	"github.com/spf13/viper"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func InsertNewPrefixDocument(request apicontracts.K8sRequestBody, nextPrefix responses.NetboxPrefix) (mongodbtypes.Prefix, error) {
	newDoc := bson.M{
		"secret":   request.Secret,
		"zone":     request.Zone,
		"prefix":   nextPrefix.Prefix,
		"id":       nextPrefix.ID,
		"services": []any{request.Service},
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
	update := bson.M{
		"$addToSet": bson.M{
			"services": request.Service,
		},
	}

	opts := options.UpdateOne().SetUpsert(true)
	client := mongodb.GetClient()
	collection := client.Database(viper.GetString("mongodb.database")).Collection(viper.GetString("mongodb.collection"))
	filter := bson.M{"secret": request.Secret, "zone": request.Zone, "prefix": request.Prefix}

	var result mongodbtypes.Prefix
	err := collection.FindOne(context.Background(), filter).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("no matching prefix found")
		}
		return errors.New(err.Error())
	}

	_, err = collection.UpdateOne(context.Background(), filter, update, opts)

	if err != nil {
		return errors.New("failed to update prefix: " + err.Error())
	}

	return nil
}

// DeleteServiceFromPrefix removes a service from the services array in a prefix document.
func DeleteServiceFromPrefix(request apicontracts.K8sRequestBody) error {
	update := bson.M{
		"$pull": bson.M{
			"services": request.Service,
		},
	}

	client := mongodb.GetClient()
	collection := client.Database(viper.GetString("mongodb.database")).Collection(viper.GetString("mongodb.collection"))
	filter := bson.M{"secret": request.Secret, "zone": request.Zone, "prefix": request.Prefix}

	var result mongodbtypes.Prefix
	err := collection.FindOne(context.Background(), filter).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("no matching prefix found")
		}
		return errors.New(err.Error())
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
