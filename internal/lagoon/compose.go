package lagoon

import (
	"fmt"
	"os"

	"github.com/compose-spec/compose-go/cli"
	"github.com/compose-spec/compose-go/loader"
	composetypes "github.com/compose-spec/compose-go/types"
	goyaml "gopkg.in/yaml.v2"
)

type OriginalServiceOrder struct {
	Name  string
	Index int
}

// UnmarshaDockerComposeYAML unmarshal the lagoon.yml file into a YAML and map for consumption.
func UnmarshaDockerComposeYAML(file string, ignoreErrors, ignoreMissingEnvFiles bool, envvars map[string]string) (*composetypes.Project, []OriginalServiceOrder, error) {
	options, err := cli.NewProjectOptions([]string{file},
		cli.WithResolvedPaths(false),
		cli.WithLoadOptions(
			loader.WithSkipValidation,
			loader.WithDiscardEnvFiles,
			func(o *loader.Options) {
				o.IgnoreNonStringKeyErrors = ignoreErrors
				o.IgnoreMissingEnvFileCheck = ignoreMissingEnvFiles
			},
		),
	)
	options.Environment = envvars
	l, err := cli.ProjectFromOptions(options)
	if err != nil {
		return nil, nil, err
	}
	originalOrder, err := UnmarshalLagoonDockerComposeYAML(file)
	if err != nil {
		return nil, nil, err
	}
	return l, originalOrder, nil
}

// UnmarshalLagoonDockerComposeYAML unmarshal the docker-compose.yml file into a YAML and map for consumption.
// this uses yaml mapslice to preserve the order of the services in the docker-compose file
// as lagoon relies on this order for building images and determining the order of routes
func UnmarshalLagoonDockerComposeYAML(file string) ([]OriginalServiceOrder, error) {
	rawYAML, err := os.ReadFile(file)
	l := []OriginalServiceOrder{}
	if err != nil {
		return nil, fmt.Errorf("couldn't read %v: %v", file, err)
	}
	// unmarshal docker-compose.yml
	// use to gopkg yaml v2 for MapSlice
	m := goyaml.MapSlice{}
	goyaml.Unmarshal(rawYAML, &m)
	for _, item := range m {
		// extract the services only
		if item.Key.(string) == "services" {
			for idx, v := range item.Value.(goyaml.MapSlice) {
				l = append(l, OriginalServiceOrder{Index: idx, Name: v.Key.(string)})
			}
		}
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
