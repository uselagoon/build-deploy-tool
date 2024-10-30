package servicetypes

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var defaultBasicPort int32 = 3000

// defines all the basic service type defaults
var basic = ServiceType{
	Name: "basic",
	Ports: ServicePorts{
		CanChangePort: true, // this service has the ability to change its default port
		Ports: []corev1.ServicePort{
			{
				Port: defaultBasicPort, // this is the default port for basic service type
				TargetPort: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "http",
				},
				Protocol: corev1.ProtocolTCP,
				Name:     "http",
			},
		},
	},
	AllowAdditionalVolumes: true,
	PrimaryContainer: ServiceContainer{
		Name: "basic",
		Container: corev1.Container{
			ImagePullPolicy: corev1.PullAlways,
			SecurityContext: &corev1.SecurityContext{},
			Ports: []corev1.ContainerPort{
				{
					Name:          "http",
					ContainerPort: defaultBasicPort,
					Protocol:      corev1.ProtocolTCP,
				},
			},
			ReadinessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					TCPSocket: &corev1.TCPSocketAction{
						Port: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: defaultBasicPort,
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
							IntVal: defaultBasicPort,
						},
					},
				},
				InitialDelaySeconds: 60,
				TimeoutSeconds:      10,
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

// contains all the persistent type overrides that the basic service doesn't have
var basicPersistent = ServiceType{
	Name:                     "basic-persistent",
	Ports:                    basic.Ports,
	ProvidesPersistentVolume: true,
	AllowAdditionalVolumes:   true,
	PrimaryContainer: ServiceContainer{
		Name:      basic.PrimaryContainer.Name,
		Container: basic.PrimaryContainer.Container,
	},
	Volumes: ServiceVolume{
		PersistentVolumeSize: "5Gi",
		PersistentVolumeType: corev1.ReadWriteMany,
		Backup:               true,
	},
}

// contains all the single type overrides that the basic service doesn't have
// basicSingle is like basic persistent except that the volume is not bulk, and the pod can only ever have 1 replica because of this
var basicSingle = ServiceType{
	Name:                     "basic-single",
	Ports:                    basic.Ports,
	ProvidesPersistentVolume: true,
	AllowAdditionalVolumes:   false,
	PrimaryContainer: ServiceContainer{
		Name:      basic.PrimaryContainer.Name,
		Container: basic.PrimaryContainer.Container,
	},
	Strategy: appsv1.DeploymentStrategy{
		Type: appsv1.RecreateDeploymentStrategyType,
	},
	Volumes: ServiceVolume{
		PersistentVolumeSize: "5Gi",
		PersistentVolumeType: corev1.ReadWriteOnce,
		Backup:               true,
	},
}
