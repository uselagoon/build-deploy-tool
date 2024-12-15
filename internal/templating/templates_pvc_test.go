package templating

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
							Name:                 "myservice-persist",
							OverrideName:         "myservice-persist",
							Type:                 "basic-persistent",
							DBaaSEnvironment:     "development",
							PersistentVolumeName: "myservice-persist",
							PersistentVolumeSize: "5Gi",
							CreateDefaultVolume:  true,
						},
						{
							Name:                 "myservice-persist-po",
							OverrideName:         "myservice-persist-po",
							Type:                 "basic-persistent",
							DBaaSEnvironment:     "development",
							ServicePort:          8080,
							PersistentVolumeName: "myservice-persist-po",
							PersistentVolumeSize: "5Gi",
							CreateDefaultVolume:  true,
						},
						{
							Name:                 "myservice-persist-po",
							OverrideName:         "myservice-persist-po",
							Type:                 "basic-persistent",
							DBaaSEnvironment:     "development",
							PersistentVolumeName: "myservice-persist-po",
							ServicePort:          8080,
							PersistentVolumeSize: "100Gi",
							CreateDefaultVolume:  true,
						},
					},
				},
			},
			want: "test-resources/pvc/result-basic-1.yaml",
		},
		{
			name: "test2 - cli (no volumes)",
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
			want: "test-resources/pvc/result-cli-1.yaml",
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
							PersistentVolumeName: "myservice",
							PersistentVolumeSize: "5Gi",
							CreateDefaultVolume:  true,
						},
						{
							Name:                 "myservice-size",
							OverrideName:         "myservice-size",
							Type:                 "elasticsearch",
							DBaaSEnvironment:     "development",
							PersistentVolumeName: "myservice-size",
							PersistentVolumeSize: "100Gi",
							CreateDefaultVolume:  true,
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
							PersistentVolumeName: "myservice",
							PersistentVolumeSize: "5Gi",
							CreateDefaultVolume:  true,
						},
						{
							Name:                 "myservice-size",
							OverrideName:         "myservice-size",
							Type:                 "opensearch",
							DBaaSEnvironment:     "development",
							PersistentVolumeName: "myservice-size",
							PersistentVolumeSize: "100Gi",
							CreateDefaultVolume:  true,
						},
					},
				},
			},
			want: "test-resources/pvc/result-opensearch-1.yaml",
		},
		{
			name: "test5 - postgres-single",
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
							Type:                 "postgres-single",
							DBaaSEnvironment:     "development",
							PersistentVolumeName: "myservice",
							PersistentVolumeSize: "5Gi",
							CreateDefaultVolume:  true,
						},
					},
				},
			},
			want: "test-resources/pvc/result-postgres-single-1.yaml",
		},
		{
			name: "test7 - basic rwx2rwo",
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
					RWX2RWO:         true,
					Services: []generator.ServiceValues{
						{
							Name:                 "myservice-persist",
							OverrideName:         "myservice-persist",
							Type:                 "basic-persistent",
							DBaaSEnvironment:     "development",
							PersistentVolumeName: "myservice-persist",
							PersistentVolumeSize: "5Gi",
							CreateDefaultVolume:  true,
						},
					},
				},
			},
			want: "test-resources/pvc/result-basic-3.yaml",
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
					RWX2RWO:         true,
					Services: []generator.ServiceValues{
						{
							Name:                 "myservice",
							OverrideName:         "myservice",
							Type:                 "basic-single",
							DBaaSEnvironment:     "development",
							PersistentVolumeName: "myservice",
							PersistentVolumeSize: "5Gi",
							CreateDefaultVolume:  true,
						},
					},
				},
			},
			want: "test-resources/pvc/result-basic-4.yaml",
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
			var result []byte
			for _, d := range got {
				templateBytes, err := TemplatePVC(d)
				if err != nil {
					t.Errorf("couldn't generate template  %v", err)
				}
				result = append(result, templateBytes[:]...)
			}
			if !reflect.DeepEqual(string(result), string(r1)) {
				t.Errorf("GeneratePVCTemplate() = \n%v", diff.LineDiff(string(r1), string(result)))
			}
		})
	}
}
