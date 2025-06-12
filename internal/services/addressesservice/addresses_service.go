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

// RegisterNextAvailable registers the next available address prefix for a given IPAM API request.
// It performs the following steps:
//  1. Retrieves the available prefix container from Netbox based on the request.
//  2. Constructs the payload to request the next available prefix.
//  3. Obtains the next available prefix from the container.
//  4. Registers the new address in MongoDB.
//  5. Updates the prefix information in Netbox with the new address details.
//  6. Logs the successful registration and returns a response containing the registered address.
//
// Returns an IpamApiResponse with the registered address on success, or an error if any step fails.
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

// RegisterSpecific registers a specific IP address within a given zone and IP family.
// It validates that the requested address belongs to a valid prefix for the specified zone,
// retrieves an available prefix container, registers the prefix in Netbox, and stores the address
// in MongoDB. The function also updates the prefix in Netbox with the new address information.
// Returns a successful IpamApiResponse if the operation completes, or an error if any step fails.
//
// Parameters:
//   - request: apicontracts.IpamApiRequest containing the address, zone, and IP family information.
//
// Returns:
//   - apicontracts.IpamApiResponse: Response containing a success message and the registered address.
//   - error: Error if the address is invalid for the zone or if any registration step fails.
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

// Update updates an address document in the MongoDB database based on the provided IpamApiRequest.
// It returns an IpamApiResponse containing a success message and the updated address if the operation succeeds.
// If an error occurs during the update, it returns an empty response and the error.
//
// Parameters:
//   - request: apicontracts.IpamApiRequest containing the address and update details.
//
// Returns:
//   - apicontracts.IpamApiResponse: Response with a success message and the updated address.
//   - error: Error encountered during the update operation, if any.
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

// SetServiceExpiration sets the service expiration on an address using the provided IpamApiRequest.
// It calls the mongodbservice to update the expiration and returns an IpamApiResponse with a success message
// and the address if successful, or an error if the operation fails.
//
// Parameters:
//   - request: apicontracts.IpamApiRequest containing the address and expiration details.
//
// Returns:
//   - apicontracts.IpamApiResponse: Response containing a message and the address.
//   - error: Error if setting the expiration fails.
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
