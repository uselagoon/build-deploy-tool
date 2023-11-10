package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/andreyvit/diff"
	"github.com/uselagoon/build-deploy-tool/internal/dbaasclient"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
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
		name        string
		description string
		args        args
		want        string
	}{
		{
			name: "test1 basic deployment",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "main",
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"}]`,
				envVars:         `[]`,
				lagoonYAML:      "../test-resources/template-lagoon-services/test1/lagoon.yml",
				templatePath:    "../test-resources/template-lagoon-services/output",
				configMapSha:    "abcdefg1234567890",
				imageReferences: map[string]string{
					"node": "harbor.example/example-project/main/node:latest",
				},
			},
			want: "../test-resources/template-lagoon-services/test1-results",
		},
		{
			name: "test2a nginx-php deployment",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "main",
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"}]`,
				envVars:         `[]`,
				lagoonYAML:      "../test-resources/template-lagoon-services/test2/lagoon.yml",
				templatePath:    "../test-resources/template-lagoon-services/output",
				configMapSha:    "abcdefg1234567890",
				imageReferences: map[string]string{
					"nginx":   "harbor.example/example-project/main/nginx:latest",
					"php":     "harbor.example/example-project/main/php:latest",
					"cli":     "harbor.example/example-project/main/cli:latest",
					"redis":   "harbor.example/example-project/main/redis:latest",
					"varnish": "harbor.example/example-project/main/varnish:latest",
				},
			},
			want: "../test-resources/template-lagoon-services/test2-results-a",
		},
		{
			name: "test2b nginx-php deployment - rootless",
			args: args{
				alertContact:      "alertcontact",
				statusPageID:      "statuspageid",
				projectName:       "example-project",
				environmentName:   "main",
				environmentType:   "production",
				buildType:         "branch",
				lagoonVersion:     "v2.7.x",
				branch:            "main",
				projectVars:       `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"}]`,
				envVars:           `[]`,
				rootlessWorkloads: "enabled",
				lagoonYAML:        "../test-resources/template-lagoon-services/test2/lagoon.yml",
				templatePath:      "../test-resources/template-lagoon-services/output",
				configMapSha:      "abcdefg1234567890",
				imageReferences: map[string]string{
					"nginx":   "harbor.example/example-project/main/nginx:latest",
					"php":     "harbor.example/example-project/main/php:latest",
					"cli":     "harbor.example/example-project/main/cli:latest",
					"redis":   "harbor.example/example-project/main/redis:latest",
					"varnish": "harbor.example/example-project/main/varnish:latest",
				},
			},
			want: "../test-resources/template-lagoon-services/test2-results-b",
		},
		{
			name:        "test3 - funky pvcs",
			description: "only create pvcs of the requested persistent-name in the docker-compose file",
			args: args{
				alertContact:      "alertcontact",
				statusPageID:      "statuspageid",
				projectName:       "example-project",
				environmentName:   "main",
				environmentType:   "production",
				buildType:         "branch",
				lagoonVersion:     "v2.7.x",
				branch:            "main",
				projectVars:       `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"}]`,
				envVars:           `[]`,
				rootlessWorkloads: "enabled",
				lagoonYAML:        "../test-resources/template-lagoon-services/test3/lagoon.yml",
				templatePath:      "../test-resources/template-lagoon-services/output",
				configMapSha:      "abcdefg1234567890",
				imageReferences: map[string]string{
					"lnd":        "harbor.example/example-project/main/lnd:latest",
					"thunderhub": "harbor.example/example-project/main/thunderhub:latest",
					"tor":        "harbor.example/example-project/main/tor:latest",
				},
			},
			want: "../test-resources/template-lagoon-services/test3-results",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// set the environment variables from args
			err := os.Setenv("MONITORING_ALERTCONTACT", tt.args.alertContact)
			if err != nil {
				t.Errorf("%v", err)
			}
			err = os.Setenv("MONITORING_STATUSPAGEID", tt.args.statusPageID)
			if err != nil {
				t.Errorf("%v", err)
			}
			err = os.Setenv("PROJECT", tt.args.projectName)
			if err != nil {
				t.Errorf("%v", err)
			}
			err = os.Setenv("ENVIRONMENT", tt.args.environmentName)
			if err != nil {
				t.Errorf("%v", err)
			}
			err = os.Setenv("BRANCH", tt.args.branch)
			if err != nil {
				t.Errorf("%v", err)
			}
			err = os.Setenv("PR_NUMBER", tt.args.prNumber)
			if err != nil {
				t.Errorf("%v", err)
			}
			err = os.Setenv("PR_HEAD_BRANCH", tt.args.prHeadBranch)
			if err != nil {
				t.Errorf("%v", err)
			}
			err = os.Setenv("PR_BASE_BRANCH", tt.args.prBaseBranch)
			if err != nil {
				t.Errorf("%v", err)
			}
			err = os.Setenv("ENVIRONMENT_TYPE", tt.args.environmentType)
			if err != nil {
				t.Errorf("%v", err)
			}
			err = os.Setenv("BUILD_TYPE", tt.args.buildType)
			if err != nil {
				t.Errorf("%v", err)
			}
			err = os.Setenv("ACTIVE_ENVIRONMENT", tt.args.activeEnvironment)
			if err != nil {
				t.Errorf("%v", err)
			}
			err = os.Setenv("STANDBY_ENVIRONMENT", tt.args.standbyEnvironment)
			if err != nil {
				t.Errorf("%v", err)
			}
			err = os.Setenv("CONFIG_MAP_SHA", tt.args.configMapSha)
			if err != nil {
				t.Errorf("%v", err)
			}
			err = os.Setenv("LAGOON_FASTLY_NOCACHE_SERVICE_ID", tt.args.cacheNoCache)
			if err != nil {
				t.Errorf("%v", err)
			}
			err = os.Setenv("LAGOON_PROJECT_VARIABLES", tt.args.projectVars)
			if err != nil {
				t.Errorf("%v", err)
			}
			err = os.Setenv("LAGOON_ENVIRONMENT_VARIABLES", tt.args.envVars)
			if err != nil {
				t.Errorf("%v", err)
			}
			err = os.Setenv("LAGOON_VERSION", tt.args.lagoonVersion)
			if err != nil {
				t.Errorf("%v", err)
			}
			err = os.Setenv("LAGOON_FEATURE_FLAG_DEFAULT_INGRESS_CLASS", tt.args.ingressClass)
			if err != nil {
				t.Errorf("%v", err)
			}
			err = os.Setenv("LAGOON_FEATURE_FLAG_DEFAULT_ROOTLESS_WORKLOAD", tt.args.rootlessWorkloads)
			if err != nil {
				t.Errorf("%v", err)
			}
			generator, err := generatorInput(false)
			if err != nil {
				t.Errorf("%v", err)
			}
			generator.LagoonYAML = tt.args.lagoonYAML
			generator.ImageReferences = tt.args.imageReferences
			generator.SavedTemplatesPath = tt.args.templatePath
			// add dbaasclient overrides for tests
			generator.DBaaSClient = dbaasclient.NewClient(dbaasclient.Client{
				RetryMax:     5,
				RetryWaitMin: time.Duration(10) * time.Millisecond,
				RetryWaitMax: time.Duration(50) * time.Millisecond,
			})

			savedTemplates := tt.args.templatePath
			err = os.MkdirAll(tt.args.templatePath, 0755)
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
				helpers.UnsetEnvVars([]helpers.EnvironmentVariable{{Name: "LAGOON_FEATURE_FLAG_DEFAULT_INGRESS_CLASS"}, {Name: "LAGOON_FEATURE_FLAG_ROOTLESS_WORKLOAD"}, {Name: "CONFIG_MAP_SHA"}})
			})
		})
	}
}
