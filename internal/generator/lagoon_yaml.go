package generator

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	"sigs.k8s.io/yaml"
)

func LoadAndUnmarshalLagoonYml(lagoonYml string, lagoonYmlOverride string, lagoonYmlOverrideEnvVarName string, lYAML *lagoon.YAML, projectName string, debug bool) error {

	// First we load the primary file
	if err := lagoon.UnmarshalLagoonYAML(lagoonYml, lYAML, projectName); err != nil {
		return fmt.Errorf("couldn't unmarshal file %v: %v", lagoonYml, err)
	}

	// Here we try and merge in .lagoon.yml override
	if _, err := os.Stat(lagoonYmlOverride); err == nil {
		overLagoonYaml := &lagoon.YAML{}
		if err := lagoon.UnmarshalLagoonYAML(lagoonYmlOverride, overLagoonYaml, projectName); err != nil {
			return fmt.Errorf("couldn't unmarshal file %v: %v", lagoonYmlOverride, err)
		}
		//now we merge
		if err := lagoon.MergeLagoonYAMLs(lYAML, overLagoonYaml); err != nil {
			return fmt.Errorf("unable to merge %v over %v: %v", lagoonYmlOverride, lagoonYml, err)
		}
	}
	// Now we see if there are any environment vars set for .lagoon.yml overrides
	envLagoonYamlStringBase64 := helpers.GetEnv(lagoonYmlOverrideEnvVarName, "", debug)
	if envLagoonYamlStringBase64 != "" {
		//Decode it
		envLagoonYamlString, err := base64.StdEncoding.DecodeString(envLagoonYamlStringBase64)
		if err != nil {
			return fmt.Errorf("unable to decode %v - is it base64 encoded?", lagoonYmlOverrideEnvVarName)
		}
		envLagoonYaml := &lagoon.YAML{}
		lEnvLagoonPolysite := make(map[string]interface{})

		err = yaml.Unmarshal(envLagoonYamlString, envLagoonYaml)
		if err != nil {
			return fmt.Errorf("unable to unmarshal env var %v: %v", lagoonYmlOverrideEnvVarName, err)
		}
		err = yaml.Unmarshal(envLagoonYamlString, lEnvLagoonPolysite)
		if err != nil {
			return fmt.Errorf("unable to unmarshal env var %v: %v", lagoonYmlOverrideEnvVarName, err)
		}

		if _, ok := lEnvLagoonPolysite[projectName]; ok {
			s, _ := yaml.Marshal(lEnvLagoonPolysite[projectName])
			_ = yaml.Unmarshal(s, &envLagoonYaml)
		}
		//now we merge
		if err := lagoon.MergeLagoonYAMLs(lYAML, envLagoonYaml); err != nil {
			return fmt.Errorf("unable to merge LAGOON_YAML_OVERRIDE over %v: %v", lagoonYml, err)
		}
	}
	return nil
}
