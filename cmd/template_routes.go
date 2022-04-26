package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/uselagoon/lagoon-routegen/internal/helpers"
	"github.com/uselagoon/lagoon-routegen/internal/lagoon"
	routeTemplater "github.com/uselagoon/lagoon-routegen/internal/templating/routes"
	"sigs.k8s.io/yaml"
)

var lagoonYml, environmentName, projectName, activeEnvironment, standbyEnvironment, environmentType string
var projectVariables, environmentVariables, monitoringStatusPageID, monitoringContact string
var templateValues, savedTemplates, lagoonFastlyCacheNoCahce, lagoonFastlyServiceID string
var monitoringEnabled bool

var routeGeneration = &cobra.Command{
	Use:     "routes",
	Aliases: []string{"route", "rs", "r"},
	Short:   "Generate the ingress/route templates for a Lagoon build",
	RunE: func(cmd *cobra.Command, args []string) error {
		activeEnv := false
		standbyEnv := false

		monitoringContact = helpers.GetEnv("MONITORING_ALERTCONTACT", monitoringContact)
		monitoringStatusPageID = helpers.GetEnv("MONITORING_STATUSPAGEID", monitoringStatusPageID)

		projectName = helpers.GetEnv("PROJECT", projectName)
		environmentName = helpers.GetEnv("BRANCH", environmentName)
		environmentType = helpers.GetEnv("ENVIRONMENT_TYPE", environmentType)
		activeEnvironment = helpers.GetEnv("ACTIVE_ENVIRONMENT", activeEnvironment)
		standbyEnvironment = helpers.GetEnv("STANDBY_ENVIRONMENT", standbyEnvironment)

		lagoonFastlyCacheNoCahce = helpers.GetEnv("LAGOON_FASTLY_NOCACHE_SERVICE_ID", lagoonFastlyCacheNoCahce)
		lagoonFastlyServiceID = helpers.GetEnv("ROUTE_FASTLY_SERVICE_ID", lagoonFastlyServiceID)

		savedTemplates = helpers.GetEnv("YAML_FOLDER", savedTemplates)

		monitoringEnabled = false
		if environmentType == "production" {
			monitoringEnabled = true
			if environmentName == activeEnvironment {
				activeEnv = true
			}
			if environmentName == standbyEnvironment {
				standbyEnv = true
			}
		}
		// get the project and environment variables
		projectVariables = helpers.GetEnv("LAGOON_PROJECT_VARIABLES", projectVariables)
		environmentVariables = helpers.GetEnv("LAGOON_ENVIRONMENT_VARIABLES", environmentVariables)

		// unmarshal and then merge the two so there is only 1 set of variables to iterate over
		projectVars := []lagoon.EnvironmentVariable{}
		envVars := []lagoon.EnvironmentVariable{}
		json.Unmarshal([]byte(projectVariables), &projectVars)
		json.Unmarshal([]byte(environmentVariables), &envVars)
		lagoonEnvVars := lagoon.MergeVariables(projectVars, envVars)

		// read the routes from the API
		var apiRoutes lagoon.RoutesV2
		lagoonRoutesJSON, err := lagoon.GetLagoonVariable("LAGOON_ROUTES_JSON", lagoonEnvVars)
		if lagoonRoutesJSON != nil {
			// if the routesJSON is populated, then attempt to decode and unmarshal it
			rawJSONStr, _ := base64.StdEncoding.DecodeString(lagoonRoutesJSON.Value)
			rawJSON := []byte(rawJSONStr)
			err := json.Unmarshal(rawJSON, &apiRoutes)
			if err != nil {
				return fmt.Errorf("couldn't unmarshal routes from Lagoon API, is it actually JSON that has been base64 encoded?: %v", err)
			}
		}

		var lYAML lagoon.YAML
		rawYAML, err := os.ReadFile(lagoonYml)
		if err != nil {
			return fmt.Errorf("couldn't read %v: %v", lagoonYml, err)
		}
		err = yaml.Unmarshal(rawYAML, &lYAML)
		if err != nil {
			return fmt.Errorf("couldn't unmarshal %v: %v", lagoonYml, err)
		}
		// because lagoonyaml is not really good yaml, unmarshal polysite into an unknown struct to check
		lPolysite := make(map[string]interface{})
		err = yaml.Unmarshal(rawYAML, &lPolysite)
		if err != nil {
			return fmt.Errorf("couldn't unmarshal %v: %v", lagoonYml, err)
		}

		// handle the active/standby routes
		activeStanbyRoutes := &lagoon.RoutesV2{}
		if lYAML.ProductionRoutes != nil {
			if activeEnv == true {
				if lYAML.ProductionRoutes.Active != nil {
					if lYAML.ProductionRoutes.Active.Routes != nil {
						for _, routeMap := range lYAML.ProductionRoutes.Active.Routes {
							routeTemplater.GenerateRouteStructure(activeStanbyRoutes, routeMap, lagoonEnvVars, true)
						}
					}
				}
			}
			if standbyEnv == true {
				if lYAML.ProductionRoutes.Standby != nil {
					if lYAML.ProductionRoutes.Standby.Routes != nil {
						for _, routeMap := range lYAML.ProductionRoutes.Standby.Routes {
							routeTemplater.GenerateRouteStructure(activeStanbyRoutes, routeMap, lagoonEnvVars, true)
						}
					}
				}
			}
		}
		lagoonValuesFile := routeTemplater.ReadValuesFile(fmt.Sprintf("%s/%s", templateValues, "values.yaml"))
		// generate the templates for active/standby routes
		for _, route := range activeStanbyRoutes.Routes {
			fmt.Println(fmt.Sprintf("Generating Active/Standby Ingress manifest for %s", route.Domain))
			templateYAML := routeTemplater.GenerateKubeTemplate(route, lagoonValuesFile, monitoringContact, monitoringStatusPageID, monitoringEnabled)
			routeTemplater.WriteTemplateFile(fmt.Sprintf("%s/%s.yaml", savedTemplates, route.Domain), templateYAML)
		}

		// handle routes from the .lagoon.yml and the API specifically
		newRoutes := &lagoon.RoutesV2{}
		if _, ok := lPolysite[projectName]; ok {
			// POLYSITE: if this is polysite, then the `projectname` routes block should be defined
			strA, _ := yaml.Marshal(lPolysite[projectName])
			var lYAMLPolysite lagoon.YAML
			err = yaml.Unmarshal(strA, &lYAMLPolysite)
			if err != nil {
				return fmt.Errorf("couldn't unmarshal for polysite %v: %v", strA, err)
			}
			for _, routeMap := range lYAMLPolysite.Environments[environmentName].Routes {
				routeTemplater.GenerateRouteStructure(newRoutes, routeMap, lagoonEnvVars, false)
			}
		} else {
			// otherwise it just uses the default environment name
			for _, routeMap := range lYAML.Environments[environmentName].Routes {
				routeTemplater.GenerateRouteStructure(newRoutes, routeMap, lagoonEnvVars, false)
			}
		}
		// merge routes from the API on top of the routes from the `.lagoon.yml`
		finalRoutes := routeTemplater.MergeRouteStructures(*newRoutes, apiRoutes)
		// generate the templates
		for _, route := range finalRoutes.Routes {
			fmt.Println(fmt.Sprintf("Generating Ingress manifest for %s", route.Domain))
			templateYAML := routeTemplater.GenerateKubeTemplate(route, lagoonValuesFile, monitoringContact, monitoringStatusPageID, monitoringEnabled)
			routeTemplater.WriteTemplateFile(fmt.Sprintf("%s/%s.yaml", savedTemplates, route.Domain), templateYAML)
		}
		return nil
	},
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
	routeGeneration.Flags().StringVarP(&lagoonFastlyCacheNoCahce, "fastly-cache-no-cache-id", "F", "",
		"The fastly cache no cache service ID to use")
	routeGeneration.Flags().StringVarP(&lagoonFastlyServiceID, "fastly-service-id", "f", "",
		"The fastly service ID to use")
}
