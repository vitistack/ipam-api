package utils

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/viper"
	"github.com/vitistack/ipam-api/internal/responses"
	"github.com/vitistack/ipam-api/internal/services/netboxservice"
)

func GetNetboxConfig() ([]responses.NetboxPrefix, error) {
	netboxURL := viper.GetString("netbox.url")
	netboxToken := viper.GetString("netbox.token")

	if netboxURL == "" || netboxToken == "" {
		return nil, errors.New("Netbox URL or token is not configured")
	}

	containers, err := netboxservice.GetPrefixes("?status=container&tag=vitistack-container")

	if err != nil {
		return nil, errors.New("failed to connect to Netbox: " + err.Error())
	}

	json, err := json.Marshal(containers)
	if err != nil {
		return nil, errors.New("failed to marshal Netbox containers: " + err.Error())
	}
	fmt.Println("Netbox Containers JSON:", string(json))

	return containers, nil
}
