package helpers

import (
	"testing"
)

func TestGetMD5HashWithNewLine(t *testing.T) {
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
			if got := GetMD5HashWithNewLine(tt.args.text); got != tt.want {
				t.Errorf("getMD5Hash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStrPtr(t *testing.T) {
	type args struct {
		str string
	}
	t1 := "true"
	t2 := "false"
	tests := []struct {
		name string
		args args
		want *string
	}{
		{
			name: "test1",
			args: args{
				str: "true",
			},
			want: &t1,
		},
		{
			name: "test2",
			args: args{
				str: "false",
			},
			want: &t2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StrPtr(tt.args.str); *got != *tt.want {
				t.Errorf("StrPtr() = %v, want %v", *got, *tt.want)
			}
		})
	}
}

func TestBoolPtr(t *testing.T) {
	type args struct {
		b bool
	}
	t1 := true
	t2 := false
	tests := []struct {
		name string
		args args
		want *bool
	}{
		{
			name: "test1",
			args: args{
				b: true,
			},
			want: &t1,
		},
		{
			name: "test2",
			args: args{
				b: false,
			},
			want: &t2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BoolPtr(tt.args.b); *got != *tt.want {
				t.Errorf("BoolPtr() = %v, want %v", *got, *tt.want)
			}
		})
	}
}
