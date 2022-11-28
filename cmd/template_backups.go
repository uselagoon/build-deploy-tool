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
		k8upVersion, err := cmd.Flags().GetString("version")
		if err != nil {
			return fmt.Errorf("error reading domain flag: %v", err)
		}
		generator, err := generatorInput(true)
		if err != nil {
			return err
		}
		generator.BackupConfiguration.K8upVersion = k8upVersion
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
	helpers.WriteTemplateFile(fmt.Sprintf("%s/%s.yaml", savedTemplates, "k8up-lagoon-backup-schedule"), templateYAML)
	return nil
}

func init() {
	templateCmd.AddCommand(backupGeneration)
	backupGeneration.Flags().StringP("version", "", "v1", "The version of k8up used.")
}
