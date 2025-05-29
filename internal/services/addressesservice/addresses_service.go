package addressesservice

import (
	"fmt"
	"strconv"

	"github.com/vitistack/ipam-api/internal/services/mongodbservice"
	"github.com/vitistack/ipam-api/internal/services/netboxservice"
	"github.com/vitistack/ipam-api/pkg/models/apicontracts"
)

func Register(request apicontracts.K8sRequestBody) (apicontracts.K8sRequestResponse, error) {
	container, err := netboxservice.GetAvailablePrefixContainer(request)

	if err != nil {
		fmt.Println("Error retrieving prefix container:", err)
		return apicontracts.K8sRequestResponse{}, err
	}

	containerId := container.ID
	payload := apicontracts.GetNextPrefixPayload(request)
	nextPrefix, err := netboxservice.GetNextPrefixFromContainer(strconv.Itoa(containerId), payload)

	if err != nil {
		return apicontracts.K8sRequestResponse{}, err
	}

	addressDocument, err := mongodbservice.RegisterAddress(request, nextPrefix)

	if err != nil {
		return apicontracts.K8sRequestResponse{}, err
	}

	updatePayload := apicontracts.GetUpdatePrefixPayload(nextPrefix, addressDocument)
	err = netboxservice.UpdateNetboxPrefix(strconv.Itoa(nextPrefix.ID), updatePayload)

	if err != nil {
		return apicontracts.K8sRequestResponse{}, err
	}

	return apicontracts.K8sRequestResponse{
		Message: "Address registered successfully",
		Secret:  request.Secret,
		Zone:    request.Zone,
		Address: nextPrefix.Prefix,
	}, nil

}

func Update(request apicontracts.K8sRequestBody) (apicontracts.K8sRequestResponse, error) {
	err := mongodbservice.UpdateAddressDocument(request)
	if err != nil {
		return apicontracts.K8sRequestResponse{}, err
	}

	return apicontracts.K8sRequestResponse{
		Message: "Address updated successfully",
		Secret:  request.Secret,
		Zone:    request.Zone,
		Address: request.Address,
	}, nil
}

func SetServiceExpiration(request apicontracts.K8sRequestBody) (apicontracts.K8sRequestResponse, error) {
	err := mongodbservice.SetServiceExpirationOnAddress(request)
	if err != nil {
		return apicontracts.K8sRequestResponse{}, err
	}

	return apicontracts.K8sRequestResponse{
		Message: "Service expiration set successfully",
		Secret:  request.Secret,
		Zone:    request.Zone,
		Address: request.Address,
	}, nil
}
