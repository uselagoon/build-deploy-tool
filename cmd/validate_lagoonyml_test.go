package cmd

import (
	"encoding/base64"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	"sigs.k8s.io/yaml"
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
		{
			name: "test 6 - Invalid lagoon override should fail",
			args: args{
				lagoonYml:         "../test-resources/validate-lagoon-yml/test6/lagoon.yml",
				lagoonOverrideYml: "../test-resources/validate-lagoon-yml/test6/lagoon-override.yml",
				lYAML:             &lagoon.YAML{},
				projectName:       "",
				debug:             false,
			},
			wantErr: true,
		},
		{
			name: "test 7 - Invalid lagoon override env var should fail",
			args: args{
				lagoonYml:                "../test-resources/validate-lagoon-yml/test6/lagoon.yml",
				lagoonOverrideEnvVarFile: "../test-resources/validate-lagoon-yml/test6/lagoon-override.yml",
				lYAML:                    &lagoon.YAML{},
				projectName:              "",
				debug:                    false,
			},
			wantErr: true,
		},
		{
			name: "multiline cronjobs should fail validation",
			args: args{
				lagoonYml:   "../test-resources/validate-lagoon-yml/cronjobs/lagoon.yml",
				lYAML:       &lagoon.YAML{},
				projectName: "",
				debug:       false,
			},
			wantErr: true,
		},
		{
			name: "merged multiline cronjobs should fail validation",
			args: args{
				lagoonYml:                "../test-resources/validate-lagoon-yml/cronjobs/lagoon.yml",
				lagoonOverrideYml:        "../test-resources/validate-lagoon-yml/cronjobs/lagoon-override.yml",
				lagoonOverrideEnvVarFile: "../test-resources/validate-lagoon-yml/cronjobs/lagoon-override-env.yml",
				lYAML:                    &lagoon.YAML{},
				projectName:              "",
				debug:                    false,
			},
			wantErr: true,
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

			if err := ValidateLagoonYml(tt.args.lagoonYml, tt.args.lagoonOverrideYml, testEnvVar, tt.args.lYAML, tt.args.projectName, tt.args.debug); err != nil {
				// if we expect a validation error, that's good, we get out of here.
				if tt.wantErr {
					if tt.args.debug {
						fmt.Printf("Test '%v' failed with error: %v", tt.name, err)
					}
					return
				} else {
					t.Errorf("ValidateLagoonYml() error = %v, wantErr %v", err, tt.wantErr)
				}
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

func TestMultilineCronjobs(t *testing.T) {
	var l lagoon.YAML
	if err := generator.LoadAndUnmarshalLagoonYml("../test-resources/validate-lagoon-yml/cronjobs/multiline-cronjobs.lagoon.yml", "", "", &l, "", false); err != nil {
		t.Fatalf("couldn't load and unmarshal YAML: %v", err)
	}

	for _, e := range l.Environments {
		for _, lagoonCronjob := range e.Cronjobs {
			t.Run(lagoonCronjob.Name, func(tt *testing.T) {
				err := ValidateCronjob(&lagoonCronjob)

				tt.Log(err)
				if err == nil {
					tt.Fatalf("expected error, but got nil")
				}
			})
		}
	}

}

func TestSinglelineCronjobs(t *testing.T) {
	var l lagoon.YAML
	if err := generator.LoadAndUnmarshalLagoonYml("../test-resources/validate-lagoon-yml/cronjobs/singleline-cronjobs.lagoon.yml", "", "", &l, "", false); err != nil {
		t.Fatalf("couldn't load and unmarshal YAML: %v", err)
	}

	for _, e := range l.Environments {
		for _, lagoonCronjob := range e.Cronjobs {
			t.Run(lagoonCronjob.Name, func(tt *testing.T) {
				fmt.Printf("%q\n", lagoonCronjob.Command)
				err := ValidateCronjob(&lagoonCronjob)

				if err != nil {
					tt.Fatalf("unexpected error %v", err)
				}
			})
		}
	}

}
