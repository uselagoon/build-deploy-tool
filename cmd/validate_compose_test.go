package cmd

import (
	"testing"
)

func TestValidateDockerCompose(t *testing.T) {
	type args struct {
		file         string
		ignoreErrors bool
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "test1 complex docker-compose",
			args: args{
				file: "../test-resources/docker-compose/test3/docker-compose.yml",
			},
		},
		{
			name: "test2 complex docker-compose",
			args: args{
				file: "../test-resources/docker-compose/test4/docker-compose.yml",
			},
		},
		{
			name: "test3 complex docker-compose",
			args: args{
				file: "../test-resources/docker-compose/test5/docker-compose.yml",
			},
		},
		{
			name: "test4 complex docker-compose",
			args: args{
				file: "../test-resources/docker-compose/test6/docker-compose.yml",
			},
		},
		{
			name: "test5 check an invalid docker-compose",
			args: args{
				file: "../test-resources/docker-compose/test7/docker-compose.yml",
			},
			wantErr:    true,
			wantErrMsg: "Non-string key in x-site-branch: <nil>",
		},
		{
			name: "test6 check an invalid docker-compose (same as test5 but ignoring the errors)",
			args: args{
				file:         "../test-resources/docker-compose/test8/docker-compose.yml",
				ignoreErrors: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateDockerCompose(tt.args.file, tt.args.ignoreErrors); err != nil {
				if tt.wantErr {
					if err.Error() != tt.wantErrMsg {
						t.Errorf("ValidateDockerCompose() error = %v, wantErr %v", err, tt.wantErr)
					}
				} else {
					t.Errorf("ValidateDockerCompose() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}
