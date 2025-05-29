package apicontracts

import (
	"slices"
	"time"

	"github.com/spf13/viper"
	"github.com/vitistack/ipam-api/internal/responses"
	"github.com/vitistack/ipam-api/pkg/models/mongodbtypes"
)

type Service struct {
	ServiceName         string     `json:"service_name" bson:"service_name" validate:"required" example:"service1"`
	ServiceId           string     `json:"service_id" bson:"service_id" validate:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	ClusterId           string     `json:"cluster_id" bson:"cluster_id" validate:"required,min=8,max=64"`
	RetentionPeriodDays int        `json:"retention_period_days" bson:"retention_period_days"`
	ExpiresAt           *time.Time `json:"expires_at,omitempty" bson:"expires_at,omitempty"`
}

type K8sRequestBody struct {
	Secret   string  `json:"secret" validate:"required,min=8,max=64" example:"a_secret_value"`
	Zone     string  `json:"zone" validate:"required" example:"inet"`
	IpFamily int     `json:"ip_family" bson:"ip_family" validate:"required,oneof=4 6" example:"4"`
	Address  string  `json:"address"`
	Service  Service `json:"service"`
}

type K8sRequestResponse struct {
	Message string `json:"message"`
	Secret  string `json:"secret"`
	Zone    string `json:"zone"`
	Address string `json:"address"`
}

func (r *K8sRequestBody) IsValidZone() bool {
	allowed := []string{"inet", "helsenett-private", "helsenett-public"}
	return slices.Contains(allowed, r.Zone)
}

func (r *K8sRequestBody) ZonePrefixes() []string {
	switch {
	case r.Zone == "inet" && r.IpFamily == 4:
		return viper.GetStringSlice("netbox.prefix_containers.inet_v4")
	case r.Zone == "inet" && r.IpFamily == 6:
		return viper.GetStringSlice("netbox.prefix_containers.inet_v6")
	// case r.Zone == "helsenett-private":
	// 	return viper.GetStringSlice("netbox.prefix_containers.helsenett-private")
	// case r.Zone == "helsenett-public":
	// 	return viper.GetStringSlice("netbox.prefix_containers.helsenett-public")
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

func GetNextPrefixPayload(request K8sRequestBody) NextPrefixPayload {
	var prefixLength int
	if request.IpFamily == 4 {
		prefixLength = 32
	} else if request.IpFamily == 6 {
		prefixLength = 128
	}

	return NextPrefixPayload{
		PrefixLength: prefixLength,
		CustomFields: CustomFields{
			Domain:  "na",
			Env:     "na",
			Infra:   "na",
			Purpose: "na",
		},
	}

}

func GetUpdatePrefixPayload(nextPrefix responses.NetboxPrefix, mongoPrefix mongodbtypes.Address) UpdatePrefixPayload {
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
