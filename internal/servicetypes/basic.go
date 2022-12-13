package servicetypes

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var basic = ServiceType{
	Name: "basic",
	Ports: ServicePorts{
		CanChangePort: true,
		Ports: []corev1.ServicePort{
			{
				Port: 3000,
				TargetPort: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "http",
				},
				Protocol: corev1.ProtocolTCP,
				Name:     "http",
			},
		},
	},
	PrimaryContainer: ServiceContainer{
		Name:            "basic",
		ImagePullPolicy: corev1.PullAlways,
		Container: corev1.Container{
			Ports: []corev1.ContainerPort{
				{
					Name:          "http",
					ContainerPort: 3000,
					Protocol:      corev1.ProtocolTCP,
				},
			},
			ReadinessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					TCPSocket: &corev1.TCPSocketAction{
						Port: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: 3000,
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
							IntVal: 3000,
						},
					},
				},
				InitialDelaySeconds: 60,
				TimeoutSeconds:      10,
			},
			EnvFrom: []corev1.EnvFromSource{
				{
					ConfigMapRef: &corev1.ConfigMapEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "lagoon-env",
						},
					},
				},
			},
		},
	},
}

var basicPersistent = ServiceType{
	Name:  "basic-persistent",
	Ports: basic.Ports,
	Volumes: ServiceVolume{
		PersistentVolumeSize: "5Gi",
		PersistentVolumeType: corev1.ReadWriteMany,
		Backup:               true,
	},
}
