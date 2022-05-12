package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	routeTemplater "github.com/uselagoon/build-deploy-tool/internal/templating/routes"
)

var lagoonYml, environmentName, projectName, activeEnvironment, standbyEnvironment, environmentType string
var buildType, lagoonVersion, branch, prNumber, prHeadBranch, prBaseBranch string
var projectVariables, environmentVariables, monitoringStatusPageID, monitoringContact string
var templateValues, savedTemplates, fastlyCacheNoCahce, fastlyServiceID, fastlyAPISecretPrefix string
var monitoringEnabled, checkValuesFile bool

var routeGeneration = &cobra.Command{
	Use:     "ingress",
	Aliases: []string{"i"},
	Short:   "Generate the ingress templates for a Lagoon build",
	RunE: func(cmd *cobra.Command, args []string) error {
		return IngressTemplateGeneration(true)
	},
}

// IngressTemplateGeneration .
func IngressTemplateGeneration(debug bool) error {
	activeEnv := false
	standbyEnv := false

	lagoonEnvVars := []lagoon.EnvironmentVariable{}
	lagoonValues := lagoon.BuildValues{}
	lYAML := lagoon.YAML{}
	lPolysite := make(map[string]interface{})
	collectIngressVariablesValues(debug, &activeEnv, &standbyEnv, &lagoonEnvVars, &lagoonValues, &lYAML, &lPolysite)

	// read the routes from the API
	apiRoutes, err := getRoutesFromAPIEnvVar(lagoonEnvVars, debug)
	if err != nil {
		return fmt.Errorf("couldn't unmarshal routes from Lagoon API, is it actually JSON that has been base64 encoded?: %v", err)
	}

	// handle routes from the .lagoon.yml and the API specifically
	finalRoutes, err := generateAndMerge(*apiRoutes, lagoonEnvVars, lPolysite, lYAML, lagoonValues)
	if err != nil {
		return fmt.Errorf("couldn't generate and merge routes: %v", err)
	}

	// generate the templates
	for _, route := range finalRoutes.Routes {
		if debug {
			fmt.Println(fmt.Sprintf("Templating ingress manifest for %s to %s", route.Domain, fmt.Sprintf("%s/%s.yaml", savedTemplates, route.Domain)))
		}
		templateYAML := routeTemplater.GenerateKubeTemplate(route, lagoonValues, monitoringContact, monitoringStatusPageID, monitoringEnabled)
		routeTemplater.WriteTemplateFile(fmt.Sprintf("%s/%s.yaml", savedTemplates, route.Domain), templateYAML)
	}

	if activeEnv || standbyEnv {
		// active/standby routes should not be changed by any environment defined routes.
		// generate the templates for these independently of any previously generated routes,
		// this WILL overwrite previously created templates ensuring that anything defined in the `production_routes`
		// section are created correctly ensuring active/standby will work
		activeStanbyRoutes := generateActiveStandby(activeEnv, standbyEnv, lagoonEnvVars, lYAML)
		// generate the templates for active/standby routes separately to normal routes
		for _, route := range activeStanbyRoutes.Routes {
			if debug {
				fmt.Println(fmt.Sprintf("Templating active/standby ingress manifest for %s to %s", route.Domain, fmt.Sprintf("%s/%s.yaml", savedTemplates, route.Domain)))
			}
			templateYAML := routeTemplater.GenerateKubeTemplate(route, lagoonValues, monitoringContact, monitoringStatusPageID, monitoringEnabled)
			routeTemplater.WriteTemplateFile(fmt.Sprintf("%s/%s.yaml", savedTemplates, route.Domain), templateYAML)
		}
	}
	return nil
}

func init() {
	templateCmd.AddCommand(routeGeneration)
	routeGeneration.Flags().StringVarP(&lagoonYml, "lagoon-yml", "l", ".lagoon.yml",
		"The .lagoon.yml file to read")
	routeGeneration.Flags().StringVarP(&projectName, "project-name", "p", "",
		"The project name")
	routeGeneration.Flags().StringVarP(&environmentName, "environment-name", "e", "",
		"The environment name to check")
	routeGeneration.Flags().StringVarP(&environmentType, "environment-type", "E", "",
		"The type of environment (development or production)")
	routeGeneration.Flags().StringVarP(&buildType, "build-type", "d", "",
		"The type of build (branch, pullrequest, promote)")
	routeGeneration.Flags().StringVarP(&branch, "branch", "b", "",
		"The name of the branch")
	routeGeneration.Flags().StringVarP(&prNumber, "pullrequest-number", "P", "",
		"The pullrequest number")
	routeGeneration.Flags().StringVarP(&prHeadBranch, "pullrequest-head-branch", "H", "",
		"The pullrequest head branch")
	routeGeneration.Flags().StringVarP(&prBaseBranch, "pullrequest-base-branch", "B", "",
		"The pullrequest base branch")
	routeGeneration.Flags().StringVarP(&lagoonVersion, "lagoon-version", "L", "",
		"The lagoon version")
	routeGeneration.Flags().StringVarP(&activeEnvironment, "active-environment", "a", "",
		"Name of the active environment if known")
	routeGeneration.Flags().StringVarP(&standbyEnvironment, "standby-environment", "s", "",
		"Name of the standby environment if known")
	routeGeneration.Flags().StringVarP(&templateValues, "template-path", "t", "/kubectl-build-deploy/",
		"Path to the template on disk")
	routeGeneration.Flags().StringVarP(&savedTemplates, "saved-templates-path", "T", "/kubectl-build-deploy/lagoon/services-routes",
		"Path to where the resulting templates are saved")
	routeGeneration.Flags().StringVarP(&monitoringContact, "monitoring-config", "M", "",
		"The monitoring contact config if known")
	routeGeneration.Flags().StringVarP(&monitoringStatusPageID, "monitoring-status-page-id", "m", "",
		"The monitoring status page ID if known")
	routeGeneration.Flags().StringVarP(&fastlyCacheNoCahce, "fastly-cache-no-cache-id", "F", "",
		"The fastly cache no cache service ID to use")
	routeGeneration.Flags().StringVarP(&fastlyServiceID, "fastly-service-id", "f", "",
		"The fastly service ID to use")
	routeGeneration.Flags().StringVarP(&fastlyAPISecretPrefix, "fastly-api-secret-prefix", "A", "fastly-api-",
		"The fastly secret prefix to use")
	routeGeneration.Flags().BoolVarP(&checkValuesFile, "check-values-file", "C", false,
		"If set, will check for the values file defined in `${template-path}/values.yaml`")
}
