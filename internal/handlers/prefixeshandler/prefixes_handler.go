package prefixeshandler

import (
	"net/http"

	"github.com/NorskHelsenett/oss-ipam-api/internal/services/prefixesservice"
	"github.com/NorskHelsenett/oss-ipam-api/pkg/models/apicontracts"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// RegisterPrefix godoc
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
func RegisterPrefix(ginContext *gin.Context) {
	var request apicontracts.K8sRequestBody
	validate := validator.New()

	err := ginContext.ShouldBindJSON(&request)

	if err != nil {
		ginContext.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse incomming request"})
		return
	}

	// if prefixRequest.Address != "" {
	// 	ginContext.JSON(http.StatusBadRequest, gin.H{"message": "For updating a prefix, use the PUT method"})
	// 	return
	// }

	if !request.IsValidZone() {
		ginContext.JSON(http.StatusBadRequest, gin.H{"message": "Invalid zone"})
		return
	}

	if request.Zone == "" || request.Secret == "" {
		ginContext.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request. Both 'zone' and 'secret' are required"})
		return
	}

	err = validate.Struct(request)

	if err != nil {
		ginContext.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request: " + err.Error()})
		return
	}

	var response apicontracts.K8sRequestResponse
	if request.Address != "" {
		response, err = prefixesservice.Update(request)
	} else {
		response, err = prefixesservice.Register(request)
	}

	if err != nil {
		ginContext.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	ginContext.JSON(http.StatusOK, response)

}

// RegisterPrefix godoc
// @Summary      Update a prefix
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
// @Router       /prefixes [PATCH]
// @Router       /prefixes [PUT]
func UpdatePrefix(ginContext *gin.Context) {
	var prefixRequest apicontracts.K8sRequestBody
	validate := validator.New()

	err := ginContext.ShouldBindJSON(&prefixRequest)

	if err != nil {
		ginContext.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse incomming request"})
		return
	}

	if !prefixRequest.IsValidZone() {
		ginContext.JSON(http.StatusBadRequest, gin.H{"message": "Invalid zone"})
		return
	}

	if prefixRequest.Zone == "" || prefixRequest.Secret == "" || prefixRequest.Address == "" {
		ginContext.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request. 'zone', 'secret' and 'prefix' are required for update"})
		return
	}

	err = validate.Struct(prefixRequest)

	if err != nil {
		ginContext.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request: " + err.Error()})
		return
	}

	response, err := prefixesservice.Update(prefixRequest)

	if err != nil {
		ginContext.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	ginContext.JSON(http.StatusOK, response)

}

// DeregisterPrefix godoc
// @Summary      Deregister a prefix
// @Schemes
// @Description  Deregister a prefix
// @Tags         prefixes
// @Accept       json
// @Produce      json
// @Param		 body body apicontracts.K8sRequestBody true "Request body"
// @Failure      400 {object} apicontracts.HTTPError
// @Failure      404 {object} apicontracts.HTTPError
// @Failure      500 {object} apicontracts.HTTPError
// @Router       /prefixes [DELETE]
func DeregisterPrefix(ginContext *gin.Context) {
	var prefixRequest apicontracts.K8sRequestBody
	validate := validator.New()

	err := ginContext.ShouldBindJSON(&prefixRequest)

	if err != nil {
		ginContext.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse incomming request"})
		return
	}

	if !prefixRequest.IsValidZone() {
		ginContext.JSON(http.StatusBadRequest, gin.H{"message": "Invalid zone"})
		return
	}

	if prefixRequest.Zone == "" || prefixRequest.Secret == "" {
		ginContext.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request. Both 'zone' and 'secret' are required"})
		return
	}

	err = validate.Struct(prefixRequest)

	if err != nil {
		ginContext.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request: " + err.Error()})
		return
	}

	response, err := prefixesservice.Deregister(prefixRequest)

	if err != nil {
		ginContext.JSON(http.StatusNotFound, gin.H{"message": "Could not deregister service: " + err.Error()})
		return
	}

	ginContext.JSON(http.StatusOK, response)

}
