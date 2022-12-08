package servicetypes

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var mariadbSingle = ServiceType{
	Name: "mariadb-single",
	Ports: ServicePorts{
		CanChangePort: true,
		Ports: []corev1.ServicePort{
			{
				Port: 3306,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 3306,
				},
				Protocol: corev1.ProtocolTCP,
				Name:     "3306-tcp",
			},
		},
	},
	Volumes: ServiceVolume{
		PersistentVolumeSize: "5Gi",
		PersistentVolumeType: corev1.ReadWriteOnce,
		PersistentVolumePath: "/var/lib/mysql",
	},
}
