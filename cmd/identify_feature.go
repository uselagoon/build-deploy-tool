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

func init() {
	identifyCmd.AddCommand(featureFlagIdentify)
	featureFlagIdentify.Flags().StringVarP(&lagoonYml, "lagoon-yml", "l", ".lagoon.yml",
		"The .lagoon.yml file to read")
	featureFlagIdentify.Flags().StringVarP(&projectName, "project-name", "p", "",
		"The project name")
	featureFlagIdentify.Flags().StringVarP(&environmentName, "environment-name", "e", "",
		"The environment name to check")
	featureFlagIdentify.Flags().StringVarP(&environmentType, "environment-type", "E", "",
		"The type of environment (development or production)")
	featureFlagIdentify.Flags().StringVarP(&buildType, "build-type", "d", "",
		"The type of build (branch, pullrequest, promote)")
	featureFlagIdentify.Flags().StringVarP(&branch, "branch", "b", "",
		"The name of the branch")
	featureFlagIdentify.Flags().StringVarP(&prNumber, "pullrequest-number", "P", "",
		"The pullrequest number")
	featureFlagIdentify.Flags().StringVarP(&prHeadBranch, "pullrequest-head-branch", "H", "",
		"The pullrequest head branch")
	featureFlagIdentify.Flags().StringVarP(&prBaseBranch, "pullrequest-base-branch", "B", "",
		"The pullrequest base branch")
	featureFlagIdentify.Flags().StringVarP(&lagoonVersion, "lagoon-version", "L", "",
		"The lagoon version")
	featureFlagIdentify.Flags().StringVarP(&activeEnvironment, "active-environment", "a", "",
		"Name of the active environment if known")
	featureFlagIdentify.Flags().StringVarP(&standbyEnvironment, "standby-environment", "s", "",
		"Name of the standby environment if known")
	featureFlagIdentify.Flags().StringVarP(&templateValues, "template-path", "t", "/kubectl-build-deploy/",
		"Path to the template on disk")
	featureFlagIdentify.Flags().StringVarP(&savedTemplates, "saved-templates-path", "T", "/kubectl-build-deploy/lagoon/services-routes",
		"Path to where the resulting templates are saved")
	featureFlagIdentify.Flags().StringVarP(&monitoringContact, "monitoring-config", "M", "",
		"The monitoring contact config if known")
	featureFlagIdentify.Flags().StringVarP(&monitoringStatusPageID, "monitoring-status-page-id", "m", "",
		"The monitoring status page ID if known")
	featureFlagIdentify.Flags().StringVarP(&fastlyCacheNoCahce, "fastly-cache-no-cache-id", "F", "",
		"The fastly cache no cache service ID to use")
	featureFlagIdentify.Flags().StringVarP(&fastlyServiceID, "fastly-service-id", "f", "",
		"The fastly service ID to use")
	featureFlagIdentify.Flags().StringVarP(&fastlyAPISecretPrefix, "fastly-api-secret-prefix", "A", "fastly-api-",
		"The fastly secret prefix to use")
	featureFlagIdentify.Flags().BoolVarP(&checkValuesFile, "check-values-file", "C", false,
		"If set, will check for the values file defined in `${template-path}/values.yaml`")
}
