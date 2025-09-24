package templating

import (
	"fmt"
	"strings"

	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/servicetypes"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// deployments and cronjobs share mostly the same pod template spec
// in cases where they differ, check if the cronjobCommand is provided to determine if the pod template spec is for a cronjob or not
func generatePodTemplateSpec(
	buildValues generator.BuildValues,
	serviceValues generator.ServiceValues,
	serviceTypeValues *servicetypes.ServiceType,
	objectMeta metav1.ObjectMeta,
	templateAnnotations map[string]string,
	name, cronjobCommand string,
) (*corev1.PodTemplateSpec, error) {
	var podTemplateSpec corev1.PodTemplateSpec
	additionalLabels := make(map[string]string)

	tpld := struct {
		ServiceValues     interface{}
		ServiceTypeValues interface{}
	}{
		serviceValues,
		serviceTypeValues,
	}

	podTemplateSpec.Spec.AutomountServiceAccountToken = &buildValues.AutoMountServiceAccountToken
	if cronjobCommand != "" {
		if serviceValues.CronjobUseSpotInstances {
			// handle spot instance label and affinity/tolerations/selectors
			additionalLabels["lagoon.sh/spot"] = "true"
			podTemplateSpec.Spec.Affinity = &corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
						{
							Weight: 1,
							Preference: corev1.NodeSelectorTerm{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{
										Key:      "lagoon.sh/spot",
										Operator: corev1.NodeSelectorOpExists,
									},
								},
							},
						},
					},
				},
			}
			podTemplateSpec.Spec.Tolerations = []corev1.Toleration{
				{
					Key:      "lagoon.sh/spot",
					Operator: corev1.TolerationOpExists,
					Effect:   corev1.TaintEffectNoSchedule,
				},
				{
					Key:      "lagoon.sh/spot",
					Operator: corev1.TolerationOpExists,
					Effect:   corev1.TaintEffectPreferNoSchedule,
				},
			}
			if serviceValues.CronjobForceSpotInstances {
				podTemplateSpec.Spec.NodeSelector = map[string]string{
					"lagoon.sh/spot": "true",
				}
			}
		}
	} else {
		if serviceValues.UseSpotInstances {
			// handle spot instance label and affinity/tolerations/selectors
			additionalLabels["lagoon.sh/spot"] = "true"
			podTemplateSpec.Spec.Affinity = &corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
						{
							Weight: 1,
							Preference: corev1.NodeSelectorTerm{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{
										Key:      "lagoon.sh/spot",
										Operator: corev1.NodeSelectorOpExists,
									},
								},
							},
						},
					},
				},
			}
			podTemplateSpec.Spec.Tolerations = []corev1.Toleration{
				{
					Key:      "lagoon.sh/spot",
					Operator: corev1.TolerationOpExists,
					Effect:   corev1.TaintEffectNoSchedule,
				},
				{
					Key:      "lagoon.sh/spot",
					Operator: corev1.TolerationOpExists,
					Effect:   corev1.TaintEffectPreferNoSchedule,
				},
			}
			if serviceValues.ForceSpotInstances {
				podTemplateSpec.Spec.NodeSelector = map[string]string{
					"lagoon.sh/spot": "true",
				}
			}
		}
	}
	// start cronjob template
	podTemplateSpec.ObjectMeta = metav1.ObjectMeta{
		Labels:      map[string]string{},
		Annotations: map[string]string{},
	}
	for key, value := range objectMeta.Labels {
		podTemplateSpec.ObjectMeta.Labels[key] = value
	}
	// add any additional annotations
	for key, value := range objectMeta.Annotations {
		podTemplateSpec.ObjectMeta.Annotations[key] = value
	}
	for key, value := range templateAnnotations {
		podTemplateSpec.ObjectMeta.Annotations[key] = value
	}

	// disable service links, this prevents some environment variables that confuse lagoon services being
	// added to the containers
	podTemplateSpec.Spec.EnableServiceLinks = helpers.BoolPtr(false)
	if serviceTypeValues.EnableServiceLinks {
		podTemplateSpec.Spec.EnableServiceLinks = helpers.BoolPtr(true)
	}
	// set the priority class
	podTemplateSpec.Spec.PriorityClassName = fmt.Sprintf("lagoon-priority-%s", buildValues.EnvironmentType)

	// handle the podescurity from rootless workloads
	if buildValues.PodSecurityContext.RunAsUser != 0 {
		podTemplateSpec.Spec.SecurityContext = &corev1.PodSecurityContext{
			RunAsUser:  helpers.Int64Ptr(buildValues.PodSecurityContext.RunAsUser),
			RunAsGroup: helpers.Int64Ptr(buildValues.PodSecurityContext.RunAsGroup),
			FSGroup:    helpers.Int64Ptr(buildValues.PodSecurityContext.FsGroup),
		}
	}
	// some services have a fsgroup override
	if serviceTypeValues.PodSecurityContext.HasDefault {
		podTemplateSpec.Spec.SecurityContext = &corev1.PodSecurityContext{
			FSGroup: helpers.Int64Ptr(serviceTypeValues.PodSecurityContext.FSGroup),
		}
	}
	if buildValues.PodSecurityContext.OnRootMismatch {
		fsGroupChangePolicy := corev1.FSGroupChangeOnRootMismatch
		if podTemplateSpec.Spec.SecurityContext != nil {
			podTemplateSpec.Spec.SecurityContext.FSGroupChangePolicy = &fsGroupChangePolicy
		} else {
			podTemplateSpec.Spec.SecurityContext = &corev1.PodSecurityContext{
				FSGroupChangePolicy: &fsGroupChangePolicy,
			}
		}
	}
	// start set up any volumes this cronjob can use
	// first handle any dynamic secret volumes that come from kubernetes secrets that are labeled
	for _, dsv := range buildValues.DynamicSecretVolumes {
		volume := corev1.Volume{
			Name: dsv.Name,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: dsv.Secret.SecretName,
					Optional:   &dsv.Secret.Optional,
				},
			},
		}
		podTemplateSpec.Spec.Volumes = append(podTemplateSpec.Spec.Volumes, volume)
	}

	// add any additional volumes with volumes
	for _, av := range serviceValues.AdditionalVolumes {
		volume := corev1.Volume{
			Name: av.Name,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: av.Name,
				},
			},
		}
		podTemplateSpec.Spec.Volumes = append(podTemplateSpec.Spec.Volumes, volume)
	}

	// if there is a persistent volume attached to this service, handle adding that here
	if serviceTypeValues.Volumes.PersistentVolumeSize != "" {
		volume := corev1.Volume{
			Name: serviceValues.PersistentVolumeName,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: serviceValues.PersistentVolumeName,
				},
			},
		}
		// if this servicetype has defaults, use them if one is not provided
		if serviceValues.PersistentVolumePath == "" {
			volume.Name = serviceValues.OverrideName
			volume.VolumeSource.PersistentVolumeClaim.ClaimName = serviceValues.OverrideName
		}
		podTemplateSpec.Spec.Volumes = append(podTemplateSpec.Spec.Volumes, volume)
	}

	// if there are any specific container volume overrides provided, handle those here
	for _, pcv := range serviceTypeValues.PrimaryContainer.Volumes {
		volume := corev1.Volume{}
		helpers.TemplateThings(tpld, pcv, &volume)
		podTemplateSpec.Spec.Volumes = append(podTemplateSpec.Spec.Volumes, volume)
	}
	for _, scv := range serviceTypeValues.SecondaryContainer.Volumes {
		volume := corev1.Volume{}
		helpers.TemplateThings(tpld, scv, &volume)
		podTemplateSpec.Spec.Volumes = append(podTemplateSpec.Spec.Volumes, volume)
	}
	// end set up any volumes this cronjob can use

	// handle any image pull secrets, add the default one first
	pullsecrets := []corev1.LocalObjectReference{
		{
			Name: generator.DefaultImagePullSecret,
		},
	}
	// then consume any from the custom provided container registries
	for _, pullsecret := range buildValues.ContainerRegistry {
		pullsecrets = append(pullsecrets, corev1.LocalObjectReference{
			Name: pullsecret.SecretName,
		})
	}
	podTemplateSpec.Spec.ImagePullSecrets = pullsecrets

	// start working out the containers to add
	// add any init container that the service may have
	if serviceTypeValues.InitContainer.Name != "" {
		enableInit := false
		init := serviceTypeValues.InitContainer
		// check if the init container has any flags required to add it
		for k, v := range buildValues.FeatureFlags {
			if init.FeatureFlags[k] == v {
				enableInit = true
			}
		}
		// otherwise if there are no flags
		if enableInit || init.FeatureFlags == nil {
			for _, svm := range serviceTypeValues.InitContainer.VolumeMounts {
				volumeMount := corev1.VolumeMount{}
				helpers.TemplateThings(tpld, svm, &volumeMount)
				init.Container.VolumeMounts = append(init.Container.VolumeMounts, volumeMount)
			}
			cmd := []string{}
			for _, c := range init.Command {
				var c2 string
				helpers.TemplateThings(tpld, c, &c2)
				cmd = append(cmd, c2)
			}
			init.Container.Command = cmd
			// init containers will more than likely contain public images, we should add a provided pull through imagecache if one is defined
			if buildValues.ImageCache != "" {
				init.Container.Image = fmt.Sprintf("%s%s", buildValues.ImageCache, init.Container.Image)
			}
			podTemplateSpec.Spec.InitContainers = append(podTemplateSpec.Spec.InitContainers, init.Container)
		}
	}

	// handle the primary container for the service type
	container := serviceTypeValues.PrimaryContainer

	// handle setting the rest of the containers specs with values from the service or build values
	container.Container.Name = name
	if val, ok := buildValues.ImageReferences[serviceValues.Name]; ok {
		container.Container.Image = val
	} else {
		return nil, fmt.Errorf("no image reference was found for primary container of service %s", serviceValues.Name)
	}

	// set up cronjobs if required
	cronjobs := ""
	for _, cronjob := range serviceValues.InPodCronjobs {
		cronjobs = fmt.Sprintf("%s%s %s\n", cronjobs, cronjob.Schedule, cronjob.Command)
	}
	container.Container.Env = append(container.Container.Env, container.EnvVars...)
	envvars := []corev1.EnvVar{}
	envvars = append(envvars, corev1.EnvVar{
		Name:  "LAGOON_GIT_SHA",
		Value: buildValues.GitSHA,
	})
	if cronjobCommand == "" {
		envvars = append(envvars, corev1.EnvVar{
			Name:  "CRONJOBS",
			Value: cronjobs,
		})
	}
	envvars = append(envvars, corev1.EnvVar{
		Name:  "SERVICE_NAME",
		Value: serviceValues.OverrideName,
	})
	container.Container.Env = append(container.Container.Env, envvars...)
	container.Container.EnvFrom = []corev1.EnvFromSource{
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "lagoon-platform-env",
				},
			},
		}, {
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "lagoon-env",
				},
			},
		},
	}
	for _, dds := range buildValues.DynamicDBaaSSecrets {
		container.Container.EnvFrom = append(container.Container.EnvFrom, corev1.EnvFromSource{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: dds,
				},
			},
		})
	}

	// mount the volumes in the primary container
	for _, dsm := range buildValues.DynamicSecretMounts {
		volumeMount := corev1.VolumeMount{
			Name:      dsm.Name,
			MountPath: dsm.MountPath,
			ReadOnly:  dsm.ReadOnly,
		}
		container.Container.VolumeMounts = append(container.Container.VolumeMounts, volumeMount)
	}
	// add any additional volumes with volumemounts
	for _, avm := range serviceValues.AdditionalVolumes {
		volumeMount := corev1.VolumeMount{
			Name:      avm.Name,
			MountPath: avm.Path,
		}
		container.Container.VolumeMounts = append(container.Container.VolumeMounts, volumeMount)
	}
	if serviceTypeValues.Volumes.PersistentVolumeSize != "" {
		volumeMount := corev1.VolumeMount{
			Name:      serviceValues.PersistentVolumeName,
			MountPath: serviceValues.PersistentVolumePath,
		}
		// if this servicetype has a default volume path, use them it if one isn't provided
		if serviceValues.PersistentVolumePath == "" {
			volumeMount.MountPath = serviceTypeValues.Volumes.PersistentVolumePath
			volumeMount.Name = serviceValues.OverrideName
		}
		container.Container.VolumeMounts = append(container.Container.VolumeMounts, volumeMount)
	}
	for _, svm := range serviceTypeValues.PrimaryContainer.VolumeMounts {
		volumeMount := corev1.VolumeMount{}
		helpers.TemplateThings(tpld, svm, &volumeMount)
		container.Container.VolumeMounts = append(container.Container.VolumeMounts, volumeMount)
	}
	if serviceValues.PersistentVolumeName != "" && serviceValues.PersistentVolumePath != "" && serviceTypeValues.Volumes.PersistentVolumeSize == "" {
		container.Container.VolumeMounts = append(container.Container.VolumeMounts, corev1.VolumeMount{
			Name:      serviceValues.PersistentVolumeName,
			MountPath: serviceValues.PersistentVolumePath,
		})
		volume := corev1.Volume{
			Name: serviceValues.PersistentVolumeName,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: serviceValues.PersistentVolumeName,
				},
			},
		}
		podTemplateSpec.Spec.Volumes = append(podTemplateSpec.Spec.Volumes, volume)
	}

	if buildValues.Resources.Limits.Memory != "" {
		if container.Container.Resources.Limits == nil {
			container.Container.Resources.Limits = corev1.ResourceList{}
		}
		container.Container.Resources.Limits[corev1.ResourceMemory] = resource.MustParse(buildValues.Resources.Limits.Memory)
	}
	if buildValues.Resources.Limits.EphemeralStorage != "" {
		if container.Container.Resources.Limits == nil {
			container.Container.Resources.Limits = corev1.ResourceList{}
		}
		container.Container.Resources.Limits[corev1.ResourceEphemeralStorage] = resource.MustParse(buildValues.Resources.Limits.EphemeralStorage)
	}
	if buildValues.Resources.Requests.EphemeralStorage != "" {
		if container.Container.Resources.Requests == nil {
			container.Container.Resources.Requests = corev1.ResourceList{}
		}
		container.Container.Resources.Requests[corev1.ResourceEphemeralStorage] = resource.MustParse(buildValues.Resources.Requests.EphemeralStorage)
	}

	if cronjobCommand != "" {
		podTemplateSpec.Spec.DNSConfig = &corev1.PodDNSConfig{
			Options: []corev1.PodDNSConfigOption{
				{
					Name:  "timeout",
					Value: helpers.StrPtr("60"),
				},
				{
					Name:  "attempts",
					Value: helpers.StrPtr("10"),
				},
			},
		}
		podTemplateSpec.Spec.RestartPolicy = corev1.RestartPolicyNever
		container.Container.Command = []string{"/lagoon/cronjob.sh", cronjobCommand}
		// strip ports from the cronjobs
		container.Container.Ports = nil
		container.Container.ReadinessProbe = nil
		container.Container.LivenessProbe = nil
	} else {
		if serviceValues.AdditionalServicePorts != nil {
			// nullify the existing ports
			container.Container.Ports = []corev1.ContainerPort{}
			// start compose service port override templating here
			for idx, addPort := range serviceValues.AdditionalServicePorts {
				port := corev1.ContainerPort{
					Name:          fmt.Sprintf("tcp-%d", addPort.ServicePort.Target),
					ContainerPort: int32(addPort.ServicePort.Target),
					Protocol:      corev1.Protocol(strings.ToUpper(addPort.ServicePort.Protocol)),
				}
				if idx == 0 {
					// first port in the docker compose file should be a tcp based port (default unless udp or other is provided in the docker-compose file)
					// the first port in the list will be used for any liveness/readiness probes, and will override the default option the container has
					switch addPort.ServicePort.Protocol {
					case "tcp":
						container.Container.ReadinessProbe = &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								TCPSocket: &corev1.TCPSocketAction{
									Port: intstr.IntOrString{
										Type:   intstr.Int,
										IntVal: int32(addPort.ServicePort.Target),
									},
								},
							},
							InitialDelaySeconds: 1, // @TODO make the timeout/delays configurable?
							TimeoutSeconds:      1,
						}
						container.Container.LivenessProbe = &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								TCPSocket: &corev1.TCPSocketAction{
									Port: intstr.IntOrString{
										Type:   intstr.Int,
										IntVal: int32(addPort.ServicePort.Target),
									},
								},
							},
							InitialDelaySeconds: 60, // @TODO make the timeout/delays configurable?
							TimeoutSeconds:      10,
						}
					default:
						return nil, fmt.Errorf("first port defined is not a tcp port, please ensure the first port is tcp")
					}
				}
				switch addPort.ServicePort.Protocol {
				case "udp":
					port.Name = fmt.Sprintf("udp-%d", addPort.ServicePort.Target)
				}
				// set the ports into the container
				container.Container.Ports = append(container.Container.Ports, port)
			}
		} else {
			// otherwise if the service has a default port, and it can be changed, handle changing it here
			if serviceTypeValues.Ports.CanChangePort {
				// check if the port override is defined
				if serviceValues.ServicePort != 0 {
					// and change the port in the container definition to suit
					container.Container.ReadinessProbe.ProbeHandler.TCPSocket.Port.IntVal = serviceValues.ServicePort
					container.Container.LivenessProbe.ProbeHandler.TCPSocket.Port.IntVal = serviceValues.ServicePort
					container.Container.Ports[0].ContainerPort = serviceValues.ServicePort
				}
			}
		}
	}

	// append the final defined container to the spec
	podTemplateSpec.Spec.Containers = append(podTemplateSpec.Spec.Containers, container.Container)

	// if this service has a secondary container provided (mainly will be `nginx-php`, but could be others in the future)
	if serviceValues.LinkedService == nil && serviceTypeValues.SecondaryContainer.Name != "" {
		// if no linked service is found from the docker-compose services, drop an error
		return nil, fmt.Errorf("service type %s has a secondary container defined, but no linked service was found", serviceValues.Type)
	}
	// if a linked service is provided, and the servicetype supports it, handle that here
	if serviceValues.LinkedService != nil && serviceTypeValues.SecondaryContainer.Name != "" {
		linkedContainer := serviceTypeValues.SecondaryContainer

		// handle setting the rest of the containers specs with values from the service or build values
		linkedContainer.Container.Name = linkedContainer.Name
		if val, ok := buildValues.ImageReferences[serviceValues.LinkedService.Name]; ok {
			linkedContainer.Container.Image = val
		} else {
			return nil, fmt.Errorf("no image reference was found for secondary container %s of service %s", serviceValues.LinkedService.Name, serviceValues.Name)
		}

		linkedContainer.Container.Env = append(linkedContainer.Container.Env, linkedContainer.EnvVars...)
		envvars := []corev1.EnvVar{}
		envvars = append(envvars, corev1.EnvVar{
			Name:  "LAGOON_GIT_SHA",
			Value: buildValues.GitSHA,
		})
		// if cronjobCommand == "" {
		// 	envvars = append(envvars, corev1.EnvVar{
		// 		Name:  "CRONJOBS",
		// 		Value: cronjobs,
		// 	})
		// }
		envvars = append(envvars, corev1.EnvVar{
			Name:  "SERVICE_NAME",
			Value: serviceValues.OverrideName,
		})
		linkedContainer.Container.Env = append(linkedContainer.Container.Env, envvars...)
		linkedContainer.Container.EnvFrom = []corev1.EnvFromSource{
			{
				SecretRef: &corev1.SecretEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "lagoon-platform-env",
					},
				},
			}, {
				SecretRef: &corev1.SecretEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "lagoon-env",
					},
				},
			},
		}
		for _, dds := range buildValues.DynamicDBaaSSecrets {
			linkedContainer.Container.EnvFrom = append(linkedContainer.Container.EnvFrom, corev1.EnvFromSource{
				SecretRef: &corev1.SecretEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: dds,
					},
				},
			})
		}
		for _, dsm := range buildValues.DynamicSecretMounts {
			volumeMount := corev1.VolumeMount{
				Name:      dsm.Name,
				MountPath: dsm.MountPath,
				ReadOnly:  dsm.ReadOnly,
			}
			linkedContainer.Container.VolumeMounts = append(linkedContainer.Container.VolumeMounts, volumeMount)
		}

		// add any additional volumes with volumemounts
		for _, avm := range serviceValues.AdditionalVolumes {
			volumeMount := corev1.VolumeMount{
				Name:      avm.Name,
				MountPath: avm.Path,
			}
			linkedContainer.Container.VolumeMounts = append(linkedContainer.Container.VolumeMounts, volumeMount)
		}
		if serviceTypeValues.Volumes.PersistentVolumeSize != "" {
			volumeMount := corev1.VolumeMount{
				Name:      serviceValues.PersistentVolumeName,
				MountPath: serviceValues.PersistentVolumePath,
			}
			linkedContainer.Container.VolumeMounts = append(linkedContainer.Container.VolumeMounts, volumeMount)
		}

		for _, svm := range serviceTypeValues.SecondaryContainer.VolumeMounts {
			volumeMount := corev1.VolumeMount{}
			helpers.TemplateThings(tpld, svm, &volumeMount)
			linkedContainer.Container.VolumeMounts = append(linkedContainer.Container.VolumeMounts, volumeMount)
		}

		// set the resource limit overrides if they are provided
		if buildValues.Resources.Limits.Memory != "" {
			if linkedContainer.Container.Resources.Limits == nil {
				linkedContainer.Container.Resources.Limits = corev1.ResourceList{}
			}
			linkedContainer.Container.Resources.Limits[corev1.ResourceMemory] = resource.MustParse(buildValues.Resources.Limits.Memory)
		}
		if buildValues.Resources.Limits.EphemeralStorage != "" {
			if linkedContainer.Container.Resources.Limits == nil {
				linkedContainer.Container.Resources.Limits = corev1.ResourceList{}
			}
			linkedContainer.Container.Resources.Limits[corev1.ResourceEphemeralStorage] = resource.MustParse(buildValues.Resources.Limits.EphemeralStorage)
		}
		if buildValues.Resources.Requests.EphemeralStorage != "" {
			if linkedContainer.Container.Resources.Requests == nil {
				linkedContainer.Container.Resources.Requests = corev1.ResourceList{}
			}
			linkedContainer.Container.Resources.Requests[corev1.ResourceEphemeralStorage] = resource.MustParse(buildValues.Resources.Requests.EphemeralStorage)
		}

		podTemplateSpec.Spec.Containers = append(podTemplateSpec.Spec.Containers, linkedContainer.Container)
	}
	return &podTemplateSpec, nil
}
