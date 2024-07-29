package cmd

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/uselagoon/build-deploy-tool/internal/dbaasclient"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	"github.com/uselagoon/build-deploy-tool/internal/testdata"
)

func TestLagoonEnvTemplateGeneration(t *testing.T) {
	tests := []struct {
		name         string
		description  string
		args         testdata.TestData
		templatePath string
		want         string
		dbaasCreds   string
		vars         []helpers.EnvironmentVariable
	}{
		{
			name: "test1 basic deployment",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/basic/lagoon.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "MY_SPECIAL_VARIABLE1",
							Value: "myspecialvariable1",
							Scope: "global",
						},
						{
							Name:  "MY_SPECIAL_VARIABLE2",
							Value: "myspecialvariable2",
							Scope: "runtime",
						},
						{
							Name:  "MY_SPECIAL_VARIABLE3",
							Value: "myspecialvariable3",
							Scope: "build",
						},
						{
							Name:  "MY_SPECIAL_VARIABLE",
							Value: "myspecialvariable",
							Scope: "global",
						},
						{
							Name:  "LAGOON_SYSTEM_CORE_VERSION",
							Value: "v2.19.0",
							Scope: "internal_system",
						},
						{
							Name:  "REGISTRY_PASSWORD",
							Value: "myenvvarregistrypassword",
							Scope: "container_registry",
						},
					},
					EnvVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "MY_SPECIAL_VARIABLE2",
							Value: "myspecialvariable2-env-override",
							Scope: "global",
						},
						{
							Name:  "MY_SPECIAL_VARIABLE4",
							Value: "myspecialvariable4",
							Scope: "runtime",
						},
					},
				}, true),
			templatePath: "testoutput",
			want:         "internal/testdata/basic/configmap-templates/lagoonenv1",
		},
		{
			name: "test1 basic deployment with mariadb creds",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/basic/lagoon.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "MY_SPECIAL_VARIABLE1",
							Value: "myspecialvariable1",
							Scope: "global",
						},
						{
							Name:  "MY_SPECIAL_VARIABLE2",
							Value: "myspecialvariable2",
							Scope: "runtime",
						},
						{
							Name:  "MY_SPECIAL_VARIABLE3",
							Value: "myspecialvariable3",
							Scope: "build",
						},
						{
							Name:  "MY_SPECIAL_VARIABLE",
							Value: "myspecialvariable",
							Scope: "global",
						},
						{
							Name:  "LAGOON_SYSTEM_CORE_VERSION",
							Value: "v2.19.0",
							Scope: "internal_system",
						},
						{
							Name:  "REGISTRY_PASSWORD",
							Value: "myenvvarregistrypassword",
							Scope: "container_registry",
						},
					},
					EnvVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "MY_SPECIAL_VARIABLE2",
							Value: "myspecialvariable2-env-override",
							Scope: "global",
						},
						{
							Name:  "MY_SPECIAL_VARIABLE4",
							Value: "myspecialvariable4",
							Scope: "runtime",
						},
					},
				}, true),
			dbaasCreds:   "internal/testdata/basic/lagoonenv2-creds.json",
			templatePath: "testoutput",
			want:         "internal/testdata/basic/configmap-templates/lagoonenv2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.UnsetEnvVars(tt.vars) //unset variables before running tests
			for _, envVar := range tt.vars {
				err := os.Setenv(envVar.Name, envVar.Value)
				if err != nil {
					t.Errorf("%v", err)
				}
			}
			// set the environment variables from args
			savedTemplates := tt.templatePath
			generator, err := testdata.SetupEnvironment(*rootCmd, savedTemplates, tt.args)
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
			dbaasCreds := &DBaaSCredRefs{}
			if tt.dbaasCreds != "" {
				dbaasCreds, err = loadCredsFromFile(tt.dbaasCreds)
				if err != nil {
					t.Errorf("%v", err)
				}
				dbCreds := map[string]string{}
				for _, v := range *dbaasCreds {
					for k, v1 := range v {
						dbCreds[k] = v1
					}
				}
				generator.DBaaSVariables = dbCreds
			}
			err = LagoonEnvTemplateGeneration(generator, "")
			if err != nil {
				t.Errorf("%v", err)
			}

			files, err := os.ReadDir(savedTemplates)
			if err != nil {
				t.Errorf("couldn't read directory %v: %v", savedTemplates, err)
			}
			results, err := os.ReadDir(tt.want)
			if err != nil {
				t.Errorf("couldn't read directory %v: %v", tt.want, err)
			}
			if len(files) != len(results) {
				for _, f := range files {
					f1, err := os.ReadFile(fmt.Sprintf("%s/%s", savedTemplates, f.Name()))
					if err != nil {
						t.Errorf("couldn't read file %v: %v", savedTemplates, err)
					}
					fmt.Println(string(f1))
				}
				t.Errorf("number of generated templates doesn't match results %v/%v: %v", len(files), len(results), err)
			}
			fCount := 0
			for _, f := range files {
				for _, r := range results {
					if f.Name() == r.Name() {
						fCount++
						f1, err := os.ReadFile(fmt.Sprintf("%s/%s", savedTemplates, f.Name()))
						if err != nil {
							t.Errorf("couldn't read file %v: %v", savedTemplates, err)
						}
						r1, err := os.ReadFile(fmt.Sprintf("%s/%s", tt.want, f.Name()))
						if err != nil {
							t.Errorf("couldn't read file %v: %v", tt.want, err)
						}
						if !reflect.DeepEqual(f1, r1) {
							t.Errorf("LagoonEnvTemplateGeneration() = \n%v", diff.LineDiff(string(r1), string(f1)))
						}
					}
				}
			}
			if fCount != len(files) {
				for _, f := range files {
					f1, err := os.ReadFile(fmt.Sprintf("%s/%s", savedTemplates, f.Name()))
					if err != nil {
						t.Errorf("couldn't read file %v: %v", savedTemplates, err)
					}
					fmt.Println(string(f1))
				}
				t.Errorf("resulting templates do not match")
			}
			t.Cleanup(func() {
				helpers.UnsetEnvVars(tt.vars)
			})
		})
	}
}
