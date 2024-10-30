package services

import (
	"os"
	"reflect"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/compose-spec/compose-go/types"
	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"sigs.k8s.io/yaml"
)

func TestGenerateServiceTemplate(t *testing.T) {
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
					Environment:     "environment-name",
					EnvironmentType: "production",
					Namespace:       "myexample-project-environment-name",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "environment-name",
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
							Name:             "myservice-persist",
							OverrideName:     "myservice-persist",
							Type:             "basic-persistent",
							DBaaSEnvironment: "development",
						},
						{
							Name:             "myservice-persist-po",
							OverrideName:     "myservice-persist-po",
							Type:             "basic-persistent",
							DBaaSEnvironment: "development",
							ServicePort:      8080,
						},
					},
				},
			},
			want: "test-resources/service/result-basic-1.yaml",
		},
		{
			name: "test2 - cli",
			args: args{
				buildValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-name",
					EnvironmentType: "production",
					Namespace:       "myexample-project-environment-name",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "environment-name",
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
			want: "test-resources/service/result-cli-1.yaml",
		},
		{
			name: "test3 - elasticsearch",
			args: args{
				buildValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-name",
					EnvironmentType: "production",
					Namespace:       "myexample-project-environment-name",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "environment-name",
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
			want: "test-resources/service/result-elasticsearch-1.yaml",
		},
		{
			name: "test4 - opensearch",
			args: args{
				buildValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-name",
					EnvironmentType: "production",
					Namespace:       "myexample-project-environment-name",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "environment-name",
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
			want: "test-resources/service/result-opensearch-1.yaml",
		},
		{
			name: "test5 - basic compose ports",
			args: args{
				buildValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-name",
					EnvironmentType: "production",
					Namespace:       "myexample-project-environment-name",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "environment-name",
					Services: []generator.ServiceValues{
						{
							Name:             "myservice-po",
							OverrideName:     "myservice-po",
							Type:             "basic",
							DBaaSEnvironment: "development",
							AdditionalServicePorts: []generator.AdditionalServicePort{
								{
									ServicePort: types.ServicePortConfig{
										Target:   8191,
										Protocol: "tcp",
									},
									ServiceName: "myservice-po-8191",
								},
							},
						},
						{
							Name:             "myservice-persist-po",
							OverrideName:     "myservice-persist-po",
							Type:             "basic-persistent",
							DBaaSEnvironment: "development",
							AdditionalServicePorts: []generator.AdditionalServicePort{
								{
									ServicePort: types.ServicePortConfig{
										Target:   8191,
										Protocol: "tcp",
									},
									ServiceName: "myservice-persist-po-8191",
								},
								{
									ServicePort: types.ServicePortConfig{
										Target:   8192,
										Protocol: "tcp",
									},
									ServiceName: "myservice-persist-po-8192",
								},
							},
						},
					},
				},
			},
			want: "test-resources/service/result-basic-2.yaml",
		},
		{
			name: "test6 - nginx-php",
			args: args{
				buildValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-name",
					EnvironmentType: "production",
					Namespace:       "myexample-project-environment-name",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "environment-name",
					Services: []generator.ServiceValues{
						{
							Name:             "myservice",
							OverrideName:     "myservice",
							Type:             "nginx-php",
							DBaaSEnvironment: "development",
							LinkedService: &generator.ServiceValues{
								Name:             "myservice-php",
								OverrideName:     "myservice",
								Type:             "nginx-php",
								DBaaSEnvironment: "production",
							},
						},
					},
				},
			},
			want: "test-resources/service/result-nginx-php-1.yaml",
		},
		{
			name: "test-basic-single",
			args: args{
				buildValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-name",
					EnvironmentType: "production",
					Namespace:       "myexample-project-environment-name",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "environment-name",
					Services: []generator.ServiceValues{
						{
							Name:             "myservice",
							OverrideName:     "myservice",
							Type:             "basic-single",
							DBaaSEnvironment: "development",
						},
					},
				},
			},
			want: "test-resources/service/result-basic-3.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateServiceTemplate(tt.args.buildValues)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateServiceTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			r1, err := os.ReadFile(tt.want)
			if err != nil {
				t.Errorf("couldn't read file %v: %v", tt.want, err)
			}
			separator := []byte("---\n")
			var result []byte
			for _, d := range got {
				sBytes, err := yaml.Marshal(d)
				if err != nil {
					t.Errorf("couldn't generate template  %v", err)
				}
				restoreResult := append(separator[:], sBytes[:]...)
				result = append(result, restoreResult[:]...)
			}
			if !reflect.DeepEqual(string(result), string(r1)) {
				t.Errorf("GenerateServiceTemplate() = \n%v", diff.LineDiff(string(r1), string(result)))
			}
		})
	}
}
