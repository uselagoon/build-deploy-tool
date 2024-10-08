package generator

import (
	"testing"

	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
)

func Test_checkDuplicateCronjobs(t *testing.T) {
	type args struct {
		cronjobs []lagoon.Cronjob
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test1 - no duplicate cronjob names",
			args: args{
				cronjobs: []lagoon.Cronjob{
					{
						Name:     "drush uli",
						Command:  "drush uli",
						Service:  "cli",
						Schedule: "5 * * * *",
					},
					{
						Name:     "drush cron",
						Command:  "drush cron",
						Service:  "cli",
						Schedule: "5 * * * *",
					},
					{
						Name:     "drush cr",
						Command:  "drush cr",
						Service:  "cli",
						Schedule: "5 * * * *",
					},
				},
			},
		},
		{
			name: "test2 - duplicate cronjob names",
			args: args{
				cronjobs: []lagoon.Cronjob{
					{
						Name:     "drush uli",
						Command:  "drush uli",
						Service:  "cli",
						Schedule: "5 * * * *",
					},
					{
						Name:     "drush cr",
						Command:  "drush cr",
						Service:  "cli",
						Schedule: "5,25,2 4 * * *",
					},
					{
						Name:     "drush cron",
						Command:  "drush cron",
						Service:  "cli",
						Schedule: "5 * * * *",
					},
					{
						Name:     "drush cr",
						Command:  "drush cr",
						Service:  "cli",
						Schedule: "5 * * * *",
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkDuplicateCronjobs(tt.args.cronjobs); (err != nil) != tt.wantErr {
				t.Errorf("checkDuplicateCronjobs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_determineRefreshImage(t *testing.T) {
	type args struct {
		serviceName string
		imageName   string
		envVars     []lagoon.EnvironmentVariable
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Identity function",
			args: args{
				serviceName: "testservice",
				imageName:   "image/name:latest",
				envVars:     nil,
			},
			want:    "image/name:latest",
			wantErr: false,
		},
		{
			name: "Fails with no matching variable in envvars",
			args: args{
				serviceName: "testservice",
				imageName:   "image/name:${NOENVVAR}",
				envVars:     nil,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Fails with variable missing curly brackets",
			args: args{
				serviceName: "testservice",
				imageName:   "image/name:$NOENVVAR",
				envVars:     nil,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Tag with simple arg - fallback to default",
			args: args{
				serviceName: "testservice",
				imageName:   "image/name:${ENVVAR:-sometag}",
				envVars:     nil,
			},
			want:    "image/name:sometag",
			wantErr: false,
		},
		{
			name: "Tag with env var that works",
			args: args{
				serviceName: "testservice",
				imageName:   "image/name:${ENVVAR:-sometag}",
				envVars: []lagoon.EnvironmentVariable{
					{
						Name:  "ENVVAR",
						Value: "injectedTag",
					},
				},
			},
			want:    "image/name:injectedTag",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, errs := determineRefreshImage(tt.args.serviceName, tt.args.imageName, tt.args.envVars)
			if len(errs) > 0 && !tt.wantErr {
				for idx, err := range errs {
					t.Errorf("determineRefreshImage() error = %v, wantErr %v", err, tt.wantErr)
					if idx+1 == len(errs) {
						return
					}
				}
			}
			if got != tt.want && !tt.wantErr {
				t.Errorf("determineRefreshImage() got = %v, want %v", got, tt.want)
			}
		})
	}
}
