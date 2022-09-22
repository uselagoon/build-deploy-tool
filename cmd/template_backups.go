package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	generator "github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	backuptemplate "github.com/uselagoon/build-deploy-tool/internal/templating/backups"
)

var backupGeneration = &cobra.Command{
	Use:     "backup-schedule",
	Aliases: []string{"schedule", "bs"},
	Short:   "Generate the backup schedule templates for a Lagoon build",
	RunE: func(cmd *cobra.Command, args []string) error {
		generator, err := generatorInput(true)
		if err != nil {
			return err
		}
		return BackupTemplateGeneration(generator)
	},
}

// BackupTemplateGeneration .
func BackupTemplateGeneration(g generator.GeneratorInput,
) error {
	lagoonBuild, err := generator.NewGenerator(
		g,
	)
	if err != nil {
		return err
	}
	savedTemplates := g.SavedTemplatesPath

	templateYAML, err := backuptemplate.GenerateBackupSchedule(*lagoonBuild.BuildValues)
	if err != nil {
		return fmt.Errorf("couldn't generate template: %v", err)
	}
	helpers.WriteTemplateFile(fmt.Sprintf("%s/%s.yaml", savedTemplates, "backups"), templateYAML)
	return nil
}

func init() {
	templateCmd.AddCommand(backupGeneration)
}
