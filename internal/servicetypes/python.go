package servicetypes

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var python = ServiceType{
	Name: "python",
	Ports: ServicePorts{
		CanChangePort: true,
		Ports: []corev1.ServicePort{
			{
				Port: 8800,
				TargetPort: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "http",
				},
				Protocol: corev1.ProtocolTCP,
				Name:     "http",
			},
		},
	},
}

var pythonPersistent = ServiceType{
	Name:  "python-persistent",
	Ports: python.Ports,
	Volumes: ServiceVolume{
		PersistentVolumeSize: "5Gi",
		PersistentVolumeType: corev1.ReadWriteMany,
		Backup:               true,
	},
}
