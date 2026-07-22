package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/uselagoon/build-deploy-tool/internal/collector"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/identify"
	"github.com/uselagoon/build-deploy-tool/internal/k8s"
	"github.com/uselagoon/machinery/api/schema"
)

type LagoonServices struct {
	Services []schema.EnvironmentService `json:"services"`
	Volumes  []schema.EnvironmentVolume  `json:"volumes"`
}

var lagoonServiceIdentify = &cobra.Command{
	Use:     "lagoon-services",
	Aliases: []string{"ls"},
	Short:   "Identify the lagoon services for a Lagoon build",
	RunE: func(cmd *cobra.Command, args []string) error {
		gen, err := GenerateInput(*rootCmd, false)
		if err != nil {
			return err
		}
		client, err := k8s.NewClient()
		if err != nil {
			return err
		}
		// create a collector
		col := collector.NewCollector(client)
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
		services, _, _, _, _, _, _, _, err := identify.GetCurrentState(col, gen)
		if err != nil {
			return err
		}
		servs, _ := json.Marshal(services)
		fmt.Println(string(servs))
		return nil
	},
}

func init() {
	identifyCmd.AddCommand(lagoonServiceIdentify)
}
