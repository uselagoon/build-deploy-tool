package lagoon

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/loader"
	composetypes "github.com/compose-spec/compose-go/v2/types"
	"github.com/sirupsen/logrus"
	goyaml "gopkg.in/yaml.v2"
	goyamlv3 "gopkg.in/yaml.v3"
	utilvalidation "k8s.io/apimachinery/pkg/util/validation"
)

type OriginalServiceOrder struct {
	Name  string
	Index int
}

// UnmarshaDockerComposeYAML unmarshal the lagoon.yml file into a YAML and map for consumption.
func UnmarshaDockerComposeYAML(file string, ignoreErrors, ignoreMissingEnvFiles bool, envvars map[string]string) (*composetypes.Project, []OriginalServiceOrder, error) {
	// disable logging output for now, maybe capture this in a buffer for later analysis
	logrus.SetOutput(io.Discard)
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
	if err != nil {
		return nil, nil, err
	}
	options.Environment = envvars
	l, err := cli.ProjectFromOptions(context.Background(), options)
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
				if err := CheckServiceNameValidity(v); err != nil {
					return nil, err
				}
				l = append(l, OriginalServiceOrder{Index: idx, Name: v.Key.(string)})
			}
		}
	}
	return l, nil
}

// use goyamlv3 that newer versions of compose-go uses to validate
func ValidateUnmarshalDockerComposeYAML(file string) error {
	rawYAML, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("couldn't read %v: %v", file, err)
	}
	var m interface{}
	err = goyamlv3.Unmarshal(rawYAML, &m)
	if err != nil {
		return err
	}
	return nil
}

// Checks the validity of the service name against the RFC1035 DNS label standard
func CheckServiceNameValidity(v goyaml.MapItem) error {
	// go over the service map looking for the labels slice
	for _, s := range v.Value.(goyaml.MapSlice) {
		if s.Key == "labels" {
			// go over the labels looking for the lagoon.type label
			for _, label := range s.Value.(goyaml.MapSlice) {
				// check if the lagoon.type != none
				if label.Key == "lagoon.type" && label.Value != "none" {
					if err := utilvalidation.IsDNS1035Label(v.Key.(string)); err != nil {
						return errors.New("Service name is invalid. Please refer to the documentation regarding service naming requirements")
					}
				}
			}
		}
	}
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
