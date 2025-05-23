package cmd

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/uselagoon/build-deploy-tool/internal/dbaasclient"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	"github.com/uselagoon/build-deploy-tool/internal/testdata"

	// changes the testing to source from root so paths to test resources must be defined from repo root
	_ "github.com/uselagoon/build-deploy-tool/internal/testing"
)

func TestBackupTemplateGeneration(t *testing.T) {
	tests := []struct {
		name         string
		description  string
		args         testdata.TestData
		templatePath string
		want         string
		emptyDir     bool // if no templates are generated, then there will be a .gitkeep file in there
		wantErr      bool
	}{
		{
			name: "test1 - change the image registry used for prebackup pods",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/complex/lagoon.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_IMAGECACHE_REGISTRY",
							Value: "imagecache.example.com",
							Scope: "global",
						},
					},
				}, true),
			templatePath: "testoutput",
			want:         "internal/testdata/complex/backup-templates/backup-1",
		},
		{
			name: "test2 - custom dev only schedule but global config change enabled",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					EnvironmentType: "development",
					LagoonYAML:      "internal/testdata/node/lagoon.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_CUSTOM_BACKUP_CONFIG",
							Value: "enabled",
							Scope: "global",
						},
						{
							Name:  "LAGOON_BACKUP_DEV_SCHEDULE",
							Value: "1,31 23 * * *",
							Scope: "build",
						},
					},
				}, true),
			templatePath: "testoutput",
			want:         "internal/testdata/node/backup-templates/backup-1",
		},
		{
			name: "test3 - custom dev only schedule but global config change not configured (use defaults)",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					EnvironmentType: "development",
					LagoonYAML:      "internal/testdata/node/lagoon.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_BACKUP_DEV_SCHEDULE",
							Value: "1,31 23 * * *",
							Scope: "build",
						},
					},
				}, true),
			templatePath: "testoutput",
			want:         "internal/testdata/node/backup-templates/backup-2",
		},
		{
			name: "test4 - custom schedule and custom backup keys",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "pr-123",
					EnvironmentType: "development",
					BuildType:       "pullrequest",
					PRNumber:        "123",
					PRHeadBranch:    "main",
					PRBaseBranch:    "main2",
					PRHeadSHA:       "a1b2c3",
					PRBaseSHA:       "1a2b3c",
					LagoonYAML:      "internal/testdata/node/lagoon.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_CUSTOM_BACKUP_CONFIG",
							Value: "enabled",
							Scope: "global",
						},
						{Name: "LAGOON_FEATURE_FLAG_CUSTOM_BACKUP_CONFIG", Value: "enabled", Scope: "global"},
						{Name: "LAGOON_BAAS_CUSTOM_BACKUP_ACCESS_KEY", Value: "abcdefg", Scope: "build"},
						{Name: "LAGOON_BAAS_CUSTOM_BACKUP_SECRET_KEY", Value: "abcdefg1234567", Scope: "build"},
						{Name: "LAGOON_BACKUP_DEV_SCHEDULE", Value: "1,31 23 * * *", Scope: "build"},
						{Name: "LAGOON_BACKUP_PR_SCHEDULE", Value: "3,33 12 * * *", Scope: "build"},
					},
				}, true),
			templatePath: "testoutput",
			want:         "internal/testdata/node/backup-templates/backup-3",
		},
		{
			name: "test5 - custom schedule and custom restore keys",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "pr-123",
					EnvironmentType: "development",
					BuildType:       "pullrequest",
					PRNumber:        "123",
					PRHeadBranch:    "main",
					PRBaseBranch:    "main2",
					PRHeadSHA:       "a1b2c3",
					PRBaseSHA:       "1a2b3c",
					LagoonYAML:      "internal/testdata/node/lagoon.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_CUSTOM_BACKUP_CONFIG",
							Value: "enabled",
							Scope: "global",
						},
						{Name: "LAGOON_FEATURE_FLAG_CUSTOM_BACKUP_CONFIG", Value: "enabled", Scope: "global"},
						{Name: "LAGOON_BAAS_CUSTOM_RESTORE_ACCESS_KEY", Value: "abcdefg", Scope: "build"},
						{Name: "LAGOON_BAAS_CUSTOM_RESTORE_SECRET_KEY", Value: "abcdefg1234567", Scope: "build"},
						{Name: "LAGOON_BACKUP_DEV_SCHEDULE", Value: "1,31 23 * * *", Scope: "build"},
						{Name: "LAGOON_BACKUP_PR_SCHEDULE", Value: "3,33 12 * * *", Scope: "build"},
					},
				}, true),
			templatePath: "testoutput",
			want:         "internal/testdata/node/backup-templates/backup-4",
		},
		{
			name: "test6 - generic backup",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					EnvironmentType: "production",
					LagoonYAML:      "internal/testdata/node/lagoon.yml",
				}, true),
			templatePath: "testoutput",
			want:         "internal/testdata/node/backup-templates/backup-5",
		},
		{
			name: "test7 - changed default backup schedule",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:           "example-project",
					EnvironmentName:       "main",
					Branch:                "main",
					EnvironmentType:       "production",
					DefaultBackupSchedule: "M */6 * * *",
					LagoonYAML:            "internal/testdata/node/lagoon.yml",
				}, true),
			templatePath: "testoutput",
			want:         "internal/testdata/node/backup-templates/backup-6",
		},
		{
			name: "test8 - change the image registry used for prebackup pods k8upv2",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/complex/lagoon.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_IMAGECACHE_REGISTRY",
							Value: "imagecache.example.com",
							Scope: "global",
						},
					},
					BuildPodVariables: []helpers.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_DEFAULT_K8UP_V2",
							Value: "enabled",
						},
					},
				}, true),
			templatePath: "testoutput",
			want:         "internal/testdata/complex/backup-templates/backup-2",
		},
		{
			name: "test9 - nothing to backup so no schedule",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/node/lagoon.nostorage.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_IMAGECACHE_REGISTRY",
							Value: "imagecache.example.com",
							Scope: "global",
						},
					},
					BuildPodVariables: []helpers.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_DEFAULT_K8UP_V2",
							Value: "enabled",
						},
					},
				}, true),
			templatePath: "testdata/output",
			emptyDir:     true,
			want:         "internal/testdata/node/backup-templates/backup-7",
		}, {
			name: "test10 - generic backup with random check prune feature flags",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					EnvironmentType: "production",
					LagoonYAML:      "internal/testdata/node/lagoon.yml",
					EnvVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_K8UP_WEEKLY_RANDOM_CHECK",
							Value: "enabled",
							Scope: "global",
						},
						{
							Name:  "LAGOON_FEATURE_FLAG_K8UP_WEEKLY_RANDOM_PRUNE",
							Value: "enabled",
							Scope: "global",
						},
					},
				}, true),
			templatePath: "testdata/output",
			want:         "internal/testdata/node/backup-templates/backup-8",
		},
		{
			name: "test11 - custom prod schedule",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					EnvironmentType: "production",
					LagoonYAML:      "internal/testdata/node/lagoon.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_CUSTOM_BACKUP_CONFIG",
							Value: "enabled",
							Scope: "global",
						},
						{Name: "LAGOON_BACKUP_PROD_SCHEDULE", Value: "2,21 22 * * *", Scope: "build"},
					},
				}, true),
			templatePath: "testoutput",
			want:         "internal/testdata/node/backup-templates/backup-9",
		},
		{
			name:        "test-generic-backup-rootless-workloads",
			description: "this will generate a podsecuritycontext if the environment is configured for rootless workloads against k8up/v1 crs",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					EnvironmentType: "production",
					LagoonYAML:      "internal/testdata/node/lagoon.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_ROOTLESS_WORKLOAD",
							Value: "enabled",
							Scope: "global",
						},
					},
					BuildPodVariables: []helpers.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_DEFAULT_K8UP_V2",
							Value: "enabled",
						},
					},
				}, true),
			templatePath: "testoutput",
			want:         "internal/testdata/node/backup-templates/test-generic-backup-rootless-workloads",
		},
		{
			name:        "test-generic-backup-rootless-workloads-onrootmismatch",
			description: "this will generate a podsecuritycontext if the environment is configured for rootless workloads k8up/v1 crs",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					EnvironmentType: "production",
					LagoonYAML:      "internal/testdata/node/lagoon.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_ROOTLESS_WORKLOAD",
							Value: "enabled",
							Scope: "global",
						},
						{
							Name:  "LAGOON_FEATURE_FLAG_FS_ON_ROOT_MISMATCH",
							Value: "enabled",
							Scope: "global",
						},
					},
					BuildPodVariables: []helpers.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_DEFAULT_K8UP_V2",
							Value: "enabled",
						},
					},
				}, true),
			templatePath: "testoutput",
			want:         "internal/testdata/node/backup-templates/test-generic-backup-rootless-workloads-onrootmismatch",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.UnsetEnvVars(nil) //unset variables before running tests
			// set the environment variables from args
			savedTemplates := tt.templatePath
			generator, err := testdata.SetupEnvironment(*rootCmd, savedTemplates, tt.args)
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

			if err := BackupTemplateGeneration(generator); (err != nil) != tt.wantErr {
				t.Errorf("BackupTemplateGeneration() error = %v, wantErr %v", err, tt.wantErr)
			}
			files, err := os.ReadDir(savedTemplates)
			if err != nil {
				t.Errorf("couldn't read directory %v: %v", savedTemplates, err)
			}
			results, err := os.ReadDir(tt.want)
			if err != nil {
				t.Errorf("couldn't read directory %v: %v", tt.want, err)
			}
			resultSize := 0
			if !tt.emptyDir {
				results, err = os.ReadDir(tt.want)
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
							t.Errorf("BackupTemplateGeneration() = \n%v", diff.LineDiff(string(r1), string(f1)))
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
				helpers.UnsetEnvVars(tt.args.BuildPodVariables)
			})
		})
	}
}
