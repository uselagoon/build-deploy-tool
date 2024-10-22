package servicetypes

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var defaultValkeyPort int32 = 6379

var valkey = ServiceType{
	Name: "valkey",
	Ports: ServicePorts{
		Ports: []corev1.ServicePort{
			{
				Port: defaultValkeyPort,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: defaultValkeyPort,
				},
				Protocol: corev1.ProtocolTCP,
				Name:     fmt.Sprintf("%d-tcp", defaultValkeyPort),
			},
		},
	},
	PrimaryContainer: ServiceContainer{
		Name: "valkey",
		Container: corev1.Container{
			ImagePullPolicy: corev1.PullAlways,
			SecurityContext: &corev1.SecurityContext{},
			Ports: []corev1.ContainerPort{
				{
					Name:          fmt.Sprintf("%d-tcp", defaultValkeyPort),
					ContainerPort: defaultValkeyPort,
					Protocol:      corev1.ProtocolTCP,
				},
			},
			ReadinessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					TCPSocket: &corev1.TCPSocketAction{
						Port: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: defaultValkeyPort,
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
							IntVal: defaultValkeyPort,
						},
					},
				},
				InitialDelaySeconds: 120,
				TimeoutSeconds:      1,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("10m"),
					corev1.ResourceMemory: resource.MustParse("10Mi"),
				},
			},
		},
	},
}

var valkeyPersistent = ServiceType{
	Name:                     "valkey-persistent",
	Ports:                    valkey.Ports,
	ProvidesPersistentVolume: true,
	PrimaryContainer: ServiceContainer{
		Name:      valkey.PrimaryContainer.Name,
		Container: valkey.PrimaryContainer.Container,
		EnvVars: []corev1.EnvVar{
			{
				Name:  "VALKEY_FLAVOR",
				Value: "persistent",
			},
		},
	},
	Volumes: ServiceVolume{
		PersistentVolumeSize: "5Gi",
		PersistentVolumeType: corev1.ReadWriteOnce,
		PersistentVolumePath: "/data",
		BackupConfiguration: BackupConfiguration{
			Command:       `/bin/sh -c "timeout 5400 tar -cf - -C {{ if .ServiceValues.PersistentVolumePath }}{{.ServiceValues.PersistentVolumePath}}{{else}}{{.ServiceTypeValues.Volumes.PersistentVolumePath}}{{end}} ."`,
			FileExtension: ".{{ .ServiceValues.OverrideName }}.tar",
		},
	},
	Strategy: appsv1.DeploymentStrategy{
		Type: appsv1.RecreateDeploymentStrategyType,
	},
}
