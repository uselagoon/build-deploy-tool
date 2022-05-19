package cmd

import (
	"os"
	"reflect"
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
		name        string
		args        args
		want        string
		wantRemain  []string
		wantautoGen []string
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
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "test-resources/template-ingress/single-lagoon.yml",
				templatePath:    "test-resources/template-ingress/output",
			},
			want:        "https://example.com",
			wantRemain:  []string{"https://node-example-project-main.example.com"},
			wantautoGen: []string{"https://node-example-project-main.example.com"},
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
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "test-resources/template-ingress/single-lagoon.yml",
				templatePath:    "test-resources/template-ingress/output",
			},
			want:        "https://example.com",
			wantRemain:  []string{"https://node-example-project-main.example.com"},
			wantautoGen: []string{"https://node-example-project-main.example.com"},
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
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "test-resources/template-ingress/single-lagoon.yml",
				templatePath:    "test-resources/template-ingress/output",
			},
			want:        "https://example.com",
			wantRemain:  []string{"https://node-example-project-main.example.com"},
			wantautoGen: []string{"https://node-example-project-main.example.com"},
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
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "test-resources/template-ingress/single-lagoon.yml",
				templatePath:    "test-resources/template-ingress/output",
			},
			want:        "https://example.com",
			wantRemain:  []string{"https://node-example-project-main.example.com"},
			wantautoGen: []string{"https://node-example-project-main.example.com"},
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
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "test-resources/template-ingress/polysite-lagoon.yml",
				templatePath:    "test-resources/template-ingress/output",
			},
			want:        "https://multiproject1.com",
			wantRemain:  []string{"https://node-multiproject1-multiproject.example.com"},
			wantautoGen: []string{"https://node-multiproject1-multiproject.example.com"},
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
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "test-resources/template-ingress/polysite-lagoon.yml",
				templatePath:    "test-resources/template-ingress/output",
			},
			want:        "https://multiproject2.com",
			wantRemain:  []string{"https://node-multiproject2-multiproject.example.com"},
			wantautoGen: []string{"https://node-multiproject2-multiproject.example.com"},
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
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "test-resources/template-ingress/multi-lagoon.yml",
				templatePath:    "test-resources/template-ingress/output",
			},
			want:        "https://example.com",
			wantRemain:  []string{"https://node-example-project-main.example.com", "https://www.example.com"},
			wantautoGen: []string{"https://node-example-project-main.example.com"},
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
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "test-resources/template-ingress/multi-lagoon2.yml",
				templatePath:    "test-resources/template-ingress/output",
			},
			want:        "https://customdomain-will-be-main-domain.com",
			wantRemain:  []string{"https://node-example-project-branch-routes.example.com", "https://customdomain-will-be-not-be-main-domain.com"},
			wantautoGen: []string{"https://node-example-project-branch-routes.example.com"},
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
				lagoonYAML:        "test-resources/template-ingress/activestandby-lagoon.yml",
				templatePath:      "test-resources/template-ingress/output",
			},
			want:        "https://active.example.com",
			wantRemain:  []string{"https://node-example-project-main.example.com", "https://main.example.com"},
			wantautoGen: []string{"https://node-example-project-main.example.com"},
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
				lagoonYAML:         "test-resources/template-ingress/activestandby-lagoon.yml",
				templatePath:       "test-resources/template-ingress/output",
			},
			want:        "https://standby.example.com",
			wantRemain:  []string{"https://node-example-project-main2.example.com", "https://main2.example.com"},
			wantautoGen: []string{"https://node-example-project-main2.example.com"},
		},
		{
			name: "test11 no custom ingress",
			args: args{
				alertContact:       "alertcontact",
				statusPageID:       "statuspageid",
				projectName:        "example-project",
				environmentName:    "no-ingress",
				environmentType:    "production",
				buildType:          "branch",
				standbyEnvironment: "no-ingress",
				lagoonVersion:      "v2.7.x",
				branch:             "main2",
				projectVars:        `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"}]`,
				envVars:            `[]`,
				secretPrefix:       "fastly-api-",
				lagoonYAML:         "test-resources/template-ingress/noingress-lagoon.yml",
				templatePath:       "test-resources/template-ingress/output",
			},
			want:        "https://node-example-project-no-ingress.example.com",
			wantRemain:  []string{},
			wantautoGen: []string{"https://node-example-project-no-ingress.example.com"},
		},
		{
			name: "test12 no custom ingress",
			args: args{
				alertContact:       "alertcontact",
				statusPageID:       "statuspageid",
				projectName:        "example-project",
				environmentName:    "no-ingress",
				environmentType:    "production",
				buildType:          "branch",
				standbyEnvironment: "no-ingress",
				lagoonVersion:      "v2.7.x",
				branch:             "main2",
				projectVars:        `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"}]`,
				envVars:            `[]`,
				secretPrefix:       "fastly-api-",
				lagoonYAML:         "test-resources/template-ingress/prefixes-lagoon.yml",
				templatePath:       "test-resources/template-ingress/output",
			},
			want:       "https://node-example-project-no-ingress.example.com",
			wantRemain: []string{},
			wantautoGen: []string{
				"https://node-example-project-no-ingress.example.com",
				"https://www.node-example-project-no-ingress.example.com",
				"https://en.node-example-project-no-ingress.example.com",
				"https://de.node-example-project-no-ingress.example.com",
				"https://fi.node-example-project-no-ingress.example.com",
			},
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

			primary, remainders, autogen, err := IdentifyPrimaryIngress(false)
			if err != nil {
				t.Errorf("%v", err)
			}

			if primary != tt.want {
				t.Errorf("returned primary %v doesn't match want %v", primary, tt.want)
			}

			if !reflect.DeepEqual(remainders, tt.wantRemain) {
				t.Errorf("returned remainders %v doesn't match want %v", remainders, tt.wantRemain)
			}

			if !reflect.DeepEqual(autogen, tt.wantautoGen) {
				t.Errorf("returned autogen %v doesn't match want %v", autogen, tt.wantautoGen)
			}
		})
	}
}
