package prefixesservice

import (
	"fmt"
	"strconv"

	"github.com/NorskHelsenett/oss-ipam-api/internal/services/mongodbservice"
	"github.com/NorskHelsenett/oss-ipam-api/internal/services/netboxservice"
	"github.com/NorskHelsenett/oss-ipam-api/pkg/models/apicontracts"
)

func Register(request apicontracts.K8sRequestBody) (apicontracts.K8sRequestResponse, error) {
	if request.Prefix == "" {
		container, err := netboxservice.GetPrefixContainer(request.ZonePrefix())

		if err != nil {
			fmt.Println("Error retrieving prefix container:", err)
			return apicontracts.K8sRequestResponse{}, err
		}

		containerId := container.ID
		createPayload := map[string]any{
			"prefix_length": 32,
			"custom_fields": map[string]any{
				"domain":  "na",
				"env":     "na",
				"infra":   "na",
				"purpose": "na",
			},
		}
		nextPrefix, err := netboxservice.GetNextPrefixFromContainer(strconv.Itoa(containerId), createPayload)

		if err != nil {
			return apicontracts.K8sRequestResponse{}, err
		}

		prefixDocument, err := mongodbservice.InsertNewPrefixDocument(request, nextPrefix)

		if err != nil {
			return apicontracts.K8sRequestResponse{}, err
		}

		updatePayload := map[string]any{
			"prefix": nextPrefix.Prefix,
			"custom_fields": map[string]any{
				"k8s_uuid": prefixDocument.ID,
			},
		}

		err = netboxservice.UpdateNetboxPrefix(strconv.Itoa(nextPrefix.ID), updatePayload)

		if err != nil {
			return apicontracts.K8sRequestResponse{}, err
		}

		return apicontracts.K8sRequestResponse{
			Message: "Prefix registered successfully",
			Secret:  request.Secret,
			Zone:    request.Zone,
			Prefix:  nextPrefix.Prefix,
		}, nil
	}

	err := mongodbservice.UpdatePrefixDocument(request)
	if err != nil {
		return apicontracts.K8sRequestResponse{}, err
	}

	return apicontracts.K8sRequestResponse{
		Message: "Prefix updated successfully",
		Secret:  request.Secret,
		Zone:    request.Zone,
		Prefix:  request.Prefix,
	}, nil
}

func Deregister(request apicontracts.K8sRequestBody) (apicontracts.K8sRequestResponse, error) {
	err := mongodbservice.DeleteServiceFromPrefix(request)
	if err != nil {
		return apicontracts.K8sRequestResponse{}, err
	}

	return apicontracts.K8sRequestResponse{
		Message: "Service deregistered successfully",
		Secret:  request.Secret,
		Zone:    request.Zone,
		Prefix:  request.Prefix,
	}, nil
}
