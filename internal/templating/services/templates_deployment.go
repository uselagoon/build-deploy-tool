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

	// add the default labels
	labels := map[string]string{
		"app.kubernetes.io/managed-by": "build-deploy-tool",
		"lagoon.sh/project":            buildValues.Project,
		"lagoon.sh/environment":        buildValues.Environment,
		"lagoon.sh/environmentType":    buildValues.EnvironmentType,
		"lagoon.sh/buildType":          buildValues.BuildType,
	}

	// add the default annotations
	annotations := map[string]string{
		"lagoon.sh/version": buildValues.LagoonVersion,
	}

	// add any additional labels
	additionalLabels := map[string]string{}
	additionalAnnotations := map[string]string{}
	if buildValues.BuildType == "branch" {
		additionalAnnotations["lagoon.sh/branch"] = buildValues.Branch
	} else if buildValues.BuildType == "pullrequest" {
		additionalAnnotations["lagoon.sh/prNumber"] = buildValues.PRNumber
		additionalAnnotations["lagoon.sh/prHeadBranch"] = buildValues.PRHeadBranch
		additionalAnnotations["lagoon.sh/prBaseBranch"] = buildValues.PRBaseBranch

	}

	// for all the services that the build values generated
	// iterate over them and generate any kubernetes deployments
	for _, serviceValues := range buildValues.Services {
		if val, ok := servicetypes.ServiceTypes[serviceValues.Type]; ok {
			// if val.Volumes.PersistentVolumeSize != "" {
			// 	if serviceValues.PersistentVolumePath == "" {
			// 		return nil, fmt.Errorf("no persistent volume size defined for this service")
			// 	}
			// }
			serviceTypeValues := &servicetypes.ServiceType{}
			helpers.DeepCopy(val, serviceTypeValues)

			var deploymentBytes []byte
			additionalLabels["app.kubernetes.io/name"] = serviceTypeValues.Name
			additionalLabels["app.kubernetes.io/instance"] = serviceValues.Name
			additionalLabels["lagoon.sh/template"] = fmt.Sprintf("%s-%s", serviceTypeValues.Name, "0.1.0")
			additionalLabels["lagoon.sh/service"] = serviceValues.Name
			additionalLabels["lagoon.sh/service-type"] = serviceTypeValues.Name
			additionalAnnotations["lagoon.sh/configMapSha"] = buildValues.ConfigMapSha

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
				Labels:      labels,
				Annotations: annotations,
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

			// set up an volumes this deployment can use
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

			// handle any image pull secrets
			pullsecrets := []corev1.LocalObjectReference{}
			for _, pullsecret := range buildValues.ImagePullSecrets {
				pullsecrets = append(pullsecrets, corev1.LocalObjectReference{
					Name: pullsecret.Name,
				})
			}
			deployment.Spec.Template.Spec.ImagePullSecrets = pullsecrets

			// start working out the containers to add
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

			cronjobs := ""
			for _, cronjob := range serviceValues.InPodCronjobs {
				cronjobs = fmt.Sprintf("%s%s %s\n", cronjobs, cronjob.Schedule, cronjob.Command)
			}
			container.Container.Env = []corev1.EnvVar{
				{
					Name:  "LAGOON_GIT_SHA",
					Value: buildValues.GitSha,
				},
				{
					Name:  "CRONJOBS",
					Value: cronjobs,
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
			// append the final defined container to the spec
			deployment.Spec.Template.Spec.Containers = append(deployment.Spec.Template.Spec.Containers, container.Container)

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
