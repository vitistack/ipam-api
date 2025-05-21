package mongodbservice

import (
	"reflect"
	"testing"

	"github.com/NorskHelsenett/oss-ipam-api/internal/responses"
	"github.com/NorskHelsenett/oss-ipam-api/pkg/models/apicontracts"
	"github.com/NorskHelsenett/oss-ipam-api/pkg/models/mongodbtypes"
)

func TestInsertNewPrefixDocument(t *testing.T) {
	type args struct {
		request    apicontracts.K8sRequestBody
		nextPrefix responses.NetboxPrefix
	}
	tests := []struct {
		name    string
		args    args
		want    mongodbtypes.Prefix
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InsertNewPrefixDocument(tt.args.request, tt.args.nextPrefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("InsertNewPrefixDocument() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InsertNewPrefixDocument() = %v, want %v", got, tt.want)
			}
		})
	}
}
