package services

import (
	"fmt"

	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/servicetypes"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metavalidation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"sigs.k8s.io/yaml"
)

// GenerateDeploymentTemplate generates the lagoon template to apply.
func GenerateDeploymentTemplate(
	buildValues generator.BuildValues,
) ([]byte, error) {
	separator := []byte("---\n")
	var result []byte

	// check linked services
	checkedServices := LinkedServiceCalculator(buildValues.Services)

	// for all the services that the build values generated
	// iterate over them and generate any kubernetes deployments
	for _, serviceValues := range checkedServices {
		if val, ok := servicetypes.ServiceTypes[serviceValues.Type]; ok {
			serviceTypeValues := &servicetypes.ServiceType{}
			helpers.DeepCopy(val, serviceTypeValues)

			var deploymentBytes []byte

			// add the default labels
			labels := map[string]string{
				"app.kubernetes.io/managed-by": "build-deploy-tool",
				"lagoon.sh/project":            buildValues.Project,
				"lagoon.sh/environment":        buildValues.Environment,
				"lagoon.sh/environmentType":    buildValues.EnvironmentType,
				"lagoon.sh/buildType":          buildValues.BuildType,
				"app.kubernetes.io/name":       serviceTypeValues.Name,
				"app.kubernetes.io/instance":   serviceValues.Name,
				"lagoon.sh/template":           fmt.Sprintf("%s-%s", serviceTypeValues.Name, "0.1.0"),
				"lagoon.sh/service":            serviceValues.Name,
				"lagoon.sh/service-type":       serviceTypeValues.Name,
			}

			// add the default annotations
			annotations := map[string]string{
				"lagoon.sh/version":      buildValues.LagoonVersion,
				"lagoon.sh/configMapSha": buildValues.ConfigMapSha,
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

			deployment := &appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Deployment",
					APIVersion: fmt.Sprintf("%s/%s", appsv1.SchemeGroupVersion.Group, appsv1.SchemeGroupVersion.Version),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        serviceValues.Name,
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
					return nil, fmt.Errorf("the annotations for %s are not valid: %v", serviceValues.Name, err)
				}
			}
			// validate any labels
			if err := metavalidation.ValidateLabels(deployment.ObjectMeta.Labels, nil); err != nil {
				if len(err) != 0 {
					return nil, fmt.Errorf("the labels for %s are not valid: %v", serviceValues.Name, err)
				}
			}
			// check length of labels
			err := helpers.CheckLabelLength(deployment.ObjectMeta.Labels)
			if err != nil {
				return nil, err
			}

			// start deployment template
			depMeta := metav1.ObjectMeta{
				Labels:      deployment.ObjectMeta.Labels,
				Annotations: deployment.ObjectMeta.Annotations,
			}
			deployment.Spec.Template.ObjectMeta = depMeta
			deployment.Spec.Replicas = helpers.Int32Ptr(1)
			if serviceValues.Replicas != 0 {
				deployment.Spec.Replicas = &serviceValues.Replicas
			}
			deployment.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name":     serviceTypeValues.Name,
					"app.kubernetes.io/instance": serviceValues.Name,
				},
			}
			deployment.Spec.Strategy = serviceTypeValues.Strategy

			// disable service links, this prevents some environment variables that confuse lagoon services being
			// added to the containers
			deployment.Spec.Template.Spec.EnableServiceLinks = helpers.BoolPtr(false)
			// set the priority class
			deployment.Spec.Template.Spec.PriorityClassName = fmt.Sprintf("lagoon-priority-%s", buildValues.EnvironmentType)

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
				deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, volume)
			}
			// if there are any specific container overrides provided, handle those here
			for _, pcv := range serviceTypeValues.PrimaryContainer.Volumes {
				volume := corev1.Volume{}
				helpers.TemplateThings(serviceValues, pcv, &volume)
				deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, volume)
			}
			for _, scv := range serviceTypeValues.SecondaryContainer.Volumes {
				volume := corev1.Volume{}
				helpers.TemplateThings(serviceValues, scv, &volume)
				deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, volume)
			}
			// end set up any volumes this deployment can use

			// handle any image pull secrets
			pullsecrets := []corev1.LocalObjectReference{}
			for _, pullsecret := range buildValues.ImagePullSecrets {
				pullsecrets = append(pullsecrets, corev1.LocalObjectReference{
					Name: pullsecret.Name,
				})
			}
			deployment.Spec.Template.Spec.ImagePullSecrets = pullsecrets

			// start working out the containers to add
			// add any init containers
			if serviceTypeValues.InitContainer.Name != "" && serviceValues.PodSecurityContext.RunAsUser == 10001 {
				init := serviceTypeValues.InitContainer
				for _, svm := range serviceTypeValues.InitContainer.VolumeMounts {
					volumeMount := corev1.VolumeMount{}
					helpers.TemplateThings(serviceValues, svm, &volumeMount)
					init.Container.VolumeMounts = append(init.Container.VolumeMounts, volumeMount)
				}
				cmd := []string{}
				for _, c := range init.Command {
					var c2 string
					helpers.TemplateThings(serviceValues, c, &c2)
					cmd = append(cmd, c2)
				}
				init.Container.Command = cmd
				deployment.Spec.Template.Spec.InitContainers = append(deployment.Spec.Template.Spec.InitContainers, init.Container)
			}

			// handle the primary container
			container := serviceTypeValues.PrimaryContainer
			// if the service can change the port
			if serviceTypeValues.Ports.CanChangePort {
				// check if the port override is defined
				if serviceValues.ServicePort != 0 {
					// and change the port in the container definition to suit
					container.Container.ReadinessProbe.ProbeHandler.TCPSocket.Port.IntVal = serviceValues.ServicePort
					container.Container.LivenessProbe.ProbeHandler.TCPSocket.Port.IntVal = serviceValues.ServicePort
					container.Container.ReadinessProbe.ProbeHandler.TCPSocket.Port.IntVal = serviceValues.ServicePort
					container.Container.Ports[0].ContainerPort = serviceValues.ServicePort
				}
			}

			// handle setting the rest of the containers specs with values from the service or build values
			container.Container.Name = container.Name
			container.Container.Image = serviceValues.ImageName

			// set up cronjobs if required
			cronjobs := ""
			for _, cronjob := range serviceValues.InPodCronjobs {
				cronjobs = fmt.Sprintf("%s%s %s\n", cronjobs, cronjob.Schedule, cronjob.Command)
			}
			envvars := []corev1.EnvVar{
				{
					Name:  "LAGOON_GIT_SHA",
					Value: buildValues.GitSha,
				},
				{
					Name:  "CRONJOBS",
					Value: cronjobs,
				},
			}
			for _, envvar := range envvars {
				container.Container.Env = append(container.Container.Env, envvar)
			}
			container.Container.EnvFrom = []corev1.EnvFromSource{
				{
					ConfigMapRef: &corev1.ConfigMapEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "lagoon-env",
						},
					},
				},
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
				container.Container.VolumeMounts = append(container.Container.VolumeMounts, volumeMount)
			}
			for _, svm := range serviceTypeValues.PrimaryContainer.VolumeMounts {
				volumeMount := corev1.VolumeMount{}
				helpers.TemplateThings(serviceValues, svm, &volumeMount)
				container.Container.VolumeMounts = append(container.Container.VolumeMounts, volumeMount)
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
				linkedContainer.Container.Image = serviceValues.LinkedService.ImageName

				envvars := []corev1.EnvVar{
					{
						Name:  "LAGOON_GIT_SHA",
						Value: buildValues.GitSha,
					},
				}
				for _, envvar := range envvars {
					linkedContainer.Container.Env = append(linkedContainer.Container.Env, envvar)
				}
				linkedContainer.Container.EnvFrom = []corev1.EnvFromSource{
					{
						ConfigMapRef: &corev1.ConfigMapEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "lagoon-env",
							},
						},
					},
				}
				for _, dsm := range buildValues.DynamicSecretMounts {
					volumeMount := corev1.VolumeMount{
						Name:      dsm.Name,
						MountPath: dsm.MountPath,
						ReadOnly:  dsm.ReadOnly,
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
					helpers.TemplateThings(serviceValues, svm, &volumeMount)
					linkedContainer.Container.VolumeMounts = append(linkedContainer.Container.VolumeMounts, volumeMount)
				}
				deployment.Spec.Template.Spec.Containers = append(deployment.Spec.Template.Spec.Containers, linkedContainer.Container)
			}

			// end deployment template

			deploymentBytes, err = yaml.Marshal(deployment)
			if err != nil {
				return nil, err
			}

			// @TODO: we should review this in the future when we stop doing `kubectl apply` in the builds :)
			// add the seperator to the template so that it can be `kubectl apply` in bulk as part
			// of the current build process
			// join all dbaas-consumer templates together
			restoreResult := append(separator[:], deploymentBytes[:]...)
			result = append(result, restoreResult[:]...)
		}
	}
	return result, nil
}

