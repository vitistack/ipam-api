package mongodbservice

import (
	"context"
	"reflect"
	"testing"

	"github.com/spf13/viper"
	"github.com/vitistack/ipam-api/internal/utils"
	"github.com/vitistack/ipam-api/pkg/models/apicontracts"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver/drivertest"
)

func TestVerifyIPOwner(t *testing.T) {
	viper.Set("enc_key", "0123456789abcdef")
	viper.Set("enc_iv", "abcdef0123456789")
	t.Cleanup(viper.Reset)

	request := apicontracts.VerifyIPOwnerRequest{
		Address: "192.0.2.1/32", Zone: "inet", Secret: "test-secret",
		NamespaceID: "namespace-1", ClusterID: "cluster-1",
	}
	filter, err := ipOwnerFilter(request)
	if err != nil {
		t.Fatal(err)
	}
	secret, err := utils.DeterministicDecrypt(filter["secret"].(string))
	if err != nil || secret != request.Secret {
		t.Fatalf("stored secret must decrypt to request secret: %v", err)
	}
	expected := bson.M{
		"address": request.Address,
		"zone":    request.Zone,
		"secret":  filter["secret"],
		"services": bson.M{"$elemMatch": bson.M{
			"cluster_id": request.ClusterID, "namespace_id": request.NamespaceID,
		}},
	}
	if !reflect.DeepEqual(filter, expected) {
		t.Fatalf("unexpected ownership filter: %v", filter)
	}

	tests := []struct {
		name    string
		batch   bson.A
		failure bool
		want    bool
	}{
		{name: "matching owner", batch: bson.A{bson.D{{Key: "_id", Value: bson.NewObjectID()}}}, want: true},
		{name: "no matching owner", batch: bson.A{}},
		{name: "database failure", failure: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := bson.D{
				{Key: "ok", Value: 1},
				{Key: "cursor", Value: bson.D{
					{Key: "id", Value: int64(0)},
					{Key: "ns", Value: "test.addresses"},
					{Key: "firstBatch", Value: test.batch},
				}},
			}
			if test.failure {
				response = bson.D{{Key: "ok", Value: 0}, {Key: "code", Value: 2}, {Key: "errmsg", Value: "query failed"}}
			}
			clientOptions := options.Client()
			clientOptions.Deployment = drivertest.NewMockDeployment(response)
			client, err := mongo.Connect(clientOptions)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := client.Disconnect(context.Background()); err != nil {
					t.Error(err)
				}
			})
			got, err := verifyIPOwner(context.Background(), client.Database("test").Collection("addresses"), request)
			if got != test.want || (err != nil) != test.failure {
				t.Fatalf("got (%v, %v), want (%v, error=%v)", got, err, test.want, test.failure)
			}
		})
	}

	viper.Set("enc_key", "invalid")
	if _, err := ipOwnerFilter(request); err == nil {
		t.Fatal("expected encryption configuration error")
	}
}
