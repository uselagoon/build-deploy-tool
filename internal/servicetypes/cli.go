package servicetypes

import (
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var cli = ServiceType{
	Name: "cli",
	PrimaryContainer: ServiceContainer{
		Name:            "cli",
		ImagePullPolicy: corev1.PullAlways,
		Volumes: []corev1.Volume{
			{
				Name: "lagoon-sshkey",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						DefaultMode: helpers.Int32Ptr(420),
						SecretName:  "lagoon-sshkey",
					},
				},
			},
		},
		Container: corev1.Container{
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "lagoon-sshkey",
					ReadOnly:  true,
					MountPath: "/var/run/secrets/lagoon/sshkey/",
				},
			},
			Ports: []corev1.ContainerPort{
				{
					Name:          "http",
					ContainerPort: 3000,
					Protocol:      corev1.ProtocolTCP,
				},
			},
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
					corev1.ResourceMemory: resource.MustParse("100M"),
				},
			},
		},
	},
}

var cliPersistent = ServiceType{
	Name: "cli-persistent",
	PrimaryContainer: ServiceContainer{
		Name:            cli.PrimaryContainer.Name,
		ImagePullPolicy: cli.PrimaryContainer.ImagePullPolicy,
		Container:       cli.PrimaryContainer.Container,
		Volumes: append(cli.PrimaryContainer.Volumes, corev1.Volume{
			Name: "{{ .PersistentVolumeName }}-twig",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		}),
		VolumeMounts: append(cli.PrimaryContainer.VolumeMounts, corev1.VolumeMount{
			Name:      "{{ .PersistentVolumeName }}-twig",
			MountPath: "{{ .PersistentVolumePath }}/php",
		}),
	},
}
