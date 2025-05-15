package apicontracts

import (
	"slices"

	"github.com/spf13/viper"
)

type Service struct {
	Name     string `json:"name"`
	Uuid     string `json:"uuid"`
	Location string `json:"location"`
}

type K8sRequestBody struct {
	Secret  string  `json:"secret"`
	Zone    string  `json:"zone"`
	Prefix  string  `json:"prefix"`
	Service Service `json:"service"`
}

func (r *K8sRequestBody) IsValidZone() bool {
	allowed := []string{"internet", "helsenett-private", "helsenett-public"}
	return slices.Contains(allowed, r.Zone)
}

func (r *K8sRequestBody) ZonePrefix() string {
	switch r.Zone {
	case "internet":
		return viper.GetString("netbox.prefix_containers.internet")
	case "helsenett-private":
		return viper.GetString("netbox.prefix_containers.helsenett-private")
	case "helsenett-public":
		return viper.GetString("netbox.prefix_containers.helsenett-public")
	default:
		return ""
	}
}
