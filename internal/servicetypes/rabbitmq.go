package servicetypes

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var defaultRabbitMQPort int32 = 5672
var defaultRabbitMQWebPort int32 = 15672

var rabbitmq = ServiceType{
	Name: "rabbitmq",
	Ports: ServicePorts{
		Ports: []corev1.ServicePort{
			{
				Port: defaultRabbitMQPort,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: defaultRabbitMQPort,
				},
				Protocol: corev1.ProtocolTCP,
				Name:     fmt.Sprintf("%d-tcp", defaultRabbitMQPort),
			},
			{
				Port: defaultRabbitMQWebPort,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: defaultRabbitMQWebPort,
				},
				Protocol: corev1.ProtocolTCP,
				Name:     fmt.Sprintf("%d-tcp", defaultRabbitMQWebPort),
			},
		},
	},
	ProvidesPersistentVolume: true,
	PrimaryContainer: ServiceContainer{
		Name: "rabbitmq",
		Container: corev1.Container{
			ImagePullPolicy: corev1.PullAlways,
			SecurityContext: &corev1.SecurityContext{},
			Ports: []corev1.ContainerPort{
				{
					Name:          fmt.Sprintf("%d-tcp", defaultRabbitMQPort),
					ContainerPort: defaultRabbitMQPort,
					Protocol:      corev1.ProtocolTCP,
				},
			},
			ReadinessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					TCPSocket: &corev1.TCPSocketAction{
						Port: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: defaultRabbitMQPort,
						},
					},
				},
				InitialDelaySeconds: 1,
				PeriodSeconds:       3,
			},
			LivenessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					TCPSocket: &corev1.TCPSocketAction{
						Port: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: defaultRabbitMQPort,
						},
					},
				},
				InitialDelaySeconds: 90,
				TimeoutSeconds:      3,
				FailureThreshold:    5,
			},
			Env: []corev1.EnvVar{
				{
					Name:  "RABBITMQ_NODENAME",
					Value: "rabbitmq@localhost",
				},
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("10m"),
					corev1.ResourceMemory: resource.MustParse("10Mi"),
				},
			},
		},
	},
	PodSecurityContext: ServicePodSecurityContext{
		HasDefault: true,
		FSGroup:    0,
	},
	Strategy: appsv1.DeploymentStrategy{
		Type: appsv1.RecreateDeploymentStrategyType,
	},
	Volumes: ServiceVolume{
		PersistentVolumeSize: "5Gi",
		PersistentVolumeType: corev1.ReadWriteOnce,
		PersistentVolumePath: "/var/lib/rabbitmq",
		BackupConfiguration: BackupConfiguration{
			Command:       `/bin/sh -c 'tar -cf - -C "{{ if .ServiceValues.PersistentVolumePath }}{{.ServiceValues.PersistentVolumePath}}{{else}}{{.ServiceTypeValues.Volumes.PersistentVolumePath}}{{end}}" --exclude="lost\+found" . || [ $? -eq 1 ]'`,
			FileExtension: ".{{ .ServiceValues.OverrideName }}.tar",
		},
	},
}
