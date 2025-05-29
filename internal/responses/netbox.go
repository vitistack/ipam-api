package responses

type NetboxPrefix struct {
	ID           int    `json:"id"`
	Prefix       string `json:"prefix"`
	CustomFields struct {
		K8sUuid string `json:"k8s_uuid"`
	} `json:"custom_fields"`
}

type NetboxPrefixesResponse struct {
	Count    int            `json:"count"`
	Next     string         `json:"next"`
	Previous string         `json:"previous"`
	Results  []NetboxPrefix `json:"results"`
}
