package templating

import (
	"os"
	"reflect"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
)

func TestGenerateNetworkPolicy(t *testing.T) {
	type args struct {
		buildValues generator.BuildValues
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{{
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
				GitSHA:          "0",
				ConfigMapSha:    "32bf1359ac92178c8909f0ef938257b477708aa0d78a5a15ad7c2d7919adf273",
				ImageReferences: map[string]string{
					"myservice":                "harbor.example.com/example-project/environment-name/myservice@latest",
					"myservice-po":             "harbor.example.com/example-project/environment-name/myservice-po@latest",
					"myservice-persist":        "harbor.example.com/example-project/environment-name/myservice-persist@latest",
					"myservice-persist-po":     "harbor.example.com/example-project/environment-name/myservice-persist-po@latest",
					"myservice-persist-posize": "harbor.example.com/example-project/environment-name/myservice-persist-posize@latest",
				},
				Services: []generator.ServiceValues{
					{
						Name:             "myservice",
						OverrideName:     "myservice",
						Type:             "basic",
						DBaaSEnvironment: "production",
						InPodCronjobs: []lagoon.Cronjob{
							{
								Name:     "cron - inpod",
								Schedule: "M/5 * * * *",
								Command:  "drush cron",
								Service:  "basic",
							},
							{
								Name:     "cron2 - inpod",
								Schedule: "M/15 * * * *",
								Command:  "other cronjob",
								Service:  "basic",
							},
						},
						NativeCronjobs: []lagoon.Cronjob{
							{
								Name:     "cron3 - native",
								Schedule: "35 * * * *",
								Command:  "drush cron",
								Service:  "basic",
							},
						},
					},
					{
						Name:             "myservice-po",
						OverrideName:     "myservice-po",
						Type:             "basic",
						DBaaSEnvironment: "production",
						ServicePort:      8080, // template should have port changed to 8080 from 3000
						UseSpotInstances: true, // template should have spot instance label and toleration/selector/affinity
						Replicas:         2,
					},
					{
						Name:                 "myservice-persist",
						OverrideName:         "myservice-persist",
						Type:                 "basic-persistent",
						DBaaSEnvironment:     "production",
						PersistentVolumePath: "/storage/data",
						PersistentVolumeName: "basic",
					},
					{
						Name:                 "myservice-persist-po",
						OverrideName:         "myservice-persist-po",
						Type:                 "basic-persistent",
						DBaaSEnvironment:     "production",
						ServicePort:          8080,
						PersistentVolumePath: "/storage/data",
						PersistentVolumeName: "basic",
					},
					{
						Name:                 "myservice-persist-posize",
						OverrideName:         "myservice-persist-posize",
						Type:                 "basic-persistent",
						DBaaSEnvironment:     "production",
						ServicePort:          8080,
						PersistentVolumeSize: "100Gi",
						PersistentVolumePath: "/storage/data",
						PersistentVolumeName: "basic",
					},
				},
			},
		},
		want: "test-resources/netpol/result-np-1.yaml",
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateNetworkPolicy(tt.args.buildValues)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateNetworkPolicy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			r1, err := os.ReadFile(tt.want)
			if err != nil {
				t.Errorf("couldn't read file %v: %v", tt.want, err)
			}
			result, err := TemplateNetworkPolicy(got)
			if err != nil {
				t.Errorf("couldn't generate template  %v", err)
			}
			if !reflect.DeepEqual(string(result), string(r1)) {
				t.Errorf("GenerateDeploymentTemplate() = \n%v", diff.LineDiff(string(r1), string(result)))
			}
		})
	}
}
