package lagoon

import (
	"fmt"
	"os"

	"sigs.k8s.io/yaml"
)

// UnmarshaDockerComposeYAML unmarshal the lagoon.yml file into a YAML and map for consumption.
func UnmarshaDockerComposeYAML(file string, l *Compose) error {
	rawYAML, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("couldn't read %v: %v", file, err)
	}
	// docker-compose.yml
	yaml.Unmarshal(rawYAML, l)
	return nil
}

// CheckServiceLagoonLabel checks the labels in a compose service to see if the requested label exists, and returns the value if so
func CheckServiceLagoonLabel(labels map[string]string, label string) string {
	for k, v := range labels {
		if k == label {
			return v
		}
	}
	return ""
}

// Compose .
type Compose struct {
	Services map[string]Service `json:"services"`
}

// Service .
type Service struct {
	Build  ServiceBuild      `json:"build"`
	Labels map[string]string `json:"labels"`
	// Image  string            `json:"image"` //@TODO: is this used by lagoon builds?
}

// ServiceBuild .
type ServiceBuild struct {
	Context    string `json:"context"`
	Dockerfile string `json:"dockerfile"`
	// Args       map[string]string `json:"args"` //@TODO: is this used by lagoon builds?
}
