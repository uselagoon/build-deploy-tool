package backups

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/uselagoon/build-deploy-tool/internal/dbaasclient"
	"github.com/uselagoon/build-deploy-tool/internal/generator"
)

func TestGenerateBackupSchedule(t *testing.T) {
	type args struct {
		lValues generator.BuildValues
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				lValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-with-really-really-reall-3fdb",
					EnvironmentType: "production",
					Namespace:       "myexample-project-environment-with-really-really-reall-3fdb",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "environment-with-really-really-reall-3fdb",
					Backup: generator.BackupConfiguration{
						S3Endpoint:     "https://minio.endpoint",
						S3BucketName:   "my-bucket",
						S3SecretName:   "my-s3-secret",
						BackupSchedule: "50 5 * * 6",
						CheckSchedule:  "50 5 * * 6",
						PruneSchedule:  "50 5 * * 6",
						PruneRetention: generator.PruneRetention{
							Hourly:  0,
							Daily:   7,
							Weekly:  6,
							Monthly: 1,
						},
					},
				},
			},
			want: "test-resources/result-schedule1.yaml",
		},
		{
			name: "test2",
			args: args{
				lValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-with-really-really-reall-3fdb",
					EnvironmentType: "production",
					Namespace:       "myexample-project-environment-with-really-really-reall-3fdb",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "environment-with-really-really-reall-3fdb",
					Backup: generator.BackupConfiguration{
						BackupSchedule: "50 5 * * 6",
						CheckSchedule:  "50 5 * * 6",
						PruneSchedule:  "50 5 * * 6",
						PruneRetention: generator.PruneRetention{
							Hourly:  0,
							Daily:   7,
							Weekly:  6,
							Monthly: 1,
						},
					},
				},
			},
			want: "test-resources/result-schedule2.yaml",
		},
		{
			name: "test3",
			args: args{
				lValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-with-really-really-reall-3fdb",
					EnvironmentType: "production",
					Namespace:       "myexample-project-environment-with-really-really-reall-3fdb",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "environment-with-really-really-reall-3fdb",
					Backup: generator.BackupConfiguration{
						S3Endpoint:     "https://minio.endpoint",
						S3BucketName:   "my-bucket",
						S3SecretName:   "my-s3-secret",
						BackupSchedule: "50 5 * * 6",
						CheckSchedule:  "50 5 * * 6",
						PruneSchedule:  "50 5 * * 6",
						PruneRetention: generator.PruneRetention{
							Hourly:  0,
							Daily:   7,
							Weekly:  6,
							Monthly: 1,
						},
						CustomLocation: generator.CustomBackupRestoreLocation{
							BackupLocationAccessKey: "abc123",
							BackupLocationSecretKey: "abcdefghijklmnopqrstuvwxyz",
						},
					},
				},
			},
			want: "test-resources/result-schedule3.yaml",
		},
		{
			name: "test4",
			args: args{
				lValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-with-really-really-reall-3fdb",
					EnvironmentType: "production",
					Namespace:       "myexample-project-environment-with-really-really-reall-3fdb",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "environment-with-really-really-reall-3fdb",
					Backup: generator.BackupConfiguration{
						S3Endpoint:     "https://minio.endpoint",
						S3BucketName:   "my-bucket",
						S3SecretName:   "my-s3-secret",
						BackupSchedule: "50 5 * * 6",
						CheckSchedule:  "50 5 * * 6",
						PruneSchedule:  "50 5 * * 6",
						PruneRetention: generator.PruneRetention{
							Hourly:  0,
							Daily:   7,
							Weekly:  6,
							Monthly: 1,
						},
						CustomLocation: generator.CustomBackupRestoreLocation{
							RestoreLocationAccessKey: "abc123",
							RestoreLocationSecretKey: "abcdefghijklmnopqrstuvwxyz",
						},
					},
				},
			},
			want: "test-resources/result-schedule4.yaml",
		},
		{
			name: "test5",
			args: args{
				lValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-with-really-really-reall-3fdb",
					EnvironmentType: "production",
					Namespace:       "myexample-project-environment-with-really-really-reall-3fdb",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "environment-with-really-really-reall-3fdb",
					Backup: generator.BackupConfiguration{
						S3Endpoint:     "https://minio.endpoint",
						S3BucketName:   "my-bucket",
						S3SecretName:   "my-s3-secret",
						BackupSchedule: "50 5 * * 6",
						CheckSchedule:  "50 5 * * 6",
						PruneSchedule:  "50 5 * * 6",
						PruneRetention: generator.PruneRetention{
							Hourly:  0,
							Daily:   7,
							Weekly:  6,
							Monthly: 1,
						},
						CustomLocation: generator.CustomBackupRestoreLocation{
							BackupLocationAccessKey:  "abc123",
							BackupLocationSecretKey:  "abcdefghijklmnopqrstuvwxyz",
							RestoreLocationAccessKey: "abc123",
							RestoreLocationSecretKey: "abcdefghijklmnopqrstuvwxyz",
						},
					},
				},
			},
			want: "test-resources/result-schedule5.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// add dbaasclient overrides for tests
			tt.args.lValues.DBaaSClient = dbaasclient.NewClient(dbaasclient.Client{
				RetryMax:     5,
				RetryWaitMin: time.Duration(10) * time.Millisecond,
				RetryWaitMax: time.Duration(50) * time.Millisecond,
			})
			got, err := GenerateBackupSchedule(tt.args.lValues)
			if err != nil {
				t.Errorf("couldn't generate template %v: %v", tt.want, err)
			}
			r1, err := os.ReadFile(tt.want)
			if err != nil {
				t.Errorf("couldn't read file %v: %v", tt.want, err)
			}
			if !reflect.DeepEqual(string(got), string(r1)) {
				t.Errorf("GenerateBackupSchedule() = %v, want %v", string(got), string(r1))
			}
		})
	}
}
