package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
)

var featureFlagIdentify = &cobra.Command{
	Use:     "feature",
	Aliases: []string{"f"},
	Short:   "Identify if a feature flag has been enabled",
	RunE: func(cmd *cobra.Command, args []string) error {
		flagValue, err := IdentifyFeatureFlag("", false)
		if err != nil {
			return err
		}
		fmt.Println(flagValue)
		return nil
	},
}

// IdentifyFeatureFlag checks if a feature flag of given name has been set or not in a build
func IdentifyFeatureFlag(name string, debug bool) (string, error) {
	activeEnv := false
	standbyEnv := false

	lagoonEnvVars := []lagoon.EnvironmentVariable{}
	lagoonValues := lagoon.BuildValues{}
	lYAML := lagoon.YAML{}
	lCompose := lagoon.Compose{}
	lPolysite := make(map[string]interface{})
	err := collectBuildValues(debug, &activeEnv, &standbyEnv, &lagoonEnvVars, &lagoonValues, &lYAML, &lPolysite, &lCompose)
	if err != nil {
		return "", err
	}

	forceFlagVar := helpers.GetEnv(fmt.Sprintf("%s%s", "LAGOON_FEATURE_FLAG_FORCE_", name), "", debug)
	if forceFlagVar != "" {
		return forceFlagVar, nil
	}
	featureFlagVar, _ := lagoon.GetLagoonVariable(fmt.Sprintf("%s%s", "LAGOON_FEATURE_FLAG_", name), []string{"build", "global"}, lagoonEnvVars)
	if featureFlagVar != nil {
		return featureFlagVar.Value, nil
	}
	defaultFlagVar := helpers.GetEnv(fmt.Sprintf("%s%s", "LAGOON_FEATURE_FLAG_DEFAULT_", name), "", debug)
	return defaultFlagVar, nil
}
