package cmd

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	"github.com/uselagoon/build-deploy-tool/internal/testdata"
)

func TestAutogeneratedIngressGeneration(t *testing.T) {
	tests := []struct {
		name         string
		args         testdata.TestData
		templatePath string
		wantErr      bool
		emptyDir     bool // if no templates are generated, then there will be a .gitkeep file in there
		want         string
	}{
		{
			name: "test1 autogenerated route",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/node/lagoon.yml",
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/autogen-templates/ingress-1",
		},
		{
			name: "test2 no autogenerated routes but allow pullrequests",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "pr-123",
					BuildType:       "pullrequest",
					PRNumber:        "123",
					PRHeadBranch:    "main",
					PRBaseBranch:    "main2",
					EnvironmentType: "development",
					LagoonYAML:      "../internal/testdata/node/lagoon.autogen-1.yml",
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/autogen-templates/ingress-2",
		},
		{
			name: "test3 autogenerated route but no pullrequests",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "pr-123",
					BuildType:       "pullrequest",
					PRNumber:        "123",
					PRHeadBranch:    "main",
					PRBaseBranch:    "main2",
					EnvironmentType: "development",
					LagoonYAML:      "../internal/testdata/node/lagoon.autogen-2.yml",
				}, true),
			templatePath: "testdata/output",
			emptyDir:     true,
			want:         "",
		},
		{
			name: "test4 autogenerated route no service in pattern",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/node/lagoon.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_SYSTEM_ROUTER_PATTERN",
							Value: "${project}-${environment}.example.com",
							Scope: "internal_system",
						},
					},
				}, false),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/autogen-templates/ingress-3",
		},
		{
			name: "test5 autogenerated route short url",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "ci-features-control-k8s",
					EnvironmentName: "short-router-url-from-a-very-l-ebe8",
					Branch:          "short-router-url-from-a-very-long-environment-name-like-this",
					LagoonYAML:      "../internal/testdata/node/lagoon.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_SYSTEM_ROUTER_PATTERN",
							Value: "${service}.${project}-${environment}.example.com",
							Scope: "internal_system",
						},
					},
				}, false),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/autogen-templates/ingress-4",
		},
		{
			name: "test6 autogenerated routes but disabled by service label",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/node/lagoon.autogen-3.yml",
				}, true),
			templatePath: "testdata/output",
			emptyDir:     true,
			want:         "",
		},
		{
			name: "test7 no autogenerated routes but enabled by service label",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/node/lagoon.autogen-4.yml",
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/autogen-templates/ingress-5",
		},
		{
			name: "test8 autogenerated routes with fastly",
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
						{
							Name:  "LAGOON_FASTLY_AUTOGENERATED",
							Value: "enabled",
							Scope: "build",
						},
					},
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/autogen-templates/ingress-6",
		},
		{
			name: "test9 autogenerated routes with fastly specific domain",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/node/lagoon.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FASTLY_SERVICE_IDS",
							Value: "node-example-project-main.example.com:service-id:true",
							Scope: "build",
						},
						{
							Name:  "LAGOON_FASTLY_AUTOGENERATED",
							Value: "enabled",
							Scope: "build",
						},
					},
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/autogen-templates/ingress-7",
		},
		{
			name: "test10 autogenerated routes with fastly and specific secret",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/node/lagoon.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FASTLY_SERVICE_ID",
							Value: "service-id:true:secretname",
							Scope: "build",
						},
						{
							Name:  "LAGOON_FASTLY_AUTOGENERATED",
							Value: "enabled",
							Scope: "build",
						},
					},
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/autogen-templates/ingress-8",
		},
		{
			name: "test11 autogenerated route development environment",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					EnvironmentType: "development",
					LagoonYAML:      "../internal/testdata/node/lagoon.yml",
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/autogen-templates/ingress-9",
		},
		{
			name: "test12 autogenerated route development environment - no insecure redirect",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					EnvironmentType: "development",
					LagoonYAML:      "../internal/testdata/node/lagoon.autogen-5.yml",
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/autogen-templates/ingress-10",
		},
		{
			name: "test13 autogenerated route development service type override",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					EnvironmentType: "development",
					LagoonYAML:      "../internal/testdata/node/lagoon.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_SERVICE_TYPES",
							Value: "node:node-persistent",
							Scope: "build",
						},
					},
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/autogen-templates/ingress-11",
		},
		{
			name: "test14 autogenerated route development no service type override",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					EnvironmentType: "development",
					LagoonYAML:      "../internal/testdata/node/lagoon.autogen-6.yml",
				}, true),
			templatePath: "testdata/output",
			emptyDir:     true,
			want:         "",
		},
		{
			name: "test15 autogenerated route development service type override",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					EnvironmentType: "development",
					LagoonYAML:      "../internal/testdata/nginxphp/lagoon.yml",
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/nginxphp/autogen-templates/ingress-1",
		},
		{
			name: "test16 autogenerated route development service name override",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					EnvironmentType: "development",
					LagoonYAML:      "../internal/testdata/nginxphp/lagoon.servicename.yml",
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/nginxphp/autogen-templates/ingress-2",
		},
		{
			name: "test17 autogenerated route development service type override",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "ci-drush-la-control-k8s",
					EnvironmentName: "drush-first",
					Branch:          "drush-first",
					EnvironmentType: "development",
					LagoonYAML:      "../internal/testdata/complex/lagoon.small.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_SYSTEM_ROUTER_PATTERN",
							Value: "${service}.${project}.${environment}.172.18.0.3.nip.io",
							Scope: "internal_system",
						},
					},
				}, false),
			templatePath: "testdata/output",
			want:         "../internal/testdata/complex/autogen-templates/ingress-1",
		},
		{
			name: "test18 autogenerated route tls-acme disabled",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					EnvironmentType: "production",
					LagoonYAML:      "../internal/testdata/node/lagoon.autogen-7.yml",
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/autogen-templates/ingress-12",
		},
		{
			name: "test19 autogenerated routes but tls-acme disabled by service label",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					EnvironmentType: "production",
					LagoonYAML:      "../internal/testdata/node/lagoon.autogen-8.yml",
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/autogen-templates/ingress-13",
		},
		{
			name: "test20 autogenerated routes where lagoon.name of service does not match service names",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "content-example-com",
					EnvironmentName: "feature-migration",
					Branch:          "feature/migration",
					EnvironmentType: "development",
					LagoonYAML:      "../internal/testdata/complex/lagoon.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_SYSTEM_ROUTER_PATTERN",
							Value: "${environment}.${project}.example.com",
							Scope: "internal_system",
						},
					},
				}, false),
			templatePath: "testdata/output",
			want:         "../internal/testdata/complex/autogen-templates/ingress-2",
		},
		{
			name: "test21 autogenerated routes where docker-compose env_file has missing file references",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "test21-example-com",
					EnvironmentName: "feature",
					Branch:          "feature",
					EnvironmentType: "development",
					LagoonYAML:      "../internal/testdata/complex/lagoon.complex-1.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_SYSTEM_ROUTER_PATTERN",
							Value: "${environment}.${project}.example.com",
							Scope: "internal_system",
						},
					},
				}, false),
			templatePath: "testdata/output",
			want:         "../internal/testdata/complex/autogen-templates/ingress-3",
		},
		{
			name: "test22 autogenerated routes where should truncate long dns",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "content-abcdefghijk-net-com-co",
					EnvironmentName: "pr-123",
					EnvironmentType: "development",
					BuildType:       "pullrequest",
					PRNumber:        "123",
					PRHeadBranch:    "main",
					PRBaseBranch:    "main2",
					LagoonYAML:      "../internal/testdata/nginxphp/lagoon.nginx-1.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_SYSTEM_ROUTER_PATTERN",
							Value: "${environment}.${project}.abc1.abc.net.com.co",
							Scope: "internal_system",
						},
					},
				}, false),
			templatePath: "testdata/output",
			want:         "../internal/testdata/nginxphp/autogen-templates/ingress-3",
		},
		{
			name: "test23 autogenerated routes with fastly service, should be no fastly annotations on autogenerated route",
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
			want:         "../internal/testdata/node/autogen-templates/ingress-14",
		},
		{
			name: "test24 autogenerated routes with 'string true enabled' with fastly service, should be no fastly annotations on autogenerated route",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/node/lagoon.autogen-9.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FASTLY_SERVICE_ID",
							Value: "service-id:true",
							Scope: "build",
						},
					},
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/autogen-templates/ingress-14",
		},
		{
			name: "test25 autogenerated routes enabled globally but disabled by environment",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "autogendisabled",
					Branch:          "autogendisabled",
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
			emptyDir:     true,
			want:         "",
		},
		{
			name: "test26 polysite autogenerated routes enabled globally but disabled by environment",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "multiproject3",
					EnvironmentName: "autogendisabled",
					Branch:          "autogendisabled",
					LagoonYAML:      "../internal/testdata/node/lagoon.polysite.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FASTLY_SERVICE_ID",
							Value: "service-id:true",
							Scope: "build",
						},
					},
				}, true),
			templatePath: "testdata/output",
			emptyDir:     true,
			want:         "",
		},
		{
			name: "test27 autogenerated route unidler request verification disable",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/node/lagoon.unidler.yml",
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/node/autogen-templates/ingress-15",
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

			if err := AutogeneratedIngressGeneration(generator); (err != nil) != tt.wantErr {
				t.Errorf("AutogeneratedIngressGeneration() error = %v, wantErr %v", err, tt.wantErr)
			}

			files, err := ioutil.ReadDir(savedTemplates)
			if err != nil {
				t.Errorf("couldn't read directory %v: %v", savedTemplates, err)
			}
			resultSize := 0
			results := []fs.FileInfo{}
			if !tt.emptyDir {
				results, err = ioutil.ReadDir(tt.want)
				if err != nil {
					t.Errorf("couldn't read directory %v: %v", tt.want, err)
				}
				// .gitkeep file needs to be subtracted to equal 0
				resultSize = len(results)
			}
			if len(files) != resultSize {
				for _, f := range files {
					f1, err := os.ReadFile(fmt.Sprintf("%s/%s", savedTemplates, f.Name()))
					if err != nil {
						t.Errorf("couldn't read file %v: %v", savedTemplates, err)
					}
					fmt.Println(string(f1))
				}
				t.Errorf("number of generated templates doesn't match results %v/%v: %v", len(files), resultSize, err)
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
					fmt.Println(fmt.Sprintf("%s/%s", savedTemplates, f.Name()))
					fmt.Println(string(f1))
				}
				t.Errorf("resulting templates do not match")
			}
			t.Cleanup(func() {
				helpers.UnsetEnvVars(nil)
			})
		})
	}
}
