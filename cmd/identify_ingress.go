package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
)

var primaryIngressIdentify = &cobra.Command{
	Use:     "primary-ingress",
	Aliases: []string{"pi"},
	Short:   "Identify the primary ingress for a specific environment",
	RunE: func(cmd *cobra.Command, args []string) error {
		primary, err := IdentifyPrimaryIngress(false)
		if err != nil {
			return err
		}
		fmt.Println(primary)
		return nil
	},
}

// IdentifyPrimaryIngress .
func IdentifyPrimaryIngress(debug bool) (string, error) {
	primaryIngress := ""
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

	// read the routes from the API
	apiRoutes, err := getRoutesFromAPIEnvVar(lagoonEnvVars, debug)
	if err != nil {
		return primaryIngress, fmt.Errorf("couldn't unmarshal routes from Lagoon API, is it actually JSON that has been base64 encoded?: %v", err)
	}

	// handle routes from the .lagoon.yml and the API specifically
	finalRoutes, err := generateAndMerge(*apiRoutes, lagoonEnvVars, lPolysite, lYAML, lagoonValues)
	if err != nil {
		return primaryIngress, fmt.Errorf("couldn't generate and merge routes: %v", err)
	}

	// get the first route from the list of routes
	if len(finalRoutes.Routes) > 0 {
		primaryIngress = finalRoutes.Routes[0].Domain
	}

	if activeEnv || standbyEnv {
		// active/standby routes should not be changed by any environment defined routes.
		// generate the templates for these independently of any previously generated routes,
		// this WILL overwrite previously created templates ensuring that anything defined in the `production_routes`
		// section are created correctly ensuring active/standby will work
		activeStanbyRoutes := generateActiveStandby(activeEnv, standbyEnv, lagoonEnvVars, lYAML)
		// get the first route from the list of routes
		if len(activeStanbyRoutes.Routes) > 0 {
			primaryIngress = activeStanbyRoutes.Routes[0].Domain
		}
	}

	return primaryIngress, nil
}

func init() {
	identifyCmd.AddCommand(primaryIngressIdentify)
	primaryIngressIdentify.Flags().StringVarP(&lagoonYml, "lagoon-yml", "l", ".lagoon.yml",
		"The .lagoon.yml file to read")
	primaryIngressIdentify.Flags().StringVarP(&projectName, "project-name", "p", "",
		"The project name")
	primaryIngressIdentify.Flags().StringVarP(&environmentName, "environment-name", "e", "",
		"The environment name to check")
	primaryIngressIdentify.Flags().StringVarP(&environmentType, "environment-type", "E", "",
		"The type of environment (development or production)")
	primaryIngressIdentify.Flags().StringVarP(&buildType, "build-type", "d", "",
		"The type of build (branch, pullrequest, promote)")
	primaryIngressIdentify.Flags().StringVarP(&branch, "branch", "b", "",
		"The name of the branch")
	primaryIngressIdentify.Flags().StringVarP(&prNumber, "pullrequest-number", "P", "",
		"The pullrequest number")
	primaryIngressIdentify.Flags().StringVarP(&prHeadBranch, "pullrequest-head-branch", "H", "",
		"The pullrequest head branch")
	primaryIngressIdentify.Flags().StringVarP(&prBaseBranch, "pullrequest-base-branch", "B", "",
		"The pullrequest base branch")
	primaryIngressIdentify.Flags().StringVarP(&lagoonVersion, "lagoon-version", "L", "",
		"The lagoon version")
	primaryIngressIdentify.Flags().StringVarP(&activeEnvironment, "active-environment", "a", "",
		"Name of the active environment if known")
	primaryIngressIdentify.Flags().StringVarP(&standbyEnvironment, "standby-environment", "s", "",
		"Name of the standby environment if known")
	primaryIngressIdentify.Flags().StringVarP(&templateValues, "template-path", "t", "/kubectl-build-deploy/",
		"Path to the template on disk")
	primaryIngressIdentify.Flags().StringVarP(&savedTemplates, "saved-templates-path", "T", "/kubectl-build-deploy/lagoon/services-routes",
		"Path to where the resulting templates are saved")
	primaryIngressIdentify.Flags().StringVarP(&monitoringContact, "monitoring-config", "M", "",
		"The monitoring contact config if known")
	primaryIngressIdentify.Flags().StringVarP(&monitoringStatusPageID, "monitoring-status-page-id", "m", "",
		"The monitoring status page ID if known")
	primaryIngressIdentify.Flags().StringVarP(&fastlyCacheNoCahce, "fastly-cache-no-cache-id", "F", "",
		"The fastly cache no cache service ID to use")
	primaryIngressIdentify.Flags().StringVarP(&fastlyServiceID, "fastly-service-id", "f", "",
		"The fastly service ID to use")
	primaryIngressIdentify.Flags().StringVarP(&fastlyAPISecretPrefix, "fastly-api-secret-prefix", "A", "fastly-api-",
		"The fastly secret prefix to use")
	primaryIngressIdentify.Flags().BoolVarP(&checkValuesFile, "check-values-file", "C", false,
		"If set, will check for the values file defined in `${template-path}/values.yaml`")
}
