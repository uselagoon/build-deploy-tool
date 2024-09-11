package servicetypes

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var defaultMongoDBPort int32 = 27017

var mongodbSingle = ServiceType{
	Name:               "mongodb-single",
	EnableServiceLinks: true,
	Ports: ServicePorts{
		Ports: []corev1.ServicePort{
			{
				Port: defaultMongoDBPort,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: defaultMongoDBPort,
				},
				Protocol: corev1.ProtocolTCP,
				Name:     fmt.Sprintf("%d-tcp", defaultMongoDBPort),
			},
		},
	},
	ProvidesPersistentVolume: true,
	PrimaryContainer: ServiceContainer{
		Name: "mongodb-single",
		Container: corev1.Container{
			ImagePullPolicy: corev1.PullAlways,
			SecurityContext: &corev1.SecurityContext{},
			Ports: []corev1.ContainerPort{
				{
					Name:          fmt.Sprintf("%d-tcp", defaultMongoDBPort),
					ContainerPort: defaultMongoDBPort,
					Protocol:      corev1.ProtocolTCP,
				},
			},
			ReadinessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					TCPSocket: &corev1.TCPSocketAction{
						Port: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: defaultMongoDBPort,
						},
					},
				},
				InitialDelaySeconds: 1,
				TimeoutSeconds:      1,
			},
			LivenessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					TCPSocket: &corev1.TCPSocketAction{
						Port: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: defaultMongoDBPort,
						},
					},
				},
				InitialDelaySeconds: 120,
				PeriodSeconds:       5,
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
		PersistentVolumePath: "/data/db",
		BackupConfiguration: BackupConfiguration{
			Command:       `/bin/sh -c 'tar -cf - -C "/data/db" --exclude="lost\+found" . || [ $? -eq 1 ]'`,
			FileExtension: ".{{ .ServiceValues.OverrideName }}.tar",
		},
	},
}
