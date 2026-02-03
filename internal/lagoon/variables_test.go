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
		{
			name: "test2",
			args: args{
				project: []EnvironmentVariable{
					{
						Name:  "PROJECT_SPECIFIC_VARIABLE",
						Value: "projectvariable",
						Scope: "global",
					},
				},
				environment: []EnvironmentVariable{
					{
						Name:  "LAGOON_ROUTES_JSON",
						Value: "eyJyb3V0ZXMiOlt7ImRvbWFpbiI6InRlc3QxLmV4YW1wbGUuY29tIiwic2VydmljZSI6Im5naW54IiwidGxzLWFjbWUiOmZhbHNlLCJtb25pdG9yaW5nLXBhdGgiOiIvYnlwYXNzLWNhY2hlIn1dfQo=",
						Scope: "build",
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
					Name:  "LAGOON_ROUTES_JSON",
					Value: "eyJyb3V0ZXMiOlt7ImRvbWFpbiI6InRlc3QxLmV4YW1wbGUuY29tIiwic2VydmljZSI6Im5naW54IiwidGxzLWFjbWUiOmZhbHNlLCJtb25pdG9yaW5nLXBhdGgiOiIvYnlwYXNzLWNhY2hlIn1dfQo=",
					Scope: "build",
				},
			},
		},
		{
			name: "test3",
			args: args{
				project: []EnvironmentVariable{},
				environment: []EnvironmentVariable{
					{
						Name:  "LAGOON_ROUTES_JSON",
						Value: "eyJyb3V0ZXMiOlt7ImRvbWFpbiI6InRlc3QxLmV4YW1wbGUuY29tIiwic2VydmljZSI6Im5naW54IiwidGxzLWFjbWUiOmZhbHNlLCJtb25pdG9yaW5nLXBhdGgiOiIvYnlwYXNzLWNhY2hlIn1dfQo=",
						Scope: "build",
					},
				},
			},
			want: []EnvironmentVariable{
				{
					Name:  "LAGOON_ROUTES_JSON",
					Value: "eyJyb3V0ZXMiOlt7ImRvbWFpbiI6InRlc3QxLmV4YW1wbGUuY29tIiwic2VydmljZSI6Im5naW54IiwidGxzLWFjbWUiOmZhbHNlLCJtb25pdG9yaW5nLXBhdGgiOiIvYnlwYXNzLWNhY2hlIn1dfQo=",
					Scope: "build",
				},
			},
		},
		{
			name: "test4 internal system not unsettable",
			args: args{
				project: []EnvironmentVariable{
					{
						Name:  "LAGOON_ROUTE_QUOTA",
						Value: "1234",
						Scope: "internal_system",
					},
					{
						Name:  "LAGOON_ROUTE_QUOTA",
						Value: "4321",
						Scope: "global",
					},
				},
				environment: []EnvironmentVariable{
					{
						Name:  "LAGOON_ROUTE_QUOTA",
						Value: "123",
						Scope: "build",
					},
					{
						// this will be ignored
						Name:  "LAGOON_ROUTE_QUOTA",
						Value: "321",
						Scope: "internal_system",
					},
				},
			},
			want: []EnvironmentVariable{
				{
					Name:  "LAGOON_ROUTE_QUOTA",
					Value: "1234",
					Scope: "internal_system",
				},
				{
					Name:  "LAGOON_ROUTE_QUOTA",
					Value: "123",
					Scope: "build",
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
		scope     []string
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
				name:  "LAGOON_FASTLY_SERVICE_ID",
				scope: []string{"global"},
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
				name:  "NON_EXISTENT_VARIABLE",
				scope: []string{"global"},
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
		{
			name: "test3",
			args: args{
				name:  "LAGOON_ROUTES_JSON",
				scope: []string{"global", "build"},
				variables: []EnvironmentVariable{
					{
						Name:  "PROJECT_SPECIFIC_VARIABLE",
						Value: "projectvariable",
						Scope: "global",
					},
					{
						Name:  "LAGOON_ROUTES_JSON",
						Value: "eyJyb3V0ZXMiOlt7ImRvbWFpbiI6InRlc3QxLmV4YW1wbGUuY29tIiwic2VydmljZSI6Im5naW54IiwidGxzLWFjbWUiOmZhbHNlLCJtb25pdG9yaW5nLXBhdGgiOiIvYnlwYXNzLWNhY2hlIn1dfQo=",
						Scope: "build",
					},
				},
			},
			want: &EnvironmentVariable{
				Name:  "LAGOON_ROUTES_JSON",
				Value: "eyJyb3V0ZXMiOlt7ImRvbWFpbiI6InRlc3QxLmV4YW1wbGUuY29tIiwic2VydmljZSI6Im5naW54IiwidGxzLWFjbWUiOmZhbHNlLCJtb25pdG9yaW5nLXBhdGgiOiIvYnlwYXNzLWNhY2hlIn1dfQo=",
				Scope: "build",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetLagoonVariable(tt.args.name, tt.args.scope, tt.args.variables)
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

func TestGetBuildVariable(t *testing.T) {
	envVars := []EnvironmentVariable{
		{
			Name:  "LAGOON_FEATURE_FLAG_INSIGHTS",
			Value: "enabled",
			Scope: "global",
		},
		{
			Name:  "LAGOON_FASTLY_SERVICE_ID",
			Value: "1234567",
			Scope: "build",
		},
		{
			Name:  "PHP_MEMORY_LIMIT",
			Value: "5G",
			Scope: "runtime",
		},
	}

	tests := []struct {
		name    string
		varName string
		want    *EnvironmentVariable
		wantErr bool
	}{
		{
			name:    "test1",
			varName: "LAGOON_FASTLY_SERVICE_ID",
			want: &EnvironmentVariable{
				Name:  "LAGOON_FASTLY_SERVICE_ID",
				Value: "1234567",
				Scope: "build",
			},
			wantErr: false,
		},
		{
			name:    "test2",
			varName: "LAGOON_FEATURE_FLAG_INSIGHTS",
			want: &EnvironmentVariable{
				Name:  "LAGOON_FEATURE_FLAG_INSIGHTS",
				Value: "enabled",
				Scope: "global",
			},
			wantErr: false,
		},
		{
			name:    "test3",
			varName: "PHP_MEMORY_LIMIT",
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBuildVariable(tt.varName, envVars)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBuildVariable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBuildVariable() = %v, want %v", got, tt.want)
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
