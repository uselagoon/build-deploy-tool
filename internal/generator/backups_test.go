package generator

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
)

func Test_generateBackupValues(t *testing.T) {
	type args struct {
		buildValues     *BuildValues
		lYAML           *lagoon.YAML
		mergedVariables []lagoon.EnvironmentVariable
		debug           bool
	}
	tests := []struct {
		name    string
		args    args
		vars    []helpers.EnvironmentVariable
		wantErr bool
		want    *BuildValues
	}{
		{
			name: "test1",
			args: args{
				buildValues: &BuildValues{
					BuildType:             "branch",
					EnvironmentType:       "development",
					Project:               "example-project",
					Namespace:             "example-com-main",
					DefaultBackupSchedule: "M H(22-2) * * *",
				},
				lYAML:           &lagoon.YAML{},
				mergedVariables: []lagoon.EnvironmentVariable{},
			},
			want: &BuildValues{
				BuildType:             "branch",
				EnvironmentType:       "development",
				Project:               "example-project",
				Namespace:             "example-com-main",
				DefaultBackupSchedule: "M H(22-2) * * *",
				Backup: BackupConfiguration{
					BackupSchedule: "31 1 * * *",
					CheckSchedule:  "31 6 * * 1",
					PruneSchedule:  "31 4 * * 0",
					S3BucketName:   "baas-example-project",
					PruneRetention: PruneRetention{
						Hourly:  0,
						Daily:   7,
						Weekly:  6,
						Monthly: 1,
					},
				},
			},
		},
		{
			name: "test2 - dev schedule from lagoon api variable",
			args: args{
				buildValues: &BuildValues{
					BuildType:             "branch",
					EnvironmentType:       "development",
					Project:               "example-project",
					Namespace:             "example-com-main",
					DefaultBackupSchedule: "M H(22-2) * * *",
				},
				lYAML: &lagoon.YAML{},
				mergedVariables: []lagoon.EnvironmentVariable{
					{Name: "LAGOON_FEATURE_FLAG_CUSTOM_BACKUP_CONFIG", Value: "enabled", Scope: "global"},
					{Name: "LAGOON_BACKUP_DEV_SCHEDULE", Value: "M/15 23 * * 0-5", Scope: "build"},
				},
			},
			want: &BuildValues{
				BuildType:             "branch",
				EnvironmentType:       "development",
				Project:               "example-project",
				Namespace:             "example-com-main",
				DefaultBackupSchedule: "M H(22-2) * * *",
				Backup: BackupConfiguration{
					BackupSchedule: "1,16,31,46 23 * * 0-5",
					CheckSchedule:  "31 6 * * 1",
					PruneSchedule:  "31 4 * * 0",
					S3BucketName:   "baas-example-project",
					PruneRetention: PruneRetention{
						Hourly:  0,
						Daily:   7,
						Weekly:  6,
						Monthly: 1,
					},
				},
			},
		},
		{
			name: "test3 - dev schedule from build pod variable",
			args: args{
				buildValues: &BuildValues{
					BuildType:             "branch",
					EnvironmentType:       "development",
					Project:               "example-project",
					Namespace:             "example-com-main",
					DefaultBackupSchedule: "M H(22-2) * * *",
				},
				lYAML: &lagoon.YAML{},
				mergedVariables: []lagoon.EnvironmentVariable{
					{Name: "LAGOON_FEATURE_FLAG_CUSTOM_BACKUP_CONFIG", Value: "enabled", Scope: "global"},
				},
			},
			vars: []helpers.EnvironmentVariable{
				{Name: "LAGOON_FEATURE_BACKUP_DEV_SCHEDULE", Value: "1,16,31,46 23 * * 0-5"},
			},
			want: &BuildValues{
				BuildType:             "branch",
				EnvironmentType:       "development",
				Project:               "example-project",
				Namespace:             "example-com-main",
				DefaultBackupSchedule: "M H(22-2) * * *",
				Backup: BackupConfiguration{
					BackupSchedule: "1,16,31,46 23 * * 0-5",
					CheckSchedule:  "31 6 * * 1",
					PruneSchedule:  "31 4 * * 0",
					S3BucketName:   "baas-example-project",
					PruneRetention: PruneRetention{
						Hourly:  0,
						Daily:   7,
						Weekly:  6,
						Monthly: 1,
					},
				},
			},
		},
		{
			name: "test4- pr schedule from lagoon api variable",
			args: args{
				buildValues: &BuildValues{
					BuildType:             "pullrequest",
					EnvironmentType:       "development",
					Project:               "example-project",
					Namespace:             "example-com-main",
					DefaultBackupSchedule: "M H(22-2) * * *",
				},
				lYAML: &lagoon.YAML{},
				mergedVariables: []lagoon.EnvironmentVariable{
					{Name: "LAGOON_FEATURE_FLAG_CUSTOM_BACKUP_CONFIG", Value: "enabled", Scope: "global"},
					{Name: "LAGOON_BACKUP_PR_SCHEDULE", Value: "M/15 23 * * 0-5", Scope: "build"},
				},
			},
			want: &BuildValues{
				BuildType:             "pullrequest",
				EnvironmentType:       "development",
				Project:               "example-project",
				Namespace:             "example-com-main",
				DefaultBackupSchedule: "M H(22-2) * * *",
				Backup: BackupConfiguration{
					BackupSchedule: "1,16,31,46 23 * * 0-5",
					CheckSchedule:  "31 6 * * 1",
					PruneSchedule:  "31 4 * * 0",
					S3BucketName:   "baas-example-project",
					PruneRetention: PruneRetention{
						Hourly:  0,
						Daily:   7,
						Weekly:  6,
						Monthly: 1,
					},
				},
			},
		},
		{
			name: "test5 - pr schedule from build pod variable",
			args: args{
				buildValues: &BuildValues{
					BuildType:             "pullrequest",
					EnvironmentType:       "development",
					Project:               "example-project",
					Namespace:             "example-com-main",
					DefaultBackupSchedule: "M H(22-2) * * *",
				},
				lYAML: &lagoon.YAML{},
				mergedVariables: []lagoon.EnvironmentVariable{
					{Name: "LAGOON_FEATURE_FLAG_CUSTOM_BACKUP_CONFIG", Value: "enabled", Scope: "global"},
				},
			},
			vars: []helpers.EnvironmentVariable{
				{Name: "LAGOON_FEATURE_BACKUP_PR_SCHEDULE", Value: "1,16,31,46 23 * * 0-5"},
			},
			want: &BuildValues{
				BuildType:             "pullrequest",
				EnvironmentType:       "development",
				Project:               "example-project",
				Namespace:             "example-com-main",
				DefaultBackupSchedule: "M H(22-2) * * *",
				Backup: BackupConfiguration{
					BackupSchedule: "1,16,31,46 23 * * 0-5",
					CheckSchedule:  "31 6 * * 1",
					PruneSchedule:  "31 4 * * 0",
					S3BucketName:   "baas-example-project",
					PruneRetention: PruneRetention{
						Hourly:  0,
						Daily:   7,
						Weekly:  6,
						Monthly: 1,
					},
				},
			},
		},
		{
			name: "test6 - pr env with dev schedule from lagoon api variable",
			args: args{
				buildValues: &BuildValues{
					BuildType:             "pullrequest",
					EnvironmentType:       "development",
					Project:               "example-project",
					Namespace:             "example-com-main",
					DefaultBackupSchedule: "M H(22-2) * * *",
				},
				lYAML: &lagoon.YAML{},
				mergedVariables: []lagoon.EnvironmentVariable{
					{Name: "LAGOON_FEATURE_FLAG_CUSTOM_BACKUP_CONFIG", Value: "enabled", Scope: "global"},
					{Name: "LAGOON_BACKUP_DEV_SCHEDULE", Value: "M/15 23 * * 0-5", Scope: "build"},
				},
			},
			want: &BuildValues{
				BuildType:             "pullrequest",
				EnvironmentType:       "development",
				Project:               "example-project",
				Namespace:             "example-com-main",
				DefaultBackupSchedule: "M H(22-2) * * *",
				Backup: BackupConfiguration{
					BackupSchedule: "1,16,31,46 23 * * 0-5",
					CheckSchedule:  "31 6 * * 1",
					PruneSchedule:  "31 4 * * 0",
					S3BucketName:   "baas-example-project",
					PruneRetention: PruneRetention{
						Hourly:  0,
						Daily:   7,
						Weekly:  6,
						Monthly: 1,
					},
				},
			},
		},
		{
			name: "test7 - pr env with dev schedule from build pod variable",
			args: args{
				buildValues: &BuildValues{
					BuildType:             "pullrequest",
					EnvironmentType:       "development",
					Project:               "example-project",
					Namespace:             "example-com-main",
					DefaultBackupSchedule: "M H(22-2) * * *",
				},
				lYAML: &lagoon.YAML{},
				mergedVariables: []lagoon.EnvironmentVariable{
					{Name: "LAGOON_FEATURE_FLAG_CUSTOM_BACKUP_CONFIG", Value: "enabled", Scope: "global"},
				},
			},
			vars: []helpers.EnvironmentVariable{
				{Name: "LAGOON_FEATURE_BACKUP_DEV_SCHEDULE", Value: "1,16,31,46 23 * * 0-5"},
			},
			want: &BuildValues{
				BuildType:             "pullrequest",
				EnvironmentType:       "development",
				Project:               "example-project",
				Namespace:             "example-com-main",
				DefaultBackupSchedule: "M H(22-2) * * *",
				Backup: BackupConfiguration{
					BackupSchedule: "1,16,31,46 23 * * 0-5",
					CheckSchedule:  "31 6 * * 1",
					PruneSchedule:  "31 4 * * 0",
					S3BucketName:   "baas-example-project",
					PruneRetention: PruneRetention{
						Hourly:  0,
						Daily:   7,
						Weekly:  6,
						Monthly: 1,
					},
				},
			},
		},
		{
			name: "test8 - production with lagoon yaml overrides",
			args: args{
				buildValues: &BuildValues{
					BuildType:             "branch",
					EnvironmentType:       "production",
					Project:               "example-project",
					Namespace:             "example-com-main",
					DefaultBackupSchedule: "M H(22-2) * * *",
				},
				lYAML: &lagoon.YAML{
					BackupRetention: lagoon.BackupRetention{
						Production: lagoon.Retention{
							Hourly:  helpers.IntPtr(10),
							Daily:   helpers.IntPtr(10),
							Weekly:  helpers.IntPtr(10),
							Monthly: helpers.IntPtr(10),
						},
					},
					BackupSchedule: lagoon.BackupSchedule{
						Production: "*/15 0-23 1-31 1-12 0-6",
					},
				},
				mergedVariables: []lagoon.EnvironmentVariable{},
			},
			want: &BuildValues{
				BuildType:             "branch",
				EnvironmentType:       "production",
				Project:               "example-project",
				Namespace:             "example-com-main",
				DefaultBackupSchedule: "M H(22-2) * * *",
				Backup: BackupConfiguration{
					BackupSchedule: "1,16,31,46 0-23 1-31 1-12 0-6",
					CheckSchedule:  "31 6 * * 1",
					PruneSchedule:  "31 4 * * 0",
					S3BucketName:   "baas-example-project",
					PruneRetention: PruneRetention{
						Hourly:  10,
						Daily:   10,
						Weekly:  10,
						Monthly: 10,
					},
				},
			},
		},
		{
			name: "test9 - custom backup configuration",
			args: args{
				buildValues: &BuildValues{
					BuildType:             "branch",
					EnvironmentType:       "development",
					Project:               "example-project",
					Namespace:             "example-com-main",
					DefaultBackupSchedule: "M H(22-2) * * *",
				},
				lYAML: &lagoon.YAML{},
				mergedVariables: []lagoon.EnvironmentVariable{
					{Name: "LAGOON_BAAS_CUSTOM_BACKUP_ACCESS_KEY", Value: "abcdefg", Scope: "build"},
					{Name: "LAGOON_BAAS_CUSTOM_BACKUP_SECRET_KEY", Value: "a1b2c3d4e5f6g7h8i9", Scope: "build"},
				},
			},
			want: &BuildValues{
				BuildType:             "branch",
				EnvironmentType:       "development",
				Project:               "example-project",
				Namespace:             "example-com-main",
				DefaultBackupSchedule: "M H(22-2) * * *",
				Backup: BackupConfiguration{
					S3SecretName: "lagoon-baas-custom-backup-credentials",
					S3BucketName: "baas-example-project",
					CustomLocation: CustomBackupRestoreLocation{
						BackupLocationAccessKey: "abcdefg",
						BackupLocationSecretKey: "a1b2c3d4e5f6g7h8i9",
					},
					BackupSchedule: "31 1 * * *",
					CheckSchedule:  "31 6 * * 1",
					PruneSchedule:  "31 4 * * 0",
					PruneRetention: PruneRetention{
						Hourly:  0,
						Daily:   7,
						Weekly:  6,
						Monthly: 1,
					},
				},
			},
		},
		{
			name: "test10 - custom backup configuration with endpoint and bucket",
			args: args{
				buildValues: &BuildValues{
					BuildType:             "branch",
					EnvironmentType:       "development",
					Project:               "example-project",
					Namespace:             "example-com-main",
					DefaultBackupSchedule: "M H(22-2) * * *",
				},
				lYAML: &lagoon.YAML{},
				mergedVariables: []lagoon.EnvironmentVariable{
					{Name: "LAGOON_BAAS_CUSTOM_BACKUP_ACCESS_KEY", Value: "abcdefg", Scope: "build"},
					{Name: "LAGOON_BAAS_CUSTOM_BACKUP_SECRET_KEY", Value: "a1b2c3d4e5f6g7h8i9", Scope: "build"},
					{Name: "LAGOON_BAAS_CUSTOM_BACKUP_ENDPOINT", Value: "https://minio.example.com", Scope: "build"},
					{Name: "LAGOON_BAAS_CUSTOM_BACKUP_BUCKET", Value: "my-bucket", Scope: "build"},
				},
			},
			want: &BuildValues{
				BuildType:             "branch",
				EnvironmentType:       "development",
				Project:               "example-project",
				Namespace:             "example-com-main",
				DefaultBackupSchedule: "M H(22-2) * * *",
				Backup: BackupConfiguration{
					S3SecretName: "lagoon-baas-custom-backup-credentials",
					S3Endpoint:   "https://minio.example.com",
					S3BucketName: "my-bucket",
					CustomLocation: CustomBackupRestoreLocation{
						BackupLocationAccessKey: "abcdefg",
						BackupLocationSecretKey: "a1b2c3d4e5f6g7h8i9",
					},
					BackupSchedule: "31 1 * * *",
					CheckSchedule:  "31 6 * * 1",
					PruneSchedule:  "31 4 * * 0",
					PruneRetention: PruneRetention{
						Hourly:  0,
						Daily:   7,
						Weekly:  6,
						Monthly: 1,
					},
				},
			},
		},
		{
			name: "test11 - custom restore configuration",
			args: args{
				buildValues: &BuildValues{
					BuildType:             "branch",
					EnvironmentType:       "development",
					Project:               "example-project",
					Namespace:             "example-com-main",
					DefaultBackupSchedule: "M H(22-2) * * *",
				},
				lYAML: &lagoon.YAML{},
				mergedVariables: []lagoon.EnvironmentVariable{
					{Name: "LAGOON_BAAS_CUSTOM_RESTORE_ACCESS_KEY", Value: "abcdefg", Scope: "build"},
					{Name: "LAGOON_BAAS_CUSTOM_RESTORE_SECRET_KEY", Value: "a1b2c3d4e5f6g7h8i9", Scope: "build"},
				},
			},
			want: &BuildValues{
				BuildType:             "branch",
				EnvironmentType:       "development",
				Project:               "example-project",
				Namespace:             "example-com-main",
				DefaultBackupSchedule: "M H(22-2) * * *",
				Backup: BackupConfiguration{
					CustomLocation: CustomBackupRestoreLocation{
						RestoreLocationAccessKey: "abcdefg",
						RestoreLocationSecretKey: "a1b2c3d4e5f6g7h8i9",
					},
					BackupSchedule: "31 1 * * *",
					CheckSchedule:  "31 6 * * 1",
					PruneSchedule:  "31 4 * * 0",
					S3BucketName:   "baas-example-project",
					PruneRetention: PruneRetention{
						Hourly:  0,
						Daily:   7,
						Weekly:  6,
						Monthly: 1,
					},
				},
			},
		},
		{
			name: "test12 - custom restore and backup configuration with endpoint and bucket",
			args: args{
				buildValues: &BuildValues{
					BuildType:             "branch",
					EnvironmentType:       "development",
					Project:               "example-project",
					Namespace:             "example-com-main",
					DefaultBackupSchedule: "M H(22-2) * * *",
				},
				lYAML: &lagoon.YAML{},
				mergedVariables: []lagoon.EnvironmentVariable{
					{Name: "LAGOON_BAAS_CUSTOM_BACKUP_ACCESS_KEY", Value: "abcdefg", Scope: "build"},
					{Name: "LAGOON_BAAS_CUSTOM_BACKUP_SECRET_KEY", Value: "a1b2c3d4e5f6g7h8i9", Scope: "build"},
					{Name: "LAGOON_BAAS_CUSTOM_BACKUP_ENDPOINT", Value: "https://minio.example.com", Scope: "build"},
					{Name: "LAGOON_BAAS_CUSTOM_BACKUP_BUCKET", Value: "my-bucket", Scope: "build"},
					{Name: "LAGOON_BAAS_CUSTOM_RESTORE_ACCESS_KEY", Value: "abcdefg", Scope: "build"},
					{Name: "LAGOON_BAAS_CUSTOM_RESTORE_SECRET_KEY", Value: "a1b2c3d4e5f6g7h8i9", Scope: "build"},
				},
			},
			want: &BuildValues{
				BuildType:             "branch",
				EnvironmentType:       "development",
				Project:               "example-project",
				Namespace:             "example-com-main",
				DefaultBackupSchedule: "M H(22-2) * * *",
				Backup: BackupConfiguration{
					S3SecretName: "lagoon-baas-custom-backup-credentials",
					S3Endpoint:   "https://minio.example.com",
					S3BucketName: "my-bucket",
					CustomLocation: CustomBackupRestoreLocation{
						BackupLocationAccessKey:  "abcdefg",
						BackupLocationSecretKey:  "a1b2c3d4e5f6g7h8i9",
						RestoreLocationAccessKey: "abcdefg",
						RestoreLocationSecretKey: "a1b2c3d4e5f6g7h8i9",
					},
					BackupSchedule: "31 1 * * *",
					CheckSchedule:  "31 6 * * 1",
					PruneSchedule:  "31 4 * * 0",
					PruneRetention: PruneRetention{
						Hourly:  0,
						Daily:   7,
						Weekly:  6,
						Monthly: 1,
					},
				},
			},
		},
		{
			name: "test13 - K8UP_WEEKLY_RANDOM_FEATURE_FLAG enabled",
			args: args{
				buildValues: &BuildValues{
					BuildType:             "branch",
					EnvironmentType:       "development",
					Project:               "example-project",
					Namespace:             "example-com-main",
					DefaultBackupSchedule: "M H(22-2) * * *",
				},
				lYAML:           &lagoon.YAML{},
				mergedVariables: []lagoon.EnvironmentVariable{},
			},
			vars: []helpers.EnvironmentVariable{
				{Name: "K8UP_WEEKLY_RANDOM_FEATURE_FLAG", Value: "enabled"},
			},
			want: &BuildValues{
				BuildType:             "branch",
				EnvironmentType:       "development",
				Project:               "example-project",
				Namespace:             "example-com-main",
				DefaultBackupSchedule: "M H(22-2) * * *",
				Backup: BackupConfiguration{
					BackupSchedule: "31 1 * * *",
					CheckSchedule:  "@weekly-random",
					PruneSchedule:  "@weekly-random",
					S3BucketName:   "baas-example-project",
					PruneRetention: PruneRetention{
						Hourly:  0,
						Daily:   7,
						Weekly:  6,
						Monthly: 1,
					},
				},
			},
		},
		{
			name: "test14 - K8UP_WEEKLY_RANDOM_FEATURE_FLAG set but not enabled/valid",
			args: args{
				buildValues: &BuildValues{
					BuildType:             "branch",
					EnvironmentType:       "development",
					Project:               "example-project",
					Namespace:             "example-com-main",
					DefaultBackupSchedule: "M H(22-2) * * *",
				},
				lYAML:           &lagoon.YAML{},
				mergedVariables: []lagoon.EnvironmentVariable{},
			},
			vars: []helpers.EnvironmentVariable{
				{Name: "K8UP_WEEKLY_RANDOM_FEATURE_FLAG", Value: "jkhk"},
			},
			want: &BuildValues{
				BuildType:             "branch",
				EnvironmentType:       "development",
				Project:               "example-project",
				Namespace:             "example-com-main",
				DefaultBackupSchedule: "M H(22-2) * * *",
				Backup: BackupConfiguration{
					BackupSchedule: "31 1 * * *",
					CheckSchedule:  "31 6 * * 1",
					PruneSchedule:  "31 4 * * 0",
					S3BucketName:   "baas-example-project",
					PruneRetention: PruneRetention{
						Hourly:  0,
						Daily:   7,
						Weekly:  6,
						Monthly: 1,
					},
				},
			},
		},
		{
			name: "test15 - baas bucket name",
			args: args{
				buildValues: &BuildValues{
					BuildType:             "branch",
					EnvironmentType:       "development",
					Project:               "example-project",
					Namespace:             "example-com-main",
					DefaultBackupSchedule: "M H(22-2) * * *",
				},
				lYAML: &lagoon.YAML{},
				mergedVariables: []lagoon.EnvironmentVariable{
					{
						Name:  "LAGOON_BAAS_BUCKET_NAME",
						Value: "mybucket",
						Scope: "global",
					},
				},
			},
			want: &BuildValues{
				BuildType:             "branch",
				EnvironmentType:       "development",
				Project:               "example-project",
				Namespace:             "example-com-main",
				DefaultBackupSchedule: "M H(22-2) * * *",
				Backup: BackupConfiguration{
					BackupSchedule: "31 1 * * *",
					CheckSchedule:  "31 6 * * 1",
					PruneSchedule:  "31 4 * * 0",
					S3BucketName:   "mybucket",
					PruneRetention: PruneRetention{
						Hourly:  0,
						Daily:   7,
						Weekly:  6,
						Monthly: 1,
					},
				},
			},
		},
		{
			name: "test16 - baas bucket name with shared",
			args: args{
				buildValues: &BuildValues{
					BuildType:             "branch",
					EnvironmentType:       "development",
					Project:               "example-project",
					Namespace:             "example-com-main",
					DefaultBackupSchedule: "M H(22-2) * * *",
				},
				lYAML: &lagoon.YAML{},
				mergedVariables: []lagoon.EnvironmentVariable{
					{
						Name:  "LAGOON_BAAS_BUCKET_NAME",
						Value: "mybucket",
						Scope: "global",
					},
					{
						Name:  "LAGOON_SYSTEM_PROJECT_SHARED_BUCKET",
						Value: "baas-cluster-bucket",
						Scope: "internal_system",
					},
				},
			},
			want: &BuildValues{
				BuildType:             "branch",
				EnvironmentType:       "development",
				Project:               "example-project",
				Namespace:             "example-com-main",
				DefaultBackupSchedule: "M H(22-2) * * *",
				Backup: BackupConfiguration{
					BackupSchedule: "31 1 * * *",
					CheckSchedule:  "31 6 * * 1",
					PruneSchedule:  "31 4 * * 0",
					S3BucketName:   "mybucket",
					PruneRetention: PruneRetention{
						Hourly:  0,
						Daily:   7,
						Weekly:  6,
						Monthly: 1,
					},
				},
			},
		},
		{
			name: "test17 - shared baas bucket name",
			args: args{
				buildValues: &BuildValues{
					BuildType:             "branch",
					EnvironmentType:       "development",
					Project:               "example-project",
					Namespace:             "example-com-main",
					DefaultBackupSchedule: "M H(22-2) * * *",
				},
				lYAML: &lagoon.YAML{},
				mergedVariables: []lagoon.EnvironmentVariable{
					{
						Name:  "LAGOON_SYSTEM_PROJECT_SHARED_BUCKET",
						Value: "baas-cluster-bucket",
						Scope: "internal_system",
					},
				},
			},
			want: &BuildValues{
				BuildType:             "branch",
				EnvironmentType:       "development",
				Project:               "example-project",
				Namespace:             "example-com-main",
				DefaultBackupSchedule: "M H(22-2) * * *",
				Backup: BackupConfiguration{
					BackupSchedule: "31 1 * * *",
					CheckSchedule:  "31 6 * * 1",
					PruneSchedule:  "31 4 * * 0",
					S3BucketName:   "baas-cluster-bucket/baas-example-project",
					PruneRetention: PruneRetention{
						Hourly:  0,
						Daily:   7,
						Weekly:  6,
						Monthly: 1,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, envVar := range tt.vars {
				err := os.Setenv(envVar.Name, envVar.Value)
				if err != nil {
					t.Errorf("%v", err)
				}
			}
			if err := generateBackupValues(tt.args.buildValues, tt.args.lYAML, tt.args.mergedVariables, tt.args.debug); (err != nil) != tt.wantErr {
				t.Errorf("generateBackupValues() error = %v, wantErr %v", err, tt.wantErr)
			}
			lValues, _ := json.Marshal(tt.args.buildValues)
			wValues, _ := json.Marshal(tt.want)
			if !reflect.DeepEqual(string(lValues), string(wValues)) {
				t.Errorf("GenerateBackupSchedule() = %v, want %v", string(lValues), string(wValues))
			}
			t.Cleanup(func() {
				helpers.UnsetEnvVars(tt.vars)
			})
		})
	}
}
