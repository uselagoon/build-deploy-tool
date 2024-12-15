package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	generator "github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	servicestemplates "github.com/uselagoon/build-deploy-tool/internal/templating"
)

var routeGeneration = &cobra.Command{
	Use:     "ingress",
	Aliases: []string{"i"},
	Short:   "Generate the ingress templates for a Lagoon build",
	RunE: func(cmd *cobra.Command, args []string) error {
		generator, err := generator.GenerateInput(*rootCmd, true)
		if err != nil {
			return err
		}
		return IngressTemplateGeneration(generator)
	},
}

// IngressTemplateGeneration .
func IngressTemplateGeneration(g generator.GeneratorInput) error {
	lagoonBuild, err := generator.NewGenerator(
		g,
	)
	if err != nil {
		return err
	}
	savedTemplates := g.SavedTemplatesPath
	// generate the templates
	for _, route := range lagoonBuild.MainRoutes.Routes {
		if g.Debug {
			fmt.Printf("Templating ingress manifest for %s to %s\n", route.Domain, fmt.Sprintf("%s/%s.yaml", savedTemplates, route.Domain))
		}
		ingress, err := servicestemplates.GenerateIngressTemplate(route, *lagoonBuild.BuildValues)
		if err != nil {
			return fmt.Errorf("couldn't generate template: %v", err)
		}
		templateYAML, err := servicestemplates.TemplateIngress(ingress)
		if err != nil {
			return fmt.Errorf("couldn't generate template: %v", err)
		}
		helpers.WriteTemplateFile(fmt.Sprintf("%s/%s.yaml", savedTemplates, route.Domain), templateYAML)
	}
	if *lagoonBuild.ActiveEnvironment || *lagoonBuild.StandbyEnvironment {
		// active/standby routes should not be changed by any environment defined routes.
		// generate the templates for these independently of any previously generated routes,
		// this WILL overwrite previously created templates ensuring that anything defined in the `production_routes`
		// section are created correctly ensuring active/standby will work
		// generate the templates for active/standby routes separately to normal routes
		for _, route := range lagoonBuild.ActiveStandbyRoutes.Routes {
			if g.Debug {
				fmt.Printf("Templating active/standby ingress manifest for %s to %s\n", route.Domain, fmt.Sprintf("%s/%s.yaml", savedTemplates, route.Domain))
			}
			ingress, err := servicestemplates.GenerateIngressTemplate(route, *lagoonBuild.BuildValues)
			if err != nil {
				return fmt.Errorf("couldn't generate template: %v", err)
			}
			templateYAML, err := servicestemplates.TemplateIngress(ingress)
			if err != nil {
				return fmt.Errorf("couldn't generate template: %v", err)
			}
			helpers.WriteTemplateFile(fmt.Sprintf("%s/%s.yaml", savedTemplates, route.Domain), templateYAML)
		}
	}
	return nil
}

func init() {
	templateCmd.AddCommand(routeGeneration)
}
