package servicetypes

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type ServiceType struct {
	Name               string
	Ports              ServicePorts
	Volumes            ServiceVolume
	Strategy           appsv1.DeploymentStrategy
	PrimaryContainer   ServiceContainer
	InitContainer      ServiceContainer
	SecondaryContainer ServiceContainer
	PodSecurityContext ServicePodSecurityContext
}

type ServicePodSecurityContext struct {
	HasDefault bool
	FSGroup    int64
}

type ServiceContainer struct {
	Name            string
	ImagePullPolicy corev1.PullPolicy
	Container       corev1.Container
	// define additional volumes here, can leverage 'go template' with generator.ServiceValues
	Volumes      []corev1.Volume
	VolumeMounts []corev1.VolumeMount
	Command      []string
}

type ServiceVolume struct {
	PersistentVolumeSize string
	PersistentVolumePath string
	PersistentVolumeType corev1.PersistentVolumeAccessMode
	Backup               bool
}

type ServicePorts struct {
	CanChangePort bool
	Ports         []corev1.ServicePort
}

// this is a map that maps the lagoon service-type that can be provided in the `lagoon.type` label to the default values for that service
var ServiceTypes = map[string]ServiceType{
	"basic":                basic,
	"basic-persistent":     basicPersistent,
	"cli":                  cli,
	"cli-persistent":       cliPersistent,
	"elasticsearch":        elasticsearch,
	"opensearch":           opensearch,
	"mariadb-single":       mariadbSingle,
	"mongodb-single":       mongodbSingle,
	"postgres-single":      postgresSingle,
	"node":                 node,
	"node-persistent":      nodePersistent,
	"python":               python,
	"python-persistent":    pythonPersistent,
	"nginx":                nginx,
	"nginx-php":            nginxPHP,
	"nginx-php-persistent": nginxPHPPersistent,
}
