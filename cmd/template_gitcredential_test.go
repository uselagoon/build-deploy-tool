package cmd

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"

	// changes the testing to source from root so paths to test resources must be defined from repo root
	_ "github.com/uselagoon/build-deploy-tool/internal/testing"
)

func TestTemplateGitCredential(t *testing.T) {
	tests := []struct {
		name               string
		vars               []helpers.EnvironmentVariable
		want               string
		resultFilename     string
		testResultfilename string
		wantFile           bool
		wantErr            bool
	}{
		{
			name: "test1 check if variables are defined",
			vars: []helpers.EnvironmentVariable{
				{
					Name:  "SOURCE_REPOSITORY",
					Value: "https://example.com/lagoon-demo.git",
				},
				{
					Name:  "LAGOON_ENVIRONMENT_VARIABLES",
					Value: `[{"name":"LAGOON_GIT_HTTPS_USERNAME","scope":"build","value":"user1"},{"name":"LAGOON_GIT_HTTPS_PASSWORD","scope":"build","value":"somep@ssword"}]`,
				},
			},
			resultFilename:     "test1",
			testResultfilename: "internal/testdata/git-credentials/test1",
			wantFile:           true,
			want:               "store",
		},
		{
			name: "test2 check if variable are defined (no username)",
			vars: []helpers.EnvironmentVariable{
				{
					Name:  "SOURCE_REPOSITORY",
					Value: "https://example.com/lagoon-demo.git",
				},
				{
					Name:  "LAGOON_ENVIRONMENT_VARIABLES",
					Value: `[{"name":"LAGOON_GIT_HTTPS_PASSWORD","scope":"build","value":"somep@ssword"}]`,
				},
			},
			wantErr: true,
			want:    "",
		},
		{
			name: "test3 check if variable are defined (no password)",
			vars: []helpers.EnvironmentVariable{
				{
					Name:  "SOURCE_REPOSITORY",
					Value: "https://example.com/lagoon-demo.git",
				},
				{
					Name:  "LAGOON_ENVIRONMENT_VARIABLES",
					Value: `[{"name":"LAGOON_GIT_HTTPS_USERNAME","scope":"build","value":"user1"}]`,
				},
			},
			wantErr: true,
			want:    "",
		},
		{
			name: "test4 no username or password",
			vars: []helpers.EnvironmentVariable{
				{
					Name:  "SOURCE_REPOSITORY",
					Value: "https://example.com/lagoon-demo.git",
				},
				{
					Name:  "LAGOON_ENVIRONMENT_VARIABLES",
					Value: `[]`,
				},
			},
			want: "",
		},
		{
			name: "test5 ssh pass through",
			vars: []helpers.EnvironmentVariable{
				{
					Name:  "SOURCE_REPOSITORY",
					Value: "ssh://git@example.com/lagoon-demo.git",
				},
				{
					Name:  "LAGOON_ENVIRONMENT_VARIABLES",
					Value: `[]`,
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, envVar := range tt.vars {
				err := os.Setenv(envVar.Name, envVar.Value)
				if err != nil {
					t.Errorf("%v", err)
				}
			}
			tempResults := "testoutput"
			err := os.MkdirAll(tempResults, 0755)
			if err != nil {
				t.Errorf("couldn't create directory %v: %v", tempResults, err)
			}
			defer os.RemoveAll(tempResults)
			got, err := TemplateGitCredential(fmt.Sprintf("%s/%s", tempResults, tt.resultFilename))
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
				helpers.UnsetEnvVars(tt.vars)
			})
		})
	}
}
