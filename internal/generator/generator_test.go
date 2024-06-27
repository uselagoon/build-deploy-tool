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

func TestCheckAdminFeatureFlag(t *testing.T) {
	type args struct {
		key   string
		debug bool
	}
	tests := []struct {
		name string
		vars []helpers.EnvironmentVariable
		args args
		want string
	}{
		{
			name: "test1 - container memory limit",
			vars: []helpers.EnvironmentVariable{
				{
					Name:  "ADMIN_LAGOON_FEATURE_FLAG_CONTAINER_MEMORY_LIMIT",
					Value: "16Gi",
				},
			},
			args: args{
				key: "CONTAINER_MEMORY_LIMIT",
			},
			want: "16Gi",
		},
		{
			name: "test2 - ephemeral storage requests",
			vars: []helpers.EnvironmentVariable{
				{
					Name:  "ADMIN_LAGOON_FEATURE_FLAG_EPHEMERAL_STORAGE_REQUESTS",
					Value: "16Gi",
				},
			},
			args: args{
				key: "EPHEMERAL_STORAGE_REQUESTS",
			},
			want: "16Gi",
		},
		{
			name: "test2 - ephemeral storage limit",
			vars: []helpers.EnvironmentVariable{
				{
					Name:  "ADMIN_LAGOON_FEATURE_FLAG_CONTAINER_MEMORY_LIMIT",
					Value: "disabled",
				},
				{
					Name:  "ADMIN_LAGOON_FEATURE_FLAG_EPHEMERAL_STORAGE_LIMIT",
					Value: "160Gi",
				},
			},
			args: args{
				key: "EPHEMERAL_STORAGE_LIMIT",
			},
			want: "160Gi",
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
			if got := CheckAdminFeatureFlag(tt.args.key, tt.args.debug); got != tt.want {
				t.Errorf("CheckAdminFeatureFlag() = %v, want %v", got, tt.want)
			}
			t.Cleanup(func() {
				helpers.UnsetEnvVars(tt.vars)
			})
		})
	}
}

func TestValidateResourceQuantity(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				s: "100m",
			},
			wantErr: false,
		},
		{
			name: "test2",
			args: args{
				s: "100M",
			},
			wantErr: false,
		},
		{
			name: "test3",
			args: args{
				s: "10Gi",
			},
			wantErr: false,
		},
		{
			name: "test4",
			args: args{
				s: "aa11",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateResourceQuantity(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("ValidateResourceQuantity() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
