package identify

import (
	"context"
	"strings"

	mariadbv1 "github.com/amazeeio/dbaas-operator/apis/mariadb/v1"
	mongodbv1 "github.com/amazeeio/dbaas-operator/apis/mongodb/v1"
	postgresv1 "github.com/amazeeio/dbaas-operator/apis/postgres/v1"
	"github.com/uselagoon/build-deploy-tool/internal/collector"
	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/machinery/api/schema"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func GetCurrentState(c *collector.Collector, gen generator.GeneratorInput) (
	LagoonServices,
	[]mariadbv1.MariaDBConsumer,
	[]mongodbv1.MongoDBConsumer,
	[]postgresv1.PostgreSQLConsumer,
	[]appsv1.Deployment,
	[]corev1.PersistentVolumeClaim,
	[]corev1.Service,
	*collector.LagoonEnvState,
	error,
) {
	lagoonServices := LagoonServices{
		Services: []schema.EnvironmentService{},
		Volumes:  []schema.EnvironmentVolume{},
	}
	out, currentServices, err := LagoonServiceTemplateIdentification(gen)
	if err != nil {
		return lagoonServices, nil, nil, nil, nil, nil, nil, nil, err
	}

	dbaas, err := IdentifyDBaaSConsumers(gen)
	if err != nil {
		return lagoonServices, nil, nil, nil, nil, nil, nil, nil, err
	}

	state, err := c.Collect(context.Background(), gen.Namespace)
	if err != nil {
		return lagoonServices, nil, nil, nil, nil, nil, nil, nil, err
	}

	// add any dbaas that should exist to the current services
	for _, prov := range dbaas {
		sp := strings.Split(prov, ":")
		currentServices.Services = append(currentServices.Services, schema.EnvironmentService{
			Name: sp[0],
			Type: sp[1],
		})
	}

	mariadbMatch := false
	var mariadbDelete []mariadbv1.MariaDBConsumer
	for _, exist := range state.MariaDBConsumers.Items {
		service := schema.EnvironmentService{
			Name: exist.Name,
			Type: "mariadb-dbaas",
		}
		for _, prov := range dbaas {
			sp := strings.Split(prov, ":")
			if strings.Contains(sp[1], "mariadb-dbaas") {
				if exist.Name == sp[0] {
					mariadbMatch = true
					continue
				}
			}
		}
		if !mariadbMatch {
			service.Abandoned = true
			mariadbDelete = append(mariadbDelete, exist)
		}
		mariadbMatch = false
		lagoonServices.Services = append(lagoonServices.Services, service)
	}

	mongodbMatch := false
	var mongodbDelete []mongodbv1.MongoDBConsumer
	for _, exist := range state.MongoDBConsumers.Items {
		service := schema.EnvironmentService{
			Name: exist.Name,
			Type: "mongodb-dbaas",
		}
		for _, prov := range dbaas {
			sp := strings.Split(prov, ":")
			if strings.Contains(sp[1], "mongodb-dbaas") {
				if exist.Name == sp[0] {
					mongodbMatch = true
					continue
				}
			}
		}
		if !mongodbMatch {
			service.Abandoned = true
			mongodbDelete = append(mongodbDelete, exist)
		}
		mongodbMatch = false
		lagoonServices.Services = append(lagoonServices.Services, service)
	}

	postgresqlMatch := false
	var postgresqlDelete []postgresv1.PostgreSQLConsumer
	for _, exist := range state.PostgreSQLConsumers.Items {
		service := schema.EnvironmentService{
			Name: exist.Name,
			Type: "postgres-dbaas",
		}
		for _, prov := range dbaas {
			sp := strings.Split(prov, ":")
			if strings.Contains(sp[1], "postgres-dbaas") {
				if exist.Name == sp[0] {
					postgresqlMatch = true
					continue
				}
			}
		}
		if !postgresqlMatch {
			service.Abandoned = true
			postgresqlDelete = append(postgresqlDelete, exist)
		}
		postgresqlMatch = false
		lagoonServices.Services = append(lagoonServices.Services, service)
	}

	volMatch := false
	var volDelete []corev1.PersistentVolumeClaim
	for _, exist := range state.PVCs.Items {
		size := exist.Spec.Resources.Requests.Storage
		storeType := "block"
		if exist.Spec.StorageClassName != nil {
			switch *exist.Spec.StorageClassName {
			case "bulk":
				storeType = "bulk"
			}
		}
		kubevol := schema.EnvironmentVolume{
			Name:        exist.Name,
			StorageType: storeType,
			Type:        exist.Labels["lagoon.sh/service-type"],
			Size:        size().String(),
		}
		for _, prov := range out.Volumes {
			if exist.Name == prov {
				volMatch = true
				continue
			}
		}
		if !volMatch {
			kubevol.Abandoned = true
			volDelete = append(volDelete, exist)
		}
		volMatch = false
		lagoonServices.Volumes = append(lagoonServices.Volumes, kubevol)
	}

	servMatch := false
	var servDelete []corev1.Service
	for _, exist := range state.Services.Items {
		for _, prov := range out.Services {
			if exist.Name == prov {
				servMatch = true
				continue
			}
		}
		if !servMatch {
			servDelete = append(servDelete, exist)
		}
		servMatch = false
	}

	depMatch := false
	var depDelete []appsv1.Deployment
	for _, exist := range state.Deployments.Items {
		containers := []schema.ServiceContainer{}
		for _, c := range exist.Spec.Template.Spec.Containers {
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
			Name:       exist.Name,
			Type:       exist.Labels["lagoon.sh/service-type"],
			Containers: containers,
		}
		for _, prov := range out.Deployments {
			if exist.Name == prov {
				depMatch = true
				continue
			}
		}
		if !depMatch {
			service.Abandoned = true
			depDelete = append(depDelete, exist)
		}
		depMatch = false
		lagoonServices.Services = append(lagoonServices.Services, service)
	}

	for _, svc := range currentServices.Services {
		if !serviceExists(lagoonServices.Services, svc.Name) {
			lagoonServices.Services = append(lagoonServices.Services, svc)
		}
	}

	for _, vol := range currentServices.Volumes {
		if !volumeExists(lagoonServices.Volumes, vol.Name) {
			lagoonServices.Volumes = append(lagoonServices.Volumes, vol)
		}
	}

	return lagoonServices, mariadbDelete, mongodbDelete, postgresqlDelete, depDelete, volDelete, servDelete, state, nil
}

func serviceExists(services []schema.EnvironmentService, serviceName string) bool {
	for _, svc := range services {
		if svc.Name == serviceName {
			return true
		}
	}
	return false
}

func volumeExists(volumes []schema.EnvironmentVolume, volumeName string) bool {
	for _, vol := range volumes {
		if vol.Name == volumeName {
			return true
		}
	}
	return false
}
