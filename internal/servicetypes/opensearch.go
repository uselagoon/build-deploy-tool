package servicetypes

import (
	"fmt"

	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var defaultOpensearchPort int32 = 9200

var opensearch = ServiceType{
	Name: "opensearch-persistent",
	Ports: ServicePorts{
		Ports: []corev1.ServicePort{
			{
				Port: defaultOpensearchPort,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: defaultOpensearchPort,
				},
				Protocol: corev1.ProtocolTCP,
				Name:     fmt.Sprintf("%d-tcp", defaultOpensearchPort),
			},
		},
	},
	ProvidesPersistentVolume: true,
	PrimaryContainer: ServiceContainer{
		Name: "opensearch",
		Container: corev1.Container{
			ImagePullPolicy: corev1.PullAlways,
			SecurityContext: &corev1.SecurityContext{},
			Ports: []corev1.ContainerPort{
				{
					Name:          fmt.Sprintf("%d-tcp", defaultOpensearchPort),
					ContainerPort: defaultOpensearchPort,
					Protocol:      corev1.ProtocolTCP,
				},
			},
			ReadinessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/_cluster/health?local=true",
						Port: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: defaultOpensearchPort,
						},
					},
				},
				InitialDelaySeconds: 20,
			},
			LivenessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/_cluster/health?local=true",
						Port: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: defaultOpensearchPort,
						},
					},
				},
				InitialDelaySeconds: 120,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("10m"),
					corev1.ResourceMemory: resource.MustParse("10Mi"),
				},
			},
		},
	},
	InitContainer: ServiceContainer{
		Name: "set-max-map-count",
		Command: []string{
			"sh",
			"-c",
			`set -xe
DESIRED="262144"
CURRENT=$(sysctl -n vm.max_map_count)
if [ "$DESIRED" -gt "$CURRENT" ]; then
  sysctl -w vm.max_map_count=$DESIRED
fi`,
		},
		Container: corev1.Container{
			Name:            "set-max-map-count",
			Image:           "library/busybox:latest",
			ImagePullPolicy: corev1.PullIfNotPresent,
			SecurityContext: &corev1.SecurityContext{
				Privileged: helpers.BoolPtr(true),
				RunAsUser:  helpers.Int64Ptr(0),
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
		PersistentVolumePath: "/usr/share/opensearch/data",
		BackupConfiguration: BackupConfiguration{
			Command:       `/bin/sh -c "tar -cf - -C {{ if .ServiceValues.PersistentVolumePath }}{{.ServiceValues.PersistentVolumePath}}{{else}}{{.ServiceTypeValues.Volumes.PersistentVolumePath}}{{end}} ."`,
			FileExtension: ".{{ .ServiceValues.OverrideName }}.tar",
		},
	},
}
