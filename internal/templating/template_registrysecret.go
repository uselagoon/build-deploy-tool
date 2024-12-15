package templating

import (
	"encoding/base64"
	"fmt"

	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	corev1 "k8s.io/api/core/v1"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metavalidation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"sigs.k8s.io/yaml"
)

// GenerateRegistrySecretTemplate generates the lagoon template to apply.
func GenerateRegistrySecretTemplate(
	buildValues generator.BuildValues,
) ([]corev1.Secret, error) {
	var result []corev1.Secret

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
	for _, containerRegistry := range buildValues.ContainerRegistry {
		additionalLabels := map[string]string{}
		additionalAnnotations := map[string]string{}

		additionalLabels["app.kubernetes.io/name"] = containerRegistry.Name
		additionalLabels["app.kubernetes.io/instance"] = "internal-registry-secret"
		additionalLabels["lagoon.sh/template"] = fmt.Sprintf("internal-registry-secret-%s", "0.1.0")

		irs := &corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: corev1.SchemeGroupVersion.Version,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: containerRegistry.SecretName,
			},
			Type: corev1.SecretTypeDockerConfigJson,
			Data: map[string][]byte{
				".dockerconfigjson": []byte(fmt.Sprintf(`{"auths":{"%s":{"username":"%s","password":"%s","auth":"%s"}}}`,
					containerRegistry.URL,
					containerRegistry.Username,
					containerRegistry.Password,
					base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", containerRegistry.Username, containerRegistry.Password))))),
			},
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
				return nil, fmt.Errorf("the annotations for %s are not valid: %v", containerRegistry.Name, err)
			}
		}
		// validate any labels
		if err := metavalidation.ValidateLabels(irs.ObjectMeta.Labels, nil); err != nil {
			if len(err) != 0 {
				return nil, fmt.Errorf("the labels for %s are not valid: %v", containerRegistry.Name, err)
			}
		}
		// check length of labels
		err := helpers.CheckLabelLength(irs.ObjectMeta.Labels)
		if err != nil {
			return nil, err
		}

		// end registry secret template
		result = append(result, *irs)
	}
	return result, nil
}

func TemplateSecret(item corev1.Secret) ([]byte, error) {
	separator := []byte("---\n")
	iBytes, err := yaml.Marshal(item)
	if err != nil {
		return nil, fmt.Errorf("couldn't generate template: %v", err)
	}
	templateYAML := append(separator[:], iBytes[:]...)
	return templateYAML, nil
}
