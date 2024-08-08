package servicetypes

import (
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// defines all the cli service type defaults
var cli = ServiceType{
	Name:                   "cli",
	AllowAdditionalVolumes: true,
	PrimaryContainer: ServiceContainer{
		Name: "cli",
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
			ImagePullPolicy: corev1.PullAlways,
			SecurityContext: &corev1.SecurityContext{},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "lagoon-sshkey",
					ReadOnly:  true,
					MountPath: "/var/run/secrets/lagoon/sshkey/",
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
					corev1.ResourceMemory: resource.MustParse("10Mi"),
				},
			},
		},
	},
}

// contains all the persistent type overrides that the cli service doesn't have
var cliPersistent = ServiceType{
	Name:                     "cli-persistent",
	ConsumesPersistentVolume: true,
	AllowAdditionalVolumes:   true,
	PrimaryContainer: ServiceContainer{
		Name:      cli.PrimaryContainer.Name,
		Container: cli.PrimaryContainer.Container,
		Volumes: append(cli.PrimaryContainer.Volumes, corev1.Volume{
			Name: "{{ .ServiceValues.PersistentVolumeName }}-twig",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		}),
		VolumeMounts: append(cli.PrimaryContainer.VolumeMounts, corev1.VolumeMount{
			Name:      "{{ .ServiceValues.PersistentVolumeName }}-twig",
			MountPath: "{{ .ServiceValues.PersistentVolumePath }}/php",
		}),
	},
}
