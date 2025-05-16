package prefixescontroller

import (
	"net/http"

	"github.com/NorskHelsenett/oss-ipam-api/internal/services/prefixesservice"
	"github.com/NorskHelsenett/oss-ipam-api/pkg/models/apicontracts"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func RegisterPrefix(ginContext *gin.Context) {
	var prefixRequest apicontracts.K8sRequestBody
	validate := validator.New()

	err := ginContext.ShouldBindJSON(&prefixRequest)

	if err != nil {
		ginContext.JSON(http.StatusBadRequest, gin.H{"error": "Could not parse incomming request", "wtf": err})
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
		ginContext.JSON(http.StatusTeapot, gin.H{"error": err.Error()})
		return
	}

	ginContext.JSON(http.StatusOK, response)

}
