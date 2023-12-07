package generator

import (
	"fmt"

	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
)

const (
	defaultCheckSchedule          = "M H(5-8) * * 1"
	defaultPruneSchedule          = "M H(3-5) * * 0"
	defaultBackupSchedule         = "M H(22-2) * * *"
	hourlyDefaultBackupRetention  = 0
	dailyDefaultBackupRetention   = 7
	weeklyDefaultBackupRetention  = 6
	monthlyDefaultBackupRetention = 1

	// TODO: make this configurable
	baasBucketPrefix = "baas"
)

func generateBackupValues(
	buildValues *BuildValues,
	lYAML *lagoon.YAML,
	mergedVariables []lagoon.EnvironmentVariable,
	debug bool,
) error {
	var err error
	// builds need to calculate a new schedule from multiple places for backups
	// create a new schedule placeholder set to the default value so it can be adjusted through this
	// generator
	newBackupSchedule := defaultBackupSchedule

	customBackupConfig := CheckFeatureFlag("CUSTOM_BACKUP_CONFIG", mergedVariables, debug)
	if customBackupConfig == "enabled" {
		switch buildValues.BuildType {
		case "promote":
			if buildValues.EnvironmentType == "production" {
				lagoonBackupDevSchedule, _ := lagoon.GetLagoonVariable("LAGOON_BACKUP_PROD_SCHEDULE", []string{"build", "global"}, mergedVariables)
				devBackupSchedule := ""
				if lagoonBackupDevSchedule != nil {
					devBackupSchedule = helpers.GetEnv("LAGOON_FEATURE_BACKUP_PROD_SCHEDULE", lagoonBackupDevSchedule.Value, debug)
				} else {
					devBackupSchedule = helpers.GetEnv("LAGOON_FEATURE_BACKUP_PROD_SCHEDULE", newBackupSchedule, debug)
				}
				if devBackupSchedule != "" {
					newBackupSchedule = devBackupSchedule
				}
			}
		case "branch":
			if buildValues.EnvironmentType == "production" {
				lagoonBackupDevSchedule, _ := lagoon.GetLagoonVariable("LAGOON_BACKUP_PROD_SCHEDULE", []string{"build", "global"}, mergedVariables)
				devBackupSchedule := ""
				if lagoonBackupDevSchedule != nil {
					devBackupSchedule = helpers.GetEnv("LAGOON_FEATURE_BACKUP_PROD_SCHEDULE", lagoonBackupDevSchedule.Value, debug)
				} else {
					devBackupSchedule = helpers.GetEnv("LAGOON_FEATURE_BACKUP_PROD_SCHEDULE", newBackupSchedule, debug)
				}
				if devBackupSchedule != "" {
					newBackupSchedule = devBackupSchedule
				}
			}
			if buildValues.EnvironmentType == "development" {
				lagoonBackupDevSchedule, _ := lagoon.GetLagoonVariable("LAGOON_BACKUP_DEV_SCHEDULE", []string{"build", "global"}, mergedVariables)
				devBackupSchedule := ""
				if lagoonBackupDevSchedule != nil {
					devBackupSchedule = helpers.GetEnv("LAGOON_FEATURE_BACKUP_DEV_SCHEDULE", lagoonBackupDevSchedule.Value, debug)
				} else {
					devBackupSchedule = helpers.GetEnv("LAGOON_FEATURE_BACKUP_DEV_SCHEDULE", newBackupSchedule, debug)
				}
				if devBackupSchedule != "" {
					newBackupSchedule = devBackupSchedule
				}
			}
		case "pullrequest":
			lagoonBackupPRSchedule, _ := lagoon.GetLagoonVariable("LAGOON_BACKUP_PR_SCHEDULE", []string{"build", "global"}, mergedVariables)
			prBackupSchedule := ""
			if lagoonBackupPRSchedule != nil {
				prBackupSchedule = helpers.GetEnv("LAGOON_FEATURE_BACKUP_PR_SCHEDULE", lagoonBackupPRSchedule.Value, debug)
			} else {
				prBackupSchedule = helpers.GetEnv("LAGOON_FEATURE_BACKUP_PR_SCHEDULE", prBackupSchedule, debug)
			}
			if prBackupSchedule == "" {
				lagoonBackupDevSchedule, _ := lagoon.GetLagoonVariable("LAGOON_BACKUP_DEV_SCHEDULE", []string{"build", "global"}, mergedVariables)
				if lagoonBackupDevSchedule != nil {
					newBackupSchedule = helpers.GetEnv("LAGOON_FEATURE_BACKUP_DEV_SCHEDULE", lagoonBackupDevSchedule.Value, debug)
				} else {
					newBackupSchedule = helpers.GetEnv("LAGOON_FEATURE_BACKUP_DEV_SCHEDULE", newBackupSchedule, debug)
				}
			} else {
				newBackupSchedule = prBackupSchedule
			}
		}
	}

	buildValues.Backup.BackupSchedule, err = helpers.ConvertCrontab(buildValues.Namespace, newBackupSchedule)
	if err != nil {
		return fmt.Errorf("Unable to convert crontab for default backup schedule: %v", err)
	}

	// start: get variables from the build pod that may have been added by the controller
	flagCheckSchedule := helpers.GetEnv("K8UP_WEEKLY_RANDOM_FEATURE_FLAG", defaultCheckSchedule, debug)
	if flagCheckSchedule == "enabled" {
		buildValues.Backup.CheckSchedule = "@weekly-random"
	} else {
		buildValues.Backup.CheckSchedule, err = helpers.ConvertCrontab(fmt.Sprintf("%s", buildValues.Namespace), defaultCheckSchedule)
		if err != nil {
			return fmt.Errorf("Unable to convert crontab for default check schedule: %v", err)
		}
	}
	flagPruneSchedule := helpers.GetEnv("K8UP_WEEKLY_RANDOM_FEATURE_FLAG", defaultPruneSchedule, debug)
	if flagPruneSchedule == "enabled" {
		buildValues.Backup.PruneSchedule = "@weekly-random"
	} else {
		buildValues.Backup.PruneSchedule, err = helpers.ConvertCrontab(fmt.Sprintf("%s", buildValues.Namespace), defaultPruneSchedule)
		if err != nil {
			return fmt.Errorf("Unable to convert crontab for default prune schedule: %v", err)
		}
	}

	buildValues.Backup.PruneRetention.Hourly, err = helpers.EGetEnvInt("HOURLY_BACKUP_DEFAULT_RETENTION", hourlyDefaultBackupRetention, debug)
	if err != nil {
		return fmt.Errorf("Unable to convert hourly retention provided in the environment variable to integer")
	}
	buildValues.Backup.PruneRetention.Daily, err = helpers.EGetEnvInt("DAILY_BACKUP_DEFAULT_RETENTION", dailyDefaultBackupRetention, debug)
	if err != nil {
		return fmt.Errorf("Unable to convert daily retention provided in the environment variable to integer")
	}
	buildValues.Backup.PruneRetention.Weekly, err = helpers.EGetEnvInt("WEEKLY_BACKUP_DEFAULT_RETENTION", weeklyDefaultBackupRetention, debug)
	if err != nil {
		return fmt.Errorf("Unable to convert weekly retention provided in the environment variable to integer")
	}
	buildValues.Backup.PruneRetention.Monthly, err = helpers.EGetEnvInt("MONTHLY_BACKUP_DEFAULT_RETENTION", monthlyDefaultBackupRetention, debug)
	if err != nil {
		return fmt.Errorf("Unable to convert monthly retention provided in the environment variable to integer")
	}
	// :end

	if lYAML.BackupRetention.Production.Hourly != nil && buildValues.EnvironmentType == "production" {
		buildValues.Backup.PruneRetention.Hourly = *lYAML.BackupRetention.Production.Hourly
	}
	if lYAML.BackupRetention.Production.Daily != nil && buildValues.EnvironmentType == "production" {
		buildValues.Backup.PruneRetention.Daily = *lYAML.BackupRetention.Production.Daily
	}
	if lYAML.BackupRetention.Production.Weekly != nil && buildValues.EnvironmentType == "production" {
		buildValues.Backup.PruneRetention.Weekly = *lYAML.BackupRetention.Production.Weekly
	}
	if lYAML.BackupRetention.Production.Monthly != nil && buildValues.EnvironmentType == "production" {
		buildValues.Backup.PruneRetention.Monthly = *lYAML.BackupRetention.Production.Monthly
	}
	if lYAML.BackupSchedule.Production != "" && buildValues.EnvironmentType == "production" {
		buildValues.Backup.BackupSchedule, err = helpers.ConvertCrontab(buildValues.Namespace, lYAML.BackupSchedule.Production)
		if err != nil {
			return fmt.Errorf("Unable to convert crontab for default backup schedule from .lagoon.yml: %v", err)
		}
	}

	// work out the bucket name
	lagoonBaaSBackupBucket, _ := lagoon.GetLagoonVariable("LAGOON_BAAS_BUCKET_NAME", []string{"build", "global"}, mergedVariables)
	if lagoonBaaSBackupBucket != nil {
		buildValues.Backup.S3BucketName = lagoonBaaSBackupBucket.Value
	} else {
		lagoonSharedBaasBucket, _ := lagoon.GetLagoonVariable("LAGOON_SYSTEM_PROJECT_SHARED_BUCKET", []string{"internal_system"}, mergedVariables)
		if lagoonSharedBaasBucket != nil {
			buildValues.Backup.S3BucketName = fmt.Sprintf("%s/%s-%s", lagoonSharedBaasBucket.Value, baasBucketPrefix, buildValues.Project)
		} else {
			buildValues.Backup.S3BucketName = fmt.Sprintf("%s-%s", baasBucketPrefix, buildValues.Project)
		}
	}

	// check for custom baas backup variables in the API
	lagoonBaaSCustomBackupEndpoint, _ := lagoon.GetLagoonVariable("LAGOON_BAAS_CUSTOM_BACKUP_ENDPOINT", []string{"build", "global"}, mergedVariables)
	if lagoonBaaSCustomBackupEndpoint != nil {
		buildValues.Backup.S3Endpoint = lagoonBaaSCustomBackupEndpoint.Value
	}
	lagoonBaaSCustomBackupBucket, _ := lagoon.GetLagoonVariable("LAGOON_BAAS_CUSTOM_BACKUP_BUCKET", []string{"build", "global"}, mergedVariables)
	if lagoonBaaSCustomBackupBucket != nil {
		buildValues.Backup.S3BucketName = lagoonBaaSCustomBackupBucket.Value
	}
	lagoonBaaSCustomBackupAccessKey, _ := lagoon.GetLagoonVariable("LAGOON_BAAS_CUSTOM_BACKUP_ACCESS_KEY", []string{"build", "global"}, mergedVariables)
	lagoonBaaSCustomBackupSecretKey, _ := lagoon.GetLagoonVariable("LAGOON_BAAS_CUSTOM_BACKUP_SECRET_KEY", []string{"build", "global"}, mergedVariables)
	if lagoonBaaSCustomBackupAccessKey != nil && lagoonBaaSCustomBackupSecretKey != nil {
		buildValues.Backup.CustomLocation.BackupLocationAccessKey = lagoonBaaSCustomBackupAccessKey.Value
		buildValues.Backup.CustomLocation.BackupLocationSecretKey = lagoonBaaSCustomBackupSecretKey.Value
		buildValues.Backup.S3SecretName = "lagoon-baas-custom-backup-credentials"
	}
	// check for custom baas restore variables
	lagoonBaaSCustomRestoreAccessKey, _ := lagoon.GetLagoonVariable("LAGOON_BAAS_CUSTOM_RESTORE_ACCESS_KEY", []string{"build", "global"}, mergedVariables)
	lagoonBaaSCustomRestoreSecretKey, _ := lagoon.GetLagoonVariable("LAGOON_BAAS_CUSTOM_RESTORE_SECRET_KEY", []string{"build", "global"}, mergedVariables)
	if lagoonBaaSCustomRestoreAccessKey != nil && lagoonBaaSCustomRestoreSecretKey != nil {
		buildValues.Backup.CustomLocation.RestoreLocationAccessKey = lagoonBaaSCustomRestoreAccessKey.Value
		buildValues.Backup.CustomLocation.RestoreLocationSecretKey = lagoonBaaSCustomRestoreSecretKey.Value
	}
	return nil
}
