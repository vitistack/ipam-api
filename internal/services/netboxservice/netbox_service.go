package netboxservice

import (
	"errors"
	"strconv"
	"sync"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"

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

// GetPrefixContainer retrieves a prefix container from Netbox using the provided prefix.
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

func GetPrefixes(query string) ([]responses.NetboxPrefix, error) {
	netboxURL := viper.GetString("netbox.url")
	netboxToken := viper.GetString("netbox.token")

	restyClient := resty.New()
	var netboxResponse responses.NetboxResponse[responses.NetboxPrefix]
	url := netboxURL + "/api/ipam/prefixes/" + query
	resp, err := restyClient.R().
		SetHeader("Authorization", "Token "+netboxToken).
		SetHeader("Accept", "application/json").
		SetResult(&netboxResponse).
		Get(url)

	if err != nil {
		return []responses.NetboxPrefix{}, err
	}

	if resp.IsError() {
		return []responses.NetboxPrefix{}, err
	}

	return netboxResponse.Results, nil
}

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

// GetNextPrefixFromContainer retrieves the next available prefix from Netbox for a given container ID.
func GetNextPrefixFromContainer(containerId string, payload apicontracts.NextPrefixPayload) (responses.NetboxPrefix, error) {
	netboxURL := viper.GetString("netbox.url")
	netboxToken := viper.GetString("netbox.token")

	restyClient := resty.New()
	var response responses.NetboxPrefix
	resp, _ := restyClient.R().
		SetHeader("Authorization", "Token "+netboxToken).
		SetHeader("Accept", "application/json").
		SetBody(payload).
		SetResult(&response).
		Post(netboxURL + "/api/ipam/prefixes/" + containerId + "/available-prefixes/")

	if resp.IsError() {
		return responses.NetboxPrefix{}, errors.New(resp.String())
	}

	return responses.NetboxPrefix{
		ID:     response.ID,
		Prefix: response.Prefix,
	}, nil
}

// UpdateNetboxPrefix updates a prefix in Netbox with the provided ID and payload.
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
		return errors.New(resp.String())
	}
	return nil
}

// DeleteNetboxPrefix deletes a prefix in Netbox with the provided ID.
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
		return err
	}

	if resp.IsError() {
		return errors.New(resp.String())
	}
	return nil
}

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
	return responses.NetboxPrefix{}, errors.New("no available prefix found. add more prefixes in netbox")
}

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
		return nil, err
	}

	if resp.IsError() {
		return nil, errors.New(resp.String())
	}

	var zones []string

	for _, choice := range netboxResponse.Results[0].ExtraChoices {
		zones = append(zones, choice[0])
	}

	return zones, nil
}

// Fetches Zones and prefixes from Netbox
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

		prefixes, err := GetPrefixes("?cf_k8s_zone=" + zone)
		if err != nil {
			return errors.New("error fetching prefixes for zone " + zone + " : " + err.Error())
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

// Return the full map
// func (c *NetboxCache) All() map[string][]responses.NetboxPrefix {
// 	c.mu.RLock()
// 	defer c.mu.RUnlock()
// 	result := make(map[string][]responses.NetboxPrefix, len(c.prefixes))
// 	maps.Copy(result, c.prefixes)
// 	return result
// }
