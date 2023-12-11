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

func TestBackupTemplateGeneration(t *testing.T) {
	type args struct {
		alertContact          string
		statusPageID          string
		projectName           string
		environmentName       string
		branch                string
		prNumber              string
		prHeadBranch          string
		prBaseBranch          string
		environmentType       string
		buildType             string
		activeEnvironment     string
		standbyEnvironment    string
		cacheNoCache          string
		serviceID             string
		secretPrefix          string
		projectVars           string
		envVars               string
		lagoonVersion         string
		lagoonYAML            string
		templatePath          string
		controllerDevSchedule string
		controllerPRSchedule  string
		k8upVersion           string
		namespace             string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				namespace:       "example-project-main",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "main",
				k8upVersion:     "v1",
				projectVars:     `[{"name":"LAGOON_FEATURE_FLAG_IMAGECACHE_REGISTRY","value":"https://imagecache.example.com","scope":"global"},{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"},{"name":"LAGOON_FASTLY_SERVICE_IDS","value":"example.com:service-id:true:annotationscom","scope":"build"}]`,
				envVars:         `[]`,
				lagoonYAML:      "../test-resources/template-backups/test1/lagoon.yml",
				templatePath:    "../test-resources/template-backups/output",
			},
			want: "../test-resources/template-backups/test1-results",
		},
		{
			name: "test2",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				namespace:       "example-project-main",
				environmentType: "development",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "main",
				k8upVersion:     "v1",
				projectVars:     `[{"name":"LAGOON_FEATURE_FLAG_CUSTOM_BACKUP_CONFIG","value":"enabled","scope":"global"},{"name":"LAGOON_BACKUP_DEV_SCHEDULE","value":"1,31 23 * * *","scope":"build"},{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"},{"name":"LAGOON_FASTLY_SERVICE_IDS","value":"example.com:service-id:true:annotationscom","scope":"build"}]`,
				envVars:         `[]`,
				lagoonYAML:      "../test-resources/template-backups/test2/lagoon.yml",
				templatePath:    "../test-resources/template-backups/output",
			},
			want: "../test-resources/template-backups/test2-results",
		},
		{
			name: "test3",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				namespace:       "example-project-main",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "main",
				k8upVersion:     "v1",
				projectVars:     `[{"name":"LAGOON_BACKUP_DEV_SCHEDULE","value":"1,31 23 * * *","scope":"build"},{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"},{"name":"LAGOON_FASTLY_SERVICE_IDS","value":"example.com:service-id:true:annotationscom","scope":"build"}]`,
				envVars:         `[]`,
				lagoonYAML:      "../test-resources/template-backups/test3/lagoon.yml",
				templatePath:    "../test-resources/template-backups/output",
			},
			want: "../test-resources/template-backups/test3-results",
		},
		{
			name: "test4",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				namespace:       "example-project-main",
				environmentType: "development",
				buildType:       "pullrequest",
				prNumber:        "123",
				prHeadBranch:    "main",
				prBaseBranch:    "main2",
				lagoonVersion:   "v2.7.x",
				branch:          "main",
				k8upVersion:     "v1",
				projectVars:     `[{"name":"LAGOON_FEATURE_FLAG_CUSTOM_BACKUP_CONFIG","value":"enabled","scope":"global"},{"name":"LAGOON_BAAS_CUSTOM_BACKUP_ACCESS_KEY","value":"abcdefg","scope":"build"},{"name":"LAGOON_BAAS_CUSTOM_BACKUP_SECRET_KEY","value":"abcdefg1234567","scope":"build"},{"name":"LAGOON_BACKUP_DEV_SCHEDULE","value":"1,31 23 * * *","scope":"build"},{"name":"LAGOON_BACKUP_PR_SCHEDULE","value":"3,33 12 * * *","scope":"build"},{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"},{"name":"LAGOON_FASTLY_SERVICE_IDS","value":"example.com:service-id:true:annotationscom","scope":"build"}]`,
				envVars:         `[]`,
				lagoonYAML:      "../test-resources/template-backups/test4/lagoon.yml",
				templatePath:    "../test-resources/template-backups/output",
			},
			want: "../test-resources/template-backups/test4-results",
		},
		{
			name: "test5",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				namespace:       "example-project-main",
				environmentType: "development",
				buildType:       "pullrequest",
				prNumber:        "123",
				prHeadBranch:    "main",
				prBaseBranch:    "main2",
				lagoonVersion:   "v2.7.x",
				branch:          "main",
				k8upVersion:     "v1",
				projectVars:     `[{"name":"LAGOON_FEATURE_FLAG_CUSTOM_BACKUP_CONFIG","value":"enabled","scope":"global"},{"name":"LAGOON_BAAS_CUSTOM_RESTORE_ACCESS_KEY","value":"abcdefg","scope":"build"},{"name":"LAGOON_BAAS_CUSTOM_RESTORE_SECRET_KEY","value":"abcdefg1234567","scope":"build"},{"name":"LAGOON_BACKUP_DEV_SCHEDULE","value":"1,31 23 * * *","scope":"build"},{"name":"LAGOON_BACKUP_PR_SCHEDULE","value":"3,33 12 * * *","scope":"build"},{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"},{"name":"LAGOON_FASTLY_SERVICE_IDS","value":"example.com:service-id:true:annotationscom","scope":"build"}]`,
				envVars:         `[]`,
				lagoonYAML:      "../test-resources/template-backups/test5/lagoon.yml",
				templatePath:    "../test-resources/template-backups/output",
			},
			want: "../test-resources/template-backups/test5-results",
		},
		{
			name: "test6",
			args: args{
				alertContact:    "alertcontact",
				statusPageID:    "statuspageid",
				projectName:     "example-project",
				environmentName: "main",
				namespace:       "example-project-main",
				environmentType: "production",
				buildType:       "branch",
				lagoonVersion:   "v2.7.x",
				branch:          "main",
				k8upVersion:     "v1",
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"},{"name":"LAGOON_FASTLY_SERVICE_IDS","value":"example.com:service-id:true:annotationscom","scope":"build"}]`,
				envVars:         `[]`,
				lagoonYAML:      "../test-resources/template-backups/test6/lagoon.yml",
				templatePath:    "../test-resources/template-backups/output",
			},
			want: "../test-resources/template-backups/test6-results",
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
			err = os.Setenv("NAMESPACE", tt.args.namespace)
			if err != nil {
				t.Errorf("%v", err)
			}
			err = os.Setenv("BRANCH", tt.args.branch)
			if err != nil {
				t.Errorf("%v", err)
			}
			err = os.Setenv("LAGOON_GIT_BRANCH", tt.args.branch)
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
			err = os.Setenv("LAGOON_FEATURE_BACKUP_DEV_SCHEDULE", tt.args.controllerDevSchedule)
			if err != nil {
				t.Errorf("%v", err)
			}
			err = os.Setenv("LAGOON_FEATURE_BACKUP_PR_SCHEDULE", tt.args.controllerPRSchedule)
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
			generator.BackupConfiguration.K8upVersion = tt.args.k8upVersion

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

			if err := BackupTemplateGeneration(generator); (err != nil) != tt.wantErr {
				t.Errorf("BackupTemplateGeneration() error = %v, wantErr %v", err, tt.wantErr)
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
				helpers.UnsetEnvVars(nil)
			})
		})
	}
}
