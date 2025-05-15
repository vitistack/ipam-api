package netboxservice

import (
	"errors"
	"fmt"

	"github.com/NorskHelsenett/oss-ipam-api/internal/responses"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
)

// GetPrefixContainer retrieves a prefix container from NetBox using the provided prefix.
func GetPrefixContainer(prefix string) (responses.NetboxPrefix, error) {
	netboxURL := viper.GetString("netbox.url")
	netboxToken := viper.GetString("netbox.token")

	fmt.Println("netboxURL: ", netboxURL)
	if netboxURL == "" || netboxToken == "" {
		return responses.NetboxPrefix{}, errors.New("check your environment")
	}

	restyClient := resty.New()
	var result responses.NetboxPrefixes

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

// GetNextPrefixFromContainer retrieves the next available prefix from NetBox for a given container ID.
func GetNextPrefixFromContainer(containerId string, payload map[string]any) (responses.NetboxPrefix, error) {
	netboxURL := viper.GetString("netbox.url")
	netboxToken := viper.GetString("netbox.token")
	if netboxURL == "" || netboxToken == "" {
		return responses.NetboxPrefix{}, errors.New("check your environment")
	}

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
