package netboxservice_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NorskHelsenett/oss-ipam-api/internal/services/netboxservice"
	"github.com/NorskHelsenett/oss-ipam-api/pkg/models/apicontracts"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func setupViper(url, token string) {
	viper.Set("netbox.url", url)
	viper.Set("netbox.token", token)
}

func TestUpdateNetboxPrefix_Success(t *testing.T) {
	// Start a test server that always returns 200 OK
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Token testtoken" {
			t.Errorf("Expected Authorization header 'Token testtoken', got '%s'", got)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"prefix": "10.0.0.0/32", "id": 1521}`))
	}))
	defer server.Close()

	setupViper(server.URL, "testtoken")

	payload := apicontracts.UpdatePrefixPayload{
		Prefix: "10.0.0.0/32",
	}

	err := netboxservice.UpdateNetboxPrefix("1521", payload)
	assert.NoError(t, err)
}

func TestUpdateNetboxPrefix_ErrorFromAPI(t *testing.T) {
	// Test server returns error (simulate NetBox failure)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer server.Close()

	setupViper(server.URL, "testtoken")

	payload := apicontracts.UpdatePrefixPayload{
		Prefix: "10.0.0.0/32",
	}

	err := netboxservice.UpdateNetboxPrefix("42", payload)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "forbidden") || strings.Contains(err.Error(), "403"))
}

func TestUpdateNetboxPrefix_ErrorFromResty(t *testing.T) {
	// Invalid server URL to trigger a client error
	setupViper("http://bad host", "testtoken")
	payload := apicontracts.UpdatePrefixPayload{Prefix: "10.0.0.0/32"}
	err := netboxservice.UpdateNetboxPrefix("42", payload)
	assert.Error(t, err)
}
