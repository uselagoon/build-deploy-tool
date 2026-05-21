package servicetypes

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var worker = ServiceType{
	Name:                   "worker",
	AllowAdditionalVolumes: true,
	AllowSSHKeyMount:       true,
	PrimaryContainer: ServiceContainer{
		Name: "worker",
		Container: corev1.Container{
			ImagePullPolicy: corev1.PullAlways,
			SecurityContext: &corev1.SecurityContext{},
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
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("10m"),
					corev1.ResourceMemory: resource.MustParse("10Mi"),
				},
			},
		},
	},
}

var workerPersistent = ServiceType{
	Name:                     "worker-persistent",
	ConsumesPersistentVolume: true,
	AllowAdditionalVolumes:   true,
	AllowSSHKeyMount:         true,
	PrimaryContainer: ServiceContainer{
		Name:      worker.PrimaryContainer.Name,
		Container: worker.PrimaryContainer.Container,
	},
}
