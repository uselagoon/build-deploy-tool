package lagoon

// ProductionRoutes represents an active/standby configuration.
type ProductionRoutes struct {
	Active  *Environment `json:"active"`
	Standby *Environment `json:"standby"`
}

// Environment represents a Lagoon environment.
type Environment struct {
	Routes []map[string][]Route `json:"routes"`
}

// Environments .
type Environments map[string]Environment

// YAML represents the .lagoon.yml file.
type YAML struct {
	DockerComposeYAML string            `json:"docker-compose-yaml"`
	Environments      Environments      `json:"environments"`
	ProductionRoutes  *ProductionRoutes `json:"production_routes"`
}
