package cmd

import (
	"os"
	"reflect"
	"testing"

	"github.com/uselagoon/build-deploy-tool/internal/dbaasclient"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	"github.com/uselagoon/build-deploy-tool/internal/testdata"
)

// these tests uses the same files as the dbaas templates
func TestIdentifyDBaaSConsumers(t *testing.T) {
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
		vars         []helpers.EnvironmentVariable
		templatePath string
		want         []string
		wantErr      bool
	}{
		{
			name: "test1 - mariadb to mariadb-dbaas only",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/complex/lagoon.yml",
				}, true),
			templatePath: "testdata/output",
			want: []string{
				"mariadb:mariadb-dbaas",
			},
		},
		{
			name: "test2 - mariadb to mariadb-shared which converts to mariadb-dbaas",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/complex/lagoon.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{Name: "LAGOON_SERVICE_TYPES", Value: "mariadb:mariadb-shared", Scope: "build"},
					},
				}, true),
			templatePath: "testdata/output",
			want: []string{
				"mariadb:mariadb-dbaas",
			},
		},
		{
			name: "test3 - override provider to non-existent should result in failing dbaas check and a single pod no dbaas found",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/complex/lagoon.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{Name: "LAGOON_DBAAS_ENVIRONMENT_TYPES", Value: "mariadb:development2", Scope: "build"},
					},
				}, true),
			templatePath: "testdata/output",
			want:         []string{},
		},
		{
			name: "test4 - mariadb-single to mariadb-dbaas (using mariadb-shared to mariadb-dbaas conversion)",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/complex/lagoon.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{Name: "LAGOON_SERVICE_TYPES", Value: "mariadb:mariadb-shared", Scope: "build"},
					},
				}, true),
			templatePath: "testdata/output",
			want: []string{
				"mariadb:mariadb-dbaas",
			},
		},
		{
			name: "test5 - multiple mariadb",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/complex/lagoon.multidb.yml",
				}, true),
			templatePath: "testdata/output",
			want: []string{
				"mariadb:mariadb-dbaas",
				"mariadb2:mariadb-dbaas",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			// setup the fake dbaas server
			ts := dbaasclient.TestDBaaSHTTPServer()
			defer ts.Close()
			err = os.Setenv("DBAAS_OPERATOR_HTTP", ts.URL)
			if err != nil {
				t.Errorf("%v", err)
			}

			got, err := IdentifyDBaaSConsumers(generator)
			if (err != nil) != tt.wantErr {
				t.Errorf("IdentifyDBaaSConsumers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("IdentifyDBaaSConsumers() = %v, want %v", got, tt.want)
			}
			t.Cleanup(func() {
				helpers.UnsetEnvVars(tt.vars)
			})
		})
	}
}
