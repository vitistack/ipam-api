package responses

type NetboxPrefix struct {
	ID     int    `json:"id"`
	Prefix string `json:"prefix"`
	// Description  string `json:"description"`
	CustomFields struct {
		K8sUuid string `json:"k8s_uuid"`
	} `json:"custom_fields"`
}

type NetboxPrefixes struct {
	Count    int            `json:"count"`
	Next     string         `json:"next"`
	Previous string         `json:"previous"`
	Results  []NetboxPrefix `json:"results"`
}
