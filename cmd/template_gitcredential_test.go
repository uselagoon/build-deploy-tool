package cmd

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/andreyvit/diff"
	generator "github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/testdata"

	// changes the testing to source from root so paths to test resources must be defined from repo root
	_ "github.com/uselagoon/build-deploy-tool/internal/testing"
)

func TestTemplateGitCredential(t *testing.T) {
	tests := []struct {
		name               string
		args               testdata.TestData
		want               string
		resultFilename     string
		testResultfilename string
		wantFile           bool
		wantErr            bool
	}{
		{
			name: "test1 check if variables are defined",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/basic/lagoon.yml",
					BuildPodVariables: []helpers.EnvironmentVariable{
						{
							Name: "LAGOON_ENVIRONMENT_VARIABLES",
							Value: `[
								{"name":"GITREPO_EXAMPLE_URL","scope":"build","value":"https://example.com"},
								{"name":"GITREPO_EXAMPLE_USERNAME","scope":"build","value":"user1"},
								{"name":"GITREPO_EXAMPLE_PASSWORD","scope":"build","value":"somep@ssword"}
							]`,
						},
						{
							Name:  "SOURCE_REPOSITORY",
							Value: "https://example.com/lagoon-demo.git",
						},
					},
				}, true),
			resultFilename:     "test1",
			testResultfilename: "internal/testdata/git-credentials/test1",
			wantFile:           true,
			want:               "store",
		},
		{
			name: "test2 check if variables are defined for multiple git repositories",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/basic/lagoon.yml",
					BuildPodVariables: []helpers.EnvironmentVariable{
						{
							Name: "LAGOON_ENVIRONMENT_VARIABLES",
							Value: `[
								{"name":"GITREPO_EXAMPLE_URL","scope":"build","value":"https://example.com"},
								{"name":"GITREPO_EXAMPLE_USERNAME","scope":"build","value":"user1"},
								{"name":"GITREPO_EXAMPLE_PASSWORD","scope":"build","value":"somep@ssword"},
								{"name":"GITREPO_GITHUB_USERNAME","scope":"build","value":"ghuser1"},
								{"name":"GITREPO_GITHUB_PASSWORD","scope":"build","value":"ghsomep@ssword"}
							]`,
						},
						{
							Name:  "SOURCE_REPOSITORY",
							Value: "https://example.com/lagoon-demo.git",
						},
					},
				}, true),
			resultFilename:     "test2",
			testResultfilename: "internal/testdata/git-credentials/test2",
			wantFile:           true,
			want:               "store",
		},
		{
			name: "test3 check if variable are defined (no username)",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/basic/lagoon.yml",
					BuildPodVariables: []helpers.EnvironmentVariable{
						{
							Name:  "LAGOON_ENVIRONMENT_VARIABLES",
							Value: `[{"name":"GITREPO_EXAMPLE_PASSWORD","scope":"build","value":"somep@ssword"}]`,
						},
						{
							Name:  "SOURCE_REPOSITORY",
							Value: "https://example.com/lagoon-demo.git",
						},
					},
				}, true),
			wantErr: true,
			want:    "",
		},
		{
			name: "test4 check if variable are defined (no password)",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/basic/lagoon.yml",
					BuildPodVariables: []helpers.EnvironmentVariable{
						{
							Name:  "LAGOON_ENVIRONMENT_VARIABLES",
							Value: `[{"name":"GITREPO_EXAMPLE_USERNAME","scope":"build","value":"user1"}]`,
						},
						{
							Name:  "SOURCE_REPOSITORY",
							Value: "https://example.com/lagoon-demo.git",
						},
					},
				}, true),
			wantErr: true,
			want:    "",
		},
		{
			name: "test5 ssh pass through",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/basic/lagoon.yml",
					BuildPodVariables: []helpers.EnvironmentVariable{
						{
							Name:  "LAGOON_ENVIRONMENT_VARIABLES",
							Value: `[]`,
						},
						{
							Name:  "SOURCE_REPOSITORY",
							Value: "ssh://git@example.com/lagoon-demo.git",
						},
					},
				}, true),
			want: "",
		},
		{
			name: "test6 ssh pass through",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/basic/lagoon.yml",
					BuildPodVariables: []helpers.EnvironmentVariable{
						{
							Name:  "LAGOON_ENVIRONMENT_VARIABLES",
							Value: `[]`,
						},
						{
							Name:  "SOURCE_REPOSITORY",
							Value: "git@example.com:lagoon-demo.git",
						},
					},
				}, true),
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.UnsetEnvVars(nil) //unset variables before running tests
			for _, envVar := range tt.args.BuildPodVariables {
				err := os.Setenv(envVar.Name, envVar.Value)
				if err != nil {
					t.Errorf("%v", err)
				}
			}
			tempResults, err := os.MkdirTemp("", "testoutput")
			if err != nil {
				t.Errorf("%v", err)
			}
			generator, err := testdata.SetupEnvironment(generator.GeneratorInput{}, tempResults, tt.args)
			if err != nil {
				t.Errorf("%v", err)
			}
			defer os.RemoveAll(tempResults)
			got, err := TemplateGitCredential(generator, fmt.Sprintf("%s/%s", tempResults, tt.resultFilename))
			if (err != nil) != tt.wantErr {
				t.Errorf("TemplateGitCredential() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TemplateGitCredential() = %v, want %v", got, tt.want)
			}
			if tt.wantFile {
				f1, err := os.ReadFile(fmt.Sprintf("%s/%s", tempResults, tt.resultFilename))
				if err != nil {
					t.Errorf("couldn't read file %v: %v", tempResults, err)
				}
				r1, err := os.ReadFile(tt.testResultfilename)
				if err != nil {
					t.Errorf("couldn't read file %v: %v", tt.wantFile, err)
				}
				if !reflect.DeepEqual(f1, r1) {
					t.Errorf("TemplateGitCredential() = \n%v", diff.LineDiff(string(r1), string(f1)))
				}
			}
			t.Cleanup(func() {
				helpers.UnsetEnvVars(nil)
				helpers.UnsetEnvVars(tt.args.BuildPodVariables)
			})
		})
	}
}
