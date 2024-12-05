package backups

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/andreyvit/diff"
	"github.com/uselagoon/build-deploy-tool/internal/dbaasclient"
	"github.com/uselagoon/build-deploy-tool/internal/generator"
)

func TestGenerateBackupPodConfig(t *testing.T) {
	type args struct {
		lValues generator.BuildValues
	}
	tests := []struct {
		name        string
		description string
		args        args
		want        string
		wantErr     bool
		wantEmpty   bool
	}{
		{
			name:        "test-k8up-v1-rootless",
			description: "this will generate a podconfig if the environment is configured for rootless workloads",
			args: args{
				lValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment",
					EnvironmentType: "production",
					Namespace:       "myexample-project-environment",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "environment",
					BackupsEnabled:  true,
					Backup: generator.BackupConfiguration{
						K8upVersion: "v2",
					},
					FeatureFlags: map[string]bool{
						"rootlessworkloads": true,
					},
					PodSecurityContext: generator.PodSecurityContext{
						RunAsGroup: 0,
						RunAsUser:  10000,
						FsGroup:    10001,
					},
				},
			},
			want: "test-resources/test-k8up-v1-rootless.yaml",
		},
		{
			name:        "test-k8up-v1-rootless-onrootmismatch",
			description: "this will generate a podconfig if the environment is configured for rootless workloads",
			args: args{
				lValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment",
					EnvironmentType: "production",
					Namespace:       "myexample-project-environment",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "environment",
					BackupsEnabled:  true,
					Backup: generator.BackupConfiguration{
						K8upVersion: "v2",
					},
					FeatureFlags: map[string]bool{
						"rootlessworkloads": true,
					},
					PodSecurityContext: generator.PodSecurityContext{
						RunAsGroup:     0,
						RunAsUser:      10000,
						FsGroup:        10001,
						OnRootMismatch: true,
					},
				},
			},
			want: "test-resources/test-k8up-v1-rootless-onrootmismatch.yaml",
		},
		{
			name:        "test-k8up-v1-root",
			description: "this will not generate a podconfig if the environment is not configured for rootless workloads",
			args: args{
				lValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment",
					EnvironmentType: "production",
					Namespace:       "myexample-project-environment",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "environment",
					BackupsEnabled:  true,
					Backup: generator.BackupConfiguration{
						K8upVersion: "v2",
					},
				},
			},
			wantEmpty: true,
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
			got, err := GenerateBackupPodConfig(tt.args.lValues)
			if err != nil {
				t.Errorf("couldn't generate template %v: %v", tt.want, err)
			}
			if tt.wantEmpty && len(got) > 0 {
				t.Errorf("wanted empty, but got data:\n%v", string(got))
			}
			if !tt.wantEmpty {
				r1, err := os.ReadFile(tt.want)
				if err != nil {
					t.Errorf("couldn't read file %v: %v", tt.want, err)
				}
				if !reflect.DeepEqual(string(got), string(r1)) {
					t.Errorf("GenerateBackupPodConfig() = \n%v", diff.LineDiff(string(r1), string(got)))
				}
			}
		})
	}
}
