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

	// changes the testing to source from root so paths to test resources must be defined from repo root
	_ "github.com/uselagoon/build-deploy-tool/internal/testing"
)

func TestTemplateLagoonServices(t *testing.T) {
	tests := []struct {
		name         string
		description  string
		args         testdata.TestData
		templatePath string
		want         string
		vars         []helpers.EnvironmentVariable
	}{
		{
			name: "test1 basic deployment",
			vars: []helpers.EnvironmentVariable{
				{
					Name:  "EXTERNAL_REGISTRY_SECRETS",
					Value: "lagoon-private-registry-custom-secret",
				},
			},
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/basic/lagoon.yml",
					ImageReferences: map[string]string{
						"node": "harbor.example/example-project/main/node@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
				}, true),
			templatePath: "testoutput",
			want:         "internal/testdata/basic/service-templates/service1",
		},
		{
			name: "test2a nginx-php deployment",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/complex/lagoon.varnish.yml",
					ImageReferences: map[string]string{
						"nginx":   "harbor.example/example-project/main/nginx@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"php":     "harbor.example/example-project/main/php@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"cli":     "harbor.example/example-project/main/cli@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"redis":   "harbor.example/example-project/main/redis@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"varnish": "harbor.example/example-project/main/varnish@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
				}, true),
			templatePath: "testoutput",
			want:         "internal/testdata/complex/service-templates/service1",
		},
		{
			name: "test2b nginx-php deployment - rootless",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/complex/lagoon.varnish.yml",
					ImageReferences: map[string]string{
						"nginx":   "harbor.example/example-project/main/nginx@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"php":     "harbor.example/example-project/main/php@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"cli":     "harbor.example/example-project/main/cli@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"redis":   "harbor.example/example-project/main/redis@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"varnish": "harbor.example/example-project/main/varnish@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_ROOTLESS_WORKLOAD",
							Value: "enabled",
							Scope: "build",
						},
					},
				}, true),
			templatePath: "testoutput",
			want:         "internal/testdata/complex/service-templates/service2",
		},
		{
			name:        "test3 - funky pvcs",
			description: "only create pvcs of the requested persistent-name in the docker-compose file",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/basic/lagoon.thunderhub.yml",
					ImageReferences: map[string]string{
						"lnd":        "harbor.example/example-project/main/lnd@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"thunderhub": "harbor.example/example-project/main/thunderhub@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"tor":        "harbor.example/example-project/main/tor@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_ROOTLESS_WORKLOAD",
							Value: "enabled",
							Scope: "build",
						},
					},
				}, true),
			templatePath: "testoutput",
			want:         "internal/testdata/basic/service-templates/service2",
		},
		{
			name:        "test4 - basic-persistent with worker-persistent",
			description: "create a basic-persistent that gets a pvc and mount that volume on a worker-persistent type",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/basic/lagoon.thunderhub-2.yml",
					ImageReferences: map[string]string{
						"lnd": "harbor.example/example-project/main/lnd@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"tor": "harbor.example/example-project/main/tor@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_ROOTLESS_WORKLOAD",
							Value: "enabled",
							Scope: "build",
						},
					},
				}, true),
			templatePath: "testoutput",
			want:         "internal/testdata/basic/service-templates/service3",
		},
		{
			name: "test5 basic deployment promote",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					BuildType:       "promote",
					LagoonYAML:      "internal/testdata/basic/lagoon.yml",
					ImageReferences: map[string]string{
						"node": "harbor.example/example-project/main/node@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
				}, true),
			templatePath: "testoutput",
			want:         "internal/testdata/basic/service-templates/service4",
		},
		{
			name: "test6 basic deployment pr with isolation network policy",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "pr-123",
					BuildType:       "pullrequest",
					PRNumber:        "123",
					PRHeadBranch:    "pr-head",
					PRBaseBranch:    "pr-base",
					PRHeadSHA:       "123456",
					PRBaseSHA:       "abcdef",
					LagoonYAML:      "internal/testdata/basic/lagoon.yml",
					ImageReferences: map[string]string{
						"node": "harbor.example/example-project/pr-123/node@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_ISOLATION_NETWORK_POLICY",
							Value: "enabled",
							Scope: "build",
						},
					},
				}, true),
			templatePath: "testoutput",
			want:         "internal/testdata/basic/service-templates/service5",
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

			err = LagoonServiceTemplateGeneration(generator)
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
							t.Errorf("TemplateLagoonServices() = \n%v", diff.LineDiff(string(r1), string(f1)))
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
