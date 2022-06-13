package lagoon

import (
	"fmt"
	"os"

	"sigs.k8s.io/yaml"
)

// ProductionRoutes represents an active/standby configuration.
type ProductionRoutes struct {
	Active  *Environment `json:"active"`
	Standby *Environment `json:"standby"`
}

// Environment represents a Lagoon environment.
type Environment struct {
	AutogenerateRoutes *bool                `json:"autogenerateRoutes"`
	Types              map[string]string    `json:"types"`
	Routes             []map[string][]Route `json:"routes"`
}

// Environments .
type Environments map[string]Environment

// TaskRun .
type TaskRun struct {
	Run Task `json:"run"`
}

// Tasks .
type Tasks struct {
	Prerollout  []TaskRun `json:"pre-rollout"`
	Postrollout []TaskRun `json:"post-rollout"`
}

// YAML represents the .lagoon.yml file.
type YAML struct {
	DockerComposeYAML string            `json:"docker-compose-yaml"`
	Environments      Environments      `json:"environments"`
	ProductionRoutes  *ProductionRoutes `json:"production_routes"`
	Tasks             Tasks             `json:"tasks"`
	Routes            Routes            `json:"routes"`
}

// Routes .
type Routes struct {
	Autogenerate Autogenerate `json:"autogenerate"`
}

// Autogenerate .
type Autogenerate struct {
	Enabled           *bool    `json:"enabled"`
	AllowPullRequests *bool    `json:"allowPullRequests"`
	Insecure          string   `json:"insecure"`
	Prefixes          []string `json:"prefixes"`
	TLSAcme           *bool    `json:"tls-acme,omitempty"`
}

// UnmarshalLagoonYAML unmarshal the lagoon.yml file into a YAML and map for consumption.
func UnmarshalLagoonYAML(file string, l *YAML, p *map[string]interface{}) error {
	rawYAML, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("couldn't read %v: %v", file, err)
	}
	// lagoon.yml
	yaml.Unmarshal(rawYAML, l)
	// polysite
	yaml.Unmarshal(rawYAML, p)
	return nil
}
