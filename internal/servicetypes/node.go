package servicetypes

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var defaultNodePort int32 = 3000

var node = ServiceType{
	Name: "node",
	Ports: ServicePorts{
		Ports: []corev1.ServicePort{
			{
				Port: defaultNodePort,
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
		Name: "node",
		Container: corev1.Container{
			ImagePullPolicy: corev1.PullAlways,
			SecurityContext: &corev1.SecurityContext{},
			Ports: []corev1.ContainerPort{
				{
					Name:          "http",
					ContainerPort: defaultNodePort,
					Protocol:      corev1.ProtocolTCP,
				},
			},
			ReadinessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					TCPSocket: &corev1.TCPSocketAction{
						Port: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: defaultNodePort,
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
							IntVal: defaultNodePort,
						},
					},
				},
				InitialDelaySeconds: 60,
				TimeoutSeconds:      10,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("10m"),
					corev1.ResourceMemory: resource.MustParse("100Mi"),
				},
			},
		},
	},
}

var nodePersistent = ServiceType{
	Name:                     "node-persistent",
	Ports:                    node.Ports,
	ProvidesPersistentVolume: true,
	AllowAdditionalVolumes:   true,
	PrimaryContainer: ServiceContainer{
		Name:      node.PrimaryContainer.Name,
		Container: node.PrimaryContainer.Container,
	},
	Volumes: ServiceVolume{
		PersistentVolumeSize: "5Gi",
		PersistentVolumeType: corev1.ReadWriteMany,
		Backup:               true,
	},
}
