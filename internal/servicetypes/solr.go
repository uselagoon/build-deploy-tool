package servicetypes

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var defaultSolrPort int32 = 8983

var solr = ServiceType{
	Name: "solr-php-persistent", // this has to be like this because it is used in selectors, and is unchangeable now on existing deployed solr
	Ports: ServicePorts{
		Ports: []corev1.ServicePort{
			{
				Port: defaultSolrPort,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: defaultSolrPort,
				},
				Protocol: corev1.ProtocolTCP,
				Name:     fmt.Sprintf("%d-tcp", defaultSolrPort),
			},
		},
	},
	ProvidesPersistentVolume: true,
	PrimaryContainer: ServiceContainer{
		Name: "solr",
		Container: corev1.Container{
			ImagePullPolicy: corev1.PullAlways,
			SecurityContext: &corev1.SecurityContext{},
			Ports: []corev1.ContainerPort{
				{
					Name:          fmt.Sprintf("%d-tcp", defaultSolrPort),
					ContainerPort: defaultSolrPort,
					Protocol:      corev1.ProtocolTCP,
				},
			},
			ReadinessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					TCPSocket: &corev1.TCPSocketAction{
						Port: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: defaultSolrPort,
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
							IntVal: defaultSolrPort,
						},
					},
				},
				InitialDelaySeconds: 90,
				TimeoutSeconds:      3,
				FailureThreshold:    5,
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
		PersistentVolumePath: "/var/solr",
		BackupConfiguration: BackupConfiguration{
			Command:       `/bin/sh -c 'tar -cf - -C "{{ if .ServiceValues.PersistentVolumePath }}{{.ServiceValues.PersistentVolumePath}}{{else}}{{.ServiceTypeValues.Volumes.PersistentVolumePath}}{{end}}" --exclude="lost\+found" . || [ $? -eq 1 ]'`,
			FileExtension: ".{{ .ServiceValues.OverrideName }}.tar",
		},
	},
}
