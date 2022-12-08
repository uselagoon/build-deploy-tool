package services

import (
	"fmt"
	"strconv"

	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/servicetypes"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metavalidation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"sigs.k8s.io/yaml"
)

// GeneratePVCTemplate generates the lagoon template to apply.
func GeneratePVCTemplate(
	buildValues generator.BuildValues,
) ([]byte, error) {
	separator := []byte("---\n")
	var result []byte

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
	additionalLabels := map[string]string{}
	additionalAnnotations := map[string]string{}
	if buildValues.BuildType == "branch" {
		additionalAnnotations["lagoon.sh/branch"] = buildValues.Branch
	} else if buildValues.BuildType == "pullrequest" {
		additionalAnnotations["lagoon.sh/prNumber"] = buildValues.PRNumber
		additionalAnnotations["lagoon.sh/prHeadBranch"] = buildValues.PRHeadBranch
		additionalAnnotations["lagoon.sh/prBaseBranch"] = buildValues.PRBaseBranch

	}

	// for all the services that the build values generated
	// iterate over them and generate any kubernetes pvcs
	for _, serviceValues := range buildValues.Services {
		if val, ok := servicetypes.ServiceTypes[serviceValues.Type]; ok {
			if val.Volumes.PersistentVolumeSize != "" {
				if serviceValues.PersistentVolumeSize == "" {
					return nil, fmt.Errorf("no persistent volume size defined for this service")
				}
				serviceType := &servicetypes.ServiceType{}
				helpers.DeepCopy(val, serviceType)

				var pvcBytes []byte
				additionalLabels["app.kubernetes.io/name"] = serviceType.Name
				additionalLabels["app.kubernetes.io/instance"] = serviceValues.Name
				additionalLabels["lagoon.sh/template"] = fmt.Sprintf("%s-%s", serviceType.Name, "0.1.0")
				additionalLabels["lagoon.sh/service"] = serviceValues.Name
				additionalLabels["lagoon.sh/service-type"] = serviceType.Name

				additionalAnnotations["k8up.syn.tools/backup"] = strconv.FormatBool(val.Volumes.Backup)
				additionalAnnotations["k8up.io/backup"] = strconv.FormatBool(val.Volumes.Backup)

				pvc := &corev1.PersistentVolumeClaim{
					TypeMeta: metav1.TypeMeta{
						Kind:       corev1.SchemeGroupVersion.Group,
						APIVersion: corev1.SchemeGroupVersion.Version,
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: serviceValues.Name,
					},
				}
				pvc.ObjectMeta.Labels = labels
				pvc.ObjectMeta.Annotations = annotations
				for key, value := range additionalLabels {
					pvc.ObjectMeta.Labels[key] = value
				}
				// add any additional annotations
				for key, value := range additionalAnnotations {
					pvc.ObjectMeta.Annotations[key] = value
				}
				// validate any annotations
				if err := apivalidation.ValidateAnnotations(pvc.ObjectMeta.Annotations, nil); err != nil {
					if len(err) != 0 {
						return nil, fmt.Errorf("the annotations for %s are not valid: %v", serviceValues.Name, err)
					}
				}
				// validate any labels
				if err := metavalidation.ValidateLabels(pvc.ObjectMeta.Labels, nil); err != nil {
					if len(err) != 0 {
						return nil, fmt.Errorf("the labels for %s are not valid: %v", serviceValues.Name, err)
					}
				}
				// check length of labels
				err := helpers.CheckLabelLength(pvc.ObjectMeta.Labels)
				if err != nil {
					return nil, err
				}

				// start PVC template
				q, err := resource.ParseQuantity(serviceValues.PersistentVolumeSize)
				if err != nil {
					return nil, fmt.Errorf("provided persistent volume size is not valid: %v", err)
				}
				volumeSize, _ := q.AsInt64()
				pvc.Spec = corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						val.Volumes.PersistentVolumeType,
					},
					StorageClassName: helpers.StrPtr("bulk"),
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							"storage": *resource.NewQuantity(volumeSize, resource.BinarySI),
						},
					},
				}
				// end PVC template

				pvcBytes, err = yaml.Marshal(pvc)
				if err != nil {
					return nil, err
				}

				// @TODO: we should review this in the future when we stop doing `kubectl apply` in the builds :)
				// add the seperator to the template so that it can be `kubectl apply` in bulk as part
				// of the current build process
				// join all dbaas-consumer templates together
				restoreResult := append(separator[:], pvcBytes[:]...)
				result = append(result, restoreResult[:]...)
			}
		}
	}
	return result, nil
}
