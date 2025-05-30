package responses

type NetboxPrefix struct {
	ID     int    `json:"id"`
	Prefix string `json:"prefix"`
	Family struct {
		Value int `json:"value"`
	}
	CustomFields struct {
		K8sUuid string `json:"k8s_uuid"`
	} `json:"custom_fields"`
}

func (n NetboxPrefix) GetIpFamily() int {
	if n.Family.Value == 4 {
		return 4
	} else if n.Family.Value == 6 {
		return 6
	}
	return 0
}

// type NetboxPrefixesResponse struct {
// 	Count    int            `json:"count"`
// 	Next     string         `json:"next"`
// 	Previous string         `json:"previous"`
// 	Results  []NetboxPrefix `json:"results"`
// }

type NetboxChoiceSet struct {
	ChoicesCount int        `json:"choices_count"`
	ExtraChoices [][]string `json:"extra_choices"`
	ID           int        `json:"id"`
	Name         string     `json:"name"`
}

// type NetboxChoiceSetsResponse struct {
// 	Count    int               `json:"count"`
// 	Next     string            `json:"next"`
// 	Previous string            `json:"previous"`
// 	Results  []NetboxChoiceSet `json:"results"`
// }

type NetboxResponse[T any] struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []T    `json:"results"`
}
