package lagoon

import (
	"github.com/compose-spec/compose-go/cli"
	"github.com/compose-spec/compose-go/loader"
	composetypes "github.com/compose-spec/compose-go/types"
)

// UnmarshaDockerComposeYAML unmarshal the lagoon.yml file into a YAML and map for consumption.
func UnmarshaDockerComposeYAML(file string, ignoreErrors bool, envvars map[string]string) (*composetypes.Project, error) {
	options, err := cli.NewProjectOptions([]string{file},
		cli.WithResolvedPaths(false),
		cli.WithLoadOptions(
			loader.WithSkipValidation,
			func(o *loader.Options) {
				o.IgnoreNonStringKeyErrors = ignoreErrors
			},
		),
	)
	options.Environment = envvars
	l, err := cli.ProjectFromOptions(options)
	if err != nil {
		return nil, err
	}
	return l, nil
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
