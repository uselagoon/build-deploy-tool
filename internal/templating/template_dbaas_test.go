package templating

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/andreyvit/diff"
	"github.com/uselagoon/build-deploy-tool/internal/dbaasclient"
	"github.com/uselagoon/build-deploy-tool/internal/generator"
)

func TestGenerateDBaaSTemplate(t *testing.T) {
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
			name: "test1 - mariadb",
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
							Name:             "mariadb",
							OverrideName:     "mariadb",
							Type:             "mariadb-dbaas",
							DBaaSEnvironment: "development",
						},
					},
				},
			},
			want: "test-resources/dbaas/result-mariadb-1.yaml",
		},
		{
			name: "test2 - mongodb",
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
							Name:             "mongodb",
							OverrideName:     "mongodb",
							Type:             "mongodb-dbaas",
							DBaaSEnvironment: "development",
						},
					},
				},
			},
			want: "test-resources/dbaas/result-mongodb-1.yaml",
		},
		{
			name: "test3 - postgres",
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
							Name:             "postgres",
							OverrideName:     "postgres",
							Type:             "postgres-dbaas",
							DBaaSEnvironment: "development",
						},
					},
				},
			},
			want: "test-resources/dbaas/result-postgres-1.yaml",
		},
		{
			name: "test4 - mongo",
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
							Name:             "mongo",
							OverrideName:     "mongo",
							Type:             "mongodb-dbaas",
							DBaaSEnvironment: "development",
						},
					},
				},
			},
			want: "test-resources/dbaas/result-mongodb-2.yaml",
		},
		{
			name: "test5 - multi mariadb",
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
							Name:             "mariadb",
							OverrideName:     "mariadb",
							Type:             "mariadb-dbaas",
							DBaaSEnvironment: "development",
						},
						{
							Name:             "mariadb2",
							OverrideName:     "mariadb2",
							Type:             "mariadb-dbaas",
							DBaaSEnvironment: "development",
						},
					},
				},
			},
			want: "test-resources/dbaas/result-mariadb-2.yaml",
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
			got, err := GenerateDBaaSTemplate(tt.args.lValues)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateDBaaSTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			r1, err := os.ReadFile(tt.want)
			if err != nil {
				t.Errorf("couldn't read file %v: %v", tt.want, err)
			}
			templateYAML, err := TemplateConsumers(got)
			if err != nil {
				t.Errorf("couldn't generate template: %v", err)
			}
			if !reflect.DeepEqual(string(templateYAML), string(r1)) {
				t.Errorf("GenerateDBaaSTemplate() = \n%v", diff.LineDiff(string(r1), string(templateYAML)))
			}
		})
	}
}
