package servicetypes

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var defaultMongoDBPort int32 = 27017

var mongodbSingle = ServiceType{
	Name: "mongodb-single",
	Ports: ServicePorts{
		Ports: []corev1.ServicePort{
			{
				Port: defaultMongoDBPort,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: defaultMongoDBPort,
				},
				Protocol: corev1.ProtocolTCP,
				Name:     fmt.Sprintf("%d-tcp", defaultMongoDBPort),
			},
		},
	},
	Volumes: ServiceVolume{
		PersistentVolumeSize: "5Gi",
		PersistentVolumeType: corev1.ReadWriteOnce,
		PersistentVolumePath: "/data/db",
		BackupConfiguration: BackupConfiguration{
			Command:       `/bin/sh -c 'tar -cf - -C "/data/db" --exclude="lost\+found" . || [ $? -eq 1 ]'`,
			FileExtension: ".{{ .OverrideName }}.tar",
		},
	},
}
