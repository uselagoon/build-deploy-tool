package backups

import (
	"fmt"

	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	k8upv1alpha1 "github.com/vshn/k8up/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/yaml"
)

func GenerateBackupSchedule(
	lValues generator.BuildValues,
) ([]byte, error) {
	// generate the template spec
	s3Spec := &k8upv1alpha1.S3Spec{}
	if lValues.Backup.S3Endpoint != "" {
		s3Spec.Endpoint = lValues.Backup.S3Endpoint
	}
	if lValues.Backup.S3BucketName != "" {
		s3Spec.Bucket = lValues.Backup.S3BucketName
	}
	if lValues.Backup.S3SecretName != "" {
		s3Spec.AccessKeyIDSecretRef = &corev1.SecretKeySelector{
			Key: "access-key",
			LocalObjectReference: corev1.LocalObjectReference{
				Name: lValues.Backup.S3SecretName,
			},
		}
		s3Spec.SecretAccessKeySecretRef = &corev1.SecretKeySelector{
			Key: "secret-key",
			LocalObjectReference: corev1.LocalObjectReference{
				Name: lValues.Backup.S3SecretName,
			},
		}
	}
	// create the schedule
	schedule := &k8upv1alpha1.Schedule{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Schedule",
			APIVersion: k8upv1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "k8up-lagoon-backup-schedule",
		},
		Spec: k8upv1alpha1.ScheduleSpec{
			Backend: &k8upv1alpha1.Backend{
				RepoPasswordSecretRef: &corev1.SecretKeySelector{
					Key: "repo-pw",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "baas-repo-pw",
					},
				},
				S3: s3Spec,
			},
			Backup: &k8upv1alpha1.BackupSchedule{
				ScheduleCommon: &k8upv1alpha1.ScheduleCommon{
					Schedule: k8upv1alpha1.ScheduleDefinition(lValues.Backup.BackupSchedule),
				},
			},
			Check: &k8upv1alpha1.CheckSchedule{
				ScheduleCommon: &k8upv1alpha1.ScheduleCommon{
					Schedule: k8upv1alpha1.ScheduleDefinition(lValues.Backup.CheckSchedule),
				},
			},
			Prune: &k8upv1alpha1.PruneSchedule{
				PruneSpec: k8upv1alpha1.PruneSpec{
					Retention: k8upv1alpha1.RetentionPolicy{
						KeepHourly:  lValues.Backup.PruneRetention.Hourly,
						KeepDaily:   lValues.Backup.PruneRetention.Daily,
						KeepWeekly:  lValues.Backup.PruneRetention.Weekly,
						KeepMonthly: lValues.Backup.PruneRetention.Monthly,
					},
				},
				ScheduleCommon: &k8upv1alpha1.ScheduleCommon{
					Schedule: k8upv1alpha1.ScheduleDefinition(lValues.Backup.PruneSchedule),
				},
			},
		},
	}
	// add the default labels
	schedule.ObjectMeta.Labels = map[string]string{
		"app.kubernetes.io/name":       "k8up-schedule",
		"app.kubernetes.io/instance":   "k8up-lagoon-backup-schedule",
		"app.kubernetes.io/managed-by": "build-deploy-tool",
		"lagoon.sh/template":           fmt.Sprintf("%s-%s", "k8up-schedule", "0.1.0"),
		"lagoon.sh/service":            "k8up-lagoon-backup-schedule",
		"lagoon.sh/service-type":       "k8up-schedule",
		"lagoon.sh/project":            lValues.Project,
		"lagoon.sh/environment":        lValues.Environment,
		"lagoon.sh/environmentType":    lValues.EnvironmentType,
		"lagoon.sh/buildType":          lValues.BuildType,
	}

	// add the default annotations
	schedule.ObjectMeta.Annotations = map[string]string{
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
		schedule.ObjectMeta.Labels[key] = value
	}
	// add any additional annotations
	for key, value := range additionalAnnotations {
		schedule.ObjectMeta.Annotations[key] = value
	}

	// check length of labels
	err := helpers.CheckLabelLength(schedule.ObjectMeta.Labels)
	if err != nil {
		return nil, err
	}
	// @TODO: we should review this in the future when we stop doing `kubectl apply` in the builds :)
	// marshal the resulting ingress
	scheduleBytes, err := yaml.Marshal(schedule)
	if err != nil {
		return nil, err
	}
	// add the seperator to the template so that it can be `kubectl apply` in bulk as part
	// of the current build process
	separator := []byte("---\n")
	result := append(separator[:], scheduleBytes[:]...)
	if lValues.Backup.CustomLocation.BackupLocationAccessKey != "" && lValues.Backup.CustomLocation.BackupLocationSecretKey != "" {
		backupSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: "lagoon-baas-custom-backup-credentials",
			},
			StringData: map[string]string{
				"access-key": lValues.Backup.CustomLocation.BackupLocationAccessKey,
				"secret-key": lValues.Backup.CustomLocation.BackupLocationSecretKey,
			},
		}
		backupSecretBytes, err := yaml.Marshal(backupSecret)
		if err != nil {
			return nil, err
		}
		backupResult := append(separator[:], backupSecretBytes[:]...)
		result = append(result, backupResult[:]...)
	}
	if lValues.Backup.CustomLocation.RestoreLocationAccessKey != "" && lValues.Backup.CustomLocation.RestoreLocationSecretKey != "" {
		restoreSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: "lagoon-baas-custom-restore-credentials",
			},
			StringData: map[string]string{
				"access-key": lValues.Backup.CustomLocation.RestoreLocationAccessKey,
				"secret-key": lValues.Backup.CustomLocation.RestoreLocationSecretKey,
			},
		}
		restoreSecretBytes, err := yaml.Marshal(restoreSecret)
		if err != nil {
			return nil, err
		}
		restoreResult := append(separator[:], restoreSecretBytes[:]...)
		result = append(result, restoreResult[:]...)
	}
	return result, nil
}
