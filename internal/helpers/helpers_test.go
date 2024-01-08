package helpers

import (
	"os"
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

func TestIntPtr(t *testing.T) {
	type args struct {
		i int
	}
	t1 := 1
	t2 := 2
	tests := []struct {
		name string
		args args
		want *int
	}{
		{
			name: "test1",
			args: args{
				i: 1,
			},
			want: &t1,
		},
		{
			name: "test2",
			args: args{
				i: 2,
			},
			want: &t2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IntPtr(tt.args.i); *got != *tt.want {
				t.Errorf("IntPtr() = %v, want %v", *got, *tt.want)
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

func TestGetEnvInt(t *testing.T) {
	type args struct {
		key      string
		fallback int
		debug    bool
	}
	tests := []struct {
		name    string
		envVars map[string]string
		args    args
		want    int
	}{
		{
			name: "test1",
			args: args{
				key:      "TEST_VAR",
				fallback: 1,
				debug:    false,
			},
			want: 1,
		},
		{
			name: "test2",
			args: args{
				key:      "TEST_VAR",
				fallback: 2,
				debug:    false,
			},
			envVars: map[string]string{
				"TEST_VAR": "1",
			},
			want: 1,
		},
		{
			name: "test3",
			args: args{
				key:      "TEST_VAR",
				fallback: 2,
				debug:    false,
			},
			envVars: map[string]string{
				"TEST_VAR": "abcdef",
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		for k, v := range tt.envVars {
			os.Setenv(k, v)
		}
		t.Run(tt.name, func(t *testing.T) {
			if got := GetEnvInt(tt.args.key, tt.args.fallback, tt.args.debug); got != tt.want {
				t.Errorf("GetEnvInt() = %v, want %v", got, tt.want)
			}
		})
		t.Cleanup(func() {
			for k := range tt.envVars {
				os.Unsetenv(k)
			}
		})
	}
}

func TestEGetEnvInt(t *testing.T) {
	type args struct {
		key      string
		fallback int
		debug    bool
	}
	tests := []struct {
		name    string
		args    args
		envVars map[string]string
		want    int
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				key:      "TEST_VAR",
				fallback: 1,
				debug:    false,
			},
			want: 1,
		},
		{
			name: "test2",
			args: args{
				key:      "TEST_VAR",
				fallback: 2,
				debug:    false,
			},
			envVars: map[string]string{
				"TEST_VAR": "1",
			},
			want: 1,
		},
		{
			name: "test3",
			args: args{
				key:      "TEST_VAR",
				fallback: 2,
				debug:    false,
			},
			envVars: map[string]string{
				"TEST_VAR": "abcdef",
			},
			wantErr: true,
			want:    0,
		},
	}
	for _, tt := range tests {
		for k, v := range tt.envVars {
			os.Setenv(k, v)
		}
		t.Run(tt.name, func(t *testing.T) {
			got, err := EGetEnvInt(tt.args.key, tt.args.fallback, tt.args.debug)
			if (err != nil) != tt.wantErr {
				t.Errorf("EGetEnvInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("EGetEnvInt() = %v, want %v", got, tt.want)
			}
		})
		t.Cleanup(func() {
			for k := range tt.envVars {
				os.Unsetenv(k)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	type args struct {
		key      string
		fallback string
		debug    bool
	}
	tests := []struct {
		name    string
		args    args
		envVars map[string]string
		want    string
	}{
		{
			name: "test1",
			args: args{
				key:      "TEST_VAR",
				fallback: "abc",
				debug:    false,
			},
			want: "abc",
		},
		{
			name: "test2",
			args: args{
				key:      "TEST_VAR",
				fallback: "123",
				debug:    false,
			},
			envVars: map[string]string{
				"TEST_VAR": "abc",
			},
			want: "abc",
		},
		{
			name: "test3",
			args: args{
				key:      "TEST_VAR",
				fallback: "abcdef",
				debug:    false,
			},
			envVars: map[string]string{
				"TEST_VAR": "abcdef",
			},
			want: "abcdef",
		},
	}
	for _, tt := range tests {
		for k, v := range tt.envVars {
			os.Setenv(k, v)
		}
		t.Run(tt.name, func(t *testing.T) {
			if got := GetEnv(tt.args.key, tt.args.fallback, tt.args.debug); got != tt.want {
				t.Errorf("GetEnv() = %v, want %v", got, tt.want)
			}
		})
		t.Cleanup(func() {
			for k := range tt.envVars {
				os.Unsetenv(k)
			}
		})
	}
}

func TestGetEnvBool(t *testing.T) {
	type args struct {
		key      string
		fallback bool
		debug    bool
	}
	tests := []struct {
		name    string
		args    args
		envVars map[string]string
		want    bool
	}{
		{
			name: "test1",
			args: args{
				key:      "TEST_VAR",
				fallback: true,
				debug:    false,
			},
			want: true,
		},
		{
			name: "test2",
			args: args{
				key:      "TEST_VAR",
				fallback: false,
				debug:    false,
			},
			envVars: map[string]string{
				"TEST_VAR": "True",
			},
			want: true,
		},
		{
			name: "test3",
			args: args{
				key:      "TEST_VAR",
				fallback: true,
				debug:    false,
			},
			envVars: map[string]string{
				"TEST_VAR": "abcd12ef",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		for k, v := range tt.envVars {
			os.Setenv(k, v)
		}
		t.Run(tt.name, func(t *testing.T) {
			if got := GetEnvBool(tt.args.key, tt.args.fallback, tt.args.debug); got != tt.want {
				t.Errorf("GetEnvBool() = %v, want %v", got, tt.want)
			}
		})
		t.Cleanup(func() {
			for k := range tt.envVars {
				os.Unsetenv(k)
			}
		})
	}
}

func TestEGetEnvBool(t *testing.T) {
	type args struct {
		key      string
		fallback bool
		debug    bool
	}
	tests := []struct {
		name    string
		args    args
		envVars map[string]string
		want    bool
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				key:      "TEST_VAR",
				fallback: true,
				debug:    false,
			},
			want: true,
		},
		{
			name: "test2",
			args: args{
				key:      "TEST_VAR",
				fallback: false,
				debug:    false,
			},
			envVars: map[string]string{
				"TEST_VAR": "True",
			},
			want: true,
		},
		{
			name: "test3",
			args: args{
				key:      "TEST_VAR",
				fallback: true,
				debug:    false,
			},
			envVars: map[string]string{
				"TEST_VAR": "abcd12ef",
			},
			wantErr: true,
			want:    false,
		},
	}
	for _, tt := range tests {
		for k, v := range tt.envVars {
			os.Setenv(k, v)
		}
		t.Run(tt.name, func(t *testing.T) {
			got, err := EGetEnvBool(tt.args.key, tt.args.fallback, tt.args.debug)
			if (err != nil) != tt.wantErr {
				t.Errorf("EGetEnvBool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("EGetEnvBool() = %v, want %v", got, tt.want)
			}
		})
		t.Cleanup(func() {
			for k := range tt.envVars {
				os.Unsetenv(k)
			}
		})
	}
}

func TestCheckLabelLength(t *testing.T) {
	type args struct {
		labels map[string]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				labels: map[string]string{
					"app.kubernetes.io/instance":   "extra-long-name-f6c8a",
					"app.kubernetes.io/managed-by": "Helm",
					"app.kubernetes.io/name":       "custom-ingress",
					"helm.sh/chart":                "custom-ingress-0.1.0",
					"lagoon.sh/autogenerated":      "false",
					"lagoon.sh/buildType":          "branch",
					"lagoon.sh/environment":        "environment-with-really-really-reall-3fdb",
					"lagoon.sh/environmentType":    "development",
					"lagoon.sh/project":            "example-project",
					"lagoon.sh/service":            "extra-long-name-f6c8a",
					"lagoon.sh/service-type":       "custom-ingress",
				},
			},
		},
		{
			name: "test1",
			args: args{
				labels: map[string]string{
					"app.kubernetes.io/instance":   "extra-long-name.a-really-long-name-that-should-truncate.www.example.com",
					"app.kubernetes.io/managed-by": "Helm",
					"app.kubernetes.io/name":       "custom-ingress",
					"helm.sh/chart":                "custom-ingress-0.1.0",
					"lagoon.sh/autogenerated":      "false",
					"lagoon.sh/buildType":          "branch",
					"lagoon.sh/environment":        "environment-with-really-really-reall-3fdb",
					"lagoon.sh/environmentType":    "development",
					"lagoon.sh/project":            "example-project",
					"lagoon.sh/service":            "extra-long-name.a-really-long-name-that-should-truncate.www.example.com",
					"lagoon.sh/service-type":       "custom-ingress",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CheckLabelLength(tt.args.labels); (err != nil) != tt.wantErr {
				t.Errorf("CheckLabelLength() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetNamespace(t *testing.T) {
	type args struct {
		namespace string
		filename  string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test1 - no file",
			args: args{
				namespace: "test-namespace",
				filename:  "test-resources/test1",
			},
			want: "test-namespace",
		},
		{
			name: "test2 - test namespace file",
			args: args{
				namespace: "a-namespace",
				filename:  "test-resources/test2",
			},
			want: "test-namespace",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetNamespace(tt.args.namespace, tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNamespace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetNamespace() = %v, want %v", got, tt.want)
			}
		})
	}
}
