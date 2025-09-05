package cmd

import (
	"os"
	"testing"

	"github.com/uselagoon/build-deploy-tool/internal/dbaasclient"
	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	"github.com/uselagoon/build-deploy-tool/internal/testdata"

	// changes the testing to source from root so paths to test resources must be defined from repo root
	_ "github.com/uselagoon/build-deploy-tool/internal/testing"
)

func TestIdentifyNativeCronjobs(t *testing.T) {
	tests := []struct {
		name         string
		args         testdata.TestData
		templatePath string
		want         string
	}{
		{
			name: "test1 basic deployment",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/node/lagoon.yml",
				}, true),
			templatePath: "testoutput",
			want:         "[]",
		},
		{
			name: "test2a nginx-php deployment",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/complex/lagoon.yml",
				}, true),
			templatePath: "testoutput",
			want:         `["cronjob-cli-drush-cron2"]`,
		},
		{
			name: "test2b nginx-php deployment - rootless",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/complex/lagoon.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_ROOTLESS_WORKLOAD",
							Value: "enabled",
							Scope: "build",
						},
					},
				}, true),
			templatePath: "testoutput",
			want:         `["cronjob-cli-drush-cron2"]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.UnsetEnvVars(nil) //unset variables before running tests
			// set the environment variables from args
			savedTemplates := tt.templatePath
			generator, err := testdata.SetupEnvironment(generator.GeneratorInput{}, savedTemplates, tt.args)
			if err != nil {
				t.Errorf("%v", err)
			}

			err = os.MkdirAll(savedTemplates, 0755)
			if err != nil {
				t.Errorf("couldn't create directory %v: %v", savedTemplates, err)
			}

			defer os.RemoveAll(savedTemplates)

			ts := dbaasclient.TestDBaaSHTTPServer()
			defer ts.Close()
			err = os.Setenv("DBAAS_OPERATOR_HTTP", ts.URL)
			if err != nil {
				t.Errorf("%v", err)
			}

			got, err := IdentifyNativeCronjobs(generator)
			if err != nil {
				t.Errorf("%v", err)
			}

			if got != tt.want {
				t.Errorf("IdentifyNativeCronjobs() = %v, want %v", got, tt.want)
			}

			t.Cleanup(func() {
				helpers.UnsetEnvVars(nil)
				helpers.UnsetEnvVars(tt.args.BuildPodVariables)
			})
		})
	}
}
