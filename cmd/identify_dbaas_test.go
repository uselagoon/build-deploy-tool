package cmd

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/uselagoon/build-deploy-tool/internal/dbaasclient"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
)

// these tests uses the same files as the dbaas templates
func TestIdentifyDBaaSConsumers(t *testing.T) {
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
		want    []string
		wantErr bool
	}{
		{
			name: "test1 - mariadb to mariadb-dbaas only",
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
				lagoonYAML:      "../test-resources/template-dbaas/test1/lagoon.yml",
				templatePath:    "../test-resources/output",
			},
			want: []string{
				"mariadb:mariadb-dbaas",
			},
		},
		{
			name: "test2 - mariadb to mariadb-shared which converts to mariadb-dbaas",
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
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"},{"name":"LAGOON_SERVICE_TYPES","value":"mariadb:mariadb-shared","scope":"build"}]`,
				envVars:         `[]`,
				lagoonYAML:      "../test-resources/template-dbaas/test1/lagoon.yml",
				templatePath:    "../test-resources/output",
			},
			want: []string{
				"mariadb:mariadb-dbaas",
			},
		},
		{
			name: "test3 - override provider to non-existent should result in failing dbaas check and a single pod no dbaas found",
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
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"},{"name":"LAGOON_DBAAS_ENVIRONMENT_TYPES","value":"mariadb:development2","scope":"build"}]`,
				envVars:         `[]`,
				lagoonYAML:      "../test-resources/template-dbaas/test1/lagoon.yml",
				templatePath:    "../test-resources/output",
			},
			want: []string{},
		},
		{
			name: "test4 - mariadb-single to mariadb-dbaas (using mariadb-shared to mariadb-dbaas conversion)",
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
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"},{"name":"LAGOON_SERVICE_TYPES","value":"mariadb:mariadb-shared","scope":"build"}]`,
				envVars:         `[]`,
				lagoonYAML:      "../test-resources/template-dbaas/test2/lagoon.yml",
				templatePath:    "../test-resources/output",
			},
			want: []string{
				"mariadb:mariadb-dbaas",
			},
		},
		{
			name: "test5 - multiple mariadb",
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
				projectVars:     `[{"name":"LAGOON_SYSTEM_ROUTER_PATTERN","value":"${service}-${project}-${environment}.example.com","scope":"internal_system"},{"name":"LAGOON_SERVICE_TYPES","value":"mariadb:mariadb-shared","scope":"build"}]`,
				envVars:         `[]`,
				lagoonYAML:      "../test-resources/template-dbaas/test3/lagoon.yml",
				templatePath:    "../test-resources/output",
			},
			want: []string{
				"mariadb:mariadb-dbaas",
				"mariadb2:mariadb-dbaas",
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
			generator, err := generatorInput(false)
			if err != nil {
				t.Errorf("%v", err)
			}
			generator.LagoonYAML = tt.args.lagoonYAML
			// add dbaasclient overrides for tests
			generator.DBaaSClient = dbaasclient.NewClient(dbaasclient.Client{
				RetryMax:     5,
				RetryWaitMin: time.Duration(10) * time.Millisecond,
				RetryWaitMax: time.Duration(50) * time.Millisecond,
			})

			// setup the fake dbaas server
			ts := dbaasclient.TestDBaaSHTTPServer()
			defer ts.Close()
			err = os.Setenv("DBAAS_OPERATOR_HTTP", ts.URL)
			if err != nil {
				t.Errorf("%v", err)
			}

			got, err := IdentifyDBaaSConsumers(generator)
			if (err != nil) != tt.wantErr {
				t.Errorf("IdentifyDBaaSConsumers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("IdentifyDBaaSConsumers() = %v, want %v", got, tt.want)
			}
			t.Cleanup(func() {
				helpers.UnsetEnvVars(tt.vars)
			})
		})
	}
}
