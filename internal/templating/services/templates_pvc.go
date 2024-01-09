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
)

// GeneratePVCTemplate generates the lagoon template to apply.
func GeneratePVCTemplate(
	buildValues generator.BuildValues,
) ([]corev1.PersistentVolumeClaim, error) {
	var result []corev1.PersistentVolumeClaim

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

	// check linked services
	checkedServices := LinkedServiceCalculator(buildValues.Services)

	// for all the services that the build values generated
	// iterate over them and generate any kubernetes services
	for _, serviceValues := range checkedServices {
		if val, ok := servicetypes.ServiceTypes[serviceValues.Type]; ok {
			if val.Volumes.PersistentVolumeSize != "" {
				if serviceValues.PersistentVolumeName != "" {
					if serviceValues.PersistentVolumeName != serviceValues.OverrideName {
						// this service base volume is not needed because it is created by a different service
						// lagoon legacy templates allowed due to a funny templating issue, for multiple "basic" types to mount one volume
						// from one main service, across multiple services of the same type
						continue
					}
				}
				serviceTypeValues := &servicetypes.ServiceType{}
				helpers.DeepCopy(val, serviceTypeValues)
				persistentVolumeSize := serviceTypeValues.Volumes.PersistentVolumeSize
				if serviceValues.PersistentVolumeSize != "" {
					persistentVolumeSize = serviceValues.PersistentVolumeSize
				}
				serviceType := &servicetypes.ServiceType{}
				helpers.DeepCopy(val, serviceType)

				additionalLabels := map[string]string{}
				additionalAnnotations := map[string]string{}

				additionalLabels["app.kubernetes.io/name"] = serviceType.Name
				additionalLabels["app.kubernetes.io/instance"] = serviceValues.OverrideName
				additionalLabels["lagoon.sh/template"] = fmt.Sprintf("%s-%s", serviceType.Name, "0.1.0")
				additionalLabels["lagoon.sh/service"] = serviceValues.OverrideName
				additionalLabels["lagoon.sh/service-type"] = serviceType.Name

				additionalAnnotations["k8up.syn.tools/backup"] = strconv.FormatBool(serviceTypeValues.Volumes.Backup)
				additionalAnnotations["k8up.io/backup"] = strconv.FormatBool(serviceTypeValues.Volumes.Backup)

				pvc := &corev1.PersistentVolumeClaim{
					TypeMeta: metav1.TypeMeta{
						Kind:       "PersistentVolumeClaim",
						APIVersion: corev1.SchemeGroupVersion.Version,
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: serviceValues.OverrideName,
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
				pvc.ObjectMeta.Labels = *labelsCopy
				pvc.ObjectMeta.Annotations = *annotationsCopy
				// validate any annotations
				if err := apivalidation.ValidateAnnotations(pvc.ObjectMeta.Annotations, nil); err != nil {
					if len(err) != 0 {
						return nil, fmt.Errorf("the annotations for %s are not valid: %v", serviceValues.OverrideName, err)
					}
				}
				// validate any labels
				if err := metavalidation.ValidateLabels(pvc.ObjectMeta.Labels, nil); err != nil {
					if len(err) != 0 {
						return nil, fmt.Errorf("the labels for %s are not valid: %v", serviceValues.OverrideName, err)
					}
				}
				// check length of labels
				err := helpers.CheckLabelLength(pvc.ObjectMeta.Labels)
				if err != nil {
					return nil, err
				}

				// start PVC template
				q, err := resource.ParseQuantity(persistentVolumeSize)
				if err != nil {
					return nil, fmt.Errorf("provided persistent volume size is not valid: %v", err)
				}
				volumeSize, _ := q.AsInt64()
				pvc.Spec = corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						serviceTypeValues.Volumes.PersistentVolumeType,
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							"storage": *resource.NewQuantity(volumeSize, resource.BinarySI),
						},
					},
				}
				if serviceTypeValues.Volumes.PersistentVolumeType == corev1.ReadWriteMany {
					pvc.Spec.StorageClassName = helpers.StrPtr("bulk")
				}
				// end PVC template
				result = append(result, *pvc)
			}
		}
	}
	return result, nil
}
