package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	generator "github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	dbaasTemplater "github.com/uselagoon/build-deploy-tool/internal/templating/dbaas"
)

var dbaasGeneration = &cobra.Command{
	Use:     "dbaas",
	Aliases: []string{"db"},
	Short:   "Generate the DBaaS templates for a Lagoon build",
	RunE: func(cmd *cobra.Command, args []string) error {
		generator, err := generator.GenerateInput(*rootCmd, true)
		if err != nil {
			return err
		}
		return DBaaSTemplateGeneration(generator)
	},
}

// DBaaSTemplateGeneration .
func DBaaSTemplateGeneration(g generator.GeneratorInput,
) error {
	lagoonBuild, err := generator.NewGenerator(
		g,
	)
	if err != nil {
		return err
	}
	savedTemplates := g.SavedTemplatesPath

	templateYAML, err := dbaasTemplater.GenerateDBaaSTemplate(*lagoonBuild.BuildValues)
	if err != nil {
		return fmt.Errorf("couldn't generate template: %v", err)
	}
	helpers.WriteTemplateFile(fmt.Sprintf("%s/%s.yaml", savedTemplates, "dbaas"), templateYAML)
	if g.Debug {
		fmt.Println(fmt.Sprintf("Templating dbaas consumers to %s", fmt.Sprintf("%s/%s.yaml", savedTemplates, "dbaas")))
	}
	return nil
}

func init() {
	templateCmd.AddCommand(dbaasGeneration)
}
