package lagoon

// RoutesV2 is the new routes definition
type RoutesV2 struct {
	Routes []RouteV2 `json:"routes"`
}

// RouteV2 is the new route definition
type RouteV2 struct {
	Domain         string            `json:"domain"`
	Service        string            `json:"service"`
	TLSAcme        *bool             `json:"tls-acme"`
	Migrate        *bool             `json:"migrate,omitempty"`
	Insecure       *string           `json:"insecure,omitempty"`
	HSTS           *string           `json:"hsts,omitempty"`
	MonitoringPath string            `json:"monitoring-path,omitempty"`
	Fastly         Fastly            `json:"fastly,omitempty"`
	Annotations    map[string]string `json:"annotations"`
}
