package helpers

import (
	"testing"
)

func TestCheckDBaaSProvider(t *testing.T) {
	type args struct {
		dbaasEndpoint    string
		dbaasType        string
		dbaasEnvironment string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "test1 - environment and provider that does exist",
			args: args{
				dbaasType:        "mariadb",
				dbaasEnvironment: "production",
			},
			want: true,
		},
		{
			name: "test2 - check for an environment that doesn't exist",
			args: args{
				dbaasType:        "mariadb",
				dbaasEnvironment: "development2",
			},
			wantErr: true,
			want:    false,
		},
		{
			name: "test3 - endpoint that doesn't resolve",
			args: args{
				dbaasEndpoint:    "http://this-does-not-exist",
				dbaasType:        "mariadb",
				dbaasEnvironment: "development2",
			},
			wantErr: true,
			want:    false,
		},
		{
			name: "test4 - type that doesn't exist",
			args: args{
				dbaasType:        "mariadb2",
				dbaasEnvironment: "production",
			},
			wantErr: true,
			want:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := TestDBaaSHTTPServer()
			defer ts.Close()
			testURL := ts.URL
			if tt.args.dbaasEndpoint != "" {
				testURL = tt.args.dbaasEndpoint
			}
			got, err := CheckDBaaSProvider(testURL, tt.args.dbaasType, tt.args.dbaasEnvironment)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckDBaaSProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CheckDBaaSProvider() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckDBaaSHealth(t *testing.T) {
	type args struct {
		dbaasEndpoint string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "test1 - should respond to health check",
			wantErr: false,
		},
		{
			name: "test2 - should not responsd to health check",
			args: args{
				dbaasEndpoint: "http://this-does-not-exist",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := TestDBaaSHTTPServer()
			defer ts.Close()
			testURL := ts.URL
			if tt.args.dbaasEndpoint != "" {
				testURL = tt.args.dbaasEndpoint
			}
			if err := CheckDBaaSHealth(testURL); (err != nil) != tt.wantErr {
				t.Errorf("CheckDBaaSHealth() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
