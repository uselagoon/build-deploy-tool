package cmd

import (
	"encoding/base64"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	"os"
	"reflect"
	"sigs.k8s.io/yaml"
	"testing"
)

func TestValidateLagoonYml(t *testing.T) {
	type args struct {
		lagoonYml                string
		lagoonOverrideYml        string
		lagoonOverrideEnvVarFile string
		wantLagoonYml            string
		lYAML                    *lagoon.YAML
		projectName              string
		debug                    bool
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
			name: "test 2 - Merging files - no env vars",
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
		{
			name: "test 3 - Merging env vars - no override",
			args: args{
				lagoonYml:                "../test-resources/validate-lagoon-yml/test3/lagoon.yml",
				lagoonOverrideEnvVarFile: "../test-resources/validate-lagoon-yml/test3/lagoon-override.yml",
				wantLagoonYml:            "../test-resources/validate-lagoon-yml/test3/lagoon-final.yml",
				lYAML:                    &lagoon.YAML{},
				projectName:              "",
				debug:                    false,
			},
			wantErr: false,
		},
		{
			name: "test 4 - Merging env vars and override file",
			args: args{
				lagoonYml:                "../test-resources/validate-lagoon-yml/test4/lagoon.yml",
				lagoonOverrideYml:        "../test-resources/validate-lagoon-yml/test4/lagoon-override.yml",
				lagoonOverrideEnvVarFile: "../test-resources/validate-lagoon-yml/test4/lagoon-override-env.yml",
				wantLagoonYml:            "../test-resources/validate-lagoon-yml/test4/lagoon-final.yml",
				lYAML:                    &lagoon.YAML{},
				projectName:              "",
				debug:                    false,
			},
			wantErr: false,
		},
		{
			name: "test 5 - Overriding named task",
			args: args{
				lagoonYml:         "../test-resources/validate-lagoon-yml/test5/lagoon.yml",
				lagoonOverrideYml: "../test-resources/validate-lagoon-yml/test5/lagoon-override.yml",
				wantLagoonYml:     "../test-resources/validate-lagoon-yml/test5/lagoon-final.yml",
				lYAML:             &lagoon.YAML{},
				projectName:       "",
				debug:             false,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			const testEnvVar = "VALIDATE_LAGOON_YML_TEST_ENV"
			os.Setenv(testEnvVar, "")
			if tt.args.lagoonOverrideEnvVarFile != "" {
				lagoonOverrideEnvVarFileContents, err := os.ReadFile(tt.args.lagoonOverrideEnvVarFile)
				if err != nil {
					t.Errorf("Unable to read contents of env var test file '%v'", tt.args.lagoonOverrideEnvVarFile)
				}
				lagoonOverrideEnvVarFileContentsB64 := base64.StdEncoding.EncodeToString(lagoonOverrideEnvVarFileContents)
				os.Setenv(testEnvVar, lagoonOverrideEnvVarFileContentsB64)
			}

			if err := ValidateLagoonYml(tt.args.lagoonYml, tt.args.lagoonOverrideYml, testEnvVar, tt.args.lYAML, tt.args.projectName, tt.args.debug); (err != nil) != tt.wantErr {
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
