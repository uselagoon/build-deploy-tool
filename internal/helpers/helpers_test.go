package helpers

import (
	"reflect"
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
			name: "test1",
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

func TestGetSha256Hash(t *testing.T) {
	type args struct {
		text string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "test1",
			args: args{
				text: "project-name",
			},
			want: []byte{169, 9, 86, 41, 146, 87, 115, 234, 90, 144, 236, 164, 109, 248, 214, 230, 69, 131, 122, 227, 97, 100, 63, 130, 21, 132, 183, 169, 109, 5, 151, 35},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetSha256Hash(tt.args.text); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSha256Hash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetBase32EncodedLowercase(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{
				data: []byte{169, 9, 86, 41, 146, 87, 115, 234, 90, 144, 236, 164, 109, 248, 214, 230, 69, 131, 122, 227, 97, 100, 63, 130, 21, 132, 183, 169, 109, 5, 151, 35},
			},
			want: "veevmkmsk5z6uwuq5ssg36gw4zcyg6xdmfsd7aqvqs32s3ifs4rq====",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetBase32EncodedLowercase(tt.args.data); got != tt.want {
				t.Errorf("GetBase32EncodedLowercase() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}
