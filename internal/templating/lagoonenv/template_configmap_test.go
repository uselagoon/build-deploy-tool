package lagoonenv

import (
	"os"
	"reflect"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	"sigs.k8s.io/yaml"
)

func TestGenerateLagoonEnvConfigMap(t *testing.T) {
	type args struct {
		buildValues generator.BuildValues
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
				buildValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-name",
					EnvironmentType: "production",
					Namespace:       "myexample-project-environment-name",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "environment-name",
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
			},
			want: "test-resources/lagoon-env-1.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateLagoonEnvConfigMap(tt.args.buildValues)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateLagoonEnvConfigMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			r1, err := os.ReadFile(tt.want)
			if err != nil {
				t.Errorf("couldn't read file %v: %v", tt.want, err)
			}
			cm, err := yaml.Marshal(got)
			if err != nil {
				t.Errorf("couldn't generate template  %v", err)
			}
			if !reflect.DeepEqual(string(cm), string(r1)) {
				t.Errorf("GenerateLagoonEnvConfigMap() = \n%v", diff.LineDiff(string(r1), string(cm)))
			}
		})
	}
}
