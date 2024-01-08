package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/uselagoon/build-deploy-tool/internal/dbaasclient"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	"github.com/uselagoon/build-deploy-tool/internal/testdata"
)

func TestTemplateLagoonServices(t *testing.T) {
	type args struct {
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
		ingressClass       string
		configMapSha       string
		rootlessWorkloads  string
		projectVars        string
		envVars            string
		lagoonVersion      string
		lagoonYAML         string
		valuesFilePath     string
		templatePath       string
		imageReferences    map[string]string
	}
	tests := []struct {
		name         string
		description  string
		args         args
		args2        testdata.TestData
		templatePath string
		want         string
	}{
		{
			name: "test1 basic deployment",
			args2: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/basic/lagoon.yml",
					ImageReferences: map[string]string{
						"node": "harbor.example/example-project/main/node:latest",
					},
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/basic/service-templates/service1",
		},
		{
			name: "test2a nginx-php deployment",
			args2: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/complex/lagoon.varnish.yml",
					ImageReferences: map[string]string{
						"nginx":   "harbor.example/example-project/main/nginx:latest",
						"php":     "harbor.example/example-project/main/php:latest",
						"cli":     "harbor.example/example-project/main/cli:latest",
						"redis":   "harbor.example/example-project/main/redis:latest",
						"varnish": "harbor.example/example-project/main/varnish:latest",
					},
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/complex/service-templates/service1",
		},
		{
			name: "test2b nginx-php deployment - rootless",
			args2: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/complex/lagoon.varnish.yml",
					ImageReferences: map[string]string{
						"nginx":   "harbor.example/example-project/main/nginx:latest",
						"php":     "harbor.example/example-project/main/php:latest",
						"cli":     "harbor.example/example-project/main/cli:latest",
						"redis":   "harbor.example/example-project/main/redis:latest",
						"varnish": "harbor.example/example-project/main/varnish:latest",
					},
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_ROOTLESS_WORKLOAD",
							Value: "enabled",
							Scope: "build",
						},
					},
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/complex/service-templates/service2",
		},
		{
			name:        "test3 - funky pvcs",
			description: "only create pvcs of the requested persistent-name in the docker-compose file",
			args2: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/basic/lagoon.thunderhub.yml",
					ImageReferences: map[string]string{
						"lnd":        "harbor.example/example-project/main/lnd:latest",
						"thunderhub": "harbor.example/example-project/main/thunderhub:latest",
						"tor":        "harbor.example/example-project/main/tor:latest",
					},
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_ROOTLESS_WORKLOAD",
							Value: "enabled",
							Scope: "build",
						},
					},
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/basic/service-templates/service2",
		},
		{
			name:        "test4 - basic-persistent with worker-persistent",
			description: "create a basic-persistent that gets a pvc and mount that volume on a worker-persistent type",
			// args: args{
			// 	alertContact:      "alertcontact",
			// 	statusPageID:      "statuspageid",
			// 	projectName:       "example-project",
			// 	environmentName:   "main",
			// 	environmentType:   "production",
			// 	buildType:         "branch",
			// 	lagoonVersion:     "v2.7.x",
			// 	branch:            "main",
			// 	projectVars:       `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"}]`,
			// 	envVars:           `[]`,
			// 	rootlessWorkloads: "enabled",
			// 	lagoonYAML:        "../test-resources/template-lagoon-services/test4/lagoon.yml",
			// 	templatePath:      "../test-resources/template-lagoon-services/output",
			// 	configMapSha:      "abcdefg1234567890",
			// 	imageReferences: map[string]string{
			// 		"lnd": "harbor.example/example-project/main/lnd:latest",
			// 		"tor": "harbor.example/example-project/main/tor:latest",
			// 	},
			// },
			// templatePath: "testdata/output",
			// want:         "../test-resources/template-lagoon-services/test4-results",
			args2: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "../internal/testdata/basic/lagoon.thunderhub-2.yml",
					ImageReferences: map[string]string{
						"lnd": "harbor.example/example-project/main/lnd:latest",
						"tor": "harbor.example/example-project/main/tor:latest",
					},
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_ROOTLESS_WORKLOAD",
							Value: "enabled",
							Scope: "build",
						},
					},
				}, true),
			templatePath: "testdata/output",
			want:         "../internal/testdata/basic/service-templates/service3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// set the environment variables from args
			savedTemplates := tt.templatePath
			generator, err := testdata.SetupEnvironment(*rootCmd, savedTemplates, tt.args2)
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
				helpers.UnsetEnvVars(nil)
			})
		})
	}
}
