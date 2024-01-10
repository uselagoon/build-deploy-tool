package cmd

import (
	"strings"
	"testing"

	// changes the testing to source from root so paths to test resources must be defined from repo root
	_ "github.com/uselagoon/build-deploy-tool/internal/testing"
)

func TestValidateDockerCompose(t *testing.T) {
	type args struct {
		file                     string
		ignoreNonStringKeyErrors bool
		ignoreMissingEnvFiles    bool
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
				file: "internal/testdata/docker-compose/test3/docker-compose.yml",
			},
		},
		{
			name: "test2 complex docker-compose",
			args: args{
				file: "internal/testdata/docker-compose/test4/docker-compose.yml",
			},
		},
		{
			name: "test3 complex docker-compose",
			args: args{
				file: "internal/testdata/docker-compose/test5/docker-compose.yml",
			},
		},
		{
			name: "test4 complex docker-compose",
			args: args{
				file: "internal/testdata/docker-compose/test6/docker-compose.yml",
			},
		},
		{
			name: "test5 check an invalid docker-compose",
			args: args{
				file: "internal/testdata/docker-compose/test7/docker-compose.yml",
			},
			wantErr:    true,
			wantErrMsg: "Non-string key in x-site-branch: <nil>",
		},
		{
			name: "test6 check an invalid docker-compose (same as test5 but ignoring the errors)",
			args: args{
				file:                     "internal/testdata/docker-compose/test8/docker-compose.yml",
				ignoreNonStringKeyErrors: true,
			},
		},
		{
			name: "test7 check an valid docker-compose with missing env_files ",
			args: args{
				file: "internal/testdata/docker-compose/test10/docker-compose.yml",
			},
			wantErr:    true,
			wantErrMsg: "no such file or directory",
		},
		{
			name: "test8 check an valid docker-compose with missing env_files (same as test7 but ignoring the missing file errors)",
			args: args{
				file:                     "internal/testdata/docker-compose/test9/docker-compose.yml",
				ignoreNonStringKeyErrors: true,
				ignoreMissingEnvFiles:    true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateDockerCompose(tt.args.file, tt.args.ignoreNonStringKeyErrors, tt.args.ignoreMissingEnvFiles); err != nil {
				if tt.wantErr {
					if !strings.Contains(err.Error(), tt.wantErrMsg) {
						t.Errorf("ValidateDockerCompose() error = %v, wantErr %v", err, tt.wantErr)
					}
				} else {
					t.Errorf("ValidateDockerCompose() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}
