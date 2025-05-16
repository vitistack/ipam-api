package apicontracts

import (
	"slices"

	"github.com/spf13/viper"
)

type Service struct {
	Name     string `json:"name" validate:"required,min=8,max=64"`
	Uuid     string `json:"uuid" validate:"required"`
	Location string `json:"location" validate:"required"`
}

type K8sRequestBody struct {
	Secret  string  `json:"secret" validate:"required,min=8,max=64"`
	Zone    string  `json:"zone" validate:"required,min=8,max=64"`
	Prefix  string  `json:"prefix" validate:"omitempty,cidr"`
	Service Service `json:"service"`
}

type K8sRequestResponse struct {
	Message string `json:"message"`
	Secret  string `json:"secret"`
	Zone    string `json:"zone"`
	Prefix  string `json:"prefix"`
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
