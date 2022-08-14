package cmd

import (
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	"os"
	"reflect"
	"sigs.k8s.io/yaml"
	"testing"
)

func TestValidateLagoonYml(t *testing.T) {
	type args struct {
		lagoonYml         string
		lagoonOverrideYml string
		wantLagoonYml     string
		lYAML             *lagoon.YAML
		projectName       string
		debug             bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test 1 - Simple .lagoon.yml - testing equality",
			args: args{
				lagoonYml:     "../test-resources/validate-lagoon-yml/test1/lagoon.yml",
				wantLagoonYml: "../test-resources/validate-lagoon-yml/test1/lagoon.yml",
				lYAML:         &lagoon.YAML{},
				projectName:   "",
				debug:         false,
			},
			wantErr: false,
		},
		{
			name: "test 2 - Merging files",
			args: args{
				lagoonYml:         "../test-resources/validate-lagoon-yml/test2/lagoon.yml",
				lagoonOverrideYml: "../test-resources/validate-lagoon-yml/test2/lagoon-override.yml",
				wantLagoonYml:     "../test-resources/validate-lagoon-yml/test2/lagoon-final.yml",
				lYAML:             &lagoon.YAML{},
				projectName:       "",
				debug:             false,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateLagoonYml(tt.args.lagoonYml, tt.args.lagoonOverrideYml, "", tt.args.lYAML, tt.args.projectName, tt.args.debug); (err != nil) != tt.wantErr {
				t.Errorf("ValidateLagoonYml() error = %v, wantErr %v", err, tt.wantErr)
			}
			wantsLYAMLString, err := os.ReadFile(tt.args.wantLagoonYml)
			if err != nil {
				t.Errorf("Error loading %v wantsLagoonYml for test '%v'", tt.args.wantLagoonYml, tt.name)
				return
			}

			wantsLYAML := &lagoon.YAML{}

			err = yaml.Unmarshal(wantsLYAMLString, wantsLYAML)
			if err != nil {
				t.Errorf(err.Error())
				return
			}

			if !reflect.DeepEqual(tt.args.lYAML, wantsLYAML) {
				t.Errorf("not equal")
				return
			}

		})
	}
}
