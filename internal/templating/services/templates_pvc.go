package services

import (
	"fmt"
	"strconv"

	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
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
				pvc, err := generateDefaultPVC(buildValues, serviceValues, val, labels, annotations)
				if err != nil {
					return nil, err
				}
				if pvc != nil {
					// end PVC template
					result = append(result, *pvc)
				}
			}
		}
	}
	// check for any additional volumes
	for _, vol := range buildValues.Volumes {
		if vol.Create {
			exists := false
			// check if an existing pvc that will be created as a default persistent volume hasn't also been defined as an additional volume
			// this will prevent creating another volume named the same
			for _, cpvc := range result {
				if lagoon.GetLagoonVolumeName(cpvc.Name) == vol.Name {
					exists = true
				}
			}
			if !exists {
				pvc, err := generateAdditionalPVC(buildValues, vol, labels, annotations)
				if err != nil {
					return nil, err
				}
				result = append(result, *pvc)
			}
		}
	}
	return result, nil
}

// generateDefaultPVC default volumes have different labels/annotations to additional values, and also handle some configuration a bit differently
func generateDefaultPVC(buildValues generator.BuildValues,
	serviceValues generator.ServiceValues,
	val servicetypes.ServiceType,
	labels, annotations map[string]string,
) (*corev1.PersistentVolumeClaim, error) {
	if serviceValues.PersistentVolumeName != "" {
		if serviceValues.PersistentVolumeName != serviceValues.OverrideName {
			// this service base volume is not needed because it is created by a different service
			// lagoon legacy templates allowed due to a funny templating issue, for multiple "basic" types to mount one volume
			// from one main service, across multiple services of the same type
			return nil, nil
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

	if serviceTypeValues.Volumes.PersistentVolumeType == corev1.ReadWriteMany {
		pvc.Spec.StorageClassName = helpers.StrPtr("bulk")
	}

	// add any remaining changes that are shared between default and additional
	err := updatePVC(
		pvc,
		&buildValues,
		serviceValues.OverrideName,
		persistentVolumeSize,
		serviceTypeValues.Volumes.PersistentVolumeType,
		labels, annotations,
		additionalLabels, additionalAnnotations,
	)
	if err != nil {
		return nil, err
	}
	// end PVC template
	return pvc, nil
}

// generateAdditionalPVC additional volumes have different labels/annotations to a default, and also handle some configuration a bit differently
func generateAdditionalPVC(
	buildValues generator.BuildValues,
	additionalVolume generator.ComposeVolume,
	labels, annotations map[string]string,
) (*corev1.PersistentVolumeClaim, error) {
	additionalLabels := map[string]string{}
	additionalAnnotations := map[string]string{}

	additionalLabels["app.kubernetes.io/name"] = lagoon.GetVolumeNameFromLagoonVolume(additionalVolume.Name)
	additionalLabels["app.kubernetes.io/instance"] = additionalVolume.Name
	additionalLabels["lagoon.sh/template"] = fmt.Sprintf("%s-%s", "additional-volume", "0.1.0")
	additionalLabels["lagoon.sh/service-type"] = "additional-volume"

	additionalAnnotations["k8up.syn.tools/backup"] = "true"
	additionalAnnotations["k8up.io/backup"] = "true"

	pvc := &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: corev1.SchemeGroupVersion.Version,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: additionalVolume.Name,
		},
	}

	// add any remaining changes that are shared between default and additional
	err := updatePVC(
		pvc,
		&buildValues,
		additionalVolume.Name,
		additionalVolume.Size,
		corev1.ReadWriteMany,
		labels, annotations,
		additionalLabels, additionalAnnotations,
	)
	if err != nil {
		return nil, err
	}
	// end PVC template
	return pvc, nil
}

// handle the remaining changes to the pvc that differentiate it from a persistent default volume and an additional volume
func updatePVC(
	pvc *corev1.PersistentVolumeClaim,
	buildValues *generator.BuildValues,
	name, size string,
	mode corev1.PersistentVolumeAccessMode,
	labels, annotations, additionalLabels, additionalAnnotations map[string]string,
) error {
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
			return fmt.Errorf("the annotations for %s are not valid: %v", name, err)
		}
	}
	// validate any labels
	if err := metavalidation.ValidateLabels(pvc.ObjectMeta.Labels, nil); err != nil {
		if len(err) != 0 {
			return fmt.Errorf("the labels for %s are not valid: %v", name, err)
		}
	}
	// check length of labels
	err := helpers.CheckLabelLength(pvc.ObjectMeta.Labels)
	if err != nil {
		return err
	}

	// start PVC templating
	// this error is also checked in composeToServiceValues
	q, err := resource.ParseQuantity(size)
	if err != nil {
		return fmt.Errorf("provided persistent volume size is not valid: %v", err)
	}
	volumeSize, _ := q.AsInt64()
	pvc.Spec = corev1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{
			mode,
		},
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				"storage": *resource.NewQuantity(volumeSize, resource.BinarySI),
			},
		},
	}
	if mode == corev1.ReadWriteMany {
		pvc.Spec.StorageClassName = helpers.StrPtr("bulk")
	}
	if buildValues.RWX2RWO || buildValues.IsCI {
		// this should be a rwo volume in CI and if the rwx2rwo flag is enabled
		pvc.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{
			corev1.ReadWriteOnce,
		}
	}
	return nil
}
