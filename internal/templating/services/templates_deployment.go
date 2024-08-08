package services

import (
	"fmt"
	"sort"
	"strings"

	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/servicetypes"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metavalidation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// GenerateDeploymentTemplate generates the lagoon template to apply.
func GenerateDeploymentTemplate(
	buildValues generator.BuildValues,
) ([]appsv1.Deployment, error) {
	var deployments []appsv1.Deployment

	// check linked services
	checkedServices := LinkedServiceCalculator(buildValues.Services)

	// for all the services that the build values generated
	// iterate over them and generate any kubernetes deployments
	for _, serviceValues := range checkedServices {
		if val, ok := servicetypes.ServiceTypes[serviceValues.Type]; ok {
			serviceTypeValues := &servicetypes.ServiceType{}
			helpers.DeepCopy(val, serviceTypeValues)

			// add the default labels
			labels := map[string]string{
				"app.kubernetes.io/managed-by": "build-deploy-tool",
				"app.kubernetes.io/name":       serviceTypeValues.Name,
				"app.kubernetes.io/instance":   serviceValues.OverrideName,
				"lagoon.sh/project":            buildValues.Project,
				"lagoon.sh/environment":        buildValues.Environment,
				"lagoon.sh/environmentType":    buildValues.EnvironmentType,
				"lagoon.sh/buildType":          buildValues.BuildType,
				"lagoon.sh/template":           fmt.Sprintf("%s-%s", serviceTypeValues.Name, "0.1.0"),
				"lagoon.sh/service":            serviceValues.OverrideName,
				"lagoon.sh/service-type":       serviceTypeValues.Name,
			}

			// add the default annotations
			annotations := map[string]string{
				"lagoon.sh/version": buildValues.LagoonVersion,
			}

			// add any additional labels
			additionalLabels := make(map[string]string)
			additionalAnnotations := make(map[string]string)
			if buildValues.BuildType == "branch" {
				additionalAnnotations["lagoon.sh/branch"] = buildValues.Branch
			} else if buildValues.BuildType == "pullrequest" {
				additionalAnnotations["lagoon.sh/prNumber"] = buildValues.PRNumber
				additionalAnnotations["lagoon.sh/prHeadBranch"] = buildValues.PRHeadBranch
				additionalAnnotations["lagoon.sh/prBaseBranch"] = buildValues.PRBaseBranch
			}

			templateAnnotations := make(map[string]string)
			templateAnnotations["lagoon.sh/configMapSha"] = buildValues.ConfigMapSha
			tpld := struct {
				ServiceValues     interface{}
				ServiceTypeValues interface{}
			}{
				serviceValues,
				serviceTypeValues,
			}
			if serviceTypeValues.Volumes.BackupConfiguration.Command != "" {
				bc := servicetypes.BackupConfiguration{}
				helpers.TemplateThings(tpld, serviceTypeValues.Volumes.BackupConfiguration, &bc)
				switch buildValues.Backup.K8upVersion {
				case "v2":
					templateAnnotations["k8up.io/backupcommand"] = bc.Command
					templateAnnotations["k8up.io/file-extension"] = bc.FileExtension
				default:
					templateAnnotations["k8up.syn.tools/backupcommand"] = bc.Command
					templateAnnotations["k8up.syn.tools/file-extension"] = bc.FileExtension
				}
			}

			// create the initial deployment spec
			deployment := &appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Deployment",
					APIVersion: fmt.Sprintf("%s/%s", appsv1.SchemeGroupVersion.Group, appsv1.SchemeGroupVersion.Version),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        serviceValues.OverrideName,
					Labels:      labels,
					Annotations: annotations,
				},
			}
			deployment.ObjectMeta.Labels = labels
			deployment.ObjectMeta.Annotations = annotations

			if serviceValues.UseSpotInstances {
				// handle spot instance label and affinity/tolerations/selectors
				additionalLabels["lagoon.sh/spot"] = "true"
				deployment.Spec.Template.Spec.Affinity = &corev1.Affinity{
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
				deployment.Spec.Template.Spec.Tolerations = []corev1.Toleration{
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
					deployment.Spec.Template.Spec.NodeSelector = map[string]string{
						"lagoon.sh/spot": "true",
					}
				}
			}

			for key, value := range additionalLabels {
				deployment.ObjectMeta.Labels[key] = value
			}
			// add any additional annotations
			for key, value := range additionalAnnotations {
				deployment.ObjectMeta.Annotations[key] = value
			}
			// validate any annotations
			if err := apivalidation.ValidateAnnotations(deployment.ObjectMeta.Annotations, nil); err != nil {
				if len(err) != 0 {
					return nil, fmt.Errorf("the annotations for %s are not valid: %v", serviceValues.OverrideName, err)
				}
			}
			// validate any labels
			if err := metavalidation.ValidateLabels(deployment.ObjectMeta.Labels, nil); err != nil {
				if len(err) != 0 {
					return nil, fmt.Errorf("the labels for %s are not valid: %v", serviceValues.OverrideName, err)
				}
			}
			// check length of labels
			err := helpers.CheckLabelLength(deployment.ObjectMeta.Labels)
			if err != nil {
				return nil, err
			}

			// start deployment template
			deployment.Spec.Template.ObjectMeta = metav1.ObjectMeta{
				Labels:      map[string]string{},
				Annotations: map[string]string{},
			}
			for key, value := range deployment.ObjectMeta.Labels {
				deployment.Spec.Template.ObjectMeta.Labels[key] = value
			}
			// add any additional annotations
			for key, value := range deployment.ObjectMeta.Annotations {
				deployment.Spec.Template.ObjectMeta.Annotations[key] = value
			}
			for key, value := range templateAnnotations {
				deployment.Spec.Template.ObjectMeta.Annotations[key] = value
			}
			deployment.Spec.Replicas = helpers.Int32Ptr(1)
			if serviceValues.Replicas != 0 {
				deployment.Spec.Replicas = helpers.Int32Ptr(serviceValues.Replicas)
			}
			deployment.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name":     serviceTypeValues.Name,
					"app.kubernetes.io/instance": serviceValues.OverrideName,
				},
			}
			deployment.Spec.Strategy = serviceTypeValues.Strategy

			// disable service links, this prevents some environment variables that confuse lagoon services being
			// added to the containers
			deployment.Spec.Template.Spec.EnableServiceLinks = helpers.BoolPtr(false)
			if serviceTypeValues.EnableServiceLinks {
				deployment.Spec.Template.Spec.EnableServiceLinks = helpers.BoolPtr(true)
			}
			// set the priority class
			deployment.Spec.Template.Spec.PriorityClassName = fmt.Sprintf("lagoon-priority-%s", buildValues.EnvironmentType)

			// handle the podescurity from rootless workloads
			if buildValues.PodSecurityContext.RunAsUser != 0 {
				deployment.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{
					RunAsUser:  helpers.Int64Ptr(buildValues.PodSecurityContext.RunAsUser),
					RunAsGroup: helpers.Int64Ptr(buildValues.PodSecurityContext.RunAsGroup),
					FSGroup:    helpers.Int64Ptr(buildValues.PodSecurityContext.FsGroup),
				}
			}
			// some services have a fsgroup override
			if serviceTypeValues.PodSecurityContext.HasDefault {
				deployment.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{
					FSGroup: helpers.Int64Ptr(serviceTypeValues.PodSecurityContext.FSGroup),
				}
			}
			if buildValues.PodSecurityContext.OnRootMismatch {
				fsGroupChangePolicy := corev1.FSGroupChangeOnRootMismatch
				if deployment.Spec.Template.Spec.SecurityContext != nil {
					deployment.Spec.Template.Spec.SecurityContext.FSGroupChangePolicy = &fsGroupChangePolicy
				} else {
					deployment.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{
						FSGroupChangePolicy: &fsGroupChangePolicy,
					}
				}
			}

			// start set up any volumes this deployment can use
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
				deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, volume)
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
				deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, volume)
			}

			// if there is a persistent volume attached to this service, and no additional volumes, handle adding that here
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
				deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, volume)
			}

			// if there are any specific container volume overrides provided, handle those here
			for _, pcv := range serviceTypeValues.PrimaryContainer.Volumes {
				volume := corev1.Volume{}
				helpers.TemplateThings(tpld, pcv, &volume)
				deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, volume)
			}
			for _, scv := range serviceTypeValues.SecondaryContainer.Volumes {
				volume := corev1.Volume{}
				helpers.TemplateThings(tpld, scv, &volume)
				deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, volume)
			}

			// future handle any additional volume mounts here
			// @TODO

			// end set up any volumes this deployment can use

			// handle any image pull secrets, add the default one first
			pullsecrets := []corev1.LocalObjectReference{
				{
					Name: generator.DefaultImagePullSecret,
				},
			}
			// then consume any from the custom provided container registries
			sort.Slice(buildValues.ContainerRegistry, func(i, j int) bool {
				return buildValues.ContainerRegistry[i].Name < buildValues.ContainerRegistry[j].Name
			})
			for _, pullsecret := range buildValues.ContainerRegistry {
				pullsecrets = append(pullsecrets, corev1.LocalObjectReference{
					Name: pullsecret.SecretName,
				})
			}
			deployment.Spec.Template.Spec.ImagePullSecrets = pullsecrets

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
					// add any volume mounts to the init container as required
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
					deployment.Spec.Template.Spec.InitContainers = append(deployment.Spec.Template.Spec.InitContainers, init.Container)
				}
			}

			// handle the primary container for the service type
			container := serviceTypeValues.PrimaryContainer
			// if the service is set to consume the additional service ports from the docker-compose file (lagoon.sh/usecomposeports label on service)
			// then generate those additional service ports in the deploymeny here
			if serviceValues.AdditionalServicePorts != nil {
				// nulify the existing ports
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

			// handle setting the rest of the containers specs with values from the service or build values
			container.Container.Name = container.Name
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
			envvars := []corev1.EnvVar{
				{
					Name:  "LAGOON_GIT_SHA",
					Value: buildValues.GitSHA,
				},
				{
					Name:  "CRONJOBS",
					Value: cronjobs,
				},
				{
					Name:  "SERVICE_NAME",
					Value: serviceValues.OverrideName,
				},
			}
			// expose any container envvars as required here
			container.Container.Env = append(container.Container.Env, envvars...)
			// consume the lagoon-env configmap here
			container.Container.EnvFrom = []corev1.EnvFromSource{
				{
					ConfigMapRef: &corev1.ConfigMapEnvSource{
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
			// handle the default storage volumemount
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
			// create the container volume mounts
			for _, svm := range serviceTypeValues.PrimaryContainer.VolumeMounts {
				volumeMount := corev1.VolumeMount{}
				helpers.TemplateThings(tpld, svm, &volumeMount)
				container.Container.VolumeMounts = append(container.Container.VolumeMounts, volumeMount)
			}
			// mount the default storage volume if one exists
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
				deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, volume)
			}

			// set the resource limit overrides if htey are provided
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

			// append the final defined container to the spec
			deployment.Spec.Template.Spec.Containers = append(deployment.Spec.Template.Spec.Containers, container.Container)

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

				envvars := []corev1.EnvVar{
					{
						Name:  "LAGOON_GIT_SHA",
						Value: buildValues.GitSHA,
					},
					{
						Name:  "SERVICE_NAME",
						Value: serviceValues.OverrideName,
					},
				}
				linkedContainer.Container.Env = append(linkedContainer.Container.Env, envvars...)
				linkedContainer.Container.EnvFrom = []corev1.EnvFromSource{
					{
						ConfigMapRef: &corev1.ConfigMapEnvSource{
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
				deployment.Spec.Template.Spec.Containers = append(deployment.Spec.Template.Spec.Containers, linkedContainer.Container)
			}

			// end deployment template
			deployments = append(deployments, *deployment)
		}
	}
	return deployments, nil
}
