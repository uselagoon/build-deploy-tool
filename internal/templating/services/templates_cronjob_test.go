package services

import (
	"os"
	"reflect"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
)

func TestGenerateCronjobTemplate(t *testing.T) {
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
			name: "test1 - cli",
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
					ImagePullSecrets: []generator.ImagePullSecrets{
						{
							Name: "lagoon-internal-registry-secret",
						},
					},
					GitSha:       "0",
					ConfigMapSha: "32bf1359ac92178c8909f0ef938257b477708aa0d78a5a15ad7c2d7919adf273",
					Services: []generator.ServiceValues{
						{
							Name:             "myservice",
							OverrideName:     "myservice",
							Type:             "cli",
							DBaaSEnvironment: "production",
							ImageName:        "harbor.example.com/example-project/environment-name/myservice@latest",
							NativeCronjobs: []lagoon.Cronjob{
								{
									Name:     "cronjob-myservice-my-cronjobbb",
									Service:  "myservice",
									Command:  "sleep 300",
									Schedule: "5 2 * * *",
								},
								{
									Name:     "cronjob-myservice-my-other-cronjobbb",
									Service:  "myservice",
									Command:  "env",
									Schedule: "25 6 * * *",
								},
							},
						},
						// {
						// 	Name:                 "myservice-persist",
						// 	OverrideName:         "myservice-persist",
						// 	Type:                 "cli-persistent",
						// 	DBaaSEnvironment:     "production",
						// 	ImageName:            "harbor.example.com/example-project/environment-name/myservice-persistent@latest",
						// 	PersistentVolumePath: "/storage/data",
						// 	PersistentVolumeName: "nginx-php",
						// },
					},
				},
			},
			want: "test-resources/cronjob/result-cli-1.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateCronjobTemplate(tt.args.buildValues)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateCronjobTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			r1, err := os.ReadFile(tt.want)
			if err != nil {
				t.Errorf("couldn't read file %v: %v", tt.want, err)
			}
			if !reflect.DeepEqual(string(got), string(r1)) {
				t.Errorf("GenerateCronjobTemplate() = \n%v", diff.LineDiff(string(r1), string(got)))
			}
		})
	}
}
