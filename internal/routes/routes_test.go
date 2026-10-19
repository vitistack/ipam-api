package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestVerifyIPOwnerRejectsInvalidRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := gin.New()
	SetupRoutes(server)

	for _, path := range []string{"/verify", "/v2/verify"} {
		t.Run(path, func(t *testing.T) {
			bodies := []string{"{", "null", "{}"}
			for _, missing := range []string{"address", "zone", "secret", "namespace_id", "cluster_id"} {
				body := map[string]string{
					"address": "192.0.2.1/32", "zone": "inet", "secret": "test-secret",
					"namespace_id": "namespace-1", "cluster_id": "cluster-1",
				}
				delete(body, missing)
				encoded, err := json.Marshal(body)
				if err != nil {
					t.Fatal(err)
				}
				bodies = append(bodies, string(encoded))
			}
			for _, body := range bodies {
				request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
				request.Header.Set("Content-Type", "application/json")
				response := httptest.NewRecorder()
				server.ServeHTTP(response, request)
				if response.Code != http.StatusBadRequest {
					t.Errorf("body %s: got status %d, want 400", body, response.Code)
				}
			}
		})
	}
}
