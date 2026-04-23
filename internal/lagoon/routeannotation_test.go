package lagoon

import (
	"testing"
)

func TestRouteAnnotationServerSnippet(t *testing.T) {
	var testCases = map[string]struct {
		input string
		valid bool
	}{
		"valid rewrite": {
			input: "rewrite ^/abc(.*) https://dev.example.com/abc$1 permanent;\n",
			valid: true,
		},
		"valid rewrite twice": {
			input: "rewrite foo bar;\nrewrite something else;\n",
			valid: true,
		},
		"invalid rewrites": {
			input: "rewrites foo bar;",
			valid: false,
		},
		"invalid if block": {
			input: "if ($request_uri !~ \"^/abc\") {\n  return 301 https://dev.example.com$request_uri;\n}\n",
			valid: false,
		},
		"valid add_header": {
			input: "add_header X-Robots-Tag \"noindex, nofollow\";\n",
			valid: true,
		},
		"valid add_header no newline": {
			input: "add_header X-Robots-Tag \"noindex, nofollow\";",
			valid: true,
		},
		"valid add_header custom": {
			input: "add_header X-branch \"#main\";\n",
			valid: true,
		},
		"valid double add_header": {
			input: "add_header X-Robots-Tag \"noindex, nofollow\"; add_header X-Robots-Tag \"noindex, nofollow\";",
			valid: true,
		},
		"valid more_set_headers": {
			input: "more_set_headers \"Strict-Transport-Security: max-age=31536000\";\n",
			valid: true,
		},
		"invalid more_set_input_headers": {
			input: "more_set_input_headers \"X-Foo: bar\";\n",
			valid: false,
		},
		"valid set_real_ip_from": {
			input: "set_real_ip_from 128.128.128.128;\nset_real_ip_from 128.128.128.128;\n",
			valid: true,
		},
		"invalid set_real_ip_from": {
			input: "set_real_ip_from 128.128.128.128;\nif (true) { return 301 http://example.com;\n};\n",
			valid: false,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			for _, snippet := range []string{serverSnippet, configurationSnippet} {
				l := RoutesV2{
					Routes: []RouteV2{
						{
							Domain: "www.example.com",
							Annotations: map[string]string{
								snippet: tc.input,
							},
						},
					},
				}
				err := validateRouteAnnotations(l)
				if tc.valid {
					if err != nil {
						tt.Fatalf("unexpected error %v", err)
					}
				} else {
					if err == nil {
						tt.Fatalf("expected error, but got nil")
					}
				}
			}
		})
	}
}

func TestRouteAnnotationRestrictedSnippets(t *testing.T) {
	var testCases = map[string]struct {
		input string
		valid bool
	}{
		"restrict configuration-snippet": {
			input: "nginx.ingress.kubernetes.io/configuration-snippet",
			valid: false,
		},
		"restrict modsecurity-snippet": {
			input: "nginx.ingress.kubernetes.io/modsecurity-snippet",
			valid: false,
		},
		"restrict auth-snippet": {
			input: "nginx.ingress.kubernetes.io/auth-snippet",
			valid: false,
		},
		"allow whitelist-source-range": {
			input: "nginx.ingress.kubernetes.io/whitelist-source-range",
			valid: true,
		},
		"restrict traefik-middleware": {
			input: "traefik.ingress.kubernetes.io/router.middlewares",
			valid: false,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			l := RoutesV2{
				Routes: []RouteV2{
					{
						Domain: "www.example.com",
						Annotations: map[string]string{
							tc.input: "any value",
						},
					},
				},
			}
			err := validateRouteAnnotations(l)
			if tc.valid {
				if err != nil {
					tt.Fatalf("unexpected error %v", err)
				}
			} else {
				if err == nil {
					tt.Fatalf("expected error, but got nil")
				}
			}
		})
	}
}
