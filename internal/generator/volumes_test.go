package generator

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
)

func Test_flagDefaultVolumeCreation(t *testing.T) {
	tests := []struct {
		name        string
		buildValues *BuildValues
		want        []ServiceValues
		wantErr     bool
	}{
		{
			name: "test1",
			buildValues: &BuildValues{
				Services: []ServiceValues{
					{
						Name:                       "nginx",
						OverrideName:               "nginx",
						Type:                       "nginx-php-persistent",
						AutogeneratedRoutesEnabled: true,
						AutogeneratedRoutesTLSAcme: true,
						InPodCronjobs:              []lagoon.Cronjob{},
						NativeCronjobs:             []lagoon.Cronjob{},
						PersistentVolumeSize:       "5Gi",
						ImageBuild: &ImageBuild{
							TemporaryImage: "example-project-main-nginx",
							Context:        ".",
							DockerFile:     "../testdata/basic/docker/basic.dockerfile",
							BuildImage:     "harbor.example/example-project/main/nginx:latest",
						},
						PersistentVolumeName: "nginx",
						PersistentVolumePath: "/app/docroot/sites/default/files/",
						BackupsEnabled:       true,
					},
				},
			},
			want: []ServiceValues{
				{
					Name:                       "nginx",
					OverrideName:               "nginx",
					Type:                       "nginx-php-persistent",
					AutogeneratedRoutesEnabled: true,
					AutogeneratedRoutesTLSAcme: true,
					InPodCronjobs:              []lagoon.Cronjob{},
					NativeCronjobs:             []lagoon.Cronjob{},
					PersistentVolumeSize:       "5Gi",
					ImageBuild: &ImageBuild{
						TemporaryImage: "example-project-main-nginx",
						Context:        ".",
						DockerFile:     "../testdata/basic/docker/basic.dockerfile",
						BuildImage:     "harbor.example/example-project/main/nginx:latest",
					},
					PersistentVolumeName: "nginx",
					PersistentVolumePath: "/app/docroot/sites/default/files/",
					BackupsEnabled:       true,
					CreateDefaultVolume:  true,
				},
			},
		},
		{
			name: "test2",
			buildValues: &BuildValues{
				Services: []ServiceValues{
					{
						Name:                 "basic",
						OverrideName:         "basic",
						Type:                 "basic-persistent",
						PersistentVolumeSize: "5Gi",
						PersistentVolumeName: "basic-data",
						PersistentVolumePath: "/basic-data/",
					},
					{
						Name:                 "basic2",
						OverrideName:         "basic2",
						Type:                 "basic-persistent",
						PersistentVolumeName: "basic-data",
						PersistentVolumePath: "/basic-data/",
					},
				},
			},
			want: []ServiceValues{
				{
					Name:                 "basic",
					OverrideName:         "basic",
					Type:                 "basic-persistent",
					PersistentVolumeSize: "5Gi",
					PersistentVolumeName: "basic-data",
					PersistentVolumePath: "/basic-data/",
					CreateDefaultVolume:  true,
				},
				{
					Name:                 "basic2",
					OverrideName:         "basic2",
					Type:                 "basic-persistent",
					PersistentVolumeName: "basic-data",
					PersistentVolumePath: "/basic-data/",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := flagDefaultVolumeCreation(tt.buildValues); (err != nil) != tt.wantErr {
				t.Errorf("flagDefaultVolumeCreation() error = %v, wantErr %v", err, tt.wantErr)
			}
			f1, _ := json.MarshalIndent(tt.buildValues.Services, "", "  ")
			r1, _ := json.MarshalIndent(tt.want, "", "  ")
			if !reflect.DeepEqual(f1, r1) {
				t.Errorf("flagDefaultVolumeCreation() = \n%v", diff.LineDiff(string(r1), string(f1)))
			}
		})
	}
}
