package dbaasclient

import (
	"testing"
	"time"
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
			d := NewClient(Client{
				RetryMax:     5,
				RetryWaitMin: time.Duration(10) * time.Millisecond,
				RetryWaitMax: time.Duration(50) * time.Millisecond,
			})
			got, err := d.CheckProvider(testURL, tt.args.dbaasType, tt.args.dbaasEnvironment)
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
			name: "test2 - should not respond to health check",
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
			d := NewClient(Client{
				RetryMax:     5,
				RetryWaitMin: time.Duration(10) * time.Millisecond,
				RetryWaitMax: time.Duration(50) * time.Millisecond,
			})
			if err := d.CheckHealth(testURL); (err != nil) != tt.wantErr {
				t.Errorf("CheckDBaaSHealth() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_addProtocol(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{
				url: "dbaas.local.svc:5000",
			},
			want: "http://dbaas.local.svc:5000",
		},
		{
			name: "test2",
			args: args{
				url: "https://dbaas.local.svc:5000",
			},
			want: "https://dbaas.local.svc:5000",
		},
		{
			name: "test3",
			args: args{
				url: "http://dbaas.local.svc:5000",
			},
			want: "http://dbaas.local.svc:5000",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := addProtocol(tt.args.url); got != tt.want {
				t.Errorf("addProtocol() = %v, want %v", got, tt.want)
			}
		})
	}
}
