package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	generator "github.com/uselagoon/build-deploy-tool/internal/generator"
	servicestemplates "github.com/uselagoon/build-deploy-tool/internal/templating"
)

// EnvironmentService this should use `schema.EnvironmentService` and machinery gets extended to support ports and volumes
type EnvironmentService struct {
	Name       string             `json:"name,omitempty"`
	Type       string             `json:"type,omitempty"`
	Containers []ServiceContainer `json:"containers,omitempty"`
}

type ServiceContainer struct {
	Name    string                   `json:"name,omitempty"`
	Ports   []ServiceContainerPort   `json:"ports,omitempty"`
	Volumes []ServiceContainerVolume `json:"volumes,omitempty"`
}

type ServiceContainerVolume struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path,omitempty"`
}

type ServiceContainerPort struct {
	Port     int32  `json:"port,omitempty"`
	Protocol string `json:"protocol,omitempty"`
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
func LagoonServiceTemplateIdentification(g generator.GeneratorInput) ([]EnvironmentService, error) {

	lServices := []EnvironmentService{}
	lagoonBuild, err := generator.NewGenerator(
		g,
	)
	if err != nil {
		return nil, err
	}
	pvcs, err := servicestemplates.GeneratePVCTemplate(*lagoonBuild.BuildValues)
	if err != nil {
		return nil, fmt.Errorf("couldn't generate template: %v", err)
	}

	deployments, err := servicestemplates.GenerateDeploymentTemplate(*lagoonBuild.BuildValues)
	if err != nil {
		return nil, fmt.Errorf("couldn't generate template: %v", err)
	}
	for _, d := range deployments {
		dcs := []ServiceContainer{}
		for _, dc := range d.Spec.Template.Spec.Containers {
			dcv := []ServiceContainerVolume{}
			for _, v := range dc.VolumeMounts {
				// only add volumes that have been defined by the user to the service output
				for _, pvc := range pvcs {
					if pvc.Name == v.Name {
						dcv = append(dcv, ServiceContainerVolume{
							Name: v.Name,
							Path: v.MountPath,
						})
					}
				}
			}

			dcp := []ServiceContainerPort{}
			for _, p := range dc.Ports {
				dcp = append(dcp, ServiceContainerPort{
					Port:     p.ContainerPort,
					Protocol: string(p.Protocol),
				})
			}
			dcs = append(dcs, ServiceContainer{
				Name:    dc.Name,
				Ports:   dcp,
				Volumes: dcv,
			})
		}
		lServices = append(lServices, EnvironmentService{
			Name:       d.Name,
			Type:       d.ObjectMeta.Labels["lagoon.sh/service-type"],
			Containers: dcs,
		})
	}
	dbaas, err := servicestemplates.GenerateDBaaSTemplate(*lagoonBuild.BuildValues)
	if err != nil {
		return nil, fmt.Errorf("couldn't generate template: %v", err)
	}
	for _, db := range dbaas.MariaDB {
		lServices = append(lServices, EnvironmentService{
			Name: db.Name,
			Type: db.ObjectMeta.Labels["lagoon.sh/service-type"],
		})
	}
	for _, db := range dbaas.MongoDB {
		lServices = append(lServices, EnvironmentService{
			Name: db.Name,
			Type: db.ObjectMeta.Labels["lagoon.sh/service-type"],
		})
	}
	for _, db := range dbaas.PostgreSQL {
		lServices = append(lServices, EnvironmentService{
			Name: db.Name,
			Type: db.ObjectMeta.Labels["lagoon.sh/service-type"],
		})
	}

	return lServices, nil
}

func init() {
	identifyCmd.AddCommand(lagoonServiceIdentify)
}
