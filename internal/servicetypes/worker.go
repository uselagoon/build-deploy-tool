package servicetypes

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var defaultWorkerPort int32 = 3000

var worker = ServiceType{
	Name: "worker",
	PrimaryContainer: ServiceContainer{
		Name: "worker",
		Container: corev1.Container{
			ImagePullPolicy: corev1.PullAlways,
			ReadinessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					Exec: &corev1.ExecAction{
						Command: []string{
							"/bin/sh",
							"-c",
							"if [ -x /bin/entrypoint-readiness ]; then /bin/entrypoint-readiness; fi",
						},
					},
				},
				InitialDelaySeconds: 5,
				PeriodSeconds:       2,
				FailureThreshold:    3,
			},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "lagoon-sshkey",
					ReadOnly:  true,
					MountPath: "/var/run/secrets/lagoon/sshkey/",
				},
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("10m"),
					corev1.ResourceMemory: resource.MustParse("100M"),
				},
			},
		},
	},
}

var workerPersistent = ServiceType{
	Name: "worker-persistent",
	PrimaryContainer: ServiceContainer{
		Name:      worker.PrimaryContainer.Name,
		Container: worker.PrimaryContainer.Container,
	},
}
