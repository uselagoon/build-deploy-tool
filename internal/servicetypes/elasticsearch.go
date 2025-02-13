package servicetypes

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var defaultElasticsearchPort int32 = 9200

var elasticsearch = ServiceType{
	Name: "elasticsearch-persistent",
	Ports: ServicePorts{
		Ports: []corev1.ServicePort{
			{
				Port: defaultElasticsearchPort,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: defaultElasticsearchPort,
				},
				Protocol: corev1.ProtocolTCP,
				Name:     fmt.Sprintf("%d-tcp", defaultElasticsearchPort),
			},
		},
	},
	ProvidesPersistentVolume: true,
	PrimaryContainer: ServiceContainer{
		Name: "elasticsearch",
		Container: corev1.Container{
			ImagePullPolicy: corev1.PullAlways,
			SecurityContext: &corev1.SecurityContext{},
			Ports: []corev1.ContainerPort{
				{
					Name:          fmt.Sprintf("%d-tcp", defaultElasticsearchPort),
					ContainerPort: defaultElasticsearchPort,
					Protocol:      corev1.ProtocolTCP,
				},
			},
			ReadinessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/_cluster/health?local=true",
						Port: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: defaultElasticsearchPort,
						},
					},
				},
				InitialDelaySeconds: 20,
			},
			LivenessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/_cluster/health?local=true",
						Port: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: defaultElasticsearchPort,
						},
					},
				},
				InitialDelaySeconds: 120,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("10m"),
					corev1.ResourceMemory: resource.MustParse("10Mi"),
				},
			},
		},
	},
	PodSecurityContext: ServicePodSecurityContext{
		HasDefault: true,
		FSGroup:    0,
	},
	Strategy: appsv1.DeploymentStrategy{
		Type: appsv1.RecreateDeploymentStrategyType,
	},
	Volumes: ServiceVolume{
		PersistentVolumeSize: "5Gi",
		PersistentVolumeType: corev1.ReadWriteOnce,
		PersistentVolumePath: "/usr/share/elasticsearch/data",
		BackupConfiguration: BackupConfiguration{
			Command:       `/bin/sh -c "tar -cf - -C {{ if .ServiceValues.PersistentVolumePath }}{{.ServiceValues.PersistentVolumePath}}{{else}}{{.ServiceTypeValues.Volumes.PersistentVolumePath}}{{end}} ."`,
			FileExtension: ".{{ .ServiceValues.OverrideName }}.tar",
		},
	},
}
