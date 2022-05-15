package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
)

// collectBuildValues is used to collect variables and values that are used within a build
func collectBuildValues(debug bool, activeEnv, standbyEnv *bool,
	lagoonEnvVars *[]lagoon.EnvironmentVariable,
	lagoonValues *lagoon.BuildValues,
	lYAML *lagoon.YAML,
	lPolysite *map[string]interface{},
	lCompose *lagoon.Compose,
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
	if err := lagoon.UnmarshalLagoonYAML(lagoonYml, lYAML, lPolysite); err != nil {
		return fmt.Errorf("couldn't read file %v: %v", lagoonYml, err)
	}

	// unmarshal the docker-compose.yml file
	if err := lagoon.UnmarshaDockerComposeYAML(lYAML.DockerComposeYAML, lCompose); err != nil {
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
	// create the services map
	lagoonValues.Services = make(map[string]lagoon.ServiceValues)

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
	return nil
}
