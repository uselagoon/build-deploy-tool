package generator

import (
	"reflect"
	"testing"
)

func Test_serverSnippetAddHeader(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		snippet string
		want    map[string]string
	}{
		{
			name: "test1",
			snippet: `
set_real_ip_from 1.2.3.4/22;
set_real_ip_from 1.2.3.5/24;
set_real_ip_from 1.2.3.6/23;
set_real_ip_from 1.2.3.7/24;
more_clear_headers "x-lagoon";
add_header X-Frame-Options "DENY" always;
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
add_header X-XSS-Protection "1; mode=block" always;
add_header X-Content-Type-Options "nosniff" always;
add_header Referrer-Policy "strict-origin-when-cross-origin" always;`,
			want: map[string]string{
				"X-Frame-Options":           "DENY",
				"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
				"X-XSS-Protection":          "1; mode=block",
				"X-Content-Type-Options":    "nosniff",
				"Referrer-Policy":           "strict-origin-when-cross-origin",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := serverSnippetAddHeader(tt.snippet)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("serverSnippetAddHeader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_serverSnippetSetRealIPFrom(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		snippet string
		want    []string
	}{
		{
			name: "test1",
			snippet: `
set_real_ip_from 1.2.3.4/22;
set_real_ip_from 1.2.3.5/24;
set_real_ip_from 1.2.3.6/23;
set_real_ip_from 1.2.3.7/24;
more_clear_headers "x-lagoon";
add_header X-Frame-Options "DENY" always;
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
add_header X-XSS-Protection "1; mode=block" always;
add_header X-Content-Type-Options "nosniff" always;
add_header Referrer-Policy "strict-origin-when-cross-origin" always;`,
			want: []string{"1.2.3.4/22", "1.2.3.5/24", "1.2.3.6/23", "1.2.3.7/24"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := serverSnippetSetRealIPFrom(tt.snippet)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("serverSnippetSetRealIPFrom() = %v, want %v", got, tt.want)
			}
		})
	}
}
