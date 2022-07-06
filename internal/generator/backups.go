package generator

import (
	"fmt"

	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
)

const (
	defaultCheckSchedule          = "M H(3-6) * * 0"
	defaultPruneSchedule          = "M H(3-6) * * 0"
	defaultBackupSchedule         = "M H(22-2) * * *"
	hourlyDefaultBackupRetention  = 0
	dailyDefaultBackupRetention   = 7
	weeklyDefaultBackupRetention  = 6
	monthlyDefaultBackupRetention = 1
)

func generateBackupValues(
	lagoonValues *BuildValues,
	lYAML *lagoon.YAML,
	mergedVariables []lagoon.EnvironmentVariable,
	debug bool,
) error {
	var err error
	// builds need to calculate a new schedule from multiple places for backups
	// create a new schedule placeholder set to the default value so it can be adjusted through this
	// generator
	newBackupSchedule := defaultBackupSchedule
	switch lagoonValues.BuildType {
	case "branch":
		if lagoonValues.EnvironmentType == "development" {
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
	lagoonValues.Backup.BackupSchedule, err = helpers.ConvertCrontab(lagoonValues.Namespace, newBackupSchedule)
	if err != nil {
		return fmt.Errorf("Unable to convert crontab for default backup schedule: %v", err)
	}
	flagCheckSchedule := helpers.GetEnv("K8UP_WEEKLY_RANDOM_FEATURE_FLAG", defaultCheckSchedule, debug)
	if flagCheckSchedule == "enabled" {
		lagoonValues.Backup.CheckSchedule = "@weekly-random"
	} else {
		lagoonValues.Backup.CheckSchedule, err = helpers.ConvertCrontab(lagoonValues.Namespace, flagCheckSchedule)
		if err != nil {
			return fmt.Errorf("Unable to convert crontab for default check schedule: %v", err)
		}
	}
	flagPruneSchedule := helpers.GetEnv("K8UP_WEEKLY_RANDOM_FEATURE_FLAG", defaultPruneSchedule, debug)
	if flagPruneSchedule == "enabled" {
		lagoonValues.Backup.PruneSchedule = "@weekly-random"
	} else {
		lagoonValues.Backup.PruneSchedule, err = helpers.ConvertCrontab(lagoonValues.Namespace, flagPruneSchedule)
		if err != nil {
			return fmt.Errorf("Unable to convert crontab for default prune schedule: %v", err)
		}
	}

	lagoonValues.Backup.PruneRetention.Hourly, err = helpers.EGetEnvInt("HOURLY_BACKUP_DEFAULT_RETENTION", hourlyDefaultBackupRetention, debug)
	if err != nil {
		return fmt.Errorf("Unable to convert hourly retention provided in the environment variable to integer")
	}
	lagoonValues.Backup.PruneRetention.Daily, err = helpers.EGetEnvInt("DAILY_BACKUP_DEFAULT_RETENTION", dailyDefaultBackupRetention, debug)
	if err != nil {
		return fmt.Errorf("Unable to convert daily retention provided in the environment variable to integer")
	}
	lagoonValues.Backup.PruneRetention.Weekly, err = helpers.EGetEnvInt("WEEKLY_BACKUP_DEFAULT_RETENTION", weeklyDefaultBackupRetention, debug)
	if err != nil {
		return fmt.Errorf("Unable to convert weekly retention provided in the environment variable to integer")
	}
	lagoonValues.Backup.PruneRetention.Monthly, err = helpers.EGetEnvInt("MONTHLY_BACKUP_DEFAULT_RETENTION", monthlyDefaultBackupRetention, debug)
	if err != nil {
		return fmt.Errorf("Unable to convert monthly retention provided in the environment variable to integer")
	}

	if lYAML.BackupRetention.Production.Hourly != nil && lagoonValues.EnvironmentType == "production" {
		lagoonValues.Backup.PruneRetention.Hourly = *lYAML.BackupRetention.Production.Hourly
	}
	if lYAML.BackupRetention.Production.Daily != nil && lagoonValues.EnvironmentType == "production" {
		lagoonValues.Backup.PruneRetention.Daily = *lYAML.BackupRetention.Production.Daily
	}
	if lYAML.BackupRetention.Production.Weekly != nil && lagoonValues.EnvironmentType == "production" {
		lagoonValues.Backup.PruneRetention.Weekly = *lYAML.BackupRetention.Production.Weekly
	}
	if lYAML.BackupRetention.Production.Monthly != nil && lagoonValues.EnvironmentType == "production" {
		lagoonValues.Backup.PruneRetention.Monthly = *lYAML.BackupRetention.Production.Monthly
	}
	if lYAML.BackupSchedule.Production != "" && lagoonValues.EnvironmentType == "production" {
		lagoonValues.Backup.BackupSchedule, err = helpers.ConvertCrontab(lagoonValues.Namespace, lYAML.BackupSchedule.Production)
		if err != nil {
			return fmt.Errorf("Unable to convert crontab for default backup schedule from .lagoon.yml: %v", err)
		}
	}
	// check for custom baas backup variables
	lagoonBaaSCustomBackupEndpoint, _ := lagoon.GetLagoonVariable("LAGOON_BAAS_CUSTOM_BACKUP_ENDPOINT", []string{"build", "global"}, mergedVariables)
	if lagoonBaaSCustomBackupEndpoint != nil {
		lagoonValues.Backup.S3Endpoint = lagoonBaaSCustomBackupEndpoint.Value
	}
	lagoonBaaSCustomBackupBucket, _ := lagoon.GetLagoonVariable("LAGOON_BAAS_CUSTOM_BACKUP_BUCKET", []string{"build", "global"}, mergedVariables)
	if lagoonBaaSCustomBackupBucket != nil {
		lagoonValues.Backup.S3BucketName = lagoonBaaSCustomBackupBucket.Value
	}
	lagoonBaaSCustomBackupAccessKey, _ := lagoon.GetLagoonVariable("LAGOON_BAAS_CUSTOM_BACKUP_ACCESS_KEY", []string{"build", "global"}, mergedVariables)
	lagoonBaaSCustomBackupSecretKey, _ := lagoon.GetLagoonVariable("LAGOON_BAAS_CUSTOM_BACKUP_SECRET_KEY", []string{"build", "global"}, mergedVariables)
	if lagoonBaaSCustomBackupAccessKey != nil && lagoonBaaSCustomBackupSecretKey != nil {
		lagoonValues.Backup.CustomLocation.BackupLocationAccessKey = lagoonBaaSCustomBackupAccessKey.Value
		lagoonValues.Backup.CustomLocation.BackupLocationSecretKey = lagoonBaaSCustomBackupSecretKey.Value
		lagoonValues.Backup.S3SecretName = "lagoon-baas-custom-backup-credentials"
	}
	// check for custom baas restore variables
	lagoonBaaSCustomRestoreAccessKey, _ := lagoon.GetLagoonVariable("LAGOON_BAAS_CUSTOM_RESTORE_ACCESS_KEY", []string{"build", "global"}, mergedVariables)
	lagoonBaaSCustomRestoreSecretKey, _ := lagoon.GetLagoonVariable("LAGOON_BAAS_CUSTOM_RESTORE_SECRET_KEY", []string{"build", "global"}, mergedVariables)
	if lagoonBaaSCustomRestoreAccessKey != nil && lagoonBaaSCustomRestoreSecretKey != nil {
		lagoonValues.Backup.CustomLocation.RestoreLocationAccessKey = lagoonBaaSCustomRestoreAccessKey.Value
		lagoonValues.Backup.CustomLocation.RestoreLocationSecretKey = lagoonBaaSCustomRestoreSecretKey.Value
	}
	return nil
}
