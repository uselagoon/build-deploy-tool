package servicetypes

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var redis = ServiceType{
	Name: "redis",
	Ports: ServicePorts{
		CanChangePort: true,
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
}

var redisPersistent = ServiceType{
	Name:  "redis-persistent",
	Ports: redis.Ports,
	Volumes: ServiceVolume{
		PersistentVolumeSize: "5Gi",
		PersistentVolumeType: corev1.ReadWriteOnce,
		PersistentVolumePath: "/data",
	},
}
