package templating

import (
	"fmt"

	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	machineryns "github.com/uselagoon/machinery/utils/namespace"
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

// GenerateServiceNetworkPolicies will generate any networkpolicies required for specific services if defined in the lagoon.yml file
func GenerateServiceNetworkPolicies(
	buildValues generator.BuildValues,
) ([]networkv1.NetworkPolicy, error) {
	var nps []networkv1.NetworkPolicy

	// default is to set the network policies to whatever is at the root of the lagoon.yml if provided
	lagoonNetworkPolicies := buildValues.LagoonYAML.NetworkPolicies
	// check if the environment has specific network policies, these should be used instead
	// they aren't stacked or added to the root network policies, they will replace them
	if buildValues.LagoonYAML.Environments[buildValues.Environment].NetworkPolicies != nil {
		lagoonNetworkPolicies = buildValues.LagoonYAML.Environments[buildValues.Environment].NetworkPolicies
	}
	if lagoonNetworkPolicies != nil {
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

		for _, netpol := range lagoonNetworkPolicies {
			// add the default labels
			labels := map[string]string{
				"app.kubernetes.io/managed-by": "build-deploy-tool",
				"app.kubernetes.io/instance":   "service-network-policy",
				"app.kubernetes.io/name":       "service-network-policy",
				"lagoon.sh/template":           "service-network-policy-0.1.0",
				"lagoon.sh/project":            buildValues.Project,
				"lagoon.sh/environment":        buildValues.Environment,
				"lagoon.sh/environmentType":    buildValues.EnvironmentType,
				"lagoon.sh/buildType":          buildValues.BuildType,
				"lagoon.sh/service":            netpol.Service,
				"lagoon.sh/service-type":       "network-policy",
			}

			// work out the policies for the specific service from the .lagoon.yml file here
			var npirs []networkv1.NetworkPolicyIngressRule
			for _, pp := range netpol.Projects {
				// this generates any project specific policies
				npirs = append(npirs, generateProjectIngressRule(pp))
			}

			for _, op := range netpol.Organizations {
				// this generates any organization specific policies
				npirs = append(npirs, generateOrganizationIngressRule(op))
			}
			np := networkv1.NetworkPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       "NetworkPolicy",
					APIVersion: "networking.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:        netpol.Service,
					Labels:      labels,
					Annotations: annotations,
				},
				Spec: networkv1.NetworkPolicySpec{
					PodSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							// set the podselector to be that of the requested service, all other policies defined in the .lagoon.yml file
							// will get enerated above
							"lagoon.sh/service": netpol.Service,
						},
					},
					Ingress: npirs,
				},
			}
			nps = append(nps, np)
		}
	}
	return nps, nil
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

func generateProjectIngressRule(pp lagoon.ProjectNetworkPolicies) networkv1.NetworkPolicyIngressRule {
	namespaceSelectors := []metav1.LabelSelectorRequirement{
		{
			Key:      "lagoon.sh/project",
			Operator: metav1.LabelSelectorOpIn,
			Values:   []string{pp.Name},
		},
	}
	if pp.Environment != "" {
		environmentName := machineryns.ShortenEnvironment(pp.Name, machineryns.MakeSafe(pp.Environment))
		namespaceSelectors = append(namespaceSelectors, metav1.LabelSelectorRequirement{
			Key:      "lagoon.sh/environment",
			Operator: metav1.LabelSelectorOpIn,
			Values:   []string{environmentName},
		})
	}
	if pp.EnvironmentType != "" {
		namespaceSelectors = append(namespaceSelectors, metav1.LabelSelectorRequirement{
			Key:      "lagoon.sh/environmentType",
			Operator: metav1.LabelSelectorOpIn,
			Values:   []string{pp.EnvironmentType},
		})
	}
	if pp.ExcludeEnvironments != nil {
		lagoonEnvironments := []string{}
		for _, exEnv := range pp.ExcludeEnvironments {
			lagoonEnvironments = append(lagoonEnvironments, machineryns.ShortenEnvironment(pp.Name, machineryns.MakeSafe(exEnv.Name)))
		}
		namespaceSelectors = append(namespaceSelectors, metav1.LabelSelectorRequirement{
			Key:      "lagoon.sh/environment",
			Operator: metav1.LabelSelectorOpNotIn,
			Values:   lagoonEnvironments,
		})
	}
	if pp.ExcludePullrequests {
		namespaceSelectors = append(namespaceSelectors, metav1.LabelSelectorRequirement{
			Key:      "lagoon.sh/buildType",
			Operator: metav1.LabelSelectorOpNotIn,
			Values:   []string{"pullrequest"},
		})
	}

	return networkv1.NetworkPolicyIngressRule{
		From: []networkv1.NetworkPolicyPeer{
			{
				PodSelector: &metav1.LabelSelector{},
			},
			{
				NamespaceSelector: &metav1.LabelSelector{
					MatchExpressions: namespaceSelectors,
				},
			},
		},
	}
}

func generateOrganizationIngressRule(op lagoon.OrgNetworkPolicies) networkv1.NetworkPolicyIngressRule {
	namespaceSelectors := []metav1.LabelSelectorRequirement{
		{
			Key:      "organization.lagoon.sh/name",
			Operator: metav1.LabelSelectorOpIn,
			Values:   []string{op.Name},
		},
	}
	if op.EnvironmentType != "" {
		namespaceSelectors = append(namespaceSelectors, metav1.LabelSelectorRequirement{
			Key:      "lagoon.sh/environmentType",
			Operator: metav1.LabelSelectorOpIn,
			Values:   []string{op.EnvironmentType},
		})
	}
	if op.ExcludeProjects != nil {
		lagoonProjects := []string{}
		for _, exEnv := range op.ExcludeProjects {
			lagoonProjects = append(lagoonProjects, machineryns.MakeSafe(exEnv.Name))
		}
		namespaceSelectors = append(namespaceSelectors, metav1.LabelSelectorRequirement{
			Key:      "lagoon.sh/project",
			Operator: metav1.LabelSelectorOpNotIn,
			Values:   lagoonProjects,
		})
	}
	return networkv1.NetworkPolicyIngressRule{
		From: []networkv1.NetworkPolicyPeer{
			{
				PodSelector: &metav1.LabelSelector{},
			},
			{
				NamespaceSelector: &metav1.LabelSelector{
					MatchExpressions: namespaceSelectors,
				},
			},
		},
	}
}
