package addressesservice

import (
	"errors"
	"net"
	"strconv"
	"strings"

	"github.com/vitistack/ipam-api/internal/logger"
	"github.com/vitistack/ipam-api/internal/services/mongodbservice"
	"github.com/vitistack/ipam-api/internal/services/netboxservice"
	"github.com/vitistack/ipam-api/pkg/models/apicontracts"
)

func RegisterNextAvailable(request apicontracts.IpamApiRequest) (apicontracts.IpamApiResponse, error) {

	container, err := netboxservice.GetAvailablePrefixContainer(request)

	if err != nil {
		return apicontracts.IpamApiResponse{}, err
	}

	payload := apicontracts.GetNextPrefixPayload(request, container)

	nextPrefix, err := netboxservice.GetNextPrefixFromContainer(strconv.Itoa(container.ID), payload)

	if err != nil {
		return apicontracts.IpamApiResponse{}, err
	}

	addressDocument, err := mongodbservice.RegisterAddress(request, nextPrefix)

	if err != nil {
		return apicontracts.IpamApiResponse{}, err
	}

	updatePayload := apicontracts.GetUpdatePrefixPayload(nextPrefix, addressDocument, request)
	err = netboxservice.UpdateNetboxPrefix(strconv.Itoa(nextPrefix.ID), updatePayload)

	if err != nil {
		return apicontracts.IpamApiResponse{}, err
	}

	logger.Log.Infof("Address %s registered successfully in Netbox and MongoDB", request.Address)
	return apicontracts.IpamApiResponse{
		Message: "Address registered successfully",
		Address: nextPrefix.Prefix,
	}, nil

}

func RegisterSpecific(request apicontracts.IpamApiRequest) (apicontracts.IpamApiResponse, error) {
	zone := request.Zone + "_v" + string(request.IpFamily[len(request.IpFamily)-1])
	zonePrefixes := netboxservice.Cache.Get(zone)

	validZonePrefix := false
	for _, prefix := range zonePrefixes {
		_, ipNet, err := net.ParseCIDR(prefix.Prefix)
		if err != nil {
			continue
		}
		ip := net.ParseIP(strings.Split(request.Address, "/")[0]) // Get the base IP without CIDR notation

		if ipNet.Contains(ip) {
			validZonePrefix = true
			break
		}
	}

	if !validZonePrefix {
		return apicontracts.IpamApiResponse{}, errors.New("the requested address is not valid for the provided zone")
	}

	container, err := netboxservice.GetAvailablePrefixContainer(request)

	if err != nil {
		return apicontracts.IpamApiResponse{}, err
	}

	payload := apicontracts.GetCreatePrefixPayload(request, container)
	prefix, err := netboxservice.RegisterPrefix(payload)

	if err != nil {
		return apicontracts.IpamApiResponse{}, err
	}

	addressDocument, err := mongodbservice.RegisterAddress(request, prefix)

	if err != nil {
		return apicontracts.IpamApiResponse{}, err
	}

	updatePayload := apicontracts.GetUpdatePrefixPayload(prefix, addressDocument, request)
	err = netboxservice.UpdateNetboxPrefix(strconv.Itoa(prefix.ID), updatePayload)

	if err != nil {
		logger.Log.Infof("Failed to update %s in Netbox: %v", request.Address, err.Error())
		return apicontracts.IpamApiResponse{}, err
	}

	logger.Log.Infof("Address %s registered successfully in Netbox and MongoDB", request.Address)
	return apicontracts.IpamApiResponse{
		Message: "Address registered successfully",
		Address: request.Address,
	}, nil

}

func Update(request apicontracts.IpamApiRequest) (apicontracts.IpamApiResponse, error) {
	err := mongodbservice.UpdateAddressDocument(request)
	if err != nil {
		return apicontracts.IpamApiResponse{}, err
	}

	logger.Log.Infof("Address %s updated successfully in MongoDB", request.Address)
	return apicontracts.IpamApiResponse{
		Message: "Address updated successfully",
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
		Address: request.Address,
	}, nil
}
