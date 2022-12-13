package servicetypes

import (
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var nginx = ServiceType{
	Name: "nginx",
	Ports: ServicePorts{
		Ports: []corev1.ServicePort{
			{
				Port: 8080,
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
		Name:            "nginx",
		ImagePullPolicy: corev1.PullAlways,
		Container: corev1.Container{
			Ports: []corev1.ContainerPort{
				{
					Name:          "http",
					ContainerPort: 8080,
					Protocol:      corev1.ProtocolTCP,
				},
			},
			ReadinessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/nginx_status",
						Port: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: 50000,
						},
					},
				},
				InitialDelaySeconds: 1,
				TimeoutSeconds:      3,
			},
			LivenessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/nginx_status",
						Port: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: 50000,
						},
					},
				},
				InitialDelaySeconds: 900,
				TimeoutSeconds:      3,
				FailureThreshold:    5,
			},
		},
	},
}

var nginxPHP = ServiceType{
	Name:  "nginx-php",
	Ports: nginx.Ports,
	PrimaryContainer: ServiceContainer{
		Name:            "nginx",
		ImagePullPolicy: corev1.PullAlways,
		Container: corev1.Container{
			Ports: []corev1.ContainerPort{
				{
					Name:          "http",
					ContainerPort: 8080,
					Protocol:      corev1.ProtocolTCP,
				},
			},
			ReadinessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/nginx_status",
						Port: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: 50000,
						},
					},
				},
				InitialDelaySeconds: 1,
				TimeoutSeconds:      3,
			},
			LivenessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/nginx_status",
						Port: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: 50000,
						},
					},
				},
				InitialDelaySeconds: 900,
				TimeoutSeconds:      3,
				FailureThreshold:    5,
			},
			Env: []corev1.EnvVar{
				{
					Name:  "NGINX_FASTCGI_PASS",
					Value: "127.0.0.1",
				},
			},
		},
	},
	SecondaryContainer: ServiceContainer{
		Name:            "php",
		ImagePullPolicy: corev1.PullAlways,
		Container: corev1.Container{
			Ports: []corev1.ContainerPort{
				{
					Name:          "http",
					ContainerPort: 9000,
					Protocol:      corev1.ProtocolTCP,
				},
			},
			ReadinessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					TCPSocket: &corev1.TCPSocketAction{
						Port: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: 9000,
						},
					},
				},
				InitialDelaySeconds: 2,
				PeriodSeconds:       10,
			},
			LivenessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					TCPSocket: &corev1.TCPSocketAction{
						Port: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: 9000,
						},
					},
				},
				InitialDelaySeconds: 60,
				PeriodSeconds:       10,
			},
			Env: []corev1.EnvVar{
				{
					Name:  "NGINX_FASTCGI_PASS",
					Value: "127.0.0.1",
				},
			},
		},
	},
}

var nginxPHPPersistent = ServiceType{
	Name:  "nginx-php-persistent",
	Ports: nginx.Ports,
	Volumes: ServiceVolume{
		PersistentVolumeSize: "5Gi",
		PersistentVolumeType: corev1.ReadWriteMany,
		Backup:               true,
	},
	InitContainer: ServiceContainer{
		Name: "fix-storage-permissions",
		Command: []string{
			"sh",
			"-c",
			`set -e
SENTINEL="/storage/.lagoon-rootless-migration-complete"
if ! [ -f "$SENTINEL" ]; then
	find /storage -exec chown {{ .PodSecurityContext.RunAsUser}}:0 {} +
	find /storage -exec chmod a+r,u+w {} +
	find /storage -type d -exec chmod a+x {} +
	touch "$SENTINEL"
fi`,
		},
		Container: corev1.Container{
			Name:            "fix-storage-permissions",
			Image:           "busybox:musl",
			ImagePullPolicy: corev1.PullIfNotPresent,
			SecurityContext: &corev1.SecurityContext{
				RunAsUser: helpers.Int64Ptr(0),
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "{{ .PersistentVolumeName }}",
				MountPath: "/storage",
			},
		},
	},
	PrimaryContainer: ServiceContainer{
		Name:            nginxPHP.PrimaryContainer.Name,
		ImagePullPolicy: nginxPHP.PrimaryContainer.ImagePullPolicy,
		Container:       nginxPHP.PrimaryContainer.Container,
	},
	SecondaryContainer: ServiceContainer{
		Name:            nginxPHP.SecondaryContainer.Name,
		ImagePullPolicy: nginxPHP.SecondaryContainer.ImagePullPolicy,
		Container:       nginxPHP.SecondaryContainer.Container,
		Volumes: []corev1.Volume{
			{
				Name: "{{ .PersistentVolumeName }}-twig",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "{{ .PersistentVolumeName }}-twig",
				MountPath: "{{ .PersistentVolumeName }}/php",
			},
		},
	},
}
