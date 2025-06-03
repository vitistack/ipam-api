package addressesservice

import (
	"strconv"

	"github.com/vitistack/ipam-api/internal/services/mongodbservice"
	"github.com/vitistack/ipam-api/internal/services/netboxservice"
	"github.com/vitistack/ipam-api/pkg/models/apicontracts"
)

func Register(request apicontracts.IpamApiRequest) (apicontracts.IpamApiResponse, error) {
	container, err := netboxservice.GetAvailablePrefixContainer(request)

	if err != nil {
		return apicontracts.IpamApiResponse{}, err
	}

	payload := apicontracts.GetNextPrefixPayload(request)

	nextPrefix, err := netboxservice.GetNextPrefixFromContainer(strconv.Itoa(container.ID), payload)

	if err != nil {
		return apicontracts.IpamApiResponse{}, err
	}

	addressDocument, err := mongodbservice.RegisterAddress(request, nextPrefix)

	if err != nil {
		return apicontracts.IpamApiResponse{}, err
	}

	updatePayload := apicontracts.GetUpdatePrefixPayload(nextPrefix, addressDocument)
	err = netboxservice.UpdateNetboxPrefix(strconv.Itoa(nextPrefix.ID), updatePayload)

	if err != nil {
		return apicontracts.IpamApiResponse{}, err
	}

	return apicontracts.IpamApiResponse{
		Message: "Address registered successfully",
		Secret:  request.Secret,
		Zone:    request.Zone,
		Address: nextPrefix.Prefix,
	}, nil

}

func Update(request apicontracts.IpamApiRequest) (apicontracts.IpamApiResponse, error) {
	err := mongodbservice.UpdateAddressDocument(request)
	if err != nil {
		return apicontracts.IpamApiResponse{}, err
	}

	return apicontracts.IpamApiResponse{
		Message: "Address updated successfully",
		Secret:  request.Secret,
		Zone:    request.Zone,
		Address: request.Address,
	}, nil
}

func SetServiceExpiration(request apicontracts.IpamApiRequest) (apicontracts.IpamApiResponse, error) {
	err := mongodbservice.SetServiceExpirationOnAddress(request)
	if err != nil {
		return apicontracts.IpamApiResponse{}, err
	}

	return apicontracts.IpamApiResponse{
		Message: "Service expiration set successfully",
		Secret:  request.Secret,
		Zone:    request.Zone,
		Address: request.Address,
	}, nil
}
