package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
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
		projectVars        string
		envVars            string
		lagoonVersion      string
		lagoonYAML         string
		valuesFilePath     string
		checkValuesFile    bool
		templatePath       string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1 check LAGOON_FASTLY_SERVICE_IDS with secret and values",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectVars:     `[{"name":"LAGOON_FASTLY_SERVICE_IDS","value":"annotations.com:service-id:true:annotationscom","scope":"build"}]`,
				envVars:         `[]`,
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "test-resources/template-ingress/single-lagoon.yml",
				valuesFilePath:  "test-resources/template-ingress/test1-results",
				checkValuesFile: true,
				templatePath:    "test-resources/template-ingress/output",
			},
			want: "test-resources/template-ingress/test1-results",
		},
		{
			name: "test2 check LAGOON_FASTLY_SERVICE_IDS no secret and with values",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectVars:     `[{"name":"LAGOON_FASTLY_SERVICE_IDS","value":"annotations.com:service-id:true","scope":"build"}]`,
				envVars:         `[]`,
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "test-resources/template-ingress/single-lagoon.yml",
				valuesFilePath:  "test-resources/template-ingress/test2-results",
				checkValuesFile: true,
				templatePath:    "test-resources/template-ingress/output",
			},
			want: "test-resources/template-ingress/test2-results",
		},
		{
			name: "test3 check LAGOON_FASTLY_SERVICE_ID no secret and with values",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectVars:     `[{"name":"LAGOON_FASTLY_SERVICE_ID","value":"service-id:true","scope":"build"}]`,
				envVars:         `[]`,
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "test-resources/template-ingress/single-lagoon.yml",
				valuesFilePath:  "test-resources/template-ingress/test3-results",
				checkValuesFile: true,
				templatePath:    "test-resources/template-ingress/output",
			},
			want: "test-resources/template-ingress/test3-results",
		},
		{
			name: "test4 check LAGOON_FASTLY_SERVICE_IDS with secret no values",
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
				checkValuesFile: false,
				templatePath:    "test-resources/template-ingress/output",
			},
			want: "test-resources/template-ingress/test1-results",
		},
		{
			name: "test5 check LAGOON_FASTLY_SERVICE_IDS no secret and no values",
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
				checkValuesFile: false,
				templatePath:    "test-resources/template-ingress/output",
			},
			want: "test-resources/template-ingress/test2-results",
		},
		{
			name: "test6 check LAGOON_FASTLY_SERVICE_ID no secret and no values",
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
				checkValuesFile: false,
				templatePath:    "test-resources/template-ingress/output",
			},
			want: "test-resources/template-ingress/test3-results",
		},
		{
			name: "test7 check no fastly and with values",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectVars:     `[]`,
				envVars:         `[]`,
				secretPrefix:    "fastly-api-",
				lagoonYAML:      "test-resources/template-ingress/single-lagoon.yml",
				valuesFilePath:  "test-resources/template-ingress/test7-results",
				checkValuesFile: true,
				templatePath:    "test-resources/template-ingress/output",
			},
			want: "test-resources/template-ingress/test7-results",
		},
		{
			name: "test8 check no fastly and no values",
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
				checkValuesFile: false,
				templatePath:    "test-resources/template-ingress/output",
			},
			want: "test-resources/template-ingress/test7-results",
		},
		{
			name: "test9 multiproject1 no values",
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
				checkValuesFile: false,
				templatePath:    "test-resources/template-ingress/output",
			},
			want: "test-resources/template-ingress/test9-results",
		},
		{
			name: "test10 multiproject2 no values",
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
				checkValuesFile: false,
				templatePath:    "test-resources/template-ingress/output",
			},
			want: "test-resources/template-ingress/test10-results",
		},
		{
			name: "test11 multidomain no values",
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
				checkValuesFile: false,
				templatePath:    "test-resources/template-ingress/output",
			},
			want: "test-resources/template-ingress/test11-results",
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
			checkValuesFile = tt.args.checkValuesFile
			err = os.MkdirAll(tt.args.templatePath, 0755)
			if err != nil {
				t.Errorf("couldn't create directory %v: %v", savedTemplates, err)
			}
			savedTemplates = tt.args.templatePath
			fastlyAPISecretPrefix = tt.args.secretPrefix
			fastlyServiceID = tt.args.serviceID

			defer os.RemoveAll(savedTemplates)

			err = IngressTemplateGeneration(false)
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
			if len(files) != (len(results) - 1) {
				for _, f := range files {
					f1, err := os.ReadFile(fmt.Sprintf("%s/%s", savedTemplates, f.Name()))
					if err != nil {
						t.Errorf("couldn't read file %v: %v", savedTemplates, err)
					}
					fmt.Println(string(f1))
				}
				t.Errorf("number of generated templates doesn't match results %v/%v: %v", len(files), (len(results) - 1), err)
			}
			for _, f := range files {
				for _, r := range results {
					if f.Name() == r.Name() {
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
		})
	}
}
