package pdb

import (
	"fmt"

	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"

	policyv1 "k8s.io/api/policy/v1"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metavalidation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"

	"sigs.k8s.io/yaml"
)

func GeneratePDBTemplate(
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

	// create the pdbs
	for _, serviceValues := range lValues.Services {
		if serviceValues.ResourceWorkload != "" && (lValues.ResourceWorkloads[serviceValues.ResourceWorkload].PDB != nil) {
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
			pdb := &policyv1.PodDisruptionBudget{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PodDisruptionBudget",
					APIVersion: policyv1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("%s-pdb", serviceValues.OverrideName),
				},
				Spec: lValues.ResourceWorkloads[serviceValues.ResourceWorkload].PDB.Spec,
			}

			// set the selector target to the service that requested it)
			pdb.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"lagoon.sh/service": serviceValues.OverrideName,
				},
			}

			pdb.ObjectMeta.Labels = labels
			pdb.ObjectMeta.Annotations = annotations

			for key, value := range additionalLabels {
				pdb.ObjectMeta.Labels[key] = value
			}
			// add any additional annotations
			for key, value := range additionalAnnotations {
				pdb.ObjectMeta.Annotations[key] = value
			}
			// validate any annotations
			if err := apivalidation.ValidateAnnotations(pdb.ObjectMeta.Annotations, nil); err != nil {
				if len(err) != 0 {
					return nil, fmt.Errorf("the annotations for %s/%s are not valid: %v", "pdb", serviceValues.Name, err)
				}
			}
			// validate any labels
			if err := metavalidation.ValidateLabels(pdb.ObjectMeta.Labels, nil); err != nil {
				if len(err) != 0 {
					return nil, fmt.Errorf("the labels for %s/%s are not valid: %v", "pdb", serviceValues.Name, err)
				}
			}

			// check length of labels
			err := helpers.CheckLabelLength(pdb.ObjectMeta.Labels)
			if err != nil {
				return nil, err
			}
			// @TODO: we should review this in the future when we stop doing `kubectl apply` in the builds :)
			// marshal the resulting ingress
			pdbBytes, err := yaml.Marshal(pdb)
			if err != nil {
				return nil, err
			}
			// add the seperator to the template so that it can be `kubectl apply` in bulk as part
			// of the current build process
			restoreResult := append(separator[:], pdbBytes[:]...)
			result = append(result, restoreResult[:]...)
		}
	}
	return result, nil
}
