package servicetypes

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var defaultVarnishPort int32 = 8080
var defaultVarnishControlPort int32 = 6082

var varnish = ServiceType{
	Name: "varnish",
	Ports: ServicePorts{
		CanChangePort: true,
		Ports: []corev1.ServicePort{
			{
				Port: defaultVarnishPort,
				TargetPort: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "http",
				},
				Protocol: corev1.ProtocolTCP,
				Name:     "http",
			},
			{
				Port: defaultVarnishControlPort,
				TargetPort: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "controlport",
				},
				Protocol: corev1.ProtocolTCP,
				Name:     "controlport",
			},
		},
	},
	PrimaryContainer: ServiceContainer{
		Name: "varnish",
		Container: corev1.Container{
			ImagePullPolicy: corev1.PullAlways,
			SecurityContext: &corev1.SecurityContext{},
			Ports: []corev1.ContainerPort{
				{
					Name:          "http",
					ContainerPort: defaultVarnishPort,
					Protocol:      corev1.ProtocolTCP,
				},
				{
					Name:          "controlport",
					ContainerPort: defaultVarnishControlPort,
					Protocol:      corev1.ProtocolTCP,
				},
			},
			ReadinessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					TCPSocket: &corev1.TCPSocketAction{
						Port: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: defaultVarnishPort,
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
							IntVal: defaultVarnishPort,
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

var varnishPersistent = ServiceType{
	Name:                     "varnish-persistent",
	Ports:                    varnish.Ports,
	ProvidesPersistentVolume: true,
	PrimaryContainer: ServiceContainer{
		Name:      varnish.PrimaryContainer.Name,
		Container: varnish.PrimaryContainer.Container,
	},
	Volumes: ServiceVolume{
		PersistentVolumeSize: "5Gi",
		PersistentVolumeType: corev1.ReadWriteOnce,
		PersistentVolumePath: "/var/cache/varnish",
		BackupConfiguration: BackupConfiguration{
			Command:       `/bin/sh -c "/bin/busybox tar -cf - -C {{ if .ServiceValues.PersistentVolumePath }}{{.ServiceValues.PersistentVolumePath}}{{else}}{{.ServiceTypeValues.Volumes.PersistentVolumePath}}{{end}} ."`,
			FileExtension: ".{{ .ServiceValues.OverrideName }}.tar",
		},
	},
}
