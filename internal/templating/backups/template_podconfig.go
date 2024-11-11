package backups

import (
	"fmt"

	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"

	k8upv1 "github.com/k8up-io/k8up/v2/api/v1"
	corev1 "k8s.io/api/core/v1"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metavalidation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"

	"sigs.k8s.io/yaml"
)

func GenerateBackupPodConfig(
	lValues generator.BuildValues,
) ([]byte, error) {
	// generate the template spec

	var result []byte
	separator := []byte("---\n")

	// create the podconfig
	if lValues.BackupsEnabled {
		switch lValues.Backup.K8upVersion {
		case "v2":
			if lValues.PodSecurityContext.RunAsUser != 0 {
				podConfig := &k8upv1.PodConfig{
					TypeMeta: metav1.TypeMeta{
						Kind:       "PodConfig",
						APIVersion: k8upv1.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "k8up-rootless-workload-podconfig",
					},
					Spec: k8upv1.PodConfigSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								SecurityContext: &corev1.PodSecurityContext{
									RunAsUser:  helpers.Int64Ptr(lValues.PodSecurityContext.RunAsUser),
									RunAsGroup: helpers.Int64Ptr(lValues.PodSecurityContext.RunAsGroup),
									FSGroup:    helpers.Int64Ptr(lValues.PodSecurityContext.FsGroup),
								},
							},
						},
					},
				}
				// add the default labels
				podConfig.ObjectMeta.Labels = map[string]string{
					"app.kubernetes.io/name":       "k8up-podconfig",
					"app.kubernetes.io/instance":   "k8up-rootless-workload-podconfig",
					"app.kubernetes.io/managed-by": "build-deploy-tool",
					"lagoon.sh/template":           fmt.Sprintf("%s-%s", "k8up-podconfig", "0.1.0"),
					"lagoon.sh/service":            "k8up-rootless-workload-podconfig",
					"lagoon.sh/service-type":       "k8up-podconfig",
					"lagoon.sh/project":            lValues.Project,
					"lagoon.sh/environment":        lValues.Environment,
					"lagoon.sh/environmentType":    lValues.EnvironmentType,
					"lagoon.sh/buildType":          lValues.BuildType,
				}

				// add the default annotations
				podConfig.ObjectMeta.Annotations = map[string]string{
					"lagoon.sh/version": lValues.LagoonVersion,
				}

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
				for key, value := range additionalLabels {
					podConfig.ObjectMeta.Labels[key] = value
				}
				// add any additional annotations
				for key, value := range additionalAnnotations {
					podConfig.ObjectMeta.Annotations[key] = value
				}
				// validate any annotations
				if err := apivalidation.ValidateAnnotations(podConfig.ObjectMeta.Annotations, nil); err != nil {
					if len(err) != 0 {
						return nil, fmt.Errorf("the annotations for %s are not valid: %v", "k8up-rootless-workload-podconfig", err)
					}
				}
				// validate any labels
				if err := metavalidation.ValidateLabels(podConfig.ObjectMeta.Labels, nil); err != nil {
					if len(err) != 0 {
						return nil, fmt.Errorf("the labels for %s are not valid: %v", "k8up-rootless-workload-podconfig", err)
					}
				}

				// check length of labels
				err := helpers.CheckLabelLength(podConfig.ObjectMeta.Labels)
				if err != nil {
					return nil, err
				}
				// @TODO: we should review this in the future when we stop doing `kubectl apply` in the builds :)
				// marshal the resulting ingress
				podconfigBytes, err := yaml.Marshal(podConfig)
				if err != nil {
					return nil, err
				}
				podconfigBytes, _ = CleanupPodConfigYAML(podconfigBytes)
				// add the seperator to the template so that it can be `kubectl apply` in bulk as part
				// of the current build process
				result = append(separator[:], podconfigBytes[:]...)
			}
		}
	}
	return result, nil
}

// helper function to remove data from the yaml spec so that kubectl will apply without validation errors
// this is only needed because we use kubectl in builds for now
func CleanupPodConfigYAML(a []byte) ([]byte, error) {
	tmpMap := map[string]interface{}{}
	yaml.Unmarshal(a, &tmpMap)
	delete(tmpMap["spec"].(map[string]interface{})["template"].(map[string]interface{}), "metadata")
	delete(tmpMap["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{}), "containers")
	b, _ := yaml.Marshal(tmpMap)
	return b, nil
}
