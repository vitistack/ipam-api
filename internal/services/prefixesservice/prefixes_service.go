package prefixesservice

import (
	"fmt"
	"strconv"

	"github.com/NorskHelsenett/oss-ipam-api/internal/services/mongodbservice"
	"github.com/NorskHelsenett/oss-ipam-api/internal/services/netboxservice"
	"github.com/NorskHelsenett/oss-ipam-api/pkg/models/apicontracts"
)

func Register(request apicontracts.K8sRequestBody) (apicontracts.K8sRequestResponse, error) {
	container, err := netboxservice.GetAvailablePrefixContainer(request)

	if err != nil {
		fmt.Println("Error retrieving prefix container:", err)
		return apicontracts.K8sRequestResponse{}, err
	}

	containerId := container.ID
	payload := apicontracts.GetNextPrefixPayload()
	nextPrefix, err := netboxservice.GetNextPrefixFromContainer(strconv.Itoa(containerId), payload)

	if err != nil {
		return apicontracts.K8sRequestResponse{}, err
	}

	prefixDocument, err := mongodbservice.InsertNewPrefixDocument(request, nextPrefix)

	if err != nil {
		return apicontracts.K8sRequestResponse{}, err
	}

	updatePayload := apicontracts.GetUpdatePrefixPayload(nextPrefix, prefixDocument)
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

func Update(request apicontracts.K8sRequestBody) (apicontracts.K8sRequestResponse, error) {
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
