package responses

type NetboxPrefix struct {
	ID     int    `json:"id"`
	Prefix string `json:"prefix"`
	Family struct {
		Value int `json:"value"`
	}
	Vrf struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	Tenant struct {
		ID int `json:"id"`
	}
	Role struct {
		ID int `json:"id"`
	}
	CustomFields struct {
		Infra   string `json:"infra"`
		K8sUuid string `json:"k8s_uuid"`
		K8sZone string `json:"k8s_zone"`
	} `json:"custom_fields"`
}

func (n NetboxPrefix) GetIpFamily() int {
	switch n.Family.Value {
	case 4:
		return 4
	case 6:
		return 6
	}
	return 0
}

type NetboxChoiceSet struct {
	ChoicesCount int        `json:"choices_count"`
	ExtraChoices [][]string `json:"extra_choices"`
	ID           int        `json:"id"`
	Name         string     `json:"name"`
}

type NetboxTag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type NetboxResponse[T any] struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []T    `json:"results"`
}
