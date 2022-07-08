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
		return BackupTemplateGeneration(true)
	},
}

// BackupTemplateGeneration .
func BackupTemplateGeneration(debug bool,
) error {
	lagoonBuild, err := generator.NewGenerator(
		lagoonYml,
		projectVariables,
		environmentVariables,
		projectName,
		environmentName,
		environmentType,
		activeEnvironment,
		standbyEnvironment,
		buildType,
		branch,
		prNumber,
		prTitle,
		prHeadBranch,
		prBaseBranch,
		lagoonVersion,
		defaultBackupSchedule,
		hourlyDefaultBackupRetention,
		dailyDefaultBackupRetention,
		weeklyDefaultBackupRetention,
		monthlyDefaultBackupRetention,
		monitoringContact,
		monitoringStatusPageID,
		fastlyCacheNoCahce,
		fastlyAPISecretPrefix,
		fastlyServiceID,
		ignoreNonStringKeyErrors,
		ignoreMissingEnvFiles,
		debug,
	)
	if err != nil {
		return err
	}

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
