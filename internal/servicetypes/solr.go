package servicetypes

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var solr = ServiceType{
	Name: "solr-php-persistent", // this has to be like this because it is used in selectors, and is unchangeable now on existing deployed solr
	Ports: ServicePorts{
		CanChangePort: true,
		Ports: []corev1.ServicePort{
			{
				Port: 8983,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 8983,
				},
				Protocol: corev1.ProtocolTCP,
				Name:     "tcp-8983",
			},
		},
	},
	Volumes: ServiceVolume{
		PersistentVolumeSize: "5Gi",
		PersistentVolumeType: corev1.ReadWriteOnce,
		PersistentVolumePath: "/var/solr",
	},
}
