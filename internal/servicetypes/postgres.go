package servicetypes

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var postgresSingle = ServiceType{
	Name: "postgres-single",
	Ports: ServicePorts{
		Ports: []corev1.ServicePort{
			{
				Port: 5432,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 5432,
				},
				Protocol: corev1.ProtocolTCP,
				Name:     "5432-tcp",
			},
		},
	},
	Strategy: appsv1.DeploymentStrategy{
		Type: appsv1.RecreateDeploymentStrategyType,
	},
	Volumes: ServiceVolume{
		PersistentVolumeSize: "5Gi",
		PersistentVolumeType: corev1.ReadWriteOnce,
		PersistentVolumePath: "/var/lib/postgresql/data",
	},
}
