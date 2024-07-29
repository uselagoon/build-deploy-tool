package templating

import (
	"github.com/uselagoon/build-deploy-tool/internal/generator"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GenerateLagoonEnvSecret generates the lagoon template to apply.
func GenerateLagoonEnvSecret(
	name string,
	buildValues generator.BuildValues,
) (corev1.Secret, error) {

	// add the default labels
	labels := map[string]string{
		"app.kubernetes.io/managed-by": "build-deploy-tool",
		"app.kubernetes.io/instance":   name,
		"app.kubernetes.io/name":       name,
		"lagoon.sh/template":           "lagoon-env-0.1.0",
		"lagoon.sh/project":            buildValues.Project,
		"lagoon.sh/environment":        buildValues.Environment,
		"lagoon.sh/environmentType":    buildValues.EnvironmentType,
		"lagoon.sh/buildType":          buildValues.BuildType,
	}

	// add the default annotations
	annotations := map[string]string{}

	// add any additional labels
	if buildValues.BuildType == "branch" {
		annotations["lagoon.sh/branch"] = buildValues.Branch
	} else if buildValues.BuildType == "pullrequest" {
		annotations["lagoon.sh/prNumber"] = buildValues.PRNumber
		annotations["lagoon.sh/prHeadBranch"] = buildValues.PRHeadBranch
		annotations["lagoon.sh/prBaseBranch"] = buildValues.PRBaseBranch
	}

	lagoonEnv := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: corev1.SchemeGroupVersion.Version,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Labels:      labels,
			Annotations: annotations,
		},
	}
	// pick which values to save into the secret based on the name
	switch name {
	case "lagoon-platform-env":
		lagoonEnv.StringData = buildValues.LagoonPlatformEnvVariables
	default:
		lagoonEnv.StringData = buildValues.LagoonEnvVariables
	}

	return lagoonEnv, nil
}
