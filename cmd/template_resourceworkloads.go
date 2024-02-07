package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	generator "github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	hpatemplate "github.com/uselagoon/build-deploy-tool/internal/templating/resources/hpa"
	pdbtemplate "github.com/uselagoon/build-deploy-tool/internal/templating/resources/pdb"
)

var resourceWorkloadGeneration = &cobra.Command{
	Use:     "resource-workloads",
	Aliases: []string{"rw"},
	Short:   "Generate the resource workload templates for a Lagoon build",
	RunE: func(cmd *cobra.Command, args []string) error {
		generator, err := generator.GenerateInput(*rootCmd, true)
		if err != nil {
			return err
		}
		return ResourceWorkloadTemplateGeneration(generator)
	},
}

// IngressTemplateGeneration .
func ResourceWorkloadTemplateGeneration(g generator.GeneratorInput) error {
	lagoonBuild, err := generator.NewGenerator(
		g,
	)
	if err != nil {
		return err
	}
	savedTemplates := g.SavedTemplatesPath

	// generate the templates
	if g.Debug {
		fmt.Println(fmt.Sprintf("Templating HPA manifests to %s", fmt.Sprintf("%s/%s.yaml", savedTemplates, "hpas")))
	}
	templateYAML, err := hpatemplate.GenerateHPATemplate(*lagoonBuild.BuildValues)
	if err != nil {
		return fmt.Errorf("couldn't generate template: %v", err)
	}
	helpers.WriteTemplateFile(fmt.Sprintf("%s/%s.yaml", savedTemplates, "hpas"), templateYAML)
	if g.Debug {
		fmt.Println(fmt.Sprintf("Templating HPA manifests to %s", fmt.Sprintf("%s/%s.yaml", savedTemplates, "pdbs")))
	}
	templateYAML, err = pdbtemplate.GeneratePDBTemplate(*lagoonBuild.BuildValues)
	if err != nil {
		return fmt.Errorf("couldn't generate template: %v", err)
	}
	helpers.WriteTemplateFile(fmt.Sprintf("%s/%s.yaml", savedTemplates, "pdbs"), templateYAML)
	return nil
}

func init() {
	templateCmd.AddCommand(resourceWorkloadGeneration)
}
