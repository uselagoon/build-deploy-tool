package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	generator "github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	backuptemplate "github.com/uselagoon/build-deploy-tool/internal/templating/backups"
	"sigs.k8s.io/yaml"
)

type readReplicaValues struct {
	ReadReplicaHosts string `json:"readReplicaHosts"`
}

var backupGeneration = &cobra.Command{
	Use:     "backup-schedule",
	Aliases: []string{"schedule", "bs"},
	Short:   "Generate the backup schedule templates for a Lagoon build",
	RunE: func(cmd *cobra.Command, args []string) error {
		k8upVersion, err := cmd.Flags().GetString("version")
		if err != nil {
			return fmt.Errorf("error reading domain flag: %v", err)
		}
		generator, err := generator.GenerateInput(*rootCmd, true)
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

	// TODO: the dbaas consumers aren't known when the generator runs currently
	// so this is a small helper function to collect this from the build stage
	// this will eventually need to be collected directly by the generator or some other component
	// of the generator in the future, but since backups are the only thing that need to know this at this stage
	repServices := []generator.ServiceValues{}
	for _, s := range lagoonBuild.BuildValues.Services {
		rawYAML, err := os.ReadFile(fmt.Sprintf("/kubectl-build-deploy/%s-values.yaml", s.Name))
		if err != nil {
			repServices = append(repServices, s)
			// skip this one
			continue
		}
		dbaasValues := &readReplicaValues{}
		err = yaml.Unmarshal(rawYAML, dbaasValues)
		if err != nil {
			return fmt.Errorf("couldn't read %v: %v", fmt.Sprintf("/kubectl-build-deploy/%s-values.yaml", s.Name), err)
		}
		if dbaasValues.ReadReplicaHosts != "" {
			s.DBaasReadReplica = true
		}
		repServices = append(repServices, s)
	}
	lagoonBuild.BuildValues.Services = repServices

	// generate the backup schedule templates
	templateYAML, err := backuptemplate.GenerateBackupSchedule(*lagoonBuild.BuildValues)
	if err != nil {
		return fmt.Errorf("couldn't generate template: %v", err)
	}
	if len(templateYAML) > 0 {
		helpers.WriteTemplateFile(fmt.Sprintf("%s/%s.yaml", savedTemplates, "k8up-lagoon-backup-schedule"), templateYAML)
	}

	// generate any prebackuppod templates
	templateYAML, err = backuptemplate.GeneratePreBackupPod(*lagoonBuild.BuildValues)
	if err != nil {
		return fmt.Errorf("couldn't generate template: %v", err)
	}
	if len(templateYAML) > 0 {
		helpers.WriteTemplateFile(fmt.Sprintf("%s/%s.yaml", savedTemplates, "prebackuppods"), templateYAML)
	}
	return nil
}

func init() {
	templateCmd.AddCommand(backupGeneration)
	backupGeneration.Flags().StringP("version", "", "v1", "The version of k8up used.")
}
