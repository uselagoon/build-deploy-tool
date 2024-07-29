package templating

import (
	"os"
	"reflect"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
)

func TestGenerateLagoonEnvSecret(t *testing.T) {
	tests := []struct {
		name        string
		secretName  string
		description string
		buildValues generator.BuildValues
		want        string
		wantErr     bool
	}{
		{
			name:        "test1",
			description: "",
			secretName:  "lagoon-env",
			buildValues: generator.BuildValues{
				Project:         "example-project",
				Environment:     "environment-name",
				EnvironmentType: "production",
				Namespace:       "myexample-project-environment-name",
				BuildType:       "branch",
				LagoonVersion:   "v2.x.x",
				Kubernetes:      "generator.local",
				Branch:          "environment-name",
				LagoonEnvVariables: map[string]string{
					"MY_SPECIAL_VARIABLE1": "myspecialvariable1",
					"MY_SPECIAL_VARIABLE2": "myspecialvariable2",
					"MY_SPECIAL_VARIABLE":  "myspecialvariable",
				},
				EnvironmentVariables: []lagoon.EnvironmentVariable{
					{
						Name:  "MY_SPECIAL_VARIABLE1",
						Value: "myspecialvariable1",
						Scope: "global",
					},
					{
						Name:  "MY_SPECIAL_VARIABLE2",
						Value: "myspecialvariable2",
						Scope: "runtime",
					},
					{
						Name:  "MY_SPECIAL_VARIABLE3",
						Value: "myspecialvariable3",
						Scope: "build",
					},
					{
						Name:  "MY_SPECIAL_VARIABLE",
						Value: "myspecialvariable",
						Scope: "global",
					},
					{
						Name:  "LAGOON_SYSTEM_CORE_VERSION",
						Value: "v2.19.0",
						Scope: "internal_system",
					},
					{
						Name:  "REGISTRY_PASSWORD",
						Value: "myenvvarregistrypassword",
						Scope: "container_registry",
					},
				},
			},
			want: "test-resources/lagoonenv/lagoon-env-1.yaml",
		},
		{
			name:       "test2",
			secretName: "lagoon-platform-env",
			buildValues: generator.BuildValues{
				Project:         "example-project",
				Environment:     "environment-name",
				EnvironmentType: "production",
				Namespace:       "myexample-project-environment-name",
				BuildType:       "branch",
				LagoonVersion:   "v2.x.x",
				Kubernetes:      "generator.local",
				Branch:          "environment-name",
				LagoonEnvVariables: map[string]string{
					"MY_SPECIAL_VARIABLE1": "myspecialvariable1",
					"MY_SPECIAL_VARIABLE2": "myspecialvariable2",
					"MY_SPECIAL_VARIABLE":  "myspecialvariable",
				},
				LagoonPlatformEnvVariables: map[string]string{
					"MY_SPECIAL_VARIABLE3": "myspecialvariable3",
					"MY_SPECIAL_VARIABLE4": "myspecialvariable4",
				},
			},
			want: "test-resources/lagoonenv/lagoon-platform-env-1.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateLagoonEnvSecret(tt.secretName, tt.buildValues)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateLagoonEnvSecret() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			r1, err := os.ReadFile(tt.want)
			if err != nil {
				t.Errorf("couldn't read file %v: %v", tt.want, err)
			}
			templateBytes, err := TemplateSecret(got)
			if err != nil {
				t.Errorf("couldn't generate template: %v", err)
			}
			if !reflect.DeepEqual(string(templateBytes), string(r1)) {
				t.Errorf("GenerateLagoonEnvSecret() = \n%v", diff.LineDiff(string(r1), string(templateBytes)))
			}
		})
	}
}
