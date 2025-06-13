package netboxservice

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"

	"github.com/vitistack/ipam-api/internal/logger"
	"github.com/vitistack/ipam-api/internal/responses"
	"github.com/vitistack/ipam-api/pkg/models/apicontracts"
)

type NetboxCache struct {
	mu       sync.RWMutex
	prefixes map[string][]responses.NetboxPrefix
}

var Cache = &NetboxCache{
	prefixes: make(map[string][]responses.NetboxPrefix),
}

// GetPrefixContainer retrieves a Netbox prefix container matching the specified prefix string.
// It sends a GET request to the Netbox API using the configured URL and token, filtering by the given prefix and
// requiring the status to be "container". If exactly one matching container is found, it is returned.
// Returns an error if the request fails, if the response indicates an error, or if multiple or no containers are found.
//
// Parameters:
//   - prefix: The prefix string to search for.
//
// Returns:
//   - responses.NetboxPrefix: The matching Netbox prefix container.
//   - error: An error if the request fails or if the result is not exactly one container.
func GetPrefixContainer(prefix string) (responses.NetboxPrefix, error) {
	netboxURL := viper.GetString("netbox.url")
	netboxToken := viper.GetString("netbox.token")

	restyClient := resty.New()
	var result responses.NetboxResponse[responses.NetboxPrefix]
	resp, err := restyClient.R().
		SetHeader("Authorization", "Token "+netboxToken).
		SetHeader("Accept", "application/json").
		SetResult(&result).
		Get(netboxURL + "/api/ipam/prefixes/?prefix=" + string(prefix) + "&status=container")

	if err != nil {
		return responses.NetboxPrefix{}, err
	}

	if resp.IsError() {
		return responses.NetboxPrefix{}, err
	}

	if len(result.Results) != 1 {
		return responses.NetboxPrefix{}, errors.New("multiple or no containers matching prefix found")
	}

	container := result.Results[0]

	return container, nil
}

// GetPrefixes retrieves a list of Netbox prefixes based on the provided query parameters.
// It sends a GET request to the Netbox API using the configured URL and token,
// and returns a slice of NetboxPrefix objects or an error if the request fails.
//
// Parameters:
//   - queryParams: map[string]string of query parameters to append to the Netbox prefixes API endpoint.
//
// Returns:
//   - []responses.NetboxPrefix: A slice of NetboxPrefix objects returned by the Netbox API.
//   - error: An error if the request fails or the API returns an error response.
func GetPrefixes(queryParams map[string]string) ([]responses.NetboxPrefix, error) {
	netboxURL := viper.GetString("netbox.url")
	netboxToken := viper.GetString("netbox.token")

	restyClient := resty.New()
	var netboxResponse responses.NetboxResponse[responses.NetboxPrefix]

	resp, err := restyClient.R().
		SetHeader("Authorization", "Token "+netboxToken).
		SetHeader("Accept", "application/json").
		SetResult(&netboxResponse).
		SetQueryParams(queryParams).
		Get(netboxURL + "/api/ipam/prefixes/")

	if err != nil {
		return []responses.NetboxPrefix{}, err
	}

	if resp.IsError() {
		return []responses.NetboxPrefix{}, err
	}

	return netboxResponse.Results, nil
}

// CheckPrefixContainerAvailability queries the NetBox API to check for available prefixes
// within a specified prefix container. It takes the containerId as a string and returns
// the first available NetboxPrefix found, or an error if none are available or if the
// request fails.
//
// Parameters:
//   - containerId: The ID of the prefix container to check for available prefixes.
//
// Returns:
//   - responses.NetboxPrefix: The first available prefix found in the container.
//   - error: An error if the request fails or no prefixes are found.
func CheckPrefixContainerAvailability(containerId string) (responses.NetboxPrefix, error) {
	netboxURL := viper.GetString("netbox.url")
	netboxToken := viper.GetString("netbox.token")

	restyClient := resty.New()
	var result responses.NetboxResponse[responses.NetboxPrefix]
	resp, err := restyClient.R().
		SetHeader("Authorization", "Token "+netboxToken).
		SetHeader("Accept", "application/json").
		SetResult(&result).
		Get(netboxURL + "/api/ipam/prefixes/" + containerId + "/available-prefixes/")

	if err != nil {
		return responses.NetboxPrefix{}, err
	}

	if resp.IsError() {
		return responses.NetboxPrefix{}, err
	}

	if len(result.Results) == 0 {
		return responses.NetboxPrefix{}, errors.New("no prefixes found in the container")
	}

	return result.Results[0], nil
}

