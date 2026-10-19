package apicontracts

import (
	"encoding/json"
	"testing"
)

func TestVerifyIPOwnerResponseJSON(t *testing.T) {
	tests := []struct {
		name     string
		verified bool
		want     string
	}{
		{name: "verified", verified: true, want: `{"verified":true}`},
		{name: "not verified", verified: false, want: `{"verified":false}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			encoded, err := json.Marshal(VerifyIPOwnerResponse{Verified: test.verified})
			if err != nil {
				t.Fatal(err)
			}
			if string(encoded) != test.want {
				t.Fatalf("got %s, want %s", encoded, test.want)
			}
		})
	}
}
