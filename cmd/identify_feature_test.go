package cmd

import (
	"os"
	"testing"

	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	"github.com/uselagoon/build-deploy-tool/internal/testdata"
)

func TestIdentifyFeatureFlag(t *testing.T) {
	type args struct {
		name               string
		alertContact       string
		statusPageID       string
		projectName        string
		environmentName    string
		branch             string
		prNumber           string
		prHeadBranch       string
		prBaseBranch       string
		environmentType    string
		buildType          string
		activeEnvironment  string
		standbyEnvironment string
		cacheNoCache       string
		serviceID          string
		secretPrefix       string
		projectVars        string
		envVars            string
		lagoonVersion      string
		lagoonYAML         string
		valuesFilePath     string
		templatePath       string
	}
	tests := []struct {
		name         string
		args         testdata.TestData
		templatePath string
		varName      string
		vars         []helpers.EnvironmentVariable
		want         string
		wantErr      bool
	}{
		{
			name:    "test1 check if flag is defined in lagoon project variables",
			varName: "ROOTLESS_WORKLOAD",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/node/lagoon.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_ROOTLESS_WORKLOAD",
							Value: "enabled",
							Scope: "build",
						},
					},
				}, true),
			templatePath: "testdata/output",
			want:         "enabled",
		},
		{
			name:    "test2 check if flag is defined in lagoon environment variables",
			varName: "ROOTLESS_WORKLOAD",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/node/lagoon.yml",
					EnvVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_ROOTLESS_WORKLOAD",
							Value: "enabled",
							Scope: "build",
						},
					},
				}, true),
			templatePath: "testdata/output",
			want:         "enabled",
		},
		{
			name:    "test3 check if force flag is defined in build variables",
			varName: "ROOTLESS_WORKLOAD",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/node/lagoon.yml",
				}, true),
			templatePath: "testdata/output",
			vars: []helpers.EnvironmentVariable{
				{
					Name:  "LAGOON_FEATURE_FLAG_FORCE_ROOTLESS_WORKLOAD",
					Value: "enabled",
				},
			},
			want: "enabled",
		},
		{
			name:    "test4 check if force flag is defined in build variables and default flag is ignored",
			varName: "ROOTLESS_WORKLOAD",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/node/lagoon.yml",
				}, true),
			templatePath: "testdata/output",
			vars: []helpers.EnvironmentVariable{
				{
					Name:  "LAGOON_FEATURE_FLAG_FORCE_ROOTLESS_WORKLOAD",
					Value: "enabled",
				},
				{
					Name:  "LAGOON_FEATURE_FLAG_DEFAULT_ROOTLESS_WORKLOAD",
					Value: "disabled",
				},
			},
			want: "enabled",
		},
		{
			name:    "test5 check if force flag is defined in build variables and one defined in lagoon project variables is ignored",
			varName: "ROOTLESS_WORKLOAD",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/node/lagoon.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_ROOTLESS_WORKLOAD",
							Value: "enabled",
							Scope: "build",
						},
					},
				}, true),
			vars: []helpers.EnvironmentVariable{
				{
					Name:  "LAGOON_FEATURE_FLAG_FORCE_ROOTLESS_WORKLOAD",
					Value: "disabled",
				},
				{
					Name:  "LAGOON_FEATURE_FLAG_DEFAULT_ROOTLESS_WORKLOAD",
					Value: "disabled",
				},
			},
			want: "disabled",
		},
		{
			name:    "test6 check if default flag is ignored and lagoon project variable is used",
			varName: "ROOTLESS_WORKLOAD",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/node/lagoon.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_ROOTLESS_WORKLOAD",
							Value: "enabled",
							Scope: "build",
						},
					},
				}, true),
			vars: []helpers.EnvironmentVariable{
				{
					Name:  "LAGOON_FEATURE_FLAG_DEFAULT_ROOTLESS_WORKLOAD",
					Value: "disabled",
				},
			},
			want: "enabled",
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
			for _, envVar := range tt.vars {
				err = os.Setenv(envVar.Name, envVar.Value)
				if err != nil {
					t.Errorf("%v", err)
				}
			}
			got, err := IdentifyFeatureFlag(generator, tt.varName)
			if (err != nil) != tt.wantErr {
				t.Errorf("IdentifyFeatureFlag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IdentifyFeatureFlag() = %v, want %v", got, tt.want)
			}
			t.Cleanup(func() {
				helpers.UnsetEnvVars(tt.vars)
			})
		})
	}
}
