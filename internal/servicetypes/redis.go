package servicetypes

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var defaultRedisPort int32 = 6379

var redis = ServiceType{
	Name: "redis",
	Ports: ServicePorts{
		Ports: []corev1.ServicePort{
			{
				Port: defaultRedisPort,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: defaultRedisPort,
				},
				Protocol: corev1.ProtocolTCP,
				Name:     fmt.Sprintf("%d-tcp", defaultRedisPort),
			},
		},
	},
	PrimaryContainer: ServiceContainer{
		Name: "redis",
		Container: corev1.Container{
			ImagePullPolicy: corev1.PullAlways,
			SecurityContext: &corev1.SecurityContext{},
			Ports: []corev1.ContainerPort{
				{
					Name:          fmt.Sprintf("%d-tcp", defaultRedisPort),
					ContainerPort: defaultRedisPort,
					Protocol:      corev1.ProtocolTCP,
				},
			},
			ReadinessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					TCPSocket: &corev1.TCPSocketAction{
						Port: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: defaultRedisPort,
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
							IntVal: defaultRedisPort,
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

var redisPersistent = ServiceType{
	Name:                     "redis-persistent",
	Ports:                    redis.Ports,
	ProvidesPersistentVolume: true,
	PrimaryContainer: ServiceContainer{
		Name:      redis.PrimaryContainer.Name,
		Container: redis.PrimaryContainer.Container,
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
