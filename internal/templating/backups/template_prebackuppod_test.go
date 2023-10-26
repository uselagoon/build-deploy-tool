package backups

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/uselagoon/build-deploy-tool/internal/dbaasclient"
	"github.com/uselagoon/build-deploy-tool/internal/generator"
)

func TestGeneratePreBackupPod(t *testing.T) {
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
			name: "test1 - k8up/v1alpha1",
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
					Services: []generator.ServiceValues{
						{
							Name:             "mariadb-database",
							OverrideName:     "mariadb-database",
							Type:             "mariadb-dbaas",
							DBaaSEnvironment: "development",
							DBaasReadReplica: true,
						},
					},
					Backup: generator.BackupConfiguration{
						K8upVersion: "v1",
					},
				},
			},
			want: "test-resources/result-prebackuppod1.yaml",
		},
		{
			name: "test2 - k8up/v1alpha1",
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
					Services: []generator.ServiceValues{
						{
							Name:             "postgres-database",
							OverrideName:     "postgres-database",
							Type:             "postgres-dbaas",
							DBaaSEnvironment: "development",
							DBaasReadReplica: true,
						},
					},
					Backup: generator.BackupConfiguration{
						K8upVersion: "v1",
					},
				},
			},
			want: "test-resources/result-prebackuppod2.yaml",
		},
		{
			name: "test3 - k8up/v1alpha1",
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
					Services: []generator.ServiceValues{
						{
							Name:             "mongodb-database",
							OverrideName:     "mongodb-database",
							Type:             "mongodb-dbaas",
							DBaaSEnvironment: "development",
							DBaasReadReplica: true,
						},
					},
					Backup: generator.BackupConfiguration{
						K8upVersion: "v1",
					},
				},
			},
			want: "test-resources/result-prebackuppod3.yaml",
		},
		{
			name: "test4 - k8up/v1",
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
					Services: []generator.ServiceValues{
						{
							Name:             "mariadb-database",
							OverrideName:     "mariadb-database",
							Type:             "mariadb-dbaas",
							DBaaSEnvironment: "development",
							DBaasReadReplica: true,
						},
					},
					Backup: generator.BackupConfiguration{
						K8upVersion: "v2",
					},
				},
			},
			want: "test-resources/result-prebackuppod4.yaml",
		},
		{
			name: "test5 - multiple k8up/v1",
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
					Services: []generator.ServiceValues{
						{
							Name:             "mariadb-database",
							OverrideName:     "mariadb-database",
							Type:             "mariadb-dbaas",
							DBaaSEnvironment: "development",
							DBaasReadReplica: true,
						},
						{
							Name:             "mariadb",
							OverrideName:     "mariadb",
							Type:             "mariadb-dbaas",
							DBaaSEnvironment: "development",
							DBaasReadReplica: true,
						},
					},
					Backup: generator.BackupConfiguration{
						K8upVersion: "v2",
					},
				},
			},
			want: "test-resources/result-prebackuppod5.yaml",
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
			got, err := GeneratePreBackupPod(tt.args.lValues)
			if err != nil {
				t.Errorf("couldn't generate template %v: %v", tt.want, err)
			}
			r1, err := os.ReadFile(tt.want)
			if err != nil {
				t.Errorf("couldn't read file %v: %v", tt.want, err)
			}
			if !reflect.DeepEqual(string(got), string(r1)) {
				t.Errorf("GeneratePreBackupPod() = %v, want %v", string(got), string(r1))
			}
		})
	}
}
