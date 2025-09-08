package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/uselagoon/build-deploy-tool/internal/cleanup"
	"github.com/uselagoon/build-deploy-tool/internal/collector"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/k8s"
)

var cleanupCmd = &cobra.Command{
	Use:     "cleanup",
	Aliases: []string{"clean", "cu", "c"},
	Short:   "Cleanup old services",
	RunE: func(cmd *cobra.Command, args []string) error {
		deleteServices, err := cmd.Flags().GetBool("delete")
		if err != nil {
			return fmt.Errorf("error reading domain flag: %v", err)
		}
		client, err := k8s.NewClient()
		if err != nil {
			return err
		}
		// create a collector
		col := collector.NewCollector(client)
		gen, err := GenerateInput(*rootCmd, false)
		if err != nil {
			return err
		}
		images, err := rootCmd.PersistentFlags().GetString("images")
		if err != nil {
			return fmt.Errorf("error reading images flag: %v", err)
		}
		imageRefs, err := loadImagesFromFile(images)
		if err != nil {
			return err
		}
		namespace := helpers.GetEnv("NAMESPACE", "", false)
		namespace, err = helpers.GetNamespace(namespace, "/var/run/secrets/kubernetes.io/serviceaccount/namespace")
		if err != nil {
			return err
		}
		if namespace == "" {
			return fmt.Errorf("unable to detect namespace")
		}
		gen.Namespace = namespace
		gen.ImageReferences = imageRefs.Images
		_, _, _, _, _, _, err = cleanup.RunCleanup(col, gen, deleteServices)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	runCmd.AddCommand(cleanupCmd)
	cleanupCmd.Flags().Bool("delete", false, "flag to actually delete services")
}
