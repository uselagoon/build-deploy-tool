package cmd

import (
	"os"
	"testing"

	"github.com/uselagoon/build-deploy-tool/internal/helpers"
)

func TestIdentifyFeatureFlag(t *testing.T) {
	type args struct {
		name               string
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
		name    string
		args    args
		vars    []helpers.EnvironmentVariable
		want    string
		wantErr bool
	}{
		{
			name: "test1 check if flag is defined in lagoon project variables",
			args: args{
				name:            "ROOTLESS_WORKLOAD",
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "main",
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"},{"name":"LAGOON_FEATURE_FLAG_ROOTLESS_WORKLOAD","value":"enabled","scope":"build"}]`,
				envVars:         `[]`,
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "../test-resources/identify-feature/alltest/lagoon.yml",
				templatePath:    "../test-resources/output",
			},
			want: "enabled",
		},
		{
			name: "test2 check if flag is defined in lagoon environment variables",
			args: args{
				name:            "ROOTLESS_WORKLOAD",
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "main",
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"}]`,
				envVars:         `[{"name":"LAGOON_FEATURE_FLAG_ROOTLESS_WORKLOAD","value":"enabled","scope":"build"}]`,
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "../test-resources/identify-feature/alltest/lagoon.yml",
				templatePath:    "../test-resources/output",
			},
			want: "enabled",
		},
		{
			name: "test3 check if force flag is defined in build variables",
			args: args{
				name:            "ROOTLESS_WORKLOAD",
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
				lagoonYAML:      "../test-resources/identify-feature/alltest/lagoon.yml",
				templatePath:    "../test-resources/output",
			},
			vars: []helpers.EnvironmentVariable{
				{
					Name:  "LAGOON_FEATURE_FLAG_FORCE_ROOTLESS_WORKLOAD",
					Value: "enabled",
				},
			},
			want: "enabled",
		},
		{
			name: "test4 check if force flag is defined in build variables and default flag is ignored",
			args: args{
				name:            "ROOTLESS_WORKLOAD",
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
				lagoonYAML:      "../test-resources/identify-feature/alltest/lagoon.yml",
				templatePath:    "../test-resources/output",
			},
			vars: []helpers.EnvironmentVariable{
				{
					Name:  "LAGOON_FEATURE_FLAG_FORCE_ROOTLESS_WORKLOAD",
					Value: "enabled",
				},
				{
					Name:  "LAGOON_FEATURE_FLAG_DEFAULT_ROOTLESS_WORKLOAD",
					Value: "disabled",
				},
			},
			want: "enabled",
		},
		{
			name: "test5 check if force flag is defined in build variables and one defined in lagoon project variables is ignored",
			args: args{
				name:            "ROOTLESS_WORKLOAD",
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "main",
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"},{"name":"LAGOON_FEATURE_FLAG_ROOTLESS_WORKLOAD","value":"enabled","scope":"build"}]`,
				envVars:         `[]`,
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "../test-resources/identify-feature/alltest/lagoon.yml",
				templatePath:    "../test-resources/output",
			},
			vars: []helpers.EnvironmentVariable{
				{
					Name:  "LAGOON_FEATURE_FLAG_FORCE_ROOTLESS_WORKLOAD",
					Value: "disabled",
				},
				{
					Name:  "LAGOON_FEATURE_FLAG_DEFAULT_ROOTLESS_WORKLOAD",
					Value: "disabled",
				},
			},
			want: "disabled",
		},
		{
			name: "test6 check if default flag is ignored and lagoon project variable is used",
			args: args{
				name:            "ROOTLESS_WORKLOAD",
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "main",
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"},{"name":"LAGOON_FEATURE_FLAG_ROOTLESS_WORKLOAD","value":"enabled","scope":"build"}]`,
				envVars:         `[]`,
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "../test-resources/identify-feature/alltest/lagoon.yml",
				templatePath:    "../test-resources/output",
			},
			vars: []helpers.EnvironmentVariable{
				{
					Name:  "LAGOON_FEATURE_FLAG_DEFAULT_ROOTLESS_WORKLOAD",
					Value: "disabled",
				},
			},
			want: "enabled",
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
			lagoonYml = tt.args.lagoonYAML
			templateValues = tt.args.valuesFilePath

			savedTemplates = tt.args.templatePath
			fastlyAPISecretPrefix = tt.args.secretPrefix
			fastlyServiceID = tt.args.serviceID

			for _, envVar := range tt.vars {
				err = os.Setenv(envVar.Name, envVar.Value)
				if err != nil {
					t.Errorf("%v", err)
				}
			}
			got, err := IdentifyFeatureFlag(generatorInput(false), tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("IdentifyFeatureFlag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IdentifyFeatureFlag() = %v, want %v", got, tt.want)
			}
			t.Cleanup(func() {
				helpers.UnsetEnvVars(tt.vars)
			})
		})
	}
}
