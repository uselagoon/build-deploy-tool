package servicetypes

import (
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var opensearch = ServiceType{
	Name: "opensearch",
	Ports: ServicePorts{
		Ports: []corev1.ServicePort{
			{
				Port: 9200,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 9200,
				},
				Protocol: corev1.ProtocolTCP,
				Name:     "9200-tcp",
			},
		},
	},
	PrimaryContainer: ServiceContainer{
		Name:            "opensearch",
		ImagePullPolicy: corev1.PullAlways,
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
			Image:           "busybox:latest",
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
	Volumes: ServiceVolume{
		PersistentVolumeSize: "5Gi",
		PersistentVolumeType: corev1.ReadWriteOnce,
		PersistentVolumePath: "/usr/share/opensearch/data",
		BackupConfiguration: BackupConfiguration{
			Command:       `/bin/sh -c "tar -cf - -C /usr/share/opensearch/data ."`,
			FileExtension: ".{{ .OverrideName }}.tar",
		},
	},
}
