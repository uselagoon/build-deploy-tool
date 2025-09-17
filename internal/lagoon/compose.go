package lagoon

import (
	"fmt"
	"os"
	"strings"

	"github.com/compose-spec/compose-go/cli"
	"github.com/compose-spec/compose-go/loader"
	composetypes "github.com/compose-spec/compose-go/types"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	goyaml "gopkg.in/yaml.v2"
	goyamlv3 "gopkg.in/yaml.v3"
	utilvalidation "k8s.io/apimachinery/pkg/util/validation"
)

type OriginalServiceOrder struct {
	Name  string
	Index int
}

var (
	customVolumePrefix string = "custom-"
)

type OriginalVolumeOrder OriginalServiceOrder

// UnmarshaDockerComposeYAML unmarshal the lagoon.yml file into a YAML and map for consumption.
func UnmarshaDockerComposeYAML(file string, ignoreErrors, ignoreMissingEnvFiles bool, envvars map[string]string) (*composetypes.Project, []OriginalServiceOrder, []OriginalVolumeOrder, error) {
	lYAML := &YAML{}
	projectName := helpers.GetEnv("PROJECT", "", false)
	err := UnmarshalLagoonYAML(file, lYAML, projectName)
	if err != nil {
		return nil, nil, nil, err
	}
	if _, err := os.Stat(lYAML.DockerComposeYAML); err != nil {
		return nil, nil, nil, fmt.Errorf("docker-compose file referenced in .lagoon.yml not found")
	}
	options, err := cli.NewProjectOptions([]string{lYAML.DockerComposeYAML},
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
		return nil, nil, nil, err
	}
	options.Environment = envvars
	l, err := cli.ProjectFromOptions(options)
	if err != nil {
		return nil, nil, nil, err
	}
	for _, service := range l.Services {
		if err := CheckServiceNameValidity(service); err != nil {
			return nil, nil, nil, err
		}
	}
	originalOrder, originalVolume, err := UnmarshalLagoonDockerComposeYAML(lYAML.DockerComposeYAML)
	if err != nil {
		return nil, nil, nil, err
	}
	return l, originalOrder, originalVolume, nil
}

// UnmarshalLagoonDockerComposeYAML unmarshal the docker-compose.yml file into a YAML and map for consumption.
// this uses yaml mapslice to preserve the order of the services in the docker-compose file
// as lagoon relies on this order for building images and determining the order of routes
func UnmarshalLagoonDockerComposeYAML(file string) ([]OriginalServiceOrder, []OriginalVolumeOrder, error) {
	rawYAML, err := os.ReadFile(file)
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't read %v: %v", file, err)
	}
	ls := []OriginalServiceOrder{}
	lv := []OriginalVolumeOrder{}
	// unmarshal docker-compose.yml
	// use to gopkg yaml v2 for MapSlice
	m := goyaml.MapSlice{}
	goyaml.Unmarshal(rawYAML, &m)
	for _, item := range m {
		// extract the services only
		if item.Key.(string) == "services" {
			for idx, v := range item.Value.(goyaml.MapSlice) {
				ls = append(ls, OriginalServiceOrder{Index: idx, Name: v.Key.(string)})
			}
		}
		// extract the volumes only
		if item.Key.(string) == "volumes" {
			for idx, v := range item.Value.(goyaml.MapSlice) {
				lv = append(lv, OriginalVolumeOrder{Index: idx, Name: v.Key.(string)})
			}
		}
	}
	return ls, lv, nil
}

// use goyamlv3 that newer versions of compose-go uses to validate
func ValidateUnmarshalDockerComposeYAML(file string) error {
	lYAML := &YAML{}
	projectName := helpers.GetEnv("PROJECT", "", false)
	err := UnmarshalLagoonYAML(file, lYAML, projectName)
	if err != nil {
		return err
	}
	rawYAML, err := os.ReadFile(lYAML.DockerComposeYAML)
	if err != nil {
		return fmt.Errorf("couldn't read %v: %v", lYAML.DockerComposeYAML, err)
	}
	var m interface{}
	err = goyamlv3.Unmarshal(rawYAML, &m)
	if err != nil {
		return err
	}
	return nil
}

// Checks the validity of the service name against the RFC1035 DNS label standard
func CheckServiceNameValidity(v composetypes.ServiceConfig) error {
	// go over the service map looking for the labels slice
	for key, value := range v.Labels {
		// check if the lagoon.type != none
		if key == "lagoon.type" && value != "none" {
			if err := utilvalidation.IsDNS1035Label(v.Name); err != nil {
				return fmt.Errorf("service name is invalid. Please refer to the documentation regarding service naming requirements")
			}
		}
	}
	return nil
}

// CheckDockerComposeLagoonLabel checks the labels in a compose service to see if the requested label exists, and returns the value if so
func CheckDockerComposeLagoonLabel(labels map[string]string, label string) string {
	for k, v := range labels {
		if k == label {
			return v
		}
	}
	return ""
}

// pad volume names from the docker compose file with the compose stack name
func GetComposeVolumeName(c, n string) string {
	return fmt.Sprintf("%s_%s", c, n)
}

// trim compose stack name from volume name
func GetVolumeNameFromComposeName(c, n string) string {
	return strings.Replace(n, fmt.Sprintf("%s_", c), "", 1)
}

// pad volume names from the docker compose file with the custom volume prefix
func GetLagoonVolumeName(n string) string {
	return fmt.Sprintf("%s%s", customVolumePrefix, n)
}

// trim the custom prefix from lagoon volume names returning them to the name defined in the docker compose file
func GetVolumeNameFromLagoonVolume(n string) string {
	return strings.Replace(n, customVolumePrefix, "", 1)
}
