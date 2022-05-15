package cmd

import (
	"os"
	"testing"
)

func TestIdentifyRoute(t *testing.T) {
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
		name string
		args args
		want string
	}{
		{
			name: "test1 check LAGOON_FASTLY_SERVICE_IDS with secret no values",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "project-name1",
				environmentName: "fastly-annotations",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "fastly-annotations",
				projectVars:     `[{"name":"LAGOON_FASTLY_SERVICE_IDS","value":"annotations.com:service-id:true:annotationscom","scope":"build"}]`,
				envVars:         `[]`,
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "test-resources/template-ingress/single-lagoon.yml",
				templatePath:    "test-resources/template-ingress/output",
			},
			want: "annotations.com",
		},
		{
			name: "test2 check LAGOON_FASTLY_SERVICE_IDS no secret and no values",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "project-name2",
				environmentName: "fastly-annotations",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "fastly-annotations",
				projectVars:     `[{"name":"LAGOON_FASTLY_SERVICE_IDS","value":"annotations.com:service-id:true","scope":"build"}]`,
				envVars:         `[]`,
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "test-resources/template-ingress/single-lagoon.yml",
				templatePath:    "test-resources/template-ingress/output",
			},
			want: "annotations.com",
		},
		{
			name: "test3 check LAGOON_FASTLY_SERVICE_ID no secret and no values",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "project-name3",
				environmentName: "fastly-annotations",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "fastly-annotations",
				projectVars:     `[{"name":"LAGOON_FASTLY_SERVICE_ID","value":"service-id:true","scope":"build"}]`,
				envVars:         `[]`,
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "test-resources/template-ingress/single-lagoon.yml",
				templatePath:    "test-resources/template-ingress/output",
			},
			want: "annotations.com",
		},
		{
			name: "test4 check no fastly and no values",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "project-name7",
				environmentName: "fastly-annotations",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "fastly-annotations",
				projectVars:     `[]`,
				envVars:         `[]`,
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "test-resources/template-ingress/single-lagoon.yml",
				templatePath:    "test-resources/template-ingress/output",
			},
			want: "annotations.com",
		},
		{
			name: "test5 multiproject1 no values",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "ci-multiproject1-control-k8s",
				environmentName: "multiproject",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "multiproject",
				projectVars:     `[]`,
				envVars:         `[]`,
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "test-resources/template-ingress/polysite-lagoon.yml",
				templatePath:    "test-resources/template-ingress/output",
			},
			want: "multiproject1.com",
		},
		{
			name: "test6 multiproject2 no values",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "ci-multiproject2-control-k8s",
				environmentName: "multiproject",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "multiproject",
				projectVars:     `[]`,
				envVars:         `[]`,
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "test-resources/template-ingress/polysite-lagoon.yml",
				templatePath:    "test-resources/template-ingress/output",
			},
			want: "multiproject2.com",
		},
		{
			name: "test7 multidomain no values",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "project-name7",
				environmentName: "fastly-annotations",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "fastly-annotations",
				projectVars:     `[]`,
				envVars:         `[]`,
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "test-resources/template-ingress/multi-lagoon.yml",
				templatePath:    "test-resources/template-ingress/output",
			},
			want: "annotations.com",
		},
		{
			name: "test8 multidomain no values",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "project-name7",
				environmentName: "branch-routes",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "branch/routes",
				projectVars:     `[]`,
				envVars:         `[]`,
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "test-resources/template-ingress/multi-lagoon2.yml",
				templatePath:    "test-resources/template-ingress/output",
			},
			want: "customdomain-will-be-main-domain.com",
		},
		{
			name: "test9 active no values",
			args: args{
				alertContact:      "alertcontact",
				statusPageID:      "statuspageid",
				projectName:       "project-name7",
				environmentName:   "main",
				environmentType:   "production",
				activeEnvironment: "main",
				buildType:         "branch",
				lagoonVersion:     "v2.7.x",
				branch:            "main",
				projectVars:       `[]`,
				envVars:           `[]`,
				secretPrefix:      "fastly-api-",
				lagoonYAML:        "test-resources/template-ingress/activestandby-lagoon.yml",
				templatePath:      "test-resources/template-ingress/output",
			},
			want: "active.example.com",
		},
		{
			name: "test10 standby no values",
			args: args{
				alertContact:       "alertcontact",
				statusPageID:       "statuspageid",
				projectName:        "project-name7",
				environmentName:    "main2",
				environmentType:    "production",
				buildType:          "branch",
				standbyEnvironment: "main2",
				lagoonVersion:      "v2.7.x",
				branch:             "main2",
				projectVars:        `[]`,
				envVars:            `[]`,
				secretPrefix:       "fastly-api-",
				lagoonYAML:         "test-resources/template-ingress/activestandby-lagoon.yml",
				templatePath:       "test-resources/template-ingress/output",
			},
			want: "standby.example.com",
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

			primaryIngress, err := IdentifyPrimaryIngress(false)
			if err != nil {
				t.Errorf("%v", err)
			}

			if primaryIngress != tt.want {
				t.Errorf("returned route %v doesn't match want %v", primaryIngress, tt.want)
			}
		})
	}
}
