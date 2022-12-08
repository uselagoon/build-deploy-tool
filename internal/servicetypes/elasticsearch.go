package servicetypes

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var elasticsearch = ServiceType{
	Name: "elasticsearch",
	Ports: ServicePorts{
		CanChangePort: true,
		Ports: []corev1.ServicePort{
			{
				Port: 9200,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 9200,
				},
				Protocol: corev1.ProtocolTCP,
				Name:     "9200-tcp",
			},
		},
	},
	Volumes: ServiceVolume{
		PersistentVolumeSize: "5Gi",
		PersistentVolumeType: corev1.ReadWriteOnce,
		PersistentVolumePath: "/usr/share/elasticsearch/data",
	},
}
