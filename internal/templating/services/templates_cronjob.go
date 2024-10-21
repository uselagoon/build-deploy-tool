package services

import (
	"fmt"

	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/servicetypes"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metavalidation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
)

// GenerateCronjobTemplate generates the lagoon template to apply.
func GenerateCronjobTemplate(
	buildValues generator.BuildValues,
) ([]batchv1.CronJob, error) {
	var result []batchv1.CronJob

	// check linked services
	checkedServices := LinkedServiceCalculator(buildValues.Services)

	// for all the services that the build values generated
	// iterate over them and generate any kubernetes cronjobs
	for _, serviceValues := range checkedServices {
		if val, ok := servicetypes.ServiceTypes[serviceValues.Type]; ok {
			for _, nCronjob := range serviceValues.NativeCronjobs {
				serviceTypeValues := &servicetypes.ServiceType{}
				helpers.DeepCopy(val, serviceTypeValues)

				// add the default labels
				labels := map[string]string{
					"app.kubernetes.io/managed-by": "build-deploy-tool",
					"app.kubernetes.io/name":       fmt.Sprintf("cronjob-%s", serviceTypeValues.Name),
					"app.kubernetes.io/instance":   fmt.Sprintf("cronjob-%s", serviceValues.OverrideName),
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

				cronjob := &batchv1.CronJob{
					TypeMeta: metav1.TypeMeta{
						Kind:       "CronJob",
						APIVersion: fmt.Sprintf("%s/%s", batchv1.SchemeGroupVersion.Group, batchv1.SchemeGroupVersion.Version),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:        nCronjob.Name,
						Labels:      labels,
						Annotations: annotations,
					},
				}
				cronjob.ObjectMeta.Labels = labels
				cronjob.ObjectMeta.Annotations = annotations
				cronjob.Spec.JobTemplate.Spec.Template.Spec.DNSConfig = &corev1.PodDNSConfig{
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
				cronjob.Spec.Schedule = nCronjob.Schedule
				cronjob.Spec.ConcurrencyPolicy = batchv1.ForbidConcurrent
				cronjob.Spec.SuccessfulJobsHistoryLimit = helpers.Int32Ptr(0)
				cronjob.Spec.FailedJobsHistoryLimit = helpers.Int32Ptr(1)
				cronjob.Spec.StartingDeadlineSeconds = helpers.Int64Ptr(240)
				cronjob.Spec.JobTemplate.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyNever

				if serviceValues.CronjobUseSpotInstances {
					// handle spot instance label and affinity/tolerations/selectors
					additionalLabels["lagoon.sh/spot"] = "true"
					cronjob.Spec.JobTemplate.Spec.Template.Spec.Affinity = &corev1.Affinity{
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
					cronjob.Spec.JobTemplate.Spec.Template.Spec.Tolerations = []corev1.Toleration{
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
						cronjob.Spec.JobTemplate.Spec.Template.Spec.NodeSelector = map[string]string{
							"lagoon.sh/spot": "true",
						}
					}
				}

				for key, value := range additionalLabels {
					cronjob.ObjectMeta.Labels[key] = value
				}
				// add any additional annotations
				for key, value := range additionalAnnotations {
					cronjob.ObjectMeta.Annotations[key] = value
				}
				// validate any annotations
				if err := apivalidation.ValidateAnnotations(cronjob.ObjectMeta.Annotations, nil); err != nil {
					if len(err) != 0 {
						return nil, fmt.Errorf("the annotations for %s are not valid: %v", serviceValues.OverrideName, err)
					}
				}
				// validate any labels
				if err := metavalidation.ValidateLabels(cronjob.ObjectMeta.Labels, nil); err != nil {
					if len(err) != 0 {
						return nil, fmt.Errorf("the labels for %s are not valid: %v", serviceValues.OverrideName, err)
					}
				}
				// check length of labels
				err := helpers.CheckLabelLength(cronjob.ObjectMeta.Labels)
				if err != nil {
					return nil, err
				}

				// start cronjob template
				cronjob.Spec.JobTemplate.Spec.Template.ObjectMeta = metav1.ObjectMeta{
					Labels:      map[string]string{},
					Annotations: map[string]string{},
				}
				for key, value := range cronjob.ObjectMeta.Labels {
					cronjob.Spec.JobTemplate.Spec.Template.ObjectMeta.Labels[key] = value
				}
				// add any additional annotations
				for key, value := range cronjob.ObjectMeta.Annotations {
					cronjob.Spec.JobTemplate.Spec.Template.ObjectMeta.Annotations[key] = value
				}
				for key, value := range templateAnnotations {
					cronjob.Spec.JobTemplate.Spec.Template.ObjectMeta.Annotations[key] = value
				}

				// disable service links, this prevents some environment variables that confuse lagoon services being
				// added to the containers
				cronjob.Spec.JobTemplate.Spec.Template.Spec.EnableServiceLinks = helpers.BoolPtr(false)
				// set the priority class
				cronjob.Spec.JobTemplate.Spec.Template.Spec.PriorityClassName = fmt.Sprintf("lagoon-priority-%s", buildValues.EnvironmentType)

				// handle the podescurity from rootless workloads
				if buildValues.PodSecurityContext.RunAsUser != 0 {
					cronjob.Spec.JobTemplate.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{
						RunAsUser:  helpers.Int64Ptr(buildValues.PodSecurityContext.RunAsUser),
						RunAsGroup: helpers.Int64Ptr(buildValues.PodSecurityContext.RunAsGroup),
						FSGroup:    helpers.Int64Ptr(buildValues.PodSecurityContext.FsGroup),
					}
				}
				// some services have a fsgroup override
				if serviceTypeValues.PodSecurityContext.HasDefault {
					cronjob.Spec.JobTemplate.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{
						FSGroup: helpers.Int64Ptr(serviceTypeValues.PodSecurityContext.FSGroup),
					}
				}
				if buildValues.PodSecurityContext.OnRootMismatch {
					fsGroupChangePolicy := corev1.FSGroupChangeOnRootMismatch
					if cronjob.Spec.JobTemplate.Spec.Template.Spec.SecurityContext != nil {
						cronjob.Spec.JobTemplate.Spec.Template.Spec.SecurityContext.FSGroupChangePolicy = &fsGroupChangePolicy
					} else {
						cronjob.Spec.JobTemplate.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{
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
					cronjob.Spec.JobTemplate.Spec.Template.Spec.Volumes = append(cronjob.Spec.JobTemplate.Spec.Template.Spec.Volumes, volume)
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
					cronjob.Spec.JobTemplate.Spec.Template.Spec.Volumes = append(cronjob.Spec.JobTemplate.Spec.Template.Spec.Volumes, volume)
				}

				// if there are any specific container volume overrides provided, handle those here
				for _, pcv := range serviceTypeValues.PrimaryContainer.Volumes {
					volume := corev1.Volume{}
					helpers.TemplateThings(tpld, pcv, &volume)
					cronjob.Spec.JobTemplate.Spec.Template.Spec.Volumes = append(cronjob.Spec.JobTemplate.Spec.Template.Spec.Volumes, volume)
				}
				for _, scv := range serviceTypeValues.SecondaryContainer.Volumes {
					volume := corev1.Volume{}
					helpers.TemplateThings(tpld, scv, &volume)
					cronjob.Spec.JobTemplate.Spec.Template.Spec.Volumes = append(cronjob.Spec.JobTemplate.Spec.Template.Spec.Volumes, volume)
				}

				// future handle any additional volume mounts here
				// @TODO

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
				cronjob.Spec.JobTemplate.Spec.Template.Spec.ImagePullSecrets = pullsecrets

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
						cronjob.Spec.JobTemplate.Spec.Template.Spec.InitContainers = append(cronjob.Spec.JobTemplate.Spec.Template.Spec.InitContainers, init.Container)
					}
				}

				// handle the primary container for the service type
				container := serviceTypeValues.PrimaryContainer

				// handle setting the rest of the containers specs with values from the service or build values
				container.Container.Name = nCronjob.Name
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
				container.Container.Env = append(container.Container.Env, envvars...)
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
					cronjob.Spec.JobTemplate.Spec.Template.Spec.Volumes = append(cronjob.Spec.JobTemplate.Spec.Template.Spec.Volumes, volume)
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

				// strip ports from the cronjobs
				container.Container.Ports = nil
				container.Container.ReadinessProbe = nil
				container.Container.LivenessProbe = nil

				container.Container.Command = []string{"/lagoon/cronjob.sh", nCronjob.Command}

				// append the final defined container to the spec
				cronjob.Spec.JobTemplate.Spec.Template.Spec.Containers = append(cronjob.Spec.JobTemplate.Spec.Template.Spec.Containers, container.Container)

				// end cronjob template
				result = append(result, *cronjob)
			}
		}
	}
	return result, nil
}
