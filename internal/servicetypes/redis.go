package servicetypes

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var redis = ServiceType{
	Name: "redis",
	Ports: ServicePorts{
		Ports: []corev1.ServicePort{
			{
				Port: 6379,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 6379,
				},
				Protocol: corev1.ProtocolTCP,
				Name:     "6379-tcp",
			},
		},
	},
	PrimaryContainer: ServiceContainer{
		Name:            "redis",
		ImagePullPolicy: corev1.PullAlways,
		Container: corev1.Container{
			Ports: []corev1.ContainerPort{
				{
					Name:          "6379-tcp",
					ContainerPort: 6379,
					Protocol:      corev1.ProtocolTCP,
				},
			},
			ReadinessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					TCPSocket: &corev1.TCPSocketAction{
						Port: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: 6379,
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
							IntVal: 6379,
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
