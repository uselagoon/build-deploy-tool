package identify

import (
	"fmt"

	"github.com/uselagoon/build-deploy-tool/internal/generator"
	servicestemplates "github.com/uselagoon/build-deploy-tool/internal/templating"
)

type IdentifyServices struct {
	Deployments []string `json:"deployments,omitempty"`
	Volumes     []string `json:"volumes,omitempty"`
	Services    []string `json:"services,omitempty"`
}

// eventually replace with https://github.com/uselagoon/machinery/pull/99
type LagoonServices struct {
	Services []EnvironmentService `json:"services,omitempty"`
	Volumes  []EnvironmentVolume  `json:"volumes,omitempty"`
	// Services []schema.EnvironmentService `json:"services,omitempty"`
	// Volumes  []schema.EnvironmentVolume  `json:"volumes,omitempty"`
}

// eventually replace with https://github.com/uselagoon/machinery/pull/99
type EnvironmentService struct {
	ID         int                `json:"id,omitempty"`
	Name       string             `json:"name,omitempty"`
	Type       string             `json:"type,omitempty"`
	Updated    string             `json:"updated,omitempty"`
	Containers []ServiceContainer `json:"containers,omitempty"`
	Created    string             `json:"created,omitempty"`
	Abandoned  bool               `json:"abandoned,omitempty"` // no longer tracked in the docker-compose file
}

// eventually replace with https://github.com/uselagoon/machinery/pull/99
type ServiceContainer struct {
	Name    string          `json:"name,omitempty"`
	Volumes []VolumeMount   `json:"volumes,omitempty"`
	Ports   []ContainerPort `json:"ports,omitempty"`
}

// eventually replace with https://github.com/uselagoon/machinery/pull/99
type EnvironmentVolume struct {
	Name        string `json:"name,omitempty"`
	StorageType string `json:"storageType,omitempty"`
	Type        string `json:"type,omitempty"`
	Size        string `json:"size,omitempty"`
	Abandoned   bool   `json:"abandoned,omitempty"` // no longer tracked in the docker-compose file
}

// eventually replace with https://github.com/uselagoon/machinery/pull/99
type VolumeMount struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path,omitempty"`
}

// eventually replace with https://github.com/uselagoon/machinery/pull/99
type ContainerPort struct {
	Name string `json:"name,omitempty"`
	Port int    `json:"port,omitempty"`
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
		kubevol := EnvironmentVolume{
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
		containers := []ServiceContainer{}
		for _, c := range d.Spec.Template.Spec.Containers {
			volumes := []VolumeMount{}
			for _, v := range c.VolumeMounts {
				for _, vo := range lagoonServices.Volumes {
					if vo.Name == v.Name {
						volumes = append(volumes, VolumeMount{
							Name: v.Name,
							Path: v.MountPath,
						})
					}
				}
			}
			ports := []ContainerPort{}
			for _, p := range c.Ports {
				ports = append(ports, ContainerPort{
					Name: p.Name,
					Port: int(p.ContainerPort),
				})
			}
			containers = append(containers, ServiceContainer{
				Name:    c.Name,
				Volumes: volumes,
				Ports:   ports,
			})
		}
		service := EnvironmentService{
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
