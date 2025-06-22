package templating

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

			podTemplateSpec, _ := generatePodTemplateSpec(buildValues, serviceValues, serviceTypeValues, deployment.ObjectMeta, templateAnnotations, serviceTypeValues.PrimaryContainer.Name, "")
			// end cronjob template
			deployment.Spec.Template = *podTemplateSpec
			if buildValues.PodSpreadConstraints {
				deployment.Spec.Template.Spec.TopologySpreadConstraints = []corev1.TopologySpreadConstraint{
					{
						MaxSkew:           1,
						WhenUnsatisfiable: corev1.ScheduleAnyway,
						TopologyKey:       "kubernetes.io/hostname",
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "app.kubernetes.io/name",
									Operator: metav1.LabelSelectorOpIn,
									Values: []string{
										serviceTypeValues.Name,
									},
								},
								{
									Key:      "app.kubernetes.io/instance",
									Operator: metav1.LabelSelectorOpIn,
									Values: []string{
										serviceValues.OverrideName,
									},
								},
								{
									Key:      "lagoon.sh/project",
									Operator: metav1.LabelSelectorOpIn,
									Values: []string{
										buildValues.Project,
									},
								},
								{
									Key:      "lagoon.sh/environment",
									Operator: metav1.LabelSelectorOpIn,
									Values: []string{
										buildValues.Environment,
									},
								},
							},
						},
					},
				}
			}

			// end deployment template
			deployments = append(deployments, *deployment)
		}
	}
	return deployments, nil
}

func TemplateDeployment(item appsv1.Deployment) ([]byte, error) {
	separator := []byte("---\n")
	iBytes, err := yaml.Marshal(item)
	if err != nil {
		return nil, fmt.Errorf("couldn't generate template: %v", err)
	}
	templateYAML := append(separator[:], iBytes[:]...)
	return templateYAML, nil
}
