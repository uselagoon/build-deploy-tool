package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	generator "github.com/uselagoon/build-deploy-tool/internal/generator"
	servicestemplates "github.com/uselagoon/build-deploy-tool/internal/templating"
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
		imagesStr, err := base64.StdEncoding.DecodeString(images)
		if err != nil {
			return fmt.Errorf("error decoding images payload: %v", err)
		}
		if err := json.Unmarshal(imagesStr, &imageRefs); err != nil {
			return fmt.Errorf("error unmarshalling images payload: %v", err)
		}
		gen.ImageReferences = imageRefs.Images
		out, err := LagoonServiceTemplateIdentification(gen)
		if err != nil {
			return err
		}
		fmt.Println(out)
		return nil
	},
}

// LagoonServiceTemplateIdentification takes the output of the generator and returns a JSON payload that contains information
// about the services that lagoon will be deploying (this will be kubernetes `kind: deployment`, but lagoon calls them services ¯\_(ツ)_/¯)
// this command can be used to identify services that are deployed by the build, so that services that may remain in the environment can be identified
// and eventually removed
func LagoonServiceTemplateIdentification(g generator.GeneratorInput) ([]identifyServices, error) {

	lServices := []identifyServices{}
	lagoonBuild, err := generator.NewGenerator(
		g,
	)
	if err != nil {
		return nil, err
	}

	deployments, err := servicestemplates.GenerateDeploymentTemplate(*lagoonBuild.BuildValues)
	if err != nil {
		return nil, fmt.Errorf("couldn't generate template: %v", err)
	}
	for _, d := range deployments {
		dcs := []containers{}
		for _, dc := range d.Spec.Template.Spec.Containers {
			dcp := []ports{}
			for _, p := range dc.Ports {
				dcp = append(dcp, ports{Port: p.ContainerPort})
			}
			dcs = append(dcs, containers{
				Name:  dc.Name,
				Ports: dcp,
			})
		}
		lServices = append(lServices, identifyServices{
			Name:       d.Name,
			Type:       d.ObjectMeta.Labels["lagoon.sh/service-type"],
			Containers: dcs,
		})
	}
	return lServices, nil
}

func init() {
	identifyCmd.AddCommand(lagoonServiceIdentify)
}
