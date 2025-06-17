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

// RegisterAddress handles the registration of an IP address based on the provided IpamApiRequest.
// It checks if the service is already registered in MongoDB, verifies address availability in Netbox,
// and determines the appropriate registration or update action to perform:
//   - If no address is provided and not registered, it registers the next available address.
//   - If no address is provided but already registered, it updates the registration with the existing address.
//   - If an address is provided and available in Netbox, it registers the specific address.
//   - Otherwise, it updates the registration as default.
//
// Returns an IpamApiResponse and an error if any operation fails.
func RegisterAddress(request apicontracts.IpamApiRequest) (apicontracts.IpamApiResponse, error) {
	alreadyRegistered, err := mongodbservice.ServiceAlreadyRegistered(request)
	if err != nil {
		logger.Log.Errorf("Failed to check if service is already registered: %v", err)
		return apicontracts.IpamApiResponse{}, err
	}

	availableInNetbox := false
	if request.Address != "" {
		vrf := "nhc"
		if request.Zone == "inet" {
			vrf = "inet"
		}
		queryParams := map[string]string{
			"prefix":         request.Address,
			"present_in_vrf": vrf,
		}
		availableInNetbox, err = netboxservice.PrefixAvailable(queryParams)
		if err != nil {
			return apicontracts.IpamApiResponse{}, err
		}
	}

	if request.Address == "" && alreadyRegistered.Address == "" {
		// Not registered in MongoDB and no address provided
		response, err := RegisterNextAvailable(request)
		if err != nil {
			return apicontracts.IpamApiResponse{}, err
		}
		return response, nil
	} else if request.Address == "" && alreadyRegistered.Address != "" {
		// Already registered in MongoDB and no address provided
		request.Address = alreadyRegistered.Address
		response, err := Update(request)
		if err != nil {
			logger.Log.Errorf("Failed to register update address: %v", err)
			return apicontracts.IpamApiResponse{}, err
		}
		return response, nil
	} else if availableInNetbox && request.Address != "" {
		// Address is available in Netbox and provided in the request
		response, err := RegisterSpecific(request)
		if err != nil {
			logger.Log.Errorf("Failed to register specific address: %v", err)
			return apicontracts.IpamApiResponse{}, err
		}
		return response, nil
	} else {
		// Update as default
		response, err := Update(request)
		if err != nil {
			logger.Log.Errorf("Failed to update address: %v", err)
			return apicontracts.IpamApiResponse{}, err
		}
		return response, nil
	}
}

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

	logger.Log.Infof("Address %s registered successfully", nextPrefix.Prefix)
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
	address, err := mongodbservice.UpdateAddressDocument(request)
	if err != nil {
		return apicontracts.IpamApiResponse{}, err
	}

	logger.Log.Infof("Address %s updated successfully", request.Address)
	return apicontracts.IpamApiResponse{
		Message: "Address updated successfully",
		Address: address.Address,
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
