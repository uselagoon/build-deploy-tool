package services

import (
	"os"
	"reflect"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/uselagoon/build-deploy-tool/internal/generator"
)

func TestGeneratePVCTemplate(t *testing.T) {
	type args struct {
		buildValues generator.BuildValues
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test1 - basic",
			args: args{
				buildValues: generator.BuildValues{
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
							Name:             "myservice",
							OverrideName:     "myservice",
							Type:             "basic",
							DBaaSEnvironment: "development",
						},
						{
							Name:             "myservice-po",
							OverrideName:     "myservice-po",
							Type:             "basic",
							DBaaSEnvironment: "development",
							ServicePort:      8080,
						},
						{
							Name:                 "myservice-persist",
							OverrideName:         "myservice-persist",
							Type:                 "basic-persistent",
							DBaaSEnvironment:     "development",
							PersistentVolumeSize: "5Gi",
						},
						{
							Name:                 "myservice-persist-po",
							OverrideName:         "myservice-persist-po",
							Type:                 "basic-persistent",
							DBaaSEnvironment:     "development",
							ServicePort:          8080,
							PersistentVolumeSize: "5Gi",
						},
						{
							Name:                 "myservice-persist-po",
							OverrideName:         "myservice-persist-po",
							Type:                 "basic-persistent",
							DBaaSEnvironment:     "development",
							ServicePort:          8080,
							PersistentVolumeSize: "100Gi",
						},
					},
				},
			},
			want: "test-resources/pvc/result-basic-1.yaml",
		},
		{
			name: "test2 - cli",
			args: args{
				buildValues: generator.BuildValues{
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
							Name:             "myservice",
							OverrideName:     "myservice",
							Type:             "cli",
							DBaaSEnvironment: "development",
						},
						{
							Name:             "myservice-persist",
							OverrideName:     "myservice-persist",
							Type:             "cli-persistent",
							DBaaSEnvironment: "development",
						},
					},
				},
			},
			want: "test-resources/pvc/result-cli-1.yaml",
		},
		{
			name: "test3 - elasticsearch",
			args: args{
				buildValues: generator.BuildValues{
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
							Name:                 "myservice",
							OverrideName:         "myservice",
							Type:                 "elasticsearch",
							DBaaSEnvironment:     "development",
							PersistentVolumeSize: "5Gi",
						},
						{
							Name:                 "myservice-size",
							OverrideName:         "myservice-size",
							Type:                 "elasticsearch",
							DBaaSEnvironment:     "development",
							PersistentVolumeSize: "100Gi",
						},
					},
				},
			},
			want: "test-resources/pvc/result-elasticsearch-1.yaml",
		},
		{
			name: "test4 - opensearch",
			args: args{
				buildValues: generator.BuildValues{
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
							Name:                 "myservice",
							OverrideName:         "myservice",
							Type:                 "opensearch",
							DBaaSEnvironment:     "development",
							PersistentVolumeSize: "5Gi",
						},
						{
							Name:                 "myservice-size",
							OverrideName:         "myservice-size",
							Type:                 "opensearch",
							DBaaSEnvironment:     "development",
							PersistentVolumeSize: "100Gi",
						},
					},
				},
			},
			want: "test-resources/pvc/result-opensearch-1.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GeneratePVCTemplate(tt.args.buildValues)
			if (err != nil) != tt.wantErr {
				t.Errorf("GeneratePVCTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			r1, err := os.ReadFile(tt.want)
			if err != nil {
				t.Errorf("couldn't read file %v: %v", tt.want, err)
			}
			if !reflect.DeepEqual(string(got), string(r1)) {
				t.Errorf("GeneratePVCTemplate() = \n%v", diff.LineDiff(string(r1), string(got)))
			}
		})
	}
}
