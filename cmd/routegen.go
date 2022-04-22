package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/uselagoon/lagoon-routegen/internal/generator"
	"sigs.k8s.io/yaml"
)

var lagoonYml, environmentName, projectName, activeEnvironment, standbyEnvironment, environmentType string
var projectVariables, environmentVariables, monitoringStatusPageID, monitoringContact, lagoonRoutesJSON string
var templatePath, yamlPath string
var monitoringEnabled bool

var routeGeneration = &cobra.Command{
	Use:     "routegen",
	Aliases: []string{"r", "rg"},
	Short:   "Generate routes",
	RunE: func(cmd *cobra.Command, args []string) error {
		activeEnv := false
		standbyEnv := false

		projectVariables = getEnv("LAGOON_PROJECT_VARIABLES", projectVariables)
		environmentVariables = getEnv("LAGOON_ENVIRONMENT_VARIABLES", environmentVariables)

		monitoringContact = getEnv("MONITORING_ALERTCONTACT", monitoringContact)
		monitoringStatusPageID = getEnv("MONITORING_STATUSPAGEID", monitoringStatusPageID)

		projectName = getEnv("PROJECT", projectName)
		environmentName = getEnv("BRANCH", environmentName)
		environmentType = getEnv("ENVIRONMENT_TYPE", environmentType)
		activeEnvironment = getEnv("ACTIVE_ENVIRONMENT", activeEnvironment)
		standbyEnvironment = getEnv("STANDBY_ENVIRONMENT", standbyEnvironment)
		lagoonRoutesJSON = getEnv("LAGOON_ROUTES_JSON", lagoonRoutesJSON)

		yamlPath = getEnv("YAML_FOLDER", yamlPath)

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

		// merge the environment variables from the lagoon api
		projectVars := []generator.LagoonEnvironmentVariable{}
		envVars := []generator.LagoonEnvironmentVariable{}
		json.Unmarshal([]byte(projectVariables), &projectVars)
		json.Unmarshal([]byte(environmentVariables), &envVars)
		_ = generator.MergeVariables(projectVars, envVars)

		// read the routes from the API
		var apiRoutes generator.RoutesV2
		rawJSONStr, _ := base64.StdEncoding.DecodeString(lagoonRoutesJSON)
		rawJSON := []byte(rawJSONStr)
		err := json.Unmarshal(rawJSON, &apiRoutes)
		if err != nil {
			return fmt.Errorf("couldn't unmarshal: %v", err)
		}

		var lYAML generator.Lagoon
		rawYAML, err := os.ReadFile(lagoonYml)
		if err != nil {
			panic(fmt.Errorf("couldn't read %v: %v", lagoonYml, err))
		}
		err = yaml.Unmarshal(rawYAML, &lYAML)
		if err != nil {
			panic(fmt.Errorf("couldn't unmarshal %v: %v", lagoonYml, err))
		}
		// because lagoonyaml is not really good yaml, unmarshal polysite into an unknown struct to check
		lPolysite := make(map[string]interface{})
		err = yaml.Unmarshal(rawYAML, &lPolysite)
		if err != nil {
			panic(fmt.Errorf("couldn't unmarshal %v: %v", lagoonYml, err))
		}

		// handle the active/standby routes
		activeStanbyRoutes := &generator.RoutesV2{}
		if activeEnv == true {
			for _, routeMap := range lYAML.ProductionRoutes.Active.Routes {
				generator.GenerateRouteStructure(activeStanbyRoutes, routeMap, true)
			}
		}
		if standbyEnv == true {
			for _, routeMap := range lYAML.ProductionRoutes.Standby.Routes {
				generator.GenerateRouteStructure(activeStanbyRoutes, routeMap, true)
			}
		}
		lagoonValuesFile := generator.ReadValuesFile(fmt.Sprintf("%s/%s", templatePath, "values.yaml"))
		for _, route := range activeStanbyRoutes.Routes {
			// templateYAML, _ := generator.GenerateHelmTemplates(
			// 	fmt.Sprintf("%s/custom-ingress", templatePath),
			// 	fmt.Sprintf("%s/%s", templatePath, "values.yaml"),
			// 	fmt.Sprintf("%s/%s-values.yaml", templatePath, route.Domain),
			// 	route,
			// 	monitoringContact,
			// 	monitoringStatusPageID,
			// 	monitoringEnabled,
			// 	true,
			// )
			templateYAML := generator.GenerateKubeTemplate(route, lagoonValuesFile, monitoringContact, monitoringStatusPageID, monitoringEnabled)
			generator.WriteTemplateFile(fmt.Sprintf("%s/%s.yaml", yamlPath, route.Domain), templateYAML)
		}

		// handle routes from the .lagoon.yml and the API specifically
		newRoutes := &generator.RoutesV2{}
		if _, ok := lPolysite[projectName]; ok {
			// POLYSITE: if this is polysite, then the `projectname` routes block should be defined
			strA, _ := yaml.Marshal(lPolysite[projectName])
			var lYAMLPolysite generator.Lagoon
			err = yaml.Unmarshal(strA, &lYAMLPolysite)
			if err != nil {
				panic(fmt.Errorf("couldn't unmarshal %v: %v", strA, err))
			}
			for _, routeMap := range lYAMLPolysite.Environments[environmentName].Routes {
				generator.GenerateRouteStructure(newRoutes, routeMap, false)
			}
		} else {
			// otherwise it just uses the default environment name
			for _, routeMap := range lYAML.Environments[environmentName].Routes {
				generator.GenerateRouteStructure(newRoutes, routeMap, false)
			}
		}
		// merge routes from the API on top of the routes from the `.lagoon.yml`
		finalRoutes := generator.MergeRouteStructures(*newRoutes, apiRoutes)
		// generate the helm templates
		for _, route := range finalRoutes.Routes {
			// templateYAML, _ := generator.GenerateHelmTemplates(
			// 	fmt.Sprintf("%s/helmcharts/custom-ingress", templatePath),
			// 	fmt.Sprintf("%s/%s", templatePath, "values.yaml"),
			// 	fmt.Sprintf("%s/%s-values.yaml", templatePath, route.Domain),
			// 	route,
			// 	monitoringContact,
			// 	monitoringStatusPageID,
			// 	monitoringEnabled,
			// 	false,
			// )
			templateYAML := generator.GenerateKubeTemplate(route, lagoonValuesFile, monitoringContact, monitoringStatusPageID, monitoringEnabled)
			generator.WriteTemplateFile(fmt.Sprintf("%s/%s.yaml", yamlPath, route.Domain), templateYAML)
		}

		return nil
	},
}

func init() {
	routeGeneration.Flags().StringVarP(&lagoonYml, "lagoon-yml", "l", ".lagoon.yml",
		"The .lagoon.yml file to read")
	routeGeneration.Flags().StringVarP(&environmentName, "environment-name", "e", "",
		"The environment name to check")
	routeGeneration.Flags().StringVarP(&projectName, "project-name", "p", "",
		"The projects name")
	routeGeneration.Flags().StringVarP(&environmentType, "environment-type", "T", "",
		"The type of environment")
	routeGeneration.Flags().StringVarP(&activeEnvironment, "active-environment", "a", "",
		"Name of the active environment if known")
	routeGeneration.Flags().StringVarP(&standbyEnvironment, "standby-environment", "s", "",
		"Name of the standby environment if known")
	routeGeneration.Flags().StringVarP(&templatePath, "template-path", "P", "/kubectl-build-deploy/",
		"Path to the template on disk")
	routeGeneration.Flags().StringVarP(&yamlPath, "yaml-path", "Y", "/kubectl-build-deploy/lagoon/services-routes",
		"Path to where the resulting templates are saved")
}
