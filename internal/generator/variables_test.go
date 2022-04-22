package generator

import (
	"reflect"
	"testing"
)

func TestMergeVariables(t *testing.T) {
	type args struct {
		project     []LagoonEnvironmentVariable
		environment []LagoonEnvironmentVariable
	}
	tests := []struct {
		name string
		args args
		want []LagoonEnvironmentVariable
	}{
		{
			name: "string",
			args: args{
				project: []LagoonEnvironmentVariable{
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
				environment: []LagoonEnvironmentVariable{
					{
						Name:  "LAGOON_FASTLY_SERVICE_ID",
						Value: "1234567",
						Scope: "global",
					},
				},
			},
			want: []LagoonEnvironmentVariable{
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

func Test_getLagoonVariable(t *testing.T) {
	type args struct {
		name      string
		variables []LagoonEnvironmentVariable
	}
	tests := []struct {
		name    string
		args    args
		want    LagoonEnvironmentVariable
		wantErr bool
	}{
		{
			name: "get variable",
			args: args{
				name: "LAGOON_FASTLY_SERVICE_ID",
				variables: []LagoonEnvironmentVariable{
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
			want: LagoonEnvironmentVariable{
				Name:  "LAGOON_FASTLY_SERVICE_ID",
				Value: "1234567",
				Scope: "global",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getLagoonVariable(tt.args.name, tt.args.variables)
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
