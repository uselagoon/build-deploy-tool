package cmd

import (
	"encoding/base64"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	"github.com/uselagoon/build-deploy-tool/internal/testdata"
)

func TestTemplateRoutes(t *testing.T) {
	tests := []struct {
		name         string
		args         testdata.TestData
		templatePath string
		want         string
	}{
		{
			name: "test1 check LAGOON_FASTLY_SERVICE_IDS with secret no values",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/node/lagoon.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FASTLY_SERVICE_IDS",
							Value: "example.com:service-id:true:annotationscom",
							Scope: "build",
						},
					},
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/ingress-templates/ingress-1",
		},
		{
			name: "test2 check LAGOON_FASTLY_SERVICE_IDS no secret and no values",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/node/lagoon.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FASTLY_SERVICE_IDS",
							Value: "example.com:service-id:true",
							Scope: "build",
						},
					},
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/ingress-templates/ingress-2",
		},
		{
			name: "test3 check LAGOON_FASTLY_SERVICE_ID no secret and no values",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/node/lagoon.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FASTLY_SERVICE_ID",
							Value: "service-id:true",
							Scope: "build",
						},
					},
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/ingress-templates/ingress-3",
		},
		{
			name: "test4 check no fastly and no values",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/node/lagoon.yml",
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/ingress-templates/ingress-4",
		},
		{
			name: "test5 multiproject1 no values",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "multiproject1",
					EnvironmentName: "multiproject",
					Branch:          "multiproject",
					LagoonYAML:      "../internal/testdata/node/lagoon.polysite.yml",
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/ingress-templates/ingress-5",
		},
		{
			name: "test6 multiproject2 no values",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "multiproject2",
					EnvironmentName: "multiproject",
					Branch:          "multiproject",
					LagoonYAML:      "../internal/testdata/node/lagoon.polysite.yml",
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/ingress-templates/ingress-6",
		},
		{
			name: "test7 multidomain no values",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "tworoutes",
					Branch:          "tworoutes",
					LagoonYAML:      "../internal/testdata/node/lagoon.yml",
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/ingress-templates/ingress-7",
		},
		{
			name: "test8 multidomain no values",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "branch-routes",
					Branch:          "branch/routes",
					LagoonYAML:      "../internal/testdata/node/lagoon.yml",
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/ingress-templates/ingress-8",
		},
		{
			name: "test9 active no values",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:        "example-project",
					EnvironmentName:    "main",
					Branch:             "main",
					ActiveEnvironment:  "main",
					StandbyEnvironment: "main-sb",
					LagoonYAML:         "../internal/testdata/node/lagoon.activestandby.yml",
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/ingress-templates/ingress-9",
		},
		{
			name: "test10 standby no values",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:        "example-project",
					EnvironmentName:    "main-sb",
					Branch:             "main-sb",
					ActiveEnvironment:  "main",
					StandbyEnvironment: "main-sb",
					LagoonYAML:         "../internal/testdata/node/lagoon.activestandby.yml",
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/ingress-templates/ingress-10",
		},
		{
			name: "test11 standby no values",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "content-example-com",
					EnvironmentName: "production",
					Branch:          "production",
					LagoonYAML:      "../internal/testdata/complex/lagoon.yml",
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/complex/ingress-templates/ingress-1",
		},
		{
			name: "test12 check LAGOON_ROUTES_JSON generates ingress",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "noyamlroutes",
					Branch:          "noyamlroutes",
					LagoonYAML:      "../internal/testdata/node/lagoon.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FASTLY_SERVICE_IDS",
							Value: "example.com:service-id:true:annotationscom",
							Scope: "build",
						},
						{
							Name:  "LAGOON_ROUTES_JSON",
							Value: base64.StdEncoding.EncodeToString([]byte(`{"routes":[{"domain":"test1.example.com","service":"nginx","tls-acme":false,"monitoring-path":"/bypass-cache"}]}`)),
							Scope: "build",
						},
					},
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/ingress-templates/ingress-11",
		},
		{
			name: "test13 ingress class from default flag",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					IngressClass:    "nginx",
					LagoonYAML:      "../internal/testdata/node/lagoon.yml",
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/ingress-templates/ingress-12",
		},
		{
			name: "test14 ingress class from lagoon.yml should overwrite default and featureflag variable",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "ingressclass",
					Branch:          "ingressclass",
					IngressClass:    "nginx",
					LagoonYAML:      "../internal/testdata/node/lagoon.yml",
				}, true), templatePath: "testdata/output",
			want: "../internal/testdata/node/ingress-templates/ingress-13",
		},
		{
			name: "test15a ingress class from lagoon api project scope",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					IngressClass:    "nginx",
					LagoonYAML:      "../internal/testdata/node/lagoon.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_INGRESS_CLASS",
							Value: "custom-ingress",
							Scope: "build",
						},
					},
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/ingress-templates/ingress-14",
		},
		{
			name: "test15b ingress class from lagoon api environment scope",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/node/lagoon.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_INGRESS_CLASS",
							Value: "project-custom-ingress",
							Scope: "build",
						},
					},
					EnvVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_INGRESS_CLASS",
							Value: "custom-ingress",
							Scope: "build",
						},
					},
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/ingress-templates/ingress-15",
		},
		{
			name: "test16 hsts basic",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "hsts",
					Branch:          "hsts",
					LagoonYAML:      "../internal/testdata/node/lagoon.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FASTLY_SERVICE_IDS",
							Value: "example.com:service-id:true:annotationscom",
							Scope: "build",
						},
					},
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/ingress-templates/ingress-16",
		},
		{
			name: "test17 hsts advanced",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "hsts2",
					Branch:          "hsts2",
					LagoonYAML:      "../internal/testdata/node/lagoon.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FASTLY_SERVICE_IDS",
							Value: "example.com:service-id:true:annotationscom",
							Scope: "build",
						},
					},
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/ingress-templates/ingress-17",
		},
		{
			name: "test18 check first route has monitoring only",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "tworoutes",
					Branch:          "tworoutes",
					LagoonYAML:      "../internal/testdata/node/lagoon.yml",
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/ingress-templates/ingress-18",
		},
		{
			name: "test19 pullrequest routes",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "pr-4841",
					BuildType:       "pullrequest",
					PRNumber:        "4841",
					PRHeadBranch:    "main",
					PRBaseBranch:    "my-branch",
					EnvironmentType: "development",
					LagoonYAML:      "../internal/testdata/node/lagoon.yml",
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/ingress-templates/ingress-19",
		},
		{
			name: "test20 pullrequest routes polysite",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "pr-4841",
					BuildType:       "pullrequest",
					PRNumber:        "4841",
					PRHeadBranch:    "main",
					PRBaseBranch:    "my-branch",
					EnvironmentType: "development",
					LagoonYAML:      "../internal/testdata/node/lagoon.polysite-pr.yml",
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/ingress-templates/ingress-20",
		},
		{
			name: "test21 alternative names",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "alternativename",
					Branch:          "alternativename",
					LagoonYAML:      "../internal/testdata/node/lagoon.yml",
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/ingress-templates/ingress-21",
		},
		{
			name: "test22 check wildcard",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "wildcard",
					Branch:          "wildcard",
					LagoonYAML:      "../internal/testdata/node/lagoon.yml",
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/ingress-templates/ingress-22",
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

			err = IngressTemplateGeneration(generator)
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
