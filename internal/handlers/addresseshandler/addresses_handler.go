package addresseshandler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/vitistack/ipam-api/internal/services/addressesservice"
	"github.com/vitistack/ipam-api/pkg/models/apicontracts"
)

// RegisterAddress godoc
// @Summary      Register a prefix
// @Schemes
// @Description  Register a prefix
// @Tags         prefixes
// @Accept       json
// @Produce      json
// @Success		 200 {object} apicontracts.K8sRequestResponse
// @Param		 body body apicontracts.K8sRequestBody true "Request body"
// @Failure      400 {object} apicontracts.HTTPError
// @Failure      404 {object} apicontracts.HTTPError
// @Failure      500 {object} apicontracts.HTTPError
// @Router       /prefixes [POST]
func RegisterAddress(ginContext *gin.Context) {
	var request apicontracts.IpamApiRequest
	err := ginContext.ShouldBindJSON(&request)

	if err != nil {
		ginContext.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse incomming request"})
		return
	}

	err = ValidateRequest(&request)

	if err != nil {
		ginContext.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	var response apicontracts.IpamApiResponse
	if request.Address != "" {
		response, err = addressesservice.Update(request)
	} else {
		response, err = addressesservice.Register(request)
	}

	if err != nil {
		ginContext.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	ginContext.JSON(http.StatusOK, response)

}

// ExpireAddress godoc
// @Summary      Set expiration for a service
// @Schemes
// @Description  Set expiration for a service
// @Tags         addresses
// @Accept       json
// @Produce      json
// @Param		 body body apicontracts.K8sRequestBody true "Request body"
// @Failure      400 {object} apicontracts.HTTPError
// @Failure      404 {object} apicontracts.HTTPError
// @Failure      500 {object} apicontracts.HTTPError
// @Router       / [DELETE]
func ExpireAddress(ginContext *gin.Context) {
	var prefixRequest apicontracts.IpamApiRequest

	err := ginContext.ShouldBindJSON(&prefixRequest)

	if err != nil {
		ginContext.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse incomming request"})
		return
	}

	err = ValidateRequest(&prefixRequest)

	if err != nil {
		ginContext.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	response, err := addressesservice.SetServiceExpiration(prefixRequest)

	if err != nil {
		ginContext.JSON(http.StatusNotFound, gin.H{"message": "Could not deregister service: " + err.Error()})
		return
	}

	ginContext.JSON(http.StatusOK, response)

}

func ValidateRequest(request *apicontracts.IpamApiRequest) error {
	validate := validator.New()

	if request.Zone == "" || request.Secret == "" {
		return errors.New("both 'zone' and 'secret' are required")
	}

	if !request.IsValidZone() {
		return errors.New("invalid zone")
	}

	if request.IpFamily == "ipv6" && request.Zone != "inet" {
		return errors.New("IPv6 is only available for zone 'inet'")
	}

	err := validate.Struct(*request)

	if err != nil {
		return err
	}

	return nil
}
