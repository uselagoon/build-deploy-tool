package lagoon

import (
	"reflect"
	"testing"
)

func TestMergeVariables(t *testing.T) {
	type args struct {
		project     []EnvironmentVariable
		environment []EnvironmentVariable
	}
	tests := []struct {
		name string
		args args
		want []EnvironmentVariable
	}{
		{
			name: "test1",
			args: args{
				project: []EnvironmentVariable{
					{
						Name:  "PROJECT_SPECIFIC_VARIABLE",
						Value: "projectvariable",
						Scope: "global",
					},
					{
						Name:  "LAGOON_FASTLY_SERVICE_ID",
						Value: "abcdefg",
						Scope: "global",
					},
				},
				environment: []EnvironmentVariable{
					{
						Name:  "LAGOON_FASTLY_SERVICE_ID",
						Value: "1234567",
						Scope: "global",
					},
				},
			},
			want: []EnvironmentVariable{
				{
					Name:  "PROJECT_SPECIFIC_VARIABLE",
					Value: "projectvariable",
					Scope: "global",
				},
				{
					Name:  "LAGOON_FASTLY_SERVICE_ID",
					Value: "1234567",
					Scope: "global",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MergeVariables(tt.args.project, tt.args.environment); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeVariables() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetLagoonVariable(t *testing.T) {
	type args struct {
		name      string
		variables []EnvironmentVariable
	}
	tests := []struct {
		name    string
		args    args
		want    *EnvironmentVariable
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				name: "LAGOON_FASTLY_SERVICE_ID",
				variables: []EnvironmentVariable{
					{
						Name:  "PROJECT_SPECIFIC_VARIABLE",
						Value: "projectvariable",
						Scope: "global",
					},
					{
						Name:  "LAGOON_FASTLY_SERVICE_ID",
						Value: "1234567",
						Scope: "global",
					},
				},
			},
			want: &EnvironmentVariable{
				Name:  "LAGOON_FASTLY_SERVICE_ID",
				Value: "1234567",
				Scope: "global",
			},
			wantErr: false,
		},
		{
			name: "test2",
			args: args{
				name: "NON_EXISTENT_VARIABLE",
				variables: []EnvironmentVariable{
					{
						Name:  "PROJECT_SPECIFIC_VARIABLE",
						Value: "projectvariable",
						Scope: "global",
					},
					{
						Name:  "LAGOON_FASTLY_SERVICE_ID",
						Value: "1234567",
						Scope: "global",
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetLagoonVariable(tt.args.name, tt.args.variables)
			if (err != nil) != tt.wantErr {
				t.Errorf("getLagoonVariable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getLagoonVariable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVariableExists(t *testing.T) {
	type args struct {
		vars  *[]EnvironmentVariable
		name  string
		value string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "string",
			args: args{
				vars: &[]EnvironmentVariable{
					{
						Name:  "GENERIC_VARIABLE1",
						Value: "GENERIC_VARIABLE1",
						Scope: "global",
					},
					{
						Name:  "GENERIC_VARIABLE2",
						Value: "GENERIC_VARIABLE2",
						Scope: "global",
					},
					{
						Name:  "LAGOON_FASTLY_SERVICE_ID",
						Value: "abcdefg",
						Scope: "global",
					},
				},
				name:  "LAGOON_FASTLY_SERVICE_ID",
				value: "abcdefg",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VariableExists(tt.args.vars, tt.args.name, tt.args.value); got != tt.want {
				t.Errorf("variableExists() = %v, want %v", got, tt.want)
			}
		})
	}
}
