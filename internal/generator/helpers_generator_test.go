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
