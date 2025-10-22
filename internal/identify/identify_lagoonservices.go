package identify

import (
	"fmt"

	"github.com/uselagoon/build-deploy-tool/internal/generator"
	servicestemplates "github.com/uselagoon/build-deploy-tool/internal/templating"
	"github.com/uselagoon/machinery/api/schema"
)

type IdentifyServices struct {
	Deployments []string `json:"deployments,omitempty"`
	Volumes     []string `json:"volumes,omitempty"`
	Services    []string `json:"services,omitempty"`
}

type LagoonServices struct {
	Services []schema.EnvironmentService `json:"services,omitempty"`
	Volumes  []schema.EnvironmentVolume  `json:"volumes,omitempty"`
}

// LagoonServiceTemplateIdentification takes the output of the generator and returns a JSON payload that contains information
// about the services that lagoon will be deploying (this will be kubernetes `kind: deployment`, but lagoon calls them services ¯\_(ツ)_/¯)
// this command can be used to identify services that are deployed by the build, so that services that may remain in the environment can be identified
// and eventually removed
func LagoonServiceTemplateIdentification(g generator.GeneratorInput) (*IdentifyServices, *LagoonServices, error) {
	servicesData := IdentifyServices{}
	lagoonBuild, err := generator.NewGenerator(
		g,
	)
	if err != nil {
		return nil, nil, err
	}

	lagoonServices := &LagoonServices{}
	pvcs, err := servicestemplates.GeneratePVCTemplate(*lagoonBuild.BuildValues)
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't identify volumes: %v", err)
	}
	for _, pvc := range pvcs {
		servicesData.Volumes = append(servicesData.Volumes, pvc.Name)
		size := pvc.Spec.Resources.Requests.Storage
		storeType := "block"
		if pvc.Spec.StorageClassName != nil {
			switch *pvc.Spec.StorageClassName {
			case "bulk":
				storeType = "bulk"
			}
		}
		kubevol := schema.EnvironmentVolume{
			Name:        pvc.Name,
			StorageType: storeType,
			Type:        pvc.Labels["lagoon.sh/service-type"],
			Size:        size().String(),
		}
		lagoonServices.Volumes = append(lagoonServices.Volumes, kubevol)
	}
	deployments, err := servicestemplates.GenerateDeploymentTemplate(*lagoonBuild.BuildValues)
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't identify deployments: %v", err)
	}
	for _, d := range deployments {
		servicesData.Deployments = append(servicesData.Deployments, d.Name)
		containers := []schema.ServiceContainer{}
		for _, c := range d.Spec.Template.Spec.Containers {
			volumes := []schema.VolumeMount{}
			for _, v := range c.VolumeMounts {
				for _, vo := range lagoonServices.Volumes {
					if vo.Name == v.Name {
						volumes = append(volumes, schema.VolumeMount{
							Name: v.Name,
							Path: v.MountPath,
						})
					}
				}
			}
			ports := []schema.ContainerPort{}
			for _, p := range c.Ports {
				ports = append(ports, schema.ContainerPort{
					Name: p.Name,
					Port: int(p.ContainerPort),
				})
			}
			containers = append(containers, schema.ServiceContainer{
				Name:    c.Name,
				Volumes: volumes,
				Ports:   ports,
			})
		}
		service := schema.EnvironmentService{
			Name:       d.Name,
			Type:       d.Labels["lagoon.sh/service-type"],
			Containers: containers,
		}
		lagoonServices.Services = append(lagoonServices.Services, service)
	}
	services, err := servicestemplates.GenerateServiceTemplate(*lagoonBuild.BuildValues)
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't identify services: %v", err)
	}
	for _, service := range services {
		servicesData.Services = append(servicesData.Services, service.Name)
	}
	return &servicesData, lagoonServices, nil
}