// Sends a POST request to the NetBox API to create a new available prefix within the specified container.
// The request includes authorization and content headers, a payload as the request body, and stores the result in newPrefix.
// Parameters:
//   - netboxToken: API token for NetBox authentication.
//   - payload: The request body containing prefix details.
//   - netboxURL: Base URL of the NetBox instance.
//   - containerId: Identifier of the prefix container in NetBox.
//
// Returns:
//   - resp: The HTTP response from the NetBox API.
//   - error: Any error encountered during the request (currently ignored in this snippet).
func GetNextPrefixFromContainer(containerId string, payload apicontracts.NextPrefixPayload) (responses.NetboxPrefix, error) {
	netboxURL := viper.GetString("netbox.url")
	netboxToken := viper.GetString("netbox.token")

	restyClient := resty.New()
	var newPrefix responses.NetboxPrefix
	resp, _ := restyClient.R().
		SetHeader("Authorization", "Token "+netboxToken).
		SetHeader("Accept", "application/json").
		SetBody(payload).
		SetResult(&newPrefix).
		Post(netboxURL + "/api/ipam/prefixes/" + containerId + "/available-prefixes/")

	if resp.IsError() {
		logger.Log.Errorf("Error fetching next prefix from container %s: %s", containerId, resp.String())
		return responses.NetboxPrefix{}, errors.New(resp.String())
	}

	return newPrefix, nil
}

// UpdateNetboxPrefix updates a prefix in Netbox with the specified prefixId using the provided payload.
// It sends a PUT request to the Netbox API and returns an error if the request fails or if the response indicates an error.
//
// Parameters:
//   - prefixId: The ID of the prefix to update in Netbox.
//   - payload: The data to update the prefix with, conforming to apicontracts.UpdatePrefixPayload.
//
// Returns:
//   - error: An error if the update fails, or nil if successful.
func UpdateNetboxPrefix(prefixId string, payload apicontracts.UpdatePrefixPayload) error {
	netboxURL := viper.GetString("netbox.url")
	netboxToken := viper.GetString("netbox.token")

	restyClient := resty.New()
	var netboxResp responses.NetboxPrefix
	resp, err := restyClient.R().
		SetHeader("Authorization", "Token "+netboxToken).
		SetHeader("Accept", "application/json").
		SetBody(payload).
		SetResult(&netboxResp).
		Put(netboxURL + "/api/ipam/prefixes/" + prefixId + "/")

	if err != nil {
		return err
	}

	if resp.IsError() {
		logger.Log.Errorf("Error updating prefix %s in Netbox: %s", prefixId, resp.String())
		return errors.New(resp.String())
	}
	return nil
}

// DeleteNetboxPrefix deletes a prefix in Netbox identified by the given prefixId.
// It sends a DELETE request to the Netbox API using the configured URL and token.
// Returns an error if the request fails or if Netbox responds with an error.
func DeleteNetboxPrefix(prefixId string) error {
	netboxURL := viper.GetString("netbox.url")
	netboxToken := viper.GetString("netbox.token")

	restyClient := resty.New()
	var netboxResp responses.NetboxPrefix
	resp, err := restyClient.R().
		SetHeader("Authorization", "Token "+netboxToken).
		SetHeader("Accept", "application/json").
		SetResult(&netboxResp).
		Delete(netboxURL + "/api/ipam/prefixes/" + prefixId + "/")

	if err != nil {
		logger.Log.Errorf("Error deleting prefix %s in Netbox: %v", prefixId, err)
		return err
	}

	if resp.IsError() {
		logger.Log.Errorf("Error deleting prefix %s in Netbox: %s", prefixId, resp.String())
		return errors.New(resp.String())
	}
	return nil
}

