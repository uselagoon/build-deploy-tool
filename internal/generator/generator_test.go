package generator

import (
	"os"
	"testing"

	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
)

func TestCheckFeatureFlag(t *testing.T) {
	type args struct {
		key          string
		envVariables []lagoon.EnvironmentVariable
		debug        bool
	}
	tests := []struct {
		name string
		vars []helpers.EnvironmentVariable
		args args
		want string
	}{
		{
			name: "test1 - rootless from default",
			vars: []helpers.EnvironmentVariable{
				{
					Name:  "LAGOON_FEATURE_FLAG_DEFAULT_ROOTLESS_WORKLOAD",
					Value: "enabled",
				},
			},
			args: args{
				key: "ROOTLESS_WORKLOAD",
			},
			want: "enabled",
		},
		{
			name: "test2 - rootless from variable",
			vars: []helpers.EnvironmentVariable{
				{
					Name:  "LAGOON_FEATURE_FLAG_DEFAULT_ROOTLESS_WORKLOAD",
					Value: "enabled",
				},
			},
			args: args{
				key: "ROOTLESS_WORKLOAD",
				envVariables: []lagoon.EnvironmentVariable{
					{
						Name:  "LAGOON_FEATURE_FLAG_ROOTLESS_WORKLOAD",
						Value: "disabled",
						Scope: "build",
					},
				},
			},
			want: "disabled",
		},
		{
			name: "test2 - rootless from forced",
			vars: []helpers.EnvironmentVariable{
				{
					Name:  "LAGOON_FEATURE_FLAG_DEFAULT_ROOTLESS_WORKLOAD",
					Value: "disabled",
				},
				{
					Name:  "LAGOON_FEATURE_FLAG_FORCE_ROOTLESS_WORKLOAD",
					Value: "enabled",
				},
			},
			args: args{
				key: "ROOTLESS_WORKLOAD",
				envVariables: []lagoon.EnvironmentVariable{
					{
						Name:  "LAGOON_FEATURE_FLAG_ROOTLESS_WORKLOAD",
						Value: "disabled",
						Scope: "build",
					},
				},
			},
			want: "enabled",
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
			if got := CheckFeatureFlag(tt.args.key, tt.args.envVariables, tt.args.debug); got != tt.want {
				t.Errorf("CheckFeatureFlag() = %v, want %v", got, tt.want)
			}
			t.Cleanup(func() {
				helpers.UnsetEnvVars(tt.vars)
			})
		})
	}
}
