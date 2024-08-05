package helpers

import (
	"strings"
	"testing"
)

func TestConvertCrontab(t *testing.T) {
	type args struct {
		namespace string
		cron      string
	}
	tests := []struct {
		name       string
		args       args
		want       string
		wantErrMsg string
		wantErr    bool
	}{
		{
			name: "test1",
			args: args{
				namespace: "example-com-main",
				cron:      "M * * * *",
			},
			want: "31 * * * *",
		},
		{
			name: "test2",
			args: args{
				namespace: "example-com-main",
				cron:      "M/5 * * * *",
			},
			want: "1,6,11,16,21,26,31,36,41,46,51,56 * * * *",
		},
		{
			name: "test3",
			args: args{
				namespace: "example-com-main",
				cron:      "M H(2-4) * * *",
			},
			want: "31 3 * * *",
		},
		{
			name: "test4",
			args: args{
				namespace: "example-com-main",
				cron:      "M H(22-2) * * *",
			},
			want: "31 1 * * *",
		},
		{
			name: "test5",
			args: args{
				namespace: "example-com-main",
				cron:      "M/15 H(22-2) * * *",
			},
			want: "1,16,31,46 1 * * *",
		},
		{
			name: "test6 - invalid minutes definition",
			args: args{
				namespace: "example-com-main",
				cron:      "M/H5 H(22-2) * * *",
			},
			wantErrMsg: "cron definition 'M/H5 H(22-2) * * *' is invalid, unable to determine minutes value",
			wantErr:    true,
		},
		{
			name: "test7 - invalid hour definiton",
			args: args{
				namespace: "example-com-main",
				cron:      "M/15 H(H2-2) * * *",
			},
			wantErrMsg: "cron definition 'M/15 H(H2-2) * * *' is invalid, unable to determine hours value",
			wantErr:    true,
		},
		{
			name: "test8",
			args: args{
				namespace: "example-com-main",
				cron:      "M/15 H(22-2) 3,5 * *",
			},
			want: "1,16,31,46 1 3,5 * *",
		},
		{
			name: "test9",
			args: args{
				namespace: "example-com-main",
				cron:      "M/15 H(22-2) * 10-12 *",
			},
			want: "1,16,31,46 1 * 10-12 *",
		},
		{
			name: "test10 - invalid dayofweek range",
			args: args{
				namespace: "example-com-main",
				cron:      "M/15 H(22-2) * * 1-8",
			},
			wantErrMsg: "cron definition 'M/15 H(22-2) * * 1-8' is invalid, unable to determine day(week) value",
			wantErr:    true,
		},
		{
			name: "test11",
			args: args{
				namespace: "example-com-main",
				cron:      "15 * * * 1,2,3,6",
			},
			want: "15 * * * 1,2,3,6",
		},
		{
			name: "test12",
			args: args{
				namespace: "example-com-main",
				cron:      "15 * 1-31 * *",
			},
			want: "15 * 1-31 * *",
		},
		{
			name: "test13 - invalid day range",
			args: args{
				namespace: "example-com-main",
				cron:      "15 * 1-32 * *",
			},
			wantErrMsg: "cron definition '15 * 1-32 * *' is invalid, unable to determine days value",
			wantErr:    true,
		},
		{
			name: "test14 - set hours",
			args: args{
				namespace: "example-com-main",
				cron:      "M/15 23 * * 0-5",
			},
			want: "1,16,31,46 23 * * 0-5",
		},
		{
			name: "test15 - set day",
			args: args{
				namespace: "example-com-main",
				cron:      "M/15 * 31 * 0-5",
			},
			want: "1,16,31,46 * 31 * 0-5",
		},
		{
			name: "test16 - set month",
			args: args{
				namespace: "example-com-main",
				cron:      "M/15 * * 11 0-5",
			},
			want: "1,16,31,46 * * 11 0-5",
		},
		{
			name: "test17 - hourly interval",
			args: args{
				namespace: "example-com-main",
				cron:      "M */6 * * *",
			},
			want: "31 1,7,13,19 * * *",
		},
		{
			name: "test18 - day and month string",
			args: args{
				namespace: "example-com-main",
				cron:      "M */6 * JAN MON",
			},
			want: "31 1,7,13,19 * JAN MON",
		},
		{
			name: "test19 - whitespace",
			args: args{
				namespace: "example-com-main",
				cron:      "M * * * * ",
			},
			want: "31 * * * *",
		},
		{
			name: "test20 - not enough fields",
			args: args{
				namespace: "example-com-main",
				cron:      "*/1 * * *",
			},
			wantErrMsg: "cron definition '*/1 * * *' is invalid, 4 fields provided, required 5",
			wantErr:    true,
		},
		{
			name: "test21 - too many fields",
			args: args{
				namespace: "example-com-main",
				cron:      "*/1 * * * * 7",
			},
			wantErrMsg: "cron definition '*/1 * * * * 7' is invalid, 6 fields provided, required 5",
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertCrontab(tt.args.namespace, tt.args.cron)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("ConvertCrontab() error = %v, wantErr %v", err, tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("ConvertCrontab() error = %v, wantErr %v", err.Error(), tt.wantErrMsg)
				}
			}
			if got != tt.want {
				if !tt.wantErr {
					t.Errorf("ConvertCrontab() = %v, want %v", got, tt.want)
				} else {
					t.Errorf("ConvertCrontab() = %v, wantErr %v", got, tt.wantErr)
				}
			}
		})
	}
}

func Test_isInPodCronjob(t *testing.T) {
	type args struct {
		cron string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "test1 - every 15 minutes, should be inpod",
			args: args{
				cron: "M/15 * * 11 0-5",
			},
			want: true,
		},
		{
			name: "test2 - every minute, should be in pod",
			args: args{
				cron: "* * * 11 0-5",
			},
			want: true,
		},
		{
			name: "test3 - on the 15 minute mark, should not be in pod",
			args: args{
				cron: "15 * * 11 0-5",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsInPodCronjob(tt.args.cron)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsInPodCronjob() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsInPodCronjob() = %v, want %v", got, tt.want)
			}
		})
	}
}
