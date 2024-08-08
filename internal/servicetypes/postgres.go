package servicetypes

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var defaultPostgresPort int32 = 5432

var postgresSingle = ServiceType{
	Name:               "postgres-single",
	EnableServiceLinks: true,
	Ports: ServicePorts{
		Ports: []corev1.ServicePort{
			{
				Port: defaultPostgresPort,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: defaultPostgresPort,
				},
				Protocol: corev1.ProtocolTCP,
				Name:     fmt.Sprintf("%d-tcp", defaultPostgresPort),
			},
		},
	},
	ProvidesPersistentVolume: true,
	PrimaryContainer: ServiceContainer{
		Name: "postgres-single",
		Container: corev1.Container{
			ImagePullPolicy: corev1.PullAlways,
			SecurityContext: &corev1.SecurityContext{},
			Ports: []corev1.ContainerPort{
				{
					Name:          fmt.Sprintf("%d-tcp", defaultPostgresPort),
					ContainerPort: defaultPostgresPort,
					Protocol:      corev1.ProtocolTCP,
				},
			},
			ReadinessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					TCPSocket: &corev1.TCPSocketAction{
						Port: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: defaultPostgresPort,
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
							IntVal: defaultPostgresPort,
						},
					},
				},
				InitialDelaySeconds: 120,
				PeriodSeconds:       5,
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
		PersistentVolumePath: "/var/lib/postgresql/data",
		BackupConfiguration: BackupConfiguration{
			Command:       `/bin/sh -c "PGPASSWORD=$POSTGRES_PASSWORD pg_dump --host=localhost --port=${{ .ServiceValues.OverrideName | FixServiceName }}_SERVICE_PORT --dbname=$POSTGRES_DB --username=$POSTGRES_USER --format=t -w"`,
			FileExtension: ".{{ .ServiceValues.OverrideName }}.tar",
		},
	},
}
