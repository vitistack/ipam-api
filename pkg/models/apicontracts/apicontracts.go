package apicontracts

import (
	"slices"

	"github.com/NorskHelsenett/oss-ipam-api/internal/responses"
	"github.com/NorskHelsenett/oss-ipam-api/pkg/models/mongodbtypes"
	"github.com/spf13/viper"
)

type Service struct {
	Name     string `json:"name" validate:"required,min=8,max=64" example:"service1"`
	Uuid     string `json:"uuid" validate:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	Location string `json:"location" validate:"required" example:"location1"`
}

type K8sRequestBody struct {
	Secret  string  `json:"secret" validate:"required,min=8,max=64" example:"a_secret_value"`
	Zone    string  `json:"zone" validate:"required,min=8,max=64" example:"internet"`
	Prefix  string  `json:"prefix" validate:"omitempty,cidr" example:"10.0.0.0/32"`
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

func (r *K8sRequestBody) ZonePrefixes() []string {
	switch r.Zone {
	case "internet":
		return viper.GetStringSlice("netbox.prefix_containers.internet")
	case "helsenett-private":
		return viper.GetStringSlice("netbox.prefix_containers.helsenett-private")
	case "helsenett-public":
		return viper.GetStringSlice("netbox.prefix_containers.helsenett-public")
	default:
		return []string{}
	}
}

type CustomFields struct {
	Domain  string `json:"domain"`
	Env     string `json:"env"`
	Infra   string `json:"infra"`
	Purpose string `json:"purpose"`
	K8suuid string `json:"k8s_uuid"`
}

type NextPrefixPayload struct {
	PrefixLength int          `json:"prefix_length"`
	CustomFields CustomFields `json:"custom_fields"`
}

type UpdatePrefixPayload struct {
	Prefix       string       `json:"prefix"`
	CustomFields CustomFields `json:"custom_fields"`
}

type HTTPError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func GetNextPrefixPayload() NextPrefixPayload {
	return NextPrefixPayload{
		PrefixLength: 32,
		CustomFields: CustomFields{
			Domain:  "na",
			Env:     "na",
			Infra:   "na",
			Purpose: "na",
		},
	}
}

func GetUpdatePrefixPayload(nextPrefix responses.NetboxPrefix, mongoPrefix mongodbtypes.Prefix) UpdatePrefixPayload {
	return UpdatePrefixPayload{
		Prefix: nextPrefix.Prefix,
		CustomFields: CustomFields{
			Domain:  "na",
			Env:     "na",
			Infra:   "na",
			Purpose: "na",
			K8suuid: mongoPrefix.ID.Hex(),
		},
	}
}
