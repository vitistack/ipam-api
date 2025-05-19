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
	var prefixRequest apicontracts.K8sRequestBody
	validate := validator.New()

	err := ginContext.ShouldBindJSON(&prefixRequest)

	if err != nil {
		ginContext.JSON(http.StatusBadRequest, gin.H{"error": "Could not parse incomming request"})
		return
	}

	if !prefixRequest.IsValidZone() {
		ginContext.JSON(http.StatusBadRequest, gin.H{"error": "Invalid zone"})
		return
	}

	if prefixRequest.Zone == "" || prefixRequest.Secret == "" {
		ginContext.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request. Both 'zone' and 'secret' are required"})
		return
	}

	err = validate.Struct(prefixRequest)

	if err != nil {
		ginContext.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	response, err := prefixesservice.Register(prefixRequest)

	if err != nil {
		ginContext.JSON(http.StatusTeapot, gin.H{"error": err.Error()})
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
// @Router       /prefixes [DELETE]
func DeregisterPrefix(ginContext *gin.Context) {
	var prefixRequest apicontracts.K8sRequestBody
	validate := validator.New()

	err := ginContext.ShouldBindJSON(&prefixRequest)

	if err != nil {
		ginContext.JSON(http.StatusBadRequest, gin.H{"error": "Could not parse incomming request"})
		return
	}

	if !prefixRequest.IsValidZone() {
		ginContext.JSON(http.StatusBadRequest, gin.H{"error": "Invalid zone"})
		return
	}

	if prefixRequest.Zone == "" || prefixRequest.Secret == "" {
		ginContext.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request. Both 'zone' and 'secret' are required"})
		return
	}

	err = validate.Struct(prefixRequest)

	if err != nil {
		ginContext.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	response, err := prefixesservice.Deregister(prefixRequest)

	if err != nil {
		ginContext.JSON(http.StatusBadRequest, apicontracts.HTTPError{Code: http.StatusTeapot, Message: err.Error()})
		return
	}

	ginContext.JSON(http.StatusOK, response)

}
