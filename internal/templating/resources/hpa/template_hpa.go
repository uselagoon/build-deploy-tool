package hpa

import (
	"fmt"

	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metavalidation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"

	"sigs.k8s.io/yaml"
)

func GenerateHPATemplate(
	lValues generator.BuildValues,
) ([]byte, error) {
	// generate the template spec

	var result []byte
	separator := []byte("---\n")

	// add the default labels
	labels := map[string]string{
		"app.kubernetes.io/managed-by": "build-deploy-tool",
		"lagoon.sh/project":            lValues.Project,
		"lagoon.sh/environment":        lValues.Environment,
		"lagoon.sh/environmentType":    lValues.EnvironmentType,
		"lagoon.sh/buildType":          lValues.BuildType,
	}

	// add the default annotations
	annotations := map[string]string{
		"lagoon.sh/version": lValues.LagoonVersion,
	}

	// create the hpas
	for _, serviceValues := range lValues.Services {
		if serviceValues.ResourceWorkload != "" && (lValues.ResourceWorkloads[serviceValues.ResourceWorkload].HPA != nil) {
			// add any additional labels
			additionalLabels := map[string]string{}
			additionalAnnotations := map[string]string{}
			if lValues.BuildType == "branch" {
				additionalAnnotations["lagoon.sh/branch"] = lValues.Branch
			} else if lValues.BuildType == "pullrequest" {
				additionalAnnotations["lagoon.sh/prNumber"] = lValues.PRNumber
				additionalAnnotations["lagoon.sh/prHeadBranch"] = lValues.PRHeadBranch
				additionalAnnotations["lagoon.sh/prBaseBranch"] = lValues.PRBaseBranch
			}
			additionalLabels["app.kubernetes.io/name"] = serviceValues.Type
			additionalLabels["app.kubernetes.io/instance"] = serviceValues.OverrideName
			additionalLabels["lagoon.sh/service"] = serviceValues.OverrideName
			additionalLabels["lagoon.sh/service-type"] = serviceValues.Type
			hpa := &autoscalingv2.HorizontalPodAutoscaler{
				TypeMeta: metav1.TypeMeta{
					Kind:       "HorizontalPodAutoscaler",
					APIVersion: autoscalingv2.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("%s-hpa", serviceValues.OverrideName),
				},
				Spec: lValues.ResourceWorkloads[serviceValues.ResourceWorkload].HPA.Spec,
			}

			// set the scale target to the service that requested it
			// since all lagoon deployed services are deployments at the moment
			// default this set to the deployment kind, refactor in the future if lagoon supports
			// additional types (statefulsets/daemonsets?)
			hpa.Spec.ScaleTargetRef = autoscalingv2.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       serviceValues.OverrideName,
				APIVersion: "apps/v1",
			}

			hpa.ObjectMeta.Labels = labels
			hpa.ObjectMeta.Annotations = annotations

			for key, value := range additionalLabels {
				hpa.ObjectMeta.Labels[key] = value
			}
			// add any additional annotations
			for key, value := range additionalAnnotations {
				hpa.ObjectMeta.Annotations[key] = value
			}
			// validate any annotations
			if err := apivalidation.ValidateAnnotations(hpa.ObjectMeta.Annotations, nil); err != nil {
				if len(err) != 0 {
					return nil, fmt.Errorf("the annotations for %s/%s are not valid: %v", "hpa", serviceValues.Name, err)
				}
			}
			// validate any labels
			if err := metavalidation.ValidateLabels(hpa.ObjectMeta.Labels, nil); err != nil {
				if len(err) != 0 {
					return nil, fmt.Errorf("the labels for %s/%s are not valid: %v", "hpa", serviceValues.Name, err)
				}
			}

			// check length of labels
			err := helpers.CheckLabelLength(hpa.ObjectMeta.Labels)
			if err != nil {
				return nil, err
			}
			// @TODO: we should review this in the future when we stop doing `kubectl apply` in the builds :)
			// marshal the resulting ingress
			hpaBytes, err := yaml.Marshal(hpa)
			if err != nil {
				return nil, err
			}
			// add the seperator to the template so that it can be `kubectl apply` in bulk as part
			// of the current build process
			restoreResult := append(separator[:], hpaBytes[:]...)
			result = append(result, restoreResult[:]...)
		}
	}

	return result, nil
}
