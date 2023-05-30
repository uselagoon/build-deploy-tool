package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/uselagoon/build-deploy-tool/internal/dbaasclient"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
)

func TestTemplateRoutes(t *testing.T) {
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
		projectVars        string
		envVars            string
		lagoonVersion      string
		lagoonYAML         string
		valuesFilePath     string
		templatePath       string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1 check LAGOON_FASTLY_SERVICE_IDS with secret no values",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "main",
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"},{"name":"LAGOON_FASTLY_SERVICE_IDS","value":"example.com:service-id:true:annotationscom","scope":"build"}]`,
				envVars:         `[]`,
				lagoonYAML:      "../test-resources/template-ingress/test1/lagoon.yml",
				templatePath:    "../test-resources/template-ingress/output",
			},
			want: "../test-resources/template-ingress/test1-results",
		},
		{
			name: "test2 check LAGOON_FASTLY_SERVICE_IDS no secret and no values",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "main",
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"},{"name":"LAGOON_FASTLY_SERVICE_IDS","value":"example.com:service-id:true","scope":"build"}]`,
				envVars:         `[]`,
				lagoonYAML:      "../test-resources/template-ingress/test2/lagoon.yml",
				templatePath:    "../test-resources/template-ingress/output",
			},
			want: "../test-resources/template-ingress/test2-results",
		},
		{
			name: "test3 check LAGOON_FASTLY_SERVICE_ID no secret and no values",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "main",
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"},{"name":"LAGOON_FASTLY_SERVICE_ID","value":"service-id:true","scope":"build"}]`,
				envVars:         `[]`,
				lagoonYAML:      "../test-resources/template-ingress/test3/lagoon.yml",
				templatePath:    "../test-resources/template-ingress/output",
			},
			want: "../test-resources/template-ingress/test3-results",
		},
		{
			name: "test4 check no fastly and no values",
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
				lagoonYAML:      "../test-resources/template-ingress/test4/lagoon.yml",
				templatePath:    "../test-resources/template-ingress/output",
			},
			want: "../test-resources/template-ingress/test4-results",
		},
		{
			name: "test5 multiproject1 no values",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "multiproject1",
				environmentName: "multiproject",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "multiproject",
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"}]`,
				envVars:         `[]`,
				lagoonYAML:      "../test-resources/template-ingress/test5/lagoon.yml",
				templatePath:    "../test-resources/template-ingress/output",
			},
			want: "../test-resources/template-ingress/test5-results",
		},
		{
			name: "test6 multiproject2 no values",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "multiproject2",
				environmentName: "multiproject",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "multiproject",
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"}]`,
				envVars:         `[]`,
				lagoonYAML:      "../test-resources/template-ingress/test6/lagoon.yml",
				templatePath:    "../test-resources/template-ingress/output",
			},
			want: "../test-resources/template-ingress/test6-results",
		},
		{
			name: "test7 multidomain no values",
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
				lagoonYAML:      "../test-resources/template-ingress/test7/lagoon.yml",
				templatePath:    "../test-resources/template-ingress/output",
			},
			want: "../test-resources/template-ingress/test7-results",
		},
		{
			name: "test8 multidomain no values",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "branch-routes",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "branch/routes",
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"}]`,
				envVars:         `[]`,
				lagoonYAML:      "../test-resources/template-ingress/test8/lagoon.yml",
				templatePath:    "../test-resources/template-ingress/output",
			},
			want: "../test-resources/template-ingress/test8-results",
		},
		{
			name: "test9 active no values",
			args: args{
				alertContact:      "alertcontact",
				statusPageID:      "statuspageid",
				projectName:       "example-project",
				environmentName:   "main",
				environmentType:   "production",
				activeEnvironment: "main",
				buildType:         "branch",
				lagoonVersion:     "v2.7.x",
				branch:            "main",
				projectVars:       `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"}]`,
				envVars:           `[]`,
				secretPrefix:      "fastly-api-",
				lagoonYAML:        "../test-resources/template-ingress/test9/lagoon.yml",
				templatePath:      "../test-resources/template-ingress/output",
			},
			want: "../test-resources/template-ingress/test9-results",
		},
		{
			name: "test10 standby no values",
			args: args{
				alertContact:       "alertcontact",
				statusPageID:       "statuspageid",
				projectName:        "example-project",
				environmentName:    "main2",
				environmentType:    "production",
				buildType:          "branch",
				standbyEnvironment: "main2",
				lagoonVersion:      "v2.7.x",
				branch:             "main2",
				projectVars:        `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"}]`,
				envVars:            `[]`,
				secretPrefix:       "fastly-api-",
				lagoonYAML:         "../test-resources/template-ingress/test10/lagoon.yml",
				templatePath:       "../test-resources/template-ingress/output",
			},
			want: "../test-resources/template-ingress/test10-results",
		},
		{
			name: "test11 standby no values",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "content-example-com",
				environmentName: "production",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "production",
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${environment}.${project}.example.com","scope":"internal_system"}]`,
				envVars:         `[]`,
				lagoonYAML:      "../test-resources/template-ingress/test11/lagoon.yml",
				templatePath:    "../test-resources/template-ingress/output",
			},
			want: "../test-resources/template-ingress/test11-results",
		},
		{
			name: "test12 check LAGOON_ROUTES_JSON generates ingress",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "main",
				projectVars:     `[{"name":"LAGOON_ROUTES_JSON","value":"eyJyb3V0ZXMiOlt7ImRvbWFpbiI6InRlc3QxLmV4YW1wbGUuY29tIiwic2VydmljZSI6Im5naW54IiwidGxzLWFjbWUiOmZhbHNlLCJtb25pdG9yaW5nLXBhdGgiOiIvYnlwYXNzLWNhY2hlIn1dfQo=","scope":"build"},{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"},{"name":"LAGOON_FASTLY_SERVICE_IDS","value":"example.com:service-id:true:annotationscom","scope":"build"}]`,
				envVars:         `[]`,
				lagoonYAML:      "../test-resources/template-ingress/test12/lagoon.yml",
				templatePath:    "../test-resources/template-ingress/output",
			},
			want: "../test-resources/template-ingress/test12-results",
		},
		{
			name: "test13 ingress class from default flag",
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
				ingressClass:    "nginx",
				lagoonYAML:      "../test-resources/template-ingress/test13/lagoon.yml",
				templatePath:    "../test-resources/template-ingress/output",
			},
			want: "../test-resources/template-ingress/test13-results",
		},
		{
			name: "test14 ingress class from lagoon.yml should overwrite default and featureflag variable",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "main",
				ingressClass:    "nginx",
				projectVars:     `[{"name":"LAGOON_FEATURE_FLAG_INGRESS_CLASS","value":"nginx","scope":"build"},{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"}]`,
				envVars:         `[]`,
				lagoonYAML:      "../test-resources/template-ingress/test14/lagoon.yml",
				templatePath:    "../test-resources/template-ingress/output",
			},
			want: "../test-resources/template-ingress/test14-results",
		},
		{
			name: "test15a ingress class from lagoon api project scope",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "main",
				ingressClass:    "nginx",
				projectVars:     `[{"name":"LAGOON_FEATURE_FLAG_INGRESS_CLASS","value":"custom-ingress","scope":"build"},{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"}]`,
				envVars:         `[]`,
				lagoonYAML:      "../test-resources/template-ingress/test15/lagoon.yml",
				templatePath:    "../test-resources/template-ingress/output",
			},
			want: "../test-resources/template-ingress/test15-results",
		},
		{
			name: "test15b ingress class from lagoon api environment scope",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "main",
				ingressClass:    "nginx",
				projectVars:     `[{"name":"LAGOON_FEATURE_FLAG_INGRESS_CLASS","value":"project-custom-ingress","scope":"build"},{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"}]`,
				envVars:         `[{"name":"LAGOON_FEATURE_FLAG_INGRESS_CLASS","value":"custom-ingress","scope":"build"}]`,
				lagoonYAML:      "../test-resources/template-ingress/test15/lagoon.yml",
				templatePath:    "../test-resources/template-ingress/output",
			},
			want: "../test-resources/template-ingress/test15-results",
		},
		{
			name: "test16 hsts basic",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "main",
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"},{"name":"LAGOON_FASTLY_SERVICE_IDS","value":"example.com:service-id:true:annotationscom","scope":"build"}]`,
				envVars:         `[]`,
				lagoonYAML:      "../test-resources/template-ingress/test16/lagoon.yml",
				templatePath:    "../test-resources/template-ingress/output",
			},
			want: "../test-resources/template-ingress/test16-results",
		},
		{
			name: "test17 hsts advanced",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "main",
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"},{"name":"LAGOON_FASTLY_SERVICE_IDS","value":"example.com:service-id:true:annotationscom","scope":"build"}]`,
				envVars:         `[]`,
				lagoonYAML:      "../test-resources/template-ingress/test17/lagoon.yml",
				templatePath:    "../test-resources/template-ingress/output",
			},
			want: "../test-resources/template-ingress/test17-results",
		},
		{
			name: "test18 check first route has monitoring only",
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
				lagoonYAML:      "../test-resources/template-ingress/test18/lagoon.yml",
				templatePath:    "../test-resources/template-ingress/output",
			},
			want: "../test-resources/template-ingress/test18-results",
		},
		{
			name: "test19 pullrequest routes",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "pr-4841",
				environmentType: "development",
				buildType:       "pullrequest",
				lagoonVersion:   "v2.7.x",
				prNumber:        "4841",
				prHeadBranch:    "main",
				prBaseBranch:    "my-branch",
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"}]`,
				envVars:         `[]`,
				lagoonYAML:      "../test-resources/template-ingress/test19/lagoon.yml",
				templatePath:    "../test-resources/template-ingress/output",
			},
			want: "../test-resources/template-ingress/test19-results",
		},
		{
			name: "test20 pullrequest routes polysite",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "pr-4841",
				environmentType: "development",
				buildType:       "pullrequest",
				lagoonVersion:   "v2.7.x",
				prNumber:        "4841",
				prHeadBranch:    "main",
				prBaseBranch:    "my-branch",
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"}]`,
				envVars:         `[]`,
				lagoonYAML:      "../test-resources/template-ingress/test20/lagoon.yml",
				templatePath:    "../test-resources/template-ingress/output",
			},
			want: "../test-resources/template-ingress/test20-results",
		},
		{
			name: "test21 alternative names",
			args: args{
				alertContact:       "alertcontact",
				statusPageID:       "statuspageid",
				projectName:        "example-project",
				environmentName:    "main",
				environmentType:    "production",
				buildType:          "branch",
				standbyEnvironment: "main2",
				lagoonVersion:      "v2.7.x",
				branch:             "main",
				projectVars:        `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"}]`,
				envVars:            `[]`,
				secretPrefix:       "fastly-api-",
				lagoonYAML:         "../test-resources/template-ingress/test21/lagoon.yml",
				templatePath:       "../test-resources/template-ingress/output",
			},
			want: "../test-resources/template-ingress/test21-results",
		},
		{
			name: "test22 check wildcard",
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
				lagoonYAML:      "../test-resources/template-ingress/test22/lagoon.yml",
				templatePath:    "../test-resources/template-ingress/output",
			},
			want: "../test-resources/template-ingress/test22-results",
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
			generator, err := generatorInput(false)
			if err != nil {
				t.Errorf("%v", err)
			}
			generator.LagoonYAML = tt.args.lagoonYAML
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

			err = IngressTemplateGeneration(generator)
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
