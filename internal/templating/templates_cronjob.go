package templating

import (
	"fmt"
	"math"
	"time"

	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/servicetypes"
	batchv1 "k8s.io/api/batch/v1"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metavalidation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"sigs.k8s.io/yaml"
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
		if val, ok := servicetypes.ServiceTypes[serviceValues.Type]; ok && serviceValues.Type != "external" {
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
				cronjob.Spec.Schedule = nCronjob.Schedule
				cronjob.Spec.ConcurrencyPolicy = batchv1.ForbidConcurrent
				cronjob.Spec.SuccessfulJobsHistoryLimit = helpers.Int32Ptr(0)
				cronjob.Spec.FailedJobsHistoryLimit = helpers.Int32Ptr(1)
				cronjob.Spec.StartingDeadlineSeconds = helpers.Int64Ptr(240)

				// time has already been parsed in generator/services to check for errors
				// and the default timeout is added in generator/services
				cronjobTimeout, _ := time.ParseDuration(nCronjob.Timeout)
				cSec := int64(math.Round(cronjobTimeout.Seconds()))
				cronjob.Spec.JobTemplate.Spec.ActiveDeadlineSeconds = &cSec

				if serviceValues.CronjobUseSpotInstances {
					// handle spot instance label and affinity/tolerations/selectors
					additionalLabels["lagoon.sh/spot"] = "true"
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
				podTemplateSpec, err := generatePodTemplateSpec(buildValues, serviceValues, serviceTypeValues, cronjob.ObjectMeta, templateAnnotations, nCronjob.Name, nCronjob.Command)
				if err != nil {
					return nil, err
				}
				// end cronjob template
				cronjob.Spec.JobTemplate.Spec.Template = *podTemplateSpec
				result = append(result, *cronjob)
			}
		}
	}
	return result, nil
}

func TemplateCronjobs(item batchv1.CronJob) ([]byte, error) {
	separator := []byte("---\n")
	iBytes, err := yaml.Marshal(item)
	if err != nil {
		return nil, fmt.Errorf("couldn't generate template: %v", err)
	}
	templateYAML := append(separator[:], iBytes[:]...)
	return templateYAML, nil
}
