package templating

import (
	"fmt"

	"github.com/uselagoon/build-deploy-tool/internal/generator"
	networkv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// GenerateNetworkPolicy generates the lagoon template to apply.
func GenerateNetworkPolicy(
	buildValues generator.BuildValues,
) (*networkv1.NetworkPolicy, error) {
	// add the default labels
	labels := map[string]string{
		"app.kubernetes.io/managed-by": "build-deploy-tool",
		"app.kubernetes.io/instance":   "isolation-network-policy",
		"app.kubernetes.io/name":       "isolation-network-policy",
		"lagoon.sh/template":           "isolation-network-policy-0.1.0",
		"lagoon.sh/project":            buildValues.Project,
		"lagoon.sh/environment":        buildValues.Environment,
		"lagoon.sh/environmentType":    buildValues.EnvironmentType,
		"lagoon.sh/buildType":          buildValues.BuildType,
		"lagoon.sh/service":            "isolation-network-policy",
		"lagoon.sh/service-type":       "isolation-network-policy",
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
	np := networkv1.NetworkPolicy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NetworkPolicy",
			APIVersion: "networking.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "isolation-network-policy",
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: networkv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			Ingress: []networkv1.NetworkPolicyIngressRule{
				{
					From: []networkv1.NetworkPolicyPeer{
						{
							PodSelector: &metav1.LabelSelector{},
						},
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{
										Key:      "lagoon.sh/environment",
										Operator: metav1.LabelSelectorOpDoesNotExist,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return &np, nil
}

func TemplateNetworkPolicy(ingress *networkv1.NetworkPolicy) ([]byte, error) {
	separator := []byte("---\n")
	var templateYAML []byte
	iBytes, err := yaml.Marshal(ingress)
	if err != nil {
		return nil, fmt.Errorf("couldn't generate template: %v", err)
	}
	restoreResult := append(separator[:], iBytes[:]...)
	templateYAML = append(templateYAML, restoreResult[:]...)
	return templateYAML, nil
}
