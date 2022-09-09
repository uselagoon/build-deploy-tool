package cmd

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/uselagoon/build-deploy-tool/internal/helpers"
)

func TestAutogeneratedIngressGeneration(t *testing.T) {
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
		projectVars        string
		envVars            string
		lagoonVersion      string
		lagoonYAML         string
		valuesFilePath     string
		templatePath       string
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		emptyDir bool // if no templates are generated, then there will be a .gitkeep file in there
		want     string
	}{
		{
			name: "test1 autogenerated route",
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
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "../test-resources/template-autogenerated/test1/lagoon.yml",
				templatePath:    "../test-resources/template-autogenerated/output",
			},
			emptyDir: false,
			want:     "../test-resources/template-autogenerated/test1-results",
		},
		{
			name: "test2 no autogenerated routes but allow pullrequests",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				environmentType: "production",
				buildType:       "pullrequest",
				prNumber:        "123",
				prHeadBranch:    "main",
				prBaseBranch:    "main2",
				lagoonVersion:   "v2.7.x",
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"}]`,
				envVars:         `[]`,
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "../test-resources/template-autogenerated/test2/lagoon.yml",
				templatePath:    "../test-resources/template-autogenerated/output",
			},
			emptyDir: false,
			want:     "../test-resources/template-autogenerated/test2-results",
		},
		{
			name: "test3 autogenerated route but no pullrequests",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				environmentType: "production",
				buildType:       "pullrequest",
				prNumber:        "123",
				prHeadBranch:    "main",
				prBaseBranch:    "main2",
				lagoonVersion:   "v2.7.x",
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"}]`,
				envVars:         `[]`,
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "../test-resources/template-autogenerated/test3/lagoon.yml",
				templatePath:    "../test-resources/template-autogenerated/output",
			},
			emptyDir: true,
			want:     "",
		},
		{
			name: "test4 autogenerated route no service in pattern",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "main",
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${project}-${environment}.example.com","scope":"internal_system"}]`,
				envVars:         `[]`,
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "../test-resources/template-autogenerated/test4/lagoon.yml",
				templatePath:    "../test-resources/template-autogenerated/output",
			},
			emptyDir: false,
			want:     "../test-resources/template-autogenerated/test4-results",
		},
		{
			name: "test5 autogenerated route short url",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "ci-features-control-k8s",
				environmentName: "short-router-url-from-a-very-l-ebe8",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "short-router-url-from-a-very-long-environment-name-like-this",
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}.${project}-${environment}.example.com","scope":"internal_system"}]`,
				envVars:         `[]`,
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "../test-resources/template-autogenerated/test5/lagoon.yml",
				templatePath:    "../test-resources/template-autogenerated/output",
			},
			emptyDir: false,
			want:     "../test-resources/template-autogenerated/test5-results",
		},
		{
			name: "test6 autogenerated routes but disabled by service label",
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
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "../test-resources/template-autogenerated/test6/lagoon.yml",
				templatePath:    "../test-resources/template-autogenerated/output",
			},
			emptyDir: true,
			want:     "",
		},
		{
			name: "test7 no autogenerated routes but enabled by service label",
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
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "../test-resources/template-autogenerated/test7/lagoon.yml",
				templatePath:    "../test-resources/template-autogenerated/output",
			},
			emptyDir: false,
			want:     "../test-resources/template-autogenerated/test7-results",
		},
		{
			name: "test8 autogenerated routes with fastly",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "main",
				projectVars: `[
					{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"},
					{"name":"LAGOON_FASTLY_SERVICE_ID","value":"service-id:true","scope":"global"},
					{"name":"LAGOON_FASTLY_AUTOGENERATED","value":"enabled","scope":"global"}
					]`,
				envVars:      `[]`,
				secretPrefix: "fastly-api-",
				lagoonYAML:   "../test-resources/template-autogenerated/test8/lagoon.yml",
				templatePath: "../test-resources/template-autogenerated/output",
			},
			emptyDir: false,
			want:     "../test-resources/template-autogenerated/test8-results",
		},
		{
			name: "test9 autogenerated routes with fastly specific domain",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "main",
				projectVars: `[
					{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"},
					{"name":"LAGOON_FASTLY_SERVICE_IDS","value":"node-example-project-main.example.com:service-id:true","scope":"global"},
					{"name":"LAGOON_FASTLY_AUTOGENERATED","value":"enabled","scope":"global"}
					]`,
				envVars:      `[]`,
				secretPrefix: "fastly-api-",
				lagoonYAML:   "../test-resources/template-autogenerated/test9/lagoon.yml",
				templatePath: "../test-resources/template-autogenerated/output",
			},
			emptyDir: false,
			want:     "../test-resources/template-autogenerated/test9-results",
		},
		{
			name: "test10 autogenerated routes with fastly and specific secret",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "main",
				projectVars: `[
					{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"},
					{"name":"LAGOON_FASTLY_SERVICE_ID","value":"service-id:true:secretname","scope":"global"},
					{"name":"LAGOON_FEATURE_FLAG_FASTLY_AUTOGENERATED","value":"enabled","scope":"global"}
					]`,
				envVars:      `[]`,
				secretPrefix: "fastly-api-",
				lagoonYAML:   "../test-resources/template-autogenerated/test10/lagoon.yml",
				templatePath: "../test-resources/template-autogenerated/output",
			},
			emptyDir: false,
			want:     "../test-resources/template-autogenerated/test10-results",
		},
		{
			name: "test11 autogenerated route development environment",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				environmentType: "development",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "main",
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"}]`,
				envVars:         `[]`,
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "../test-resources/template-autogenerated/test11/lagoon.yml",
				templatePath:    "../test-resources/template-autogenerated/output",
			},
			emptyDir: false,
			want:     "../test-resources/template-autogenerated/test11-results",
		},
		{
			name: "test12 autogenerated route development environment",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				environmentType: "development",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "main",
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"}]`,
				envVars:         `[]`,
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "../test-resources/template-autogenerated/test12/lagoon.yml",
				templatePath:    "../test-resources/template-autogenerated/output",
			},
			emptyDir: false,
			want:     "../test-resources/template-autogenerated/test12-results",
		},
		{
			name: "test13 autogenerated route development service type override",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				environmentType: "development",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "main",
				projectVars: `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"},
				{"name":"LAGOON_SERVICE_TYPES","value":"node:node-persistent","scope":"build"}]`,
				envVars:      `[]`,
				secretPrefix: "fastly-api-",
				lagoonYAML:   "../test-resources/template-autogenerated/test13/lagoon.yml",
				templatePath: "../test-resources/template-autogenerated/output",
			},
			emptyDir: false,
			want:     "../test-resources/template-autogenerated/test13-results",
		},
		{
			name: "test14 autogenerated route development no service type override",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				environmentType: "development",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "main",
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"}]`,
				envVars:         `[]`,
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "../test-resources/template-autogenerated/test14/lagoon.yml",
				templatePath:    "../test-resources/template-autogenerated/output",
			},
			emptyDir: true,
			want:     "",
		},
		{
			name: "test15 autogenerated route development service type override",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				environmentType: "development",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "main",
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"}]`,
				envVars:         `[]`,
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "../test-resources/template-autogenerated/test15/lagoon.yml",
				templatePath:    "../test-resources/template-autogenerated/output",
			},
			emptyDir: false,
			want:     "../test-resources/template-autogenerated/test15-results",
		},
		{
			name: "test16 autogenerated route development service type override",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				environmentType: "development",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "main",
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"}]`,
				envVars:         `[]`,
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "../test-resources/template-autogenerated/test16/lagoon.yml",
				templatePath:    "../test-resources/template-autogenerated/output",
			},
			emptyDir: false,
			want:     "../test-resources/template-autogenerated/test16-results",
		},
		{
			name: "test17 autogenerated route development service type override",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "ci-drush-la-control-k8s",
				environmentName: "drush-first",
				environmentType: "development",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "drush-first",
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}.${project}.${environment}.172.18.0.3.nip.io","scope":"internal_system"}]`,
				envVars:         `[]`,
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "../test-resources/template-autogenerated/test17/lagoon.yml",
				templatePath:    "../test-resources/template-autogenerated/output",
			},
			emptyDir: false,
			want:     "../test-resources/template-autogenerated/test17-results",
		},
		{
			name: "test18 autogenerated route tls-acme disabled",
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
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "../test-resources/template-autogenerated/test18/lagoon.yml",
				templatePath:    "../test-resources/template-autogenerated/output",
			},
			emptyDir: false,
			want:     "../test-resources/template-autogenerated/test18-results",
		},
		{
			name: "test19 autogenerated routes but tls-acme disabled by service label",
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
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "../test-resources/template-autogenerated/test19/lagoon.yml",
				templatePath:    "../test-resources/template-autogenerated/output",
			},
			emptyDir: false,
			want:     "../test-resources/template-autogenerated/test19-results",
		},
		{
			name: "test20 autogenerated routes where lagoon.name of service does not match service names",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "content-example-com",
				environmentName: "feature-migration",
				environmentType: "development",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "feature/migration",
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${environment}.${project}.example.com","scope":"internal_system"}]`,
				envVars:         `[]`,
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "../test-resources/template-autogenerated/test20/lagoon.yml",
				templatePath:    "../test-resources/template-autogenerated/output",
			},
			emptyDir: false,
			want:     "../test-resources/template-autogenerated/test20-results",
		},
		{
			name: "test21 autogenerated routes where docker-compose env_file has missing file references",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "test21-example-com",
				environmentName: "feature",
				environmentType: "development",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "feature",
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${environment}.${project}.example.com","scope":"internal_system"}]`,
				envVars:         `[]`,
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "../test-resources/template-autogenerated/test21/lagoon.yml",
				templatePath:    "../test-resources/template-autogenerated/output",
			},
			emptyDir: false,
			want:     "../test-resources/template-autogenerated/test21-results",
		},
		{
			name: "test22 autogenerated routes where should truncate long dns",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "content-abcdefghijk-net-com-co",
				environmentName: "pr-123",
				environmentType: "development",
				lagoonVersion:   "v2.7.x",
				buildType:       "pullrequest",
				prNumber:        "123",
				prHeadBranch:    "main",
				prBaseBranch:    "main2",
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${environment}.${project}.abc1.abc.net.com.co","scope":"internal_system"}]`,
				envVars:         `[]`,
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "../test-resources/template-autogenerated/test22/lagoon.yml",
				templatePath:    "../test-resources/template-autogenerated/output",
			},
			emptyDir: false,
			want:     "../test-resources/template-autogenerated/test22-results",
		},
		{
			name: "test23 autogenerated routes with fastly service, should be no fastly annotations on autogenerated route",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "main",
				projectVars: `[
					{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"},
					{"name":"LAGOON_FASTLY_SERVICE_ID","value":"service-id:true","scope":"global"}
					]`,
				envVars:      `[]`,
				secretPrefix: "fastly-api-",
				lagoonYAML:   "../test-resources/template-autogenerated/test23/lagoon.yml",
				templatePath: "../test-resources/template-autogenerated/output",
			},
			emptyDir: false,
			want:     "../test-resources/template-autogenerated/test23-results",
		},
		{
			name: "test24 autogenerated routes with fastly service, should be no fastly annotations on autogenerated route",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "main",
				projectVars: `[
					{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"},
					{"name":"LAGOON_FASTLY_SERVICE_ID","value":"service-id:true","scope":"global"}
					]`,
				envVars:      `[]`,
				secretPrefix: "fastly-api-",
				lagoonYAML:   "../test-resources/template-autogenerated/test24/lagoon.yml",
				templatePath: "../test-resources/template-autogenerated/output",
			},
			emptyDir: false,
			want:     "../test-resources/template-autogenerated/test24-results",
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
			generator := generatorInput(false)
			generator.LagoonYAML = tt.args.lagoonYAML
			generator.FastlyAPISecretPrefix = tt.args.secretPrefix
			generator.SavedTemplatesPath = tt.args.templatePath

			savedTemplates := tt.args.templatePath
			err = os.MkdirAll(tt.args.templatePath, 0755)
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
