package generator

import (
	"testing"
)

func Test_variableExists(t *testing.T) {
	type args struct {
		vars  *[]LagoonEnvironmentVariable
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
				vars: &[]LagoonEnvironmentVariable{
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
			if got := variableExists(tt.args.vars, tt.args.name, tt.args.value); got != tt.want {
				t.Errorf("variableExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getMD5HashWithNewLine(t *testing.T) {
	type args struct {
		text string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "route hash",
			args: args{
				text: "a-really-long-name-that-should-truncate.www.example.com",
			},
			want: "7f2d0e459b080643ade429cf0bd782c6",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMD5HashWithNewLine(tt.args.text); got != tt.want {
				t.Errorf("getMD5Hash() = %v, want %v", got, tt.want)
			}
		})
	}
}
