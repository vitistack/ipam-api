package mongodbservice

import (
	"context"
	"errors"

	"github.com/NorskHelsenett/oss-ipam-api/internal/responses"
	"github.com/NorskHelsenett/oss-ipam-api/pkg/clients/mongodb"
	"github.com/NorskHelsenett/oss-ipam-api/pkg/models/apicontracts"
	"github.com/NorskHelsenett/oss-ipam-api/pkg/models/mongodbtypes"

	"go.mongodb.org/mongo-driver/v2/bson"
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
	collection := client.Database("netbox_proxy").Collection("prefixes")

	result, err := collection.InsertOne(context.Background(), newDoc)
	if err != nil {
		return mongodbtypes.Prefix{}, errors.New("failed to insert new prefix document")
	}

	var prefix mongodbtypes.Prefix
	err = collection.FindOne(context.Background(), bson.M{"_id": result.InsertedID.(bson.ObjectID)}).Decode(&prefix)

	if err != nil {
		return mongodbtypes.Prefix{}, err
	}

	return prefix, nil

}
