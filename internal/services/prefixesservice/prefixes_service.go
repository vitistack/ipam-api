package prefixesservice

import (
	"fmt"
	"strconv"

	"github.com/NorskHelsenett/oss-ipam-api/internal/services/mongodbservice"
	"github.com/NorskHelsenett/oss-ipam-api/internal/services/netboxservice"
	"github.com/NorskHelsenett/oss-ipam-api/pkg/models/apicontracts"
)

func Register(request apicontracts.K8sRequestBody) (any, error) {
	if request.Prefix == "" {
		container, err := netboxservice.GetPrefixContainer(request.ZonePrefix())

		if err != nil {
			fmt.Println("Error retrieving prefix container:", err)
			return nil, err
		}

		fmt.Println("Container ID:", container.ID)
		containerId := container.ID
		createPayload := map[string]any{
			"prefix_length": 32,
			"custom_fields": map[string]any{
				"domain":  "na",
				"env":     "na",
				"infra":   "na",
				"purpose": "na",
			},
		}
		nextPrefix, err := netboxservice.GetNextPrefixFromContainer(strconv.Itoa(containerId), createPayload)

		if err != nil {
			fmt.Println("Error retrieving next prefix:", err)
			return nil, err
		}

		prefixDocument, err := mongodbservice.InsertNewPrefixDocument(request, nextPrefix)

		if err != nil {
			return nil, err
		}

		updatePayload := map[string]any{
			"prefix": nextPrefix.Prefix,
			"custom_fields": map[string]any{
				"k8s_uuid": prefixDocument.ID,
			},
		}

		err = netboxservice.UpdateNetboxPrefix(strconv.Itoa(nextPrefix.ID), updatePayload)

		if err != nil {
			return nil, err
		}

		return nextPrefix, nil
	} else {
		return nil, nil
	}

}
