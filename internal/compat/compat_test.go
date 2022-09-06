package compat

import (
	"testing"
)

func Test_checkVersion(t *testing.T) {
	type args struct {
		version string
	}
	tests := []struct {
		name                string
		args                args
		supportedMinVersion string
		want                bool
	}{
		{
			name: "test1",
			args: args{
				version: "v2.8.4",
			},
			supportedMinVersion: "v2.9.0",
			want:                false,
		},
		{
			name: "test2",
			args: args{
				version: "v2.9.0",
			},
			supportedMinVersion: "v2.9.0",
			want:                true,
		},
		{
			name: "test3",
			args: args{
				version: "v2.9.3",
			},
			supportedMinVersion: "v2.9.0",
			want:                true,
		},
		{
			name: "test4",
			args: args{
				version: "",
			},
			supportedMinVersion: "v2.9.0",
			want:                false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkVersion(tt.args.version, tt.supportedMinVersion); got != tt.want {
				t.Errorf("checkVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckVersion(t *testing.T) {
	type args struct {
		version string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test1",
			args: args{
				version: "v1.0.0",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckVersion(tt.args.version); got != tt.want {
				t.Errorf("CheckVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
