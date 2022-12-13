package servicetypes

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var mongodbSingle = ServiceType{
	Name: "mongodb-single",
	Ports: ServicePorts{
		Ports: []corev1.ServicePort{
			{
				Port: 27017,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 27017,
				},
				Protocol: corev1.ProtocolTCP,
				Name:     "27017-tcp",
			},
		},
	},
	Volumes: ServiceVolume{
		PersistentVolumeSize: "5Gi",
		PersistentVolumeType: corev1.ReadWriteOnce,
		PersistentVolumePath: "/data/db",
	},
}
