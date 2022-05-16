package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
)

// generateActiveStandby generates the active/standby routes from the provided lagoon yaml
func generateActiveStandby(
	active bool,
	standby bool,
	envVars []lagoon.EnvironmentVariable,
	lagoonYAML lagoon.YAML,
) lagoon.RoutesV2 {
	activeStanbyRoutes := &lagoon.RoutesV2{}
	if lagoonYAML.ProductionRoutes != nil {
		if active == true {
			if lagoonYAML.ProductionRoutes.Active != nil {
				if lagoonYAML.ProductionRoutes.Active.Routes != nil {
					for _, routeMap := range lagoonYAML.ProductionRoutes.Active.Routes {
						lagoon.GenerateRoutesV2(activeStanbyRoutes, routeMap, envVars, fastlyAPISecretPrefix, true)
					}
				}
			}
		}
		if standby == true {
			if lagoonYAML.ProductionRoutes.Standby != nil {
				if lagoonYAML.ProductionRoutes.Standby.Routes != nil {
					for _, routeMap := range lagoonYAML.ProductionRoutes.Standby.Routes {
						lagoon.GenerateRoutesV2(activeStanbyRoutes, routeMap, envVars, fastlyAPISecretPrefix, true)
					}
				}
			}
		}
	}
	return *activeStanbyRoutes
}

// getRoutesFromEnvVar will collect the value of the LAGOON_ROUTES_JSON
// from provided lagoon environment variables from the API
func getRoutesFromAPIEnvVar(
	envVars []lagoon.EnvironmentVariable,
	debug bool,
) (*lagoon.RoutesV2, error) {
	apiRoutes := &lagoon.RoutesV2{}
	lagoonRoutesJSON, _ := lagoon.GetLagoonVariable("LAGOON_ROUTES_JSON", []string{"build", "global"}, envVars)
	if lagoonRoutesJSON != nil {
		if debug {
			fmt.Println("Collecting routes from environment variable LAGOON_ROUTES_JSON")
		}
		// if the routesJSON is populated, then attempt to decode and unmarshal it
		rawJSONStr, _ := base64.StdEncoding.DecodeString(lagoonRoutesJSON.Value)
		rawJSON := []byte(rawJSONStr)
		err := json.Unmarshal(rawJSON, apiRoutes)
		if err != nil {
			return nil, fmt.Errorf("couldn't unmarshal routes from Lagoon API, is it actually JSON that has been base64 encoded?: %v", err)
		}
	}
	return apiRoutes, nil
}

// generateAndMerge generates the completed custom ingress for an environment
// it generates the custom ingress from lagoon yaml and also merges in any that were
// provided by the lagoon environment variables from the API
func generateAndMerge(
	api lagoon.RoutesV2,
	envVars []lagoon.EnvironmentVariable,
	lagoonYAML lagoon.YAML,
	lagoonValues lagoon.BuildValues,
) (*lagoon.RoutesV2, error) {
	n := &lagoon.RoutesV2{} // placeholder for generated routes

	// otherwise it just uses the default environment name
	for _, routeMap := range lagoonYAML.Environments[lagoonValues.Branch].Routes {
		lagoon.GenerateRoutesV2(n, routeMap, envVars, fastlyAPISecretPrefix, false)
	}
	// merge routes from the API on top of the routes from the `.lagoon.yml`
	merged := lagoon.MergeRoutesV2(*n, api, envVars, fastlyAPISecretPrefix)
	return &merged, nil
}
