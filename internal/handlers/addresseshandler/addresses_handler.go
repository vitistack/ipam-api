package addresseshandler

import (
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/vitistack/ipam-api/internal/logger"
	"github.com/vitistack/ipam-api/internal/services/addressesservice"
	"github.com/vitistack/ipam-api/internal/services/netboxservice"
	"github.com/vitistack/ipam-api/internal/utils"
	"github.com/vitistack/ipam-api/pkg/models/apicontracts"
)

// RegisterAddress godoc
//
//	@Summary	Register an address
//	@Schemes
//	@Description	Register an address in Vitistack IPAM API
//	@Tags			addresses
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}	apicontracts.IpamApiResponse
//	@Param			body	body		apicontracts.IpamApiRequest	true	"Request body"
//	@Failure		400		{object}	apicontracts.HTTPError
//	@Failure		404		{object}	apicontracts.HTTPError
//	@Failure		500		{object}	apicontracts.HTTPError
//	@Router			/ [POST]
func RegisterAddress(ginContext *gin.Context) {
	var request apicontracts.IpamApiRequest
	err := ginContext.ShouldBindJSON(&request)

	if err != nil {
		ginContext.Error(err)
		ginContext.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse incomming request"})
		return
	}

	err = ValidateRequest(&request)

	if err != nil {
		logger.Log.Errorf("Request validation failed: %v", err)
		ginContext.Error(err)
		ginContext.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	var response apicontracts.IpamApiResponse
	httpStatus := http.StatusOK

	if err != nil {
		ginContext.Error(err)
		ginContext.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	response, err = addressesservice.RegisterAddress(request)
	if err != nil {
		logger.Log.Errorf("Failed to register address: %v", err)
		ginContext.Error(err)
		ginContext.JSON(http.StatusInternalServerError, gin.H{"message": "Could not register address: " + err.Error()})
		return
	}

	// if alreadyRegistered.Address != "" && request.Address == "" {
	// 	logger.Log.Warnln("Registered in mongodb and no address in request -> Update()")
	// 	request.Address = alreadyRegistered.Address
	// 	// If the address is already registered in MongoDB, we update it with the request address
	// 	response, err = addressesservice.Update(request)
	// } else if request.Address == "" {
	// 	logger.Log.Warnln("No address provided -> RegisterAddress()")
	// 	response, err = addressesservice.RegisterAddress(request)
	// 	httpStatus = http.StatusCreated
	// } else if alreadyRegistered.Address != "" && request.Address != "" {
	// 	logger.Log.Warnln("Registered in mongodb and address in request -> RegisterAddress()")
	// 	response, err = addressesservice.RegisterAddress(request)
	// 	httpStatus = http.StatusCreated
	// } else if alreadyRegistered.Address != "" && request.Address != "" {
	// 	logger.Log.Warnln("Registered in mongodb and address in request -> RegisterAddress()")
	// 	response, err = addressesservice.RegisterAddress(request)
	// 	httpStatus = http.StatusCreated
	// } else if alreadyRegistered.Address == "" && request.Address != "" {
	// 	logger.Log.Warnln("Not registered in mongodb and address in request -> RegisterAddress()")
	// 	// If the address is not registered in MongoDB, we register it with the request address
	// 	response, err = addressesservice.RegisterAddress(request)
	// 	httpStatus = http.StatusCreated
	// } else {
	// 	logger.Log.Warnln("default case -> Update()")
	// 	response, err = addressesservice.Update(request)
	// }

	if err != nil {
		ginContext.Error(err)
		ginContext.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	ginContext.JSON(httpStatus, response)

}

// ExpireAddress godoc
//
//	@Summary	Set expiration for a service
//	@Schemes
//	@Description	Set expiration for a service
//	@Tags			addresses
//	@Accept			json
//	@Produce		json
//	@Param			body	body		apicontracts.IpamApiRequest	true	"Request body"
//	@Success		200		{object}	apicontracts.IpamApiResponse
//	@Failure		400		{object}	apicontracts.HTTPError
//	@Failure		404		{object}	apicontracts.HTTPError
//	@Failure		500		{object}	apicontracts.HTTPError
//	@Router			/ [DELETE]
func ExpireAddress(ginContext *gin.Context) {
	var prefixRequest apicontracts.IpamApiRequest

	err := ginContext.ShouldBindJSON(&prefixRequest)

	if err != nil {
		ginContext.Error(err)
		ginContext.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse incomming request"})
		return
	}

	err = ValidateRequest(&prefixRequest)

	if err != nil {
		ginContext.Error(err)
		ginContext.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	response, err := addressesservice.SetServiceExpiration(prefixRequest)

	if err != nil {
		ginContext.Error(err)
		ginContext.JSON(http.StatusNotFound, gin.H{"message": "Could not deregister service: " + err.Error()})
		return
	}

	ginContext.JSON(http.StatusOK, response)

}

func ValidateRequest(request *apicontracts.IpamApiRequest) error {
	validate := validator.New()

	netboxZones, err := netboxservice.GetK8sZones()

	if err != nil {
		return errors.New("failed to fetch zones: " + err.Error())
	}

	err = validate.Struct(*request)

	if err != nil {
		if err.Error() == "Key: 'IpamApiRequest.IpFamily' Error:Field validation for 'IpFamily' failed on the 'oneof' tag" {
			return errors.New("invalid ip family, must be either 'ipv4' or 'ipv6'")
		}
		return err
	}

	if request.Service.RetentionPeriodDays > 30 {
		return errors.New("retention period cannot be more than 30 days")
	}

	if request.Zone == "" || request.Secret == "" {
		return errors.New("both 'zone' and 'secret' are required")
	}
	if !slices.Contains(netboxZones, request.Zone) {
		return errors.New("invalid zone '" + request.Zone + "', must be one of: " + strings.Join(netboxZones, ", "))
	}

	if request.Address != "" {
		prefixIpFamily, err := utils.IPFamilyFromPrefix(request.Address)

		if err != nil {
			return err
		}

		if prefixIpFamily != request.IpFamily {
			return errors.New("invalid ip familiy for the provided address")
		}
	}

	if request.IpFamily == "ipv6" && request.Zone != "inet" {
		return errors.New("IPv6 is only available for zone 'inet'")
	}

	return nil
}