// LinkedServiceCalculator checks the provided services to see if there are any linked services
// linked services are mostly just `nginx-php` but lagoon has the possibility to support more than this in the future
func LinkedServiceCalculator(services []generator.ServiceValues) []generator.ServiceValues {
	linkedMap := make(map[string][]generator.ServiceValues)
	retServices := []generator.ServiceValues{}
	linkedOrder := []string{}

	// go over the services twice to extract just the linked services (the override names will be the same in a linked service)
	for _, s1 := range services {
		for _, s2 := range services {
			if s1.OverrideName == s2.OverrideName && s1.Name != s2.Name {
				linkedMap[s1.OverrideName] = append(linkedMap[s1.OverrideName], s1)
				linkedOrder = helpers.AppendIfMissing(linkedOrder, s1.OverrideName)
			}
		}
	}
	// go over the services again and any that are in the services that aren't in the linked map (again the override name is the key)
	// add it as a standalone service
	for _, s1 := range services {
		if _, ok := linkedMap[s1.OverrideName]; !ok {
			retServices = append(retServices, s1)
		}
	}

	// go over the linked services and add the linkedservice to the main service
	// example would be adding the `php` service in docker-compose to the `nginx` service as a `LinkedService` definition
	// this allows the generated service values to carry across
	for _, name := range linkedOrder {
		service := generator.ServiceValues{}
		if len(linkedMap[name]) == 2 {
			for idx, s := range linkedMap[name] {
				if idx == 0 {
					service = s
				}
				if idx == 1 {
					service.LinkedService = &s
				}
			}
		}
		// then add it to the slice of services to return
		retServices = append(retServices, service)
	}
	return retServices
}
