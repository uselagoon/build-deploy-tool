package lagoonenv

import (
	"github.com/uselagoon/build-deploy-tool/internal/generator"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GenerateLagoonEnvConfigMap generates the lagoon template to apply.
func GenerateLagoonEnvConfigMap(
	buildValues generator.BuildValues,
) (corev1.ConfigMap, error) {

	// add the default labels
	labels := map[string]string{
		"app.kubernetes.io/managed-by": "build-deploy-tool",
		"app.kubernetes.io/instance":   "lagoon-env",
		"app.kubernetes.io/name":       "lagoon-env",
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

	variables := map[string]string{}

	// add variables from the project/environment/build created variables
	for _, v := range buildValues.EnvironmentVariables {
		if v.Scope == "global" || v.Scope == "runtime" {
			variables[v.Name] = v.Value
		}
	}

	// add dbaas variables to lagoon-env
	for k, v := range buildValues.DBaaSVariables {
		variables[k] = v
	}

	lagoonEnv := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: corev1.SchemeGroupVersion.Version,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "lagoon-env",
			Labels:      labels,
			Annotations: annotations,
		},
		Data: variables,
	}

	return lagoonEnv, nil
}
