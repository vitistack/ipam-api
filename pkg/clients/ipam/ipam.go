package ipam

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/vitistack/ipam-api/pkg/models/apicontracts"
)

type IPAMClient struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

// NewIPAMClient creates a new IPAM API client.
//
// baseURL is the base URL of the IPAM API (e.g. "http://localhost:3000/v2").
//
// token is the Bearer token used for authentication.
//
// Example:
//
//	client := ipam.NewIPAMClient("http://localhost:3000/v2", "your-token-goes-here")
func NewIPAMClient(baseURL, token string) *IPAMClient {
	httpClient := &http.Client{
		Timeout: 20 * time.Second,
	}

	return &IPAMClient{
		httpClient: httpClient,
		baseURL:    baseURL,
		token:      token,
	}
}

// DeleteCluster deletes all services associated with the given cluster ID.
//
// request is the IpamAPIDeleteClusterRequest containing the ID of the cluster whose services should be removed.
//
// The function returns an IpamAPIResponse if the operation succeeds.
// If the API responds with a non-2xx status code, an error is returned.
func (c *IPAMClient) DeleteCluster(request apicontracts.IpamAPIDeleteClusterRequest) (apicontracts.IpamAPIResponse, error) {
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return apicontracts.IpamAPIResponse{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/cluster", c.baseURL), bytes.NewReader(requestBytes))

	if err != nil {
		return apicontracts.IpamAPIResponse{}, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return apicontracts.IpamAPIResponse{}, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return apicontracts.IpamAPIResponse{}, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return apicontracts.IpamAPIResponse{}, errors.New(string(bodyBytes))
	}

	var apiResponse apicontracts.IpamAPIResponse
	if err := json.Unmarshal(bodyBytes, &apiResponse); err != nil {
		return apicontracts.IpamAPIResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return apiResponse, nil
}
