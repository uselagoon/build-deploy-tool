package templating

import (
	"os"
	"reflect"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/uselagoon/build-deploy-tool/internal/generator"
)

func TestGenerateRegistrySecretTemplate(t *testing.T) {
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
					Services: []generator.ServiceValues{
						{
							Name:             "myservice",
							OverrideName:     "myservice",
							Type:             "basic",
							DBaaSEnvironment: "development",
						},
						{
							Name:             "myservice-po",
							OverrideName:     "myservice-po",
							Type:             "basic",
							DBaaSEnvironment: "development",
							ServicePort:      8080,
						},
					},
					ContainerRegistry: []generator.ContainerRegistry{
						{
							Name:       "secret1",
							SecretName: "internal-registry-secret-secret1",
							Username:   "username",
							Password:   "password",
							URL:        "my.registry.example.com",
						},
					},
				},
			},
			want: "test-resources/regsecret/registry-secret1.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateRegistrySecretTemplate(tt.args.buildValues)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateRegistrySecretTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			r1, err := os.ReadFile(tt.want)
			if err != nil {
				t.Errorf("couldn't read file %v: %v", tt.want, err)
			}
			var result []byte
			for _, d := range got {
				templateBytes, err := TemplateSecret(d)
				if err != nil {
					t.Errorf("couldn't generate template  %v", err)
				}
				result = append(result, templateBytes[:]...)
			}
			if !reflect.DeepEqual(string(result), string(r1)) {
				t.Errorf("GenerateRegistrySecretTemplate() = \n%v", diff.LineDiff(string(r1), string(result)))
			}
		})
	}
}
