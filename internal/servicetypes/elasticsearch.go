package servicetypes

import (
	"fmt"

	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var defaultElasticsearchPort int32 = 9200

var elasticsearch = ServiceType{
	Name: "elasticsearch",
	Ports: ServicePorts{
		Ports: []corev1.ServicePort{
			{
				Port: defaultElasticsearchPort,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: defaultElasticsearchPort,
				},
				Protocol: corev1.ProtocolTCP,
				Name:     fmt.Sprintf("%d-tcp", defaultElasticsearchPort),
			},
		},
	},
	PrimaryContainer: ServiceContainer{
		Name:            "elasticsearch",
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
			Image:           "library/busybox:latest",
			ImagePullPolicy: corev1.PullIfNotPresent,
			SecurityContext: &corev1.SecurityContext{
				Privileged: helpers.BoolPtr(true),
				RunAsUser:  helpers.Int64Ptr(0),
			},
		},
	},
	Volumes: ServiceVolume{
		PersistentVolumeSize: "5Gi",
		PersistentVolumeType: corev1.ReadWriteOnce,
		PersistentVolumePath: "/usr/share/elasticsearch/data",
		BackupConfiguration: BackupConfiguration{
			Command:       `/bin/sh -c "tar -cf - -C {{ if .ServiceValues.PersistentVolumePath }}{{.ServiceValues.PersistentVolumePath}}{{else}}{{.ServiceTypeValues.Volumes.PersistentVolumePath}}{{end}} ."`,
			FileExtension: ".{{ .ServiceValues.OverrideName }}.tar",
		},
	},
}
