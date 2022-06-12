package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	routeTemplater "github.com/uselagoon/build-deploy-tool/internal/templating/routes"
)

var routeGeneration = &cobra.Command{
	Use:     "ingress",
	Aliases: []string{"i"},
	Short:   "Generate the ingress templates for a Lagoon build",
	RunE: func(cmd *cobra.Command, args []string) error {
		return IngressTemplateGeneration(true)
	},
}

func generateIngress(values lagoon.BuildValues, lYAML lagoon.YAML, envVars []lagoon.EnvironmentVariable, mainRoutes *lagoon.RoutesV2, debug bool) error {
	// read the routes from the API
	apiRoutes, err := getRoutesFromAPIEnvVar(envVars, debug)
	if err != nil {
		return fmt.Errorf("couldn't unmarshal routes from Lagoon API, is it actually JSON that has been base64 encoded?: %v", err)
	}

	// handle routes from the .lagoon.yml and the API specifically
	*mainRoutes, err = generateAndMerge(*apiRoutes, envVars, lYAML, values)
	if err != nil {
		return fmt.Errorf("couldn't generate and merge routes: %v", err)
	}
	return nil
}

// IngressTemplateGeneration .
func IngressTemplateGeneration(debug bool) error {
	activeEnv := false
	standbyEnv := false

	lagoonEnvVars := []lagoon.EnvironmentVariable{}
	lagoonValues := lagoon.BuildValues{}
	lYAML := lagoon.YAML{}
	autogenRoutes := &lagoon.RoutesV2{}
	mainRoutes := &lagoon.RoutesV2{}
	activeStandbyRoutes := &lagoon.RoutesV2{}
	err := collectBuildValues(debug, &activeEnv, &standbyEnv, &lagoonEnvVars, &lagoonValues, &lYAML, autogenRoutes, mainRoutes, activeStandbyRoutes, ignoreNonStringKeyErrors)
	if err != nil {
		return err
	}

	// generate the templates
	for _, route := range mainRoutes.Routes {
		if debug {
			fmt.Println(fmt.Sprintf("Templating ingress manifest for %s to %s", route.Domain, fmt.Sprintf("%s/%s.yaml", savedTemplates, route.Domain)))
		}
		templateYAML, err := routeTemplater.GenerateIngressTemplate(route, lagoonValues, monitoringContact, monitoringStatusPageID, monitoringEnabled)
		if err != nil {
			return fmt.Errorf("couldn't generate template: %v", err)
		}
		routeTemplater.WriteTemplateFile(fmt.Sprintf("%s/%s.yaml", savedTemplates, route.Domain), templateYAML)
	}
	if activeEnv || standbyEnv {
		// active/standby routes should not be changed by any environment defined routes.
		// generate the templates for these independently of any previously generated routes,
		// this WILL overwrite previously created templates ensuring that anything defined in the `production_routes`
		// section are created correctly ensuring active/standby will work
		// generate the templates for active/standby routes separately to normal routes
		for _, route := range activeStandbyRoutes.Routes {
			if debug {
				fmt.Println(fmt.Sprintf("Templating active/standby ingress manifest for %s to %s", route.Domain, fmt.Sprintf("%s/%s.yaml", savedTemplates, route.Domain)))
			}
			templateYAML, err := routeTemplater.GenerateIngressTemplate(route, lagoonValues, monitoringContact, monitoringStatusPageID, monitoringEnabled)
			if err != nil {
				return fmt.Errorf("couldn't generate template: %v", err)
			}
			routeTemplater.WriteTemplateFile(fmt.Sprintf("%s/%s.yaml", savedTemplates, route.Domain), templateYAML)
		}
	}
	return nil
}

func init() {
	templateCmd.AddCommand(routeGeneration)
}
