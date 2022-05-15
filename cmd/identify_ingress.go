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
}
