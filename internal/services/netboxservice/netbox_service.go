package netboxservice

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
	"github.com/vitistack/ipam-api/internal/responses"
	"github.com/vitistack/ipam-api/pkg/models/apicontracts"
)

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

	// if len(result.Results) != 1 {
	// 	return responses.NetboxPrefixesResponse{}, errors.New("multiple or no containers matching prefix found")
	// }

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

	client := resty.New()
	var netboxResp responses.NetboxPrefix
	resp, err := client.R().
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

	client := resty.New()
	var netboxResp responses.NetboxPrefix
	resp, err := client.R().
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

func GetAvailablePrefixContainer(request apicontracts.K8sRequestBody) (responses.NetboxPrefix, error) {
	zonePrefixes := request.ZonePrefixes()
	netboxURL := viper.GetString("netbox.url")
	netboxToken := viper.GetString("netbox.token")
	restyClient := resty.New()

	for _, prefix := range zonePrefixes {
		container, err := GetPrefixContainer(prefix)
		if err != nil {
			continue
		}
		var result []any

		resp, err := restyClient.R().
			SetHeader("Authorization", "Token "+netboxToken).
			SetHeader("Accept", "application/json").
			SetResult(&result).
			Get(netboxURL + "/api/ipam/prefixes/" + strconv.Itoa(container.ID) + "/available-prefixes/")

		if err != nil {
			continue
		}

		if resp.IsError() {
			continue
		}

		if len(result) == 0 {
			continue
		}

		return container, nil
	}
	return responses.NetboxPrefix{}, errors.New("no available prefix found. add more prefixes to config.json")
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

	fmt.Println("Zones", zones)
	return zones, nil
}
