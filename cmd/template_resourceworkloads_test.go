package cmd

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/testdata"
)

func TestResourceWorkloadTemplateGeneration(t *testing.T) {
	type args struct {
		alertContact              string
		statusPageID              string
		projectName               string
		environmentName           string
		branch                    string
		prNumber                  string
		prHeadBranch              string
		prBaseBranch              string
		environmentType           string
		buildType                 string
		activeEnvironment         string
		standbyEnvironment        string
		cacheNoCache              string
		serviceID                 string
		secretPrefix              string
		ingressClass              string
		projectVars               string
		envVars                   string
		lagoonVersion             string
		lagoonYAML                string
		valuesFilePath            string
		templatePath              string
		workloadJSONfile          string
		resourceWorkloadOverrides string
	}
	tests := []struct {
		name string
		// args args
		args                      testdata.TestData
		templatePath              string
		workloadJSONfile          string
		resourceWorkloadOverrides string
		want                      string
	}{
		{
			name: "test1 no resource workloads",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/node/lagoon.yml",
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/resource-templates/resource1",
		},
		{
			name: "test2 node hpa",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/node/lagoon.resources1.yml",
				}, true),
			templatePath:     "testdata/output",
			workloadJSONfile: "../internal/testdata/node/workload.resources1.json",
			want:             "../internal/testdata/node/resource-templates/resource2",
		},
		{
			name: "test3 nginx hpa and pdb",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/complex/lagoon.resource1.yml",
				}, true),
			templatePath:     "testdata/output",
			workloadJSONfile: "../internal/testdata/complex/workload.resources1.json",
			want:             "../internal/testdata/complex/resource-templates/resource1",
		},
		{
			name: "test4 nginx hpa and pdb with resource override from feature flag",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/complex/lagoon.yml",
				}, true),
			templatePath:              "testdata/output",
			workloadJSONfile:          "../internal/testdata/complex/workload.resources2.json",
			resourceWorkloadOverrides: "nginx-php:nginx-php-performance",
			want:                      "../internal/testdata/complex/resource-templates/resource2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// set the environment variables from args
			err := os.Setenv("LAGOON_FEATURE_FLAG_DEFAULT_WORKLOAD_RESOURCES", helpers.ReadFileBase64Encode(tt.workloadJSONfile))
			if err != nil {
				t.Errorf("%v", err)
			}
			err = os.Setenv("LAGOON_FEATURE_FLAG_DEFAULT_WORKLOAD_RESOURCE_TYPES", tt.resourceWorkloadOverrides)
			if err != nil {
				t.Errorf("%v", err)
			}
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

			err = ResourceWorkloadTemplateGeneration(generator)
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
				helpers.UnsetEnvVars([]helpers.EnvironmentVariable{{Name: "LAGOON_FEATURE_FLAG_DEFAULT_INGRESS_CLASS"}})
			})
		})
	}
}
