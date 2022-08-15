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
		return DBaaSTemplateGeneration(true)
	},
}

// DBaaSTemplateGeneration .
func DBaaSTemplateGeneration(debug bool,
) error {
	lagoonBuild, err := generator.NewGenerator(
		lagoonYml,
		lagoonYmlOverride,
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

	templateYAML, err := dbaasTemplater.GenerateDBaaSTemplate(*lagoonBuild.BuildValues)
	if err != nil {
		return fmt.Errorf("couldn't generate template: %v", err)
	}
	helpers.WriteTemplateFile(fmt.Sprintf("%s/%s.yaml", savedTemplates, "dbaas"), templateYAML)
	return nil
}

func init() {
	templateCmd.AddCommand(dbaasGeneration)
}
