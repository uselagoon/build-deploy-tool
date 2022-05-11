package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	routeTemplater "github.com/uselagoon/build-deploy-tool/internal/templating/routes"
	"sigs.k8s.io/yaml"
)

var lagoonYml, environmentName, projectName, activeEnvironment, standbyEnvironment, environmentType string
var buildType, lagoonVersion, branch, prNumber, prHeadBranch, prBaseBranch string
var projectVariables, environmentVariables, monitoringStatusPageID, monitoringContact string
var templateValues, savedTemplates, fastlyCacheNoCahce, fastlyServiceID, fastlyAPISecretPrefix string
var monitoringEnabled, checkValuesFile bool

var routeGeneration = &cobra.Command{
	Use:     "routes",
	Aliases: []string{"route", "rs", "r"},
	Short:   "Generate the ingress/route templates for a Lagoon build",
	RunE: func(cmd *cobra.Command, args []string) error {
		return RouteGeneration(true)
	},
}

// RouteGeneration .
func RouteGeneration(debug bool) error {
	activeEnv := false
	standbyEnv := false

	// environment variables will override what is provided by flags
	// the following variables have been identified as used by custom-ingress objects
	// these are available within a lagoon build as standard
	monitoringContact = helpers.GetEnv("MONITORING_ALERTCONTACT", monitoringContact, debug)
	monitoringStatusPageID = helpers.GetEnv("MONITORING_STATUSPAGEID", monitoringStatusPageID, debug)
	projectName = helpers.GetEnv("PROJECT", projectName, debug)
	environmentName = helpers.GetEnv("ENVIRONMENT", environmentName, debug)
	branch = helpers.GetEnv("BRANCH", branch, debug)
	prNumber = helpers.GetEnv("PR_NUMBER", prNumber, debug)
	prHeadBranch = helpers.GetEnv("PR_HEAD_BRANCH", prHeadBranch, debug)
	prBaseBranch = helpers.GetEnv("PR_BASE_BRANCH", prBaseBranch, debug)
	environmentType = helpers.GetEnv("ENVIRONMENT_TYPE", environmentType, debug)
	buildType = helpers.GetEnv("BUILD_TYPE", buildType, debug)
	activeEnvironment = helpers.GetEnv("ACTIVE_ENVIRONMENT", activeEnvironment, debug)
	standbyEnvironment = helpers.GetEnv("STANDBY_ENVIRONMENT", standbyEnvironment, debug)
	fastlyCacheNoCahce = helpers.GetEnv("LAGOON_FASTLY_NOCACHE_SERVICE_ID", fastlyCacheNoCahce, debug)
	fastlyServiceID = helpers.GetEnv("ROUTE_FASTLY_SERVICE_ID", fastlyServiceID, debug)
	fastlyAPISecretPrefix = helpers.GetEnv("FASTLY_API_SECRET_PREFIX", fastlyAPISecretPrefix, debug)
	lagoonVersion = helpers.GetEnv("LAGOON_VERSION", lagoonVersion, debug)
	savedTemplates = helpers.GetEnv("YAML_FOLDER", savedTemplates, debug)

	// read the .lagoon.yml file
	var lYAML lagoon.YAML
	lPolysite := make(map[string]interface{})
	if err := lagoon.UnmarshalLagoonYAML(lagoonYml, &lYAML, &lPolysite); err != nil {
		return fmt.Errorf("couldn't read file %v: %v", lagoonYml, err)
	}

	// get or generate the values file for generating route templates
	lagoonValues := lagoon.BuildValues{}
	if checkValuesFile {
		if debug {
			fmt.Println(fmt.Sprintf("Collecting values for templating from %s", fmt.Sprintf("%s/%s", templateValues, "values.yaml")))
		}
		lagoonValues = routeTemplater.ReadValuesFile(fmt.Sprintf("%s/%s", templateValues, "values.yaml"))
	} else {
		lagoonValues.Project = projectName
		lagoonValues.Environment = environmentName
		lagoonValues.EnvironmentType = environmentType
		lagoonValues.BuildType = buildType
		lagoonValues.LagoonVersion = lagoonVersion
		switch buildType {
		case "branch", "promote":
			lagoonValues.Branch = branch
		case "pullrequest":
			lagoonValues.PRNumber = prNumber
			lagoonValues.PRHeadBranch = prHeadBranch
			lagoonValues.PRBaseBranch = prBaseBranch
		}
	}

	if lagoonValues.Project == "" || lagoonValues.Environment == "" || lagoonValues.EnvironmentType == "" || lagoonValues.BuildType == "" {
		return fmt.Errorf("Missing arguments: project-name, environment-name, environment-type, or build-type not defined")
	}
	switch lagoonValues.BuildType {
	case "branch", "promote":
		if lagoonValues.Branch == "" {
			return fmt.Errorf("Missing arguments: branch not defined")
		}
	case "pullrequest":
		if lagoonValues.PRNumber == "" || lagoonValues.PRHeadBranch == "" || lagoonValues.PRBaseBranch == "" {
			return fmt.Errorf("Missing arguments: pullrequest-number, pullrequest-head-branch, or pullrequest-base-branch not defined")
		}
	}

	// get the project and environment variables
	projectVariables = helpers.GetEnv("LAGOON_PROJECT_VARIABLES", projectVariables, debug)
	environmentVariables = helpers.GetEnv("LAGOON_ENVIRONMENT_VARIABLES", environmentVariables, debug)

	// by default, environment routes are not monitored
	monitoringEnabled = false
	if lagoonValues.EnvironmentType == "production" {
		// if this is a production environment, monitoring IS enabled
		monitoringEnabled = true
		// check if the environment is active or standby
		if lagoonValues.Environment == activeEnvironment {
			activeEnv = true
		}
		if lagoonValues.Environment == standbyEnvironment {
			standbyEnv = true
		}
	}

	// unmarshal and then merge the two so there is only 1 set of variables to iterate over
	projectVars := []lagoon.EnvironmentVariable{}
	envVars := []lagoon.EnvironmentVariable{}
	json.Unmarshal([]byte(projectVariables), &projectVars)
	json.Unmarshal([]byte(environmentVariables), &envVars)
	lagoonEnvVars := lagoon.MergeVariables(projectVars, envVars)

	// read the routes from the API
	var apiRoutes lagoon.RoutesV2
	lagoonRoutesJSON, err := lagoon.GetLagoonVariable("LAGOON_ROUTES_JSON", []string{"build", "global"}, lagoonEnvVars)
	if lagoonRoutesJSON != nil {
		if debug {
			fmt.Println("Collecting routes from environment variable LAGOON_ROUTES_JSON")
		}
		// if the routesJSON is populated, then attempt to decode and unmarshal it
		rawJSONStr, _ := base64.StdEncoding.DecodeString(lagoonRoutesJSON.Value)
		rawJSON := []byte(rawJSONStr)
		err := json.Unmarshal(rawJSON, &apiRoutes)
		if err != nil {
			return fmt.Errorf("couldn't unmarshal routes from Lagoon API, is it actually JSON that has been base64 encoded?: %v", err)
		}
	}

	// handle routes from the .lagoon.yml and the API specifically
	newRoutes := &lagoon.RoutesV2{}
	if _, ok := lPolysite[lagoonValues.Project]; ok {
		// POLYSITE: if this is polysite, then the `projectname` routes block should be defined
		strA, _ := yaml.Marshal(lPolysite[lagoonValues.Project])
		var lYAMLPolysite lagoon.YAML
		err = yaml.Unmarshal(strA, &lYAMLPolysite)
		if err != nil {
			return fmt.Errorf("couldn't unmarshal for polysite %v: %v", strA, err)
		}
		for _, routeMap := range lYAMLPolysite.Environments[lagoonValues.Environment].Routes {
			lagoon.GenerateRoutesV2(newRoutes, routeMap, lagoonEnvVars, fastlyAPISecretPrefix, false)
		}
	} else {
		// otherwise it just uses the default environment name
		for _, routeMap := range lYAML.Environments[lagoonValues.Environment].Routes {
			lagoon.GenerateRoutesV2(newRoutes, routeMap, lagoonEnvVars, fastlyAPISecretPrefix, false)
		}
	}
	// merge routes from the API on top of the routes from the `.lagoon.yml`
	finalRoutes := lagoon.MergeRoutesV2(*newRoutes, apiRoutes)
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
		activeStanbyRoutes := &lagoon.RoutesV2{}
		if lYAML.ProductionRoutes != nil {
			if activeEnv == true {
				if lYAML.ProductionRoutes.Active != nil {
					if lYAML.ProductionRoutes.Active.Routes != nil {
						for _, routeMap := range lYAML.ProductionRoutes.Active.Routes {
							lagoon.GenerateRoutesV2(activeStanbyRoutes, routeMap, lagoonEnvVars, fastlyAPISecretPrefix, true)
						}
					}
				}
			}
			if standbyEnv == true {
				if lYAML.ProductionRoutes.Standby != nil {
					if lYAML.ProductionRoutes.Standby.Routes != nil {
						for _, routeMap := range lYAML.ProductionRoutes.Standby.Routes {
							lagoon.GenerateRoutesV2(activeStanbyRoutes, routeMap, lagoonEnvVars, fastlyAPISecretPrefix, true)
						}
					}
				}
			}
		}
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
	routeGeneration.Flags().BoolVarP(&checkValuesFile, "check-values-file", "C", false,
		"If set, will check for the values file defined in `${template-path}/values.yaml`")
}
