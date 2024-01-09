package servicetypes

import (
	"fmt"

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
		Name:            "redis",
		ImagePullPolicy: corev1.PullAlways,
		Container: corev1.Container{
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
				TimeoutSeconds:      10,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("10m"),
					corev1.ResourceMemory: resource.MustParse("100M"),
				},
			},
		},
	},
}

var redisPersistent = ServiceType{
	Name:  "redis-persistent",
	Ports: redis.Ports,
	Volumes: ServiceVolume{
		PersistentVolumeSize: "5Gi",
		PersistentVolumeType: corev1.ReadWriteOnce,
		PersistentVolumePath: "/data",
		BackupConfiguration: BackupConfiguration{
			Command:       `/bin/sh -c "/bin/busybox tar -cf - -C /data ."`,
			FileExtension: ".{{ .OverrideName }}.tar",
		},
	},
}