// GetAvailablePrefixContainer attempts to find and return an available prefix container for the specified IPAM API request.
// It determines the zone based on the request's zone and IP family, retrieves cached prefixes for that zone,
// and queries the NetBox API for available prefixes within each cached prefix.
// If an available prefix is found, it is returned; otherwise, an error is returned indicating no available prefix was found.
//
// Parameters:
//   - request: apicontracts.IpamApiRequest containing the zone and IP family information.
//
// Returns:
//   - responses.NetboxPrefix: The first available prefix found for the specified zone.
//   - error: An error if no available prefix is found or if an error occurs during the process.
func GetAvailablePrefixContainer(request apicontracts.IpamApiRequest) (responses.NetboxPrefix, error) {
	zone := request.Zone + "_v" + string(request.IpFamily[len(request.IpFamily)-1])
	zonePrefixes := Cache.Get(zone)

	netboxURL := viper.GetString("netbox.url")
	netboxToken := viper.GetString("netbox.token")
	restyClient := resty.New()

	for _, prefix := range zonePrefixes {
		var result []any
		resp, err := restyClient.R().
			SetHeader("Authorization", "Token "+netboxToken).
			SetHeader("Accept", "application/json").
			SetResult(&result).
			Get(netboxURL + "/api/ipam/prefixes/" + strconv.Itoa(prefix.ID) + "/available-prefixes/")

		if err != nil {
			continue
		}

		if resp.IsError() {
			continue
		}

		if len(result) == 0 {
			continue
		}

		return prefix, nil
	}
	logger.Log.Infof("No available prefix found for zone %s", zone)
	return responses.NetboxPrefix{}, fmt.Errorf("no available prefix found for zone %s", zone)
}

// GetK8sZones retrieves the list of Kubernetes zones from the Netbox API.
// It sends a GET request to the Netbox custom field choice sets endpoint, filtering for "k8s_zone_choices".
// The function returns a slice of zone names as strings, or an error if the request fails or the response is invalid.
func GetK8sZones() ([]string, error) {
	netboxURL := viper.GetString("netbox.url")
	netboxToken := viper.GetString("netbox.token")

	restyClient := resty.New()
	var netboxResponse responses.NetboxResponse[responses.NetboxChoiceSet]

	resp, err := restyClient.R().
		SetHeader("Authorization", "Token "+netboxToken).
		SetHeader("Accept", "application/json").
		SetResult(&netboxResponse).
		Get(netboxURL + "/api/extras/custom-field-choice-sets/?q=k8s_zone_choices")

	if err != nil {
		logger.Log.Errorf("Error fetching k8s zones from Netbox: %v", err)
		return nil, err
	}

	if resp.IsError() {
		logger.Log.Errorf("Error fetching k8s zones from Netbox: %s", resp.String())
		return nil, errors.New(resp.String())
	}

	var zones []string

	for _, choice := range netboxResponse.Results[0].ExtraChoices {
		zones = append(zones, choice[0])
	}

	return zones, nil
}

