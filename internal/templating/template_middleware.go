package templating

import (
	"fmt"

	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metavalidation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"sigs.k8s.io/yaml"

	traefik "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
)

// GenerateRegistrySecretTemplate generates the lagoon template to apply.
func GenerateMiddleware(
	buildValues generator.BuildValues,
) ([]traefik.Middleware, error) {
	var result []traefik.Middleware

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
	if buildValues.BuildType == "branch" {
		annotations["lagoon.sh/branch"] = buildValues.Branch
	} else if buildValues.BuildType == "pullrequest" {
		annotations["lagoon.sh/prNumber"] = buildValues.PRNumber
		annotations["lagoon.sh/prHeadBranch"] = buildValues.PRHeadBranch
		annotations["lagoon.sh/prBaseBranch"] = buildValues.PRBaseBranch
	}
	// iterate over the container registries and generate any kubernetes secrets
	for name, middleware := range buildValues.TraefikMiddlewares {
		additionalLabels := map[string]string{}
		additionalAnnotations := map[string]string{}

		additionalLabels["app.kubernetes.io/name"] = name
		additionalLabels["app.kubernetes.io/instance"] = "traefik-middleware"
		additionalLabels["lagoon.sh/template"] = fmt.Sprintf("traefik-middleware-%s", "0.1.0")

		irs := &traefik.Middleware{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Middleware",
				APIVersion: "traefik.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: middleware,
		}

		labelsCopy := &map[string]string{}
		helpers.DeepCopy(labels, labelsCopy)
		annotationsCopy := &map[string]string{}
		helpers.DeepCopy(annotations, annotationsCopy)

		for key, value := range additionalLabels {
			(*labelsCopy)[key] = value
		}
		// add any additional annotations
		for key, value := range additionalAnnotations {
			(*annotationsCopy)[key] = value
		}
		irs.ObjectMeta.Labels = *labelsCopy
		irs.ObjectMeta.Annotations = *annotationsCopy
		// validate any annotations
		if err := apivalidation.ValidateAnnotations(irs.ObjectMeta.Annotations, nil); err != nil {
			if len(err) != 0 {
				return nil, fmt.Errorf("the annotations for %s are not valid: %v", name, err)
			}
		}
		// validate any labels
		if err := metavalidation.ValidateLabels(irs.ObjectMeta.Labels, nil); err != nil {
			if len(err) != 0 {
				return nil, fmt.Errorf("the labels for %s are not valid: %v", name, err)
			}
		}
		// check length of labels
		err := helpers.CheckLabelLength(irs.ObjectMeta.Labels)
		if err != nil {
			return nil, err
		}

		// end middleware templates
		result = append(result, *irs)
	}
	return result, nil
}

func TemplateMiddleware(middleware *traefik.Middleware) ([]byte, error) {
	separator := []byte("---\n")
	var templateYAML []byte
	iBytes, err := yaml.Marshal(middleware)
	if err != nil {
		return nil, fmt.Errorf("couldn't generate template: %v", err)
	}
	restoreResult := append(separator[:], iBytes[:]...)
	templateYAML = append(templateYAML, restoreResult[:]...)
	return templateYAML, nil
}
