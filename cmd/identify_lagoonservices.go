package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	generator "github.com/uselagoon/build-deploy-tool/internal/generator"
	servicestemplates "github.com/uselagoon/build-deploy-tool/internal/templating"
)

type identifyServices struct {
	Deployments []string `json:"deployments,omitempty"`
	Volumes     []string `json:"volumes,omitempty"`
	Services    []string `json:"services,omitempty"`
}

var lagoonServiceIdentify = &cobra.Command{
	Use:     "lagoon-services",
	Aliases: []string{"ls"},
	Short:   "Identify the lagoon services for a Lagoon build",
	RunE: func(cmd *cobra.Command, args []string) error {
		gen, err := generator.GenerateInput(*rootCmd, false)
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
		gen.ImageReferences = imageRefs.Images
		out, err := LagoonServiceTemplateIdentification(gen)
		if err != nil {
			return err
		}
		b, _ := json.Marshal(out)
		fmt.Println(string(b))
		return nil
	},
}

// LagoonServiceTemplateIdentification takes the output of the generator and returns a JSON payload that contains information
// about the services that lagoon will be deploying (this will be kubernetes `kind: deployment`, but lagoon calls them services ¯\_(ツ)_/¯)
// this command can be used to identify services that are deployed by the build, so that services that may remain in the environment can be identified
// and eventually removed
func LagoonServiceTemplateIdentification(g generator.GeneratorInput) (*identifyServices, error) {

	servicesData := identifyServices{}
	lagoonBuild, err := generator.NewGenerator(
		g,
	)
	if err != nil {
		return nil, err
	}

	deployments, err := servicestemplates.GenerateDeploymentTemplate(*lagoonBuild.BuildValues)
	if err != nil {
		return nil, fmt.Errorf("couldn't identify deployments: %v", err)
	}
	for _, d := range deployments {
		servicesData.Deployments = append(servicesData.Deployments, d.ObjectMeta.Name)
	}
	pvcs, err := servicestemplates.GeneratePVCTemplate(*lagoonBuild.BuildValues)
	if err != nil {
		return nil, fmt.Errorf("couldn't identify volumes: %v", err)
	}
	for _, pvc := range pvcs {
		servicesData.Volumes = append(servicesData.Volumes, pvc.ObjectMeta.Name)
	}
	services, err := servicestemplates.GenerateServiceTemplate(*lagoonBuild.BuildValues)
	if err != nil {
		return nil, fmt.Errorf("couldn't identify services: %v", err)
	}
	for _, service := range services {
		servicesData.Services = append(servicesData.Services, service.ObjectMeta.Name)
	}
	return &servicesData, nil
}

func init() {
	identifyCmd.AddCommand(lagoonServiceIdentify)
}
