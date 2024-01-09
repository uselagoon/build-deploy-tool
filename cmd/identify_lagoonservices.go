package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	generator "github.com/uselagoon/build-deploy-tool/internal/generator"
	servicestemplates "github.com/uselagoon/build-deploy-tool/internal/templating/services"
)

type identifyServices struct {
	Name       string       `json:"name"`
	Type       string       `json:"type"`
	Containers []containers `json:"containers,omitempty"`
}

type containers struct {
	Name  string  `json:"name"`
	Ports []ports `json:"ports,omitempty"`
}

type ports struct {
	Port int32 `json:"port"`
}

var lagoonServiceIdentify = &cobra.Command{
	Use:     "lagoon-services",
	Aliases: []string{"ls"},
	Short:   "Identify the lagoon services for a Lagoon build",
	RunE: func(cmd *cobra.Command, args []string) error {
		gen, err := generator.GenerateInput(*rootCmd, true)
		if err != nil {
			return err
		}
		images, err := rootCmd.PersistentFlags().GetString("images")
		if err != nil {
			return fmt.Errorf("error reading images flag: %v", err)
		}
		var imageRefs struct {
			Images map[string]string `json:"images"`
		}
		imagesStr, _ := base64.StdEncoding.DecodeString(images)
		json.Unmarshal(imagesStr, &imageRefs)
		gen.ImageReferences = imageRefs.Images
		out, err := LagoonServiceTemplateIdentification(gen)
		if err != nil {
			return err
		}
		fmt.Println(out)
		return nil
	},
}

// LagoonServiceTemplateIdentification .
func LagoonServiceTemplateIdentification(g generator.GeneratorInput) ([]identifyServices, error) {

	lServices := []identifyServices{}
	lagoonBuild, err := generator.NewGenerator(
		g,
	)
	if err != nil {
		return nil, err
	}

	// generate the templates
	deployments, err := servicestemplates.GenerateDeploymentTemplate(*lagoonBuild.BuildValues)
	if err != nil {
		return nil, fmt.Errorf("couldn't generate template: %v", err)
	}
	if deployments != nil {
		for _, d := range deployments {
			dcs := []containers{}
			for _, dc := range d.Spec.Template.Spec.Containers {
				dcp := []ports{}
				for _, p := range dc.Ports {
					dcp = append(dcp, ports{Port: p.ContainerPort})
				}
				dcs = append(dcs, containers{Name: dc.Name, Ports: dcp})
			}
			lServices = append(lServices, identifyServices{
				Name:       d.Name,
				Type:       d.ObjectMeta.Labels["lagoon.sh/service-type"],
				Containers: dcs,
			})
		}
	}
	return lServices, nil
}

func init() {
	identifyCmd.AddCommand(lagoonServiceIdentify)
}
