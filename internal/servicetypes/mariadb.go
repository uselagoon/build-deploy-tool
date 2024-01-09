package servicetypes

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var defaultMariaDBPort int32 = 3306

var mariadbSingle = ServiceType{
	Name: "mariadb-single",
	Ports: ServicePorts{
		Ports: []corev1.ServicePort{
			{
				Port: defaultMariaDBPort,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: defaultMariaDBPort,
				},
				Protocol: corev1.ProtocolTCP,
				Name:     fmt.Sprintf("%d-tcp", defaultMariaDBPort),
			},
		},
	},
	Volumes: ServiceVolume{
		PersistentVolumeSize: "5Gi",
		PersistentVolumeType: corev1.ReadWriteOnce,
		PersistentVolumePath: "/var/lib/mysql",
		BackupConfiguration: BackupConfiguration{
			Command:       `/bin/sh -c 'mysqldump --max-allowed-packet=500M --events --routines --quick --add-locks --no-autocommit --single-transaction --all-databases'`,
			FileExtension: ".{{ .OverrideName }}.sql",
		},
	},
}
