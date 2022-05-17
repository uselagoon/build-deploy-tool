package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	"sigs.k8s.io/yaml"
)

// collectBuildValues is used to collect variables and values that are used within a build
func collectBuildValues(debug bool, activeEnv, standbyEnv *bool,
	lagoonEnvVars *[]lagoon.EnvironmentVariable,
	lagoonValues *lagoon.BuildValues,
	lYAML *lagoon.YAML,
) error {
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
	lagoonVersion = helpers.GetEnv("LAGOON_VERSION", lagoonVersion, debug)

	// read the .lagoon.yml file
	lPolysite := make(map[string]interface{})
	if err := lagoon.UnmarshalLagoonYAML(lagoonYml, lYAML, &lPolysite); err != nil {
		return fmt.Errorf("couldn't read file %v: %v", lagoonYml, err)
	}

	// if this is a polysite, then unmarshal the polysite data into a normal lagoon environments yaml
	if _, ok := lPolysite[projectName]; ok {
		s, _ := yaml.Marshal(lPolysite[projectName])
		_ = yaml.Unmarshal(s, &lYAML)
	}

	lCompose := lagoon.Compose{}
	// unmarshal the docker-compose.yml file
	if err := lagoon.UnmarshaDockerComposeYAML(lYAML.DockerComposeYAML, &lCompose); err != nil {
		return err
	}

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

	if projectName == "" || environmentName == "" || environmentType == "" || buildType == "" {
		return fmt.Errorf("Missing arguments: project-name, environment-name, environment-type, or build-type not defined")
	}
	switch buildType {
	case "branch", "promote":
		if branch == "" {
			return fmt.Errorf("Missing arguments: branch not defined")
		}
	case "pullrequest":
		if prNumber == "" || prHeadBranch == "" || prBaseBranch == "" {
			return fmt.Errorf("Missing arguments: pullrequest-number, pullrequest-head-branch, or pullrequest-base-branch not defined")
		}
	}

	// get the project and environment variables
	projectVariables = helpers.GetEnv("LAGOON_PROJECT_VARIABLES", projectVariables, debug)
	environmentVariables = helpers.GetEnv("LAGOON_ENVIRONMENT_VARIABLES", environmentVariables, debug)

	// by default, environment routes are not monitored
	monitoringEnabled = false
	if environmentType == "production" {
		// if this is a production environment, monitoring IS enabled
		monitoringEnabled = true
		// check if the environment is active or standby
		if environmentName == activeEnvironment {
			*activeEnv = true
		}
		if environmentName == standbyEnvironment {
			*standbyEnv = true
		}
	}

	// unmarshal and then merge the two so there is only 1 set of variables to iterate over
	projectVars := []lagoon.EnvironmentVariable{}
	envVars := []lagoon.EnvironmentVariable{}
	json.Unmarshal([]byte(projectVariables), &projectVars)
	json.Unmarshal([]byte(environmentVariables), &envVars)
	*lagoonEnvVars = lagoon.MergeVariables(projectVars, envVars)

	// create the services map
	lagoonValues.Services = make(map[string]lagoon.ServiceValues)
	lagoonServiceTypes, _ := lagoon.GetLagoonVariable("LAGOON_SERVICE_TYPES", []string{"build"}, *lagoonEnvVars)
	// convert docker-compose services to servicevalues
	for csName, csValues := range lCompose.Services {
		cService, err := composeToServiceValues(lYAML, lagoonServiceTypes, csName, csValues)
		if err != nil {
			return err
		}
		lagoonValues.Services[csName] = cService
	}
	return nil
}

func composeToServiceValues(lYAML *lagoon.YAML, lagoonServiceTypes *lagoon.EnvironmentVariable, csName string, csValues lagoon.Service) (lagoon.ServiceValues, error) {
	lagoonType := lagoon.CheckServiceLagoonLabel(csValues.Labels, "lagoon.type")
	if lagoonType != "" {
		if value, ok := lYAML.Environments[environmentName].Types[csName]; ok {
			lagoonType = value
		}
		if lagoonServiceTypes != nil {
			serviceTypesSplit := strings.Split(lagoonServiceTypes.Value, ",")
			for _, sType := range serviceTypesSplit {
				sTypeSplit := strings.Split(sType, ":")
				if sTypeSplit[0] == csName {
					lagoonType = sTypeSplit[1]
				}
			}
		}
		cService := lagoon.ServiceValues{
			Name: csName,
			Type: lagoonType,
		}
		serviceAutogenerated := lagoon.CheckServiceLagoonLabel(csValues.Labels, "lagoon.autogeneratedroute")
		if reflect.TypeOf(serviceAutogenerated).Kind() == reflect.String {
			vBool, err := strconv.ParseBool(serviceAutogenerated)
			if err == nil {
				cService.AutogeneratedRoutes = &vBool
			}
		}
		lagoonTypeName := lagoon.CheckServiceLagoonLabel(csValues.Labels, "lagoon.name")
		if lagoonTypeName != "" {
			// if there is a lagoon.name label on a service, then this service does not need an autogenerated route
			// as it is likely an ancillary pod for another service (the one in the `lagoon.name` label :))
			cService.AutogeneratedRoutes = helpers.BoolPtr(false)
		}
		return cService, nil
	}
	return lagoon.ServiceValues{}, fmt.Errorf("Service %s has no `lagoon.type` label in the docker-compose.yml file", csName)
}

func unsetEnvVars(localVars []struct {
	name  string
	value string
}) {
	varNames := []string{"MONITORING_ALERTCONTACT", "MONITORING_STATUSPAGEID",
		"PROJECT", "ENVIRONMENT", "BRANCH", "PR_NUMBER", "PR_HEAD_BRANCH",
		"PR_BASE_BRANCH", "ENVIRONMENT_TYPE", "BUILD_TYPE", "ACTIVE_ENVIRONMENT",
		"STANDBY_ENVIRONMENT", "LAGOON_FASTLY_NOCACHE_SERVICE_ID", "LAGOON_PROJECT_VARIABLES",
		"LAGOON_ENVIRONMENT_VARIABLES", "LAGOON_VERSION",
	}
	for _, varName := range varNames {
		os.Unsetenv(varName)
	}
	for _, varName := range localVars {
		os.Unsetenv(varName.name)
	}
}
