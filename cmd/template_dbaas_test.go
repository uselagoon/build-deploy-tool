package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/uselagoon/build-deploy-tool/internal/dbaasclient"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	"github.com/uselagoon/build-deploy-tool/internal/testdata"
)

func TestDBaaSTemplateGeneration(t *testing.T) {
	tests := []struct {
		name         string
		args         testdata.TestData
		templatePath string
		want         string
		wantErr      bool
	}{
		{
			name: "test1 - mariadb-dbaas",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/complex/lagoon.yml",
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/complex/dbaas-templates/dbaas-1",
		},
		{
			name: "test2 - mariadb-single to mariadb-dbaas (using mariadb-shared to mariadb-dbaas conversion)",
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
			want:         "../internal/testdata/complex/dbaas-templates/dbaas-2",
		},
		{
			name: "test3 - multiple mariadb",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/complex/lagoon.multidb.yml",
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/complex/dbaas-templates/dbaas-3",
		},
		{
			name: "test4 - mongo",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/node/lagoon.mongo.yml",
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/dbaas-templates/dbaas-1",
		},
		{
			name: "test5 - mongo override (the mongo should not generate because it has a mongodb-single override)",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/node/lagoon.mongo.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{Name: "LAGOON_SERVICE_TYPES", Value: "mongo:mongodb-single", Scope: "build"},
					},
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/dbaas-templates/dbaas-2",
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
			defer os.RemoveAll(savedTemplates)

			if err := DBaaSTemplateGeneration(generator); (err != nil) != tt.wantErr {
				t.Errorf("DBaaSTemplateGeneration() error = %v, wantErr %v", err, tt.wantErr)
			}
			files, err := ioutil.ReadDir(savedTemplates)
			if err != nil {
				t.Errorf("couldn't read directory %v: %v", savedTemplates, err)
			}
			results, err := ioutil.ReadDir(tt.want)
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
							fmt.Println(string(f1))
							t.Errorf("resulting templates do not match")
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
				helpers.UnsetEnvVars([]helpers.EnvironmentVariable{{Name: "DBAAS_OPERATOR_HTTP"}})
			})
		})
	}
}