// FetchPrefixContainers retrieves Kubernetes zones from Netbox, fetches associated IPv4 and IPv6 prefixes for each zone,
// and updates the NetboxCache with the collected prefix data. It organizes prefixes by zone and IP family (IPv4/IPv6).
// Returns an error if fetching zones or prefixes fails.
func (c *NetboxCache) FetchPrefixContainers() error {
	zones, err := GetK8sZones()
	if err != nil {
		return errors.New("failed to fetch zones from Netbox: " + err.Error())
	}

	zonePrefixes := make(map[string][]responses.NetboxPrefix)

	for _, zone := range zones {
		ipv4Zone := zone + "_v4"
		ipv6Zone := zone + "_v6"
		zonePrefixes[ipv4Zone] = []responses.NetboxPrefix{}
		zonePrefixes[ipv6Zone] = []responses.NetboxPrefix{}

		queryParams := map[string]string{
			"cf_k8s_zone": zone,
			"status":      "container"}

		prefixes, err := GetPrefixes(queryParams)
		if err != nil {
			logger.Log.Errorf("Error fetching prefixes for zone %s: %v", zone, err)
			return fmt.Errorf("error fetching prefixes for zone %s: %v", zone, err)
		}

		for _, prefix := range prefixes {
			switch prefix.Family.Value {
			case 4:
				zonePrefixes[ipv4Zone] = append(zonePrefixes[ipv4Zone], prefix)
			case 6:
				zonePrefixes[ipv6Zone] = append(zonePrefixes[ipv6Zone], prefix)
			}
		}
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	c.prefixes = zonePrefixes
	return nil
}

// Get returns prefixes for a given zone
func (c *NetboxCache) Get(key string) []responses.NetboxPrefix {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.prefixes[key]
}

// WaitForNetbox continuously attempts to connect to the NetBox API using the configured URL and token.
// It sends a GET request to the /api/ipam/prefixes/ endpoint, retrying every 10 seconds until a successful response is received.
// If an error occurs or a non-successful status code is returned, it logs the issue and retries.
// The function returns nil once NetBox becomes available.
func WaitForNetbox() error {
	netboxURL := viper.GetString("netbox.url")
	netboxToken := viper.GetString("netbox.token")
	delay := 10 * time.Second
	client := resty.New()

	netboxAvailable := false

	for !netboxAvailable {
		resp, err := client.R().
			SetHeader("Authorization", "Token "+netboxToken).
			SetHeader("Accept", "application/json").
			Get(netboxURL + "/api/ipam/prefixes/")

		if err == nil && resp.IsSuccess() {
			netboxAvailable = true
			return nil
		}

		if err != nil {
			logger.Log.Infof("Error reaching NetBox: %v. Retrying in %v...", err, delay)
		} else {
			logger.Log.Infof("Netbox responded with status %d. Retrying in %v...", resp.StatusCode(), delay)
		}

		time.Sleep(delay)
	}

	return nil
}

// PrefixAvailable checks if a given IP prefix is available in NetBox.
//
// It sends a GET request to the NetBox API, querying for the specified prefix
// within the "nhc" VRF. If no matching prefix is found in the response, the
// function returns true, indicating the prefix is available. If the prefix
// exists or an error occurs during the request, it returns false and an error.
//
// Parameters:
//   - request: an IpamApiRequest containing the prefix address to check.
//
// Returns:
//   - bool: true if the prefix is available, false otherwise.
//   - error: any error encountered during the API request or response handling.
func PrefixAvailable(request apicontracts.IpamApiRequest) (bool, error) {
	netboxURL := viper.GetString("netbox.url")
	netboxToken := viper.GetString("netbox.token")

	restyClient := resty.New()
	var result responses.NetboxResponse[responses.NetboxPrefix]
	resp, err := restyClient.R().
		SetHeader("Authorization", "Token "+netboxToken).
		SetHeader("Accept", "application/json").
		SetResult(&result).
		SetQueryParam("present_in_vrf", "nhc").
		SetQueryParam("prefix", request.Address).
		Get(netboxURL + "/api/ipam/prefixes/")

	if err != nil {
		return false, err
	}

	if resp.IsError() {
		return false, errors.New(resp.String())
	}

	if len(result.Results) == 0 {
		return true, nil
	}

	return false, nil
}

// RegisterPrefix sends a request to the NetBox API to create a new IP prefix using the provided payload.
// It returns the created NetboxPrefix object on success, or an error if the request fails.
//
// Parameters:
//   - payload: apicontracts.CreatePrefixPayload containing the details of the prefix to be created.
//
// Returns:
//   - responses.NetboxPrefix: The created prefix object returned by NetBox.
//   - error: An error if the request fails or the API returns an error response.
func RegisterPrefix(payload apicontracts.CreatePrefixPayload) (responses.NetboxPrefix, error) {
	netboxURL := viper.GetString("netbox.url")
	netboxToken := viper.GetString("netbox.token")

	restyClient := resty.New()
	var result responses.NetboxPrefix
	resp, err := restyClient.R().
		SetHeader("Authorization", "Token "+netboxToken).
		SetHeader("Accept", "application/json").
		SetBody(payload).
		SetResult(&result).
		Post(netboxURL + "/api/ipam/prefixes/")

	if err != nil {
		return responses.NetboxPrefix{}, err
	}

	if resp.IsError() {
		return responses.NetboxPrefix{}, errors.New(resp.String())
	}

	return result, nil
}

func GetTagId(tagName string) (int, error) {
	netboxURL := viper.GetString("netbox.url")
	netboxToken := viper.GetString("netbox.token")

	restyClient := resty.New()
	var result responses.NetboxResponse[responses.NetboxTag]
	resp, err := restyClient.R().
		SetHeader("Authorization", "Token "+netboxToken).
		SetHeader("Accept", "application/json").
		SetQueryParam("name", tagName).
		SetResult(&result).
		Get(netboxURL + "/api/extras/tags/")

	if err != nil {
		return 0, err
	}

	if resp.IsError() || len(result.Results) == 0 {
		return 0, fmt.Errorf("tag %s not found", tagName)
	}

	return result.Results[0].ID, nil
}
