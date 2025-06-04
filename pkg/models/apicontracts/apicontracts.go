package apicontracts

import (
	"time"

	"github.com/vitistack/ipam-api/internal/responses"
	"github.com/vitistack/ipam-api/pkg/models/mongodbtypes"
)

type Service struct {
	ServiceName         string     `json:"service_name" bson:"service_name" validate:"required" example:"service1"`
	NamespaceId         string     `json:"namespace_id" bson:"namespace_id" validate:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	ClusterId           string     `json:"cluster_id" bson:"cluster_id" validate:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	RetentionPeriodDays int        `json:"retention_period_days,omitempty" bson:"retention_period_days,omitempty"`
	ExpiresAt           *time.Time `json:"expires_at,omitempty" bson:"expires_at,omitempty" example:"2025-06-03 14:39:31.546230273"`
	DenyExternalCleanup bool       `json:"deny_external_cleanup,omitempty" bson:"deny_external_cleanup"`
}

type IpamApiRequest struct {
	Secret    string  `json:"secret" validate:"required,min=8,max=64" example:"a_secret_value"`
	Zone      string  `json:"zone" validate:"required" example:"inet"`
	IpFamily  string  `json:"ip_family" bson:"ip_family" validate:"required,oneof=ipv4 ipv6" example:"ipv4"`
	Address   string  `json:"address"`
	Service   Service `json:"service"`
	NewSecret string  `json:"new_secret,omitempty" bson:"new_secret,omitempty"`
}

type IpamApiResponse struct {
	Message string `json:"message"`
	Address string `json:"address"`
}

type CustomFields struct {
	Domain  string `json:"domain"`
	Env     string `json:"env"`
	Infra   string `json:"infra"`
	Purpose string `json:"purpose"`
	K8suuid string `json:"k8s_uuid"`
	K8sZone string `json:"k8s_zone"`
}

type NextPrefixPayload struct {
	PrefixLength int          `json:"prefix_length"`
	VrfId        int          `json:"vrf"`
	TenantId     int          `json:"tenant"`
	RoleId       int          `json:"role"`
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

func GetNextPrefixPayload(request IpamApiRequest, container responses.NetboxPrefix) NextPrefixPayload {
	var prefixLength int
	if request.IpFamily == "ipv4" {
		prefixLength = 32
	} else if request.IpFamily == "ipv6" {
		prefixLength = 128
	}

	return NextPrefixPayload{
		PrefixLength: prefixLength,
		VrfId:        container.Vrf.ID,
		TenantId:     container.Tenant.ID,
		RoleId:       container.Role.ID,
		CustomFields: CustomFields{
			Domain:  "na",
			Env:     "na",
			Infra:   container.CustomFields.Infra,
			Purpose: "na",
			K8sZone: request.Zone,
		},
	}

}

func GetUpdatePrefixPayload(nextPrefix responses.NetboxPrefix, mongoPrefix mongodbtypes.Address, request IpamApiRequest) UpdatePrefixPayload {
	return UpdatePrefixPayload{
		Prefix: nextPrefix.Prefix,
		CustomFields: CustomFields{
			Domain:  "na",
			Env:     "na",
			Infra:   nextPrefix.CustomFields.Infra,
			Purpose: "na",
			K8suuid: mongoPrefix.ID.Hex(),
			K8sZone: request.Zone,
		},
	}
}
