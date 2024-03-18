package cmd

import (
	"reflect"
	"testing"

	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	"github.com/uselagoon/build-deploy-tool/internal/testdata"
)

func TestIdentifyServices(t *testing.T) {
	tests := []struct {
		name         string
		args         testdata.TestData
		templatePath string
		want         []string
		wantServices []identifyServices
		wantErr      bool
	}{
		{
			name: "test1 single service",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/node/lagoon.yml",
				}, true),
			templatePath: "testdata/output",
			want:         []string{"node"},
			wantServices: []identifyServices{{Name: "node", Type: "node"}},
		},
		{
			name: "test2 complex servives",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					EnvironmentType: "development",
					LagoonYAML:      "../internal/testdata/complex/lagoon.yml",
				}, true),
			templatePath: "testdata/output",
			want:         []string{"cli", "nginx-php", "mariadb", "redis"},
			wantServices: []identifyServices{
				{Name: "cli", Type: "cli-persistent"},
				{Name: "nginx-php", Type: "nginx-php-persistent"},
				{Name: "mariadb", Type: "mariadb-dbaas"},
				{Name: "redis", Type: "redis"},
			},
		},
		{
			name: "test3 complex servives where one is removed",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					EnvironmentType: "development",
					LagoonYAML:      "../internal/testdata/complex/lagoon.yml",
					EnvVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_SERVICE_TYPES",
							Value: "redis:none",
							Scope: "build",
						},
					},
				}, true),
			templatePath: "testdata/output",
			want:         []string{"cli", "nginx-php", "mariadb"},
			wantServices: []identifyServices{
				{Name: "cli", Type: "cli-persistent"},
				{Name: "nginx-php", Type: "nginx-php-persistent"},
				{Name: "mariadb", Type: "mariadb-dbaas"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// set the environment variables from args
			savedTemplates := tt.templatePath
			generator, err := testdata.SetupEnvironment(*rootCmd, savedTemplates, tt.args)
			if err != nil {
				t.Errorf("%v", err)
			}
			got, got2, err := IdentifyServices(generator)
			if (err != nil) != tt.wantErr {
				t.Errorf("IdentifyServices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("IdentifyServices() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got2, tt.wantServices) {
				t.Errorf("IdentifyServices() = %v, want %v", got2, tt.wantServices)
			}
			t.Cleanup(func() {
				helpers.UnsetEnvVars(nil)
			})
		})
	}
}
