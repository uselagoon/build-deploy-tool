package routes

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/uselagoon/build-deploy-tool/internal/dbaasclient"
	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
)

func TestGenerateKubeTemplate(t *testing.T) {
	type args struct {
		route         lagoon.RouteV2
		values        generator.BuildValues
		activeStandby bool
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "active-standby1",
			args: args{
				route: lagoon.RouteV2{
					Domain:         "extra-long-name.a-really-long-name-that-should-truncate.www.example.com",
					LagoonService:  "nginx",
					MonitoringPath: "/",
					Insecure:       helpers.StrPtr("Redirect"),
					TLSAcme:        helpers.BoolPtr(true),
					Migrate:        helpers.BoolPtr(true),
					Annotations: map[string]string{
						"custom-annotation": "custom annotation value",
					},
					Fastly: lagoon.Fastly{
						Watch: false,
					},
				},
				values: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-with-really-really-reall-3fdb",
					EnvironmentType: "production",
					Namespace:       "myexample-project-environment-with-really-really-reall-3fdb",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "lagoon.local",
					Branch:          "environment-with-really-really-reall-3fdb",
					Monitoring: generator.MonitoringConfig{
						AlertContact: "abcdefg",
						StatusPageID: "12345",
						Enabled:      true,
					},
					Route: "https://extra-long-name.a-really-long-name-that-should-truncate.www.example.com/",
				},
				activeStandby: true,
			},
			want: "test-resources/result-active-standby1.yaml",
		},
		{
			name: "custom-ingress1",
			args: args{
				route: lagoon.RouteV2{
					Domain:         "extra-long-name.a-really-long-name-that-should-truncate.www.example.com",
					LagoonService:  "nginx",
					MonitoringPath: "/",
					Insecure:       helpers.StrPtr("Redirect"),
					TLSAcme:        helpers.BoolPtr(true),
					Migrate:        helpers.BoolPtr(false),
					Annotations: map[string]string{
						"custom-annotation": "custom annotation value",
					},
					Fastly: lagoon.Fastly{
						Watch: false,
					},
				},
				values: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-with-really-really-reall-3fdb",
					EnvironmentType: "production",
					Namespace:       "myexample-project-environment-with-really-really-reall-3fdb",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "lagoon.local",
					Branch:          "environment-with-really-really-reall-3fdb",
					Monitoring: generator.MonitoringConfig{
						AlertContact: "abcdefg",
						StatusPageID: "12345",
						Enabled:      true,
					},
					Route: "https://extra-long-name.a-really-long-name-that-should-truncate.www.example.com/",
				},
				activeStandby: false,
			},
			want: "test-resources/result-custom-ingress1.yaml",
		},
		{
			name: "custom-ingress2",
			args: args{
				route: lagoon.RouteV2{
					Domain:         "extra-long-name.a-really-long-name-that-should-truncate.www.example.com",
					LagoonService:  "nginx",
					MonitoringPath: "/",
					Insecure:       helpers.StrPtr("Redirect"),
					TLSAcme:        helpers.BoolPtr(true),
					Migrate:        helpers.BoolPtr(false),
					Annotations: map[string]string{
						"custom-annotation": "custom annotation value",
					},
					Fastly: lagoon.Fastly{
						Watch: false,
					},
				},
				values: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-with-really-really-reall-3fdb",
					EnvironmentType: "development",
					Namespace:       "myexample-project-environment-with-really-really-reall-3fdb",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "lagoon.local",
					Branch:          "environment-with-really-really-reall-3fdb",
					Monitoring: generator.MonitoringConfig{
						AlertContact: "abcdefg",
						StatusPageID: "12345",
						Enabled:      true,
					},
					Route: "https://extra-long-name.a-really-long-name-that-should-truncate.www.example.com/",
				},
				activeStandby: false,
			},
			want: "test-resources/result-custom-ingress2.yaml",
		},
		{
			name: "test3 - custom ingress with ingress class",
			args: args{
				route: lagoon.RouteV2{
					Domain:         "extra-long-name.a-really-long-name-that-should-truncate.www.example.com",
					LagoonService:  "nginx",
					MonitoringPath: "/",
					Insecure:       helpers.StrPtr("Redirect"),
					TLSAcme:        helpers.BoolPtr(true),
					Migrate:        helpers.BoolPtr(false),
					Annotations: map[string]string{
						"custom-annotation": "custom annotation value",
					},
					Fastly: lagoon.Fastly{
						Watch: false,
					},
					IngressClass: "nginx",
				},
				values: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-with-really-really-reall-3fdb",
					EnvironmentType: "development",
					Namespace:       "myexample-project-environment-with-really-really-reall-3fdb",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "lagoon.local",
					Branch:          "environment-with-really-really-reall-3fdb",
					Monitoring: generator.MonitoringConfig{
						AlertContact: "abcdefg",
						StatusPageID: "12345",
						Enabled:      true,
					},
					Route: "https://extra-long-name.a-really-long-name-that-should-truncate.www.example.com/",
				},
				activeStandby: false,
			},
			want: "test-resources/result-custom-ingress3.yaml",
		},
		{
			name: "test4 - custom ingress with ingress class and hsts",
			args: args{
				route: lagoon.RouteV2{
					Domain:         "extra-long-name.a-really-long-name-that-should-truncate.www.example.com",
					LagoonService:  "nginx",
					MonitoringPath: "/",
					Insecure:       helpers.StrPtr("Redirect"),
					TLSAcme:        helpers.BoolPtr(true),
					Migrate:        helpers.BoolPtr(false),
					Annotations: map[string]string{
						"custom-annotation": "custom annotation value",
					},
					Fastly: lagoon.Fastly{
						Watch: false,
					},
					IngressClass: "nginx",
					HSTSEnabled:  helpers.BoolPtr(true),
					HSTSMaxAge:   31536000,
				},
				values: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-with-really-really-reall-3fdb",
					EnvironmentType: "development",
					Namespace:       "myexample-project-environment-with-really-really-reall-3fdb",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "lagoon.local",
					Branch:          "environment-with-really-really-reall-3fdb",
					Monitoring: generator.MonitoringConfig{
						AlertContact: "abcdefg",
						StatusPageID: "12345",
						Enabled:      true,
					},
					Route: "https://extra-long-name.a-really-long-name-that-should-truncate.www.example.com/",
				},
				activeStandby: false,
			},
			want: "test-resources/result-custom-ingress4.yaml",
		},
		{
			name: "test5 - custom ingress with ingress class and hsts and existing config snippet",
			args: args{
				route: lagoon.RouteV2{
					Domain:         "extra-long-name.a-really-long-name-that-should-truncate.www.example.com",
					LagoonService:  "nginx",
					MonitoringPath: "/",
					Insecure:       helpers.StrPtr("Redirect"),
					TLSAcme:        helpers.BoolPtr(true),
					Migrate:        helpers.BoolPtr(false),
					Annotations: map[string]string{
						"custom-annotation": "custom annotation value",
						"nginx.ingress.kubernetes.io/configuration-snippet": "more_set_headers \"MyCustomHeader: Value\";",
					},
					Fastly: lagoon.Fastly{
						Watch: false,
					},
					IngressClass:          "nginx",
					HSTSEnabled:           helpers.BoolPtr(true),
					HSTSMaxAge:            31536000,
					HSTSIncludeSubdomains: helpers.BoolPtr(true),
					HSTSPreload:           helpers.BoolPtr(true),
				},
				values: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-with-really-really-reall-3fdb",
					EnvironmentType: "development",
					Namespace:       "myexample-project-environment-with-really-really-reall-3fdb",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "lagoon.local",
					Branch:          "environment-with-really-really-reall-3fdb",
					Monitoring: generator.MonitoringConfig{
						AlertContact: "abcdefg",
						StatusPageID: "12345",
						Enabled:      true,
					},
					Route: "https://extra-long-name.a-really-long-name-that-should-truncate.www.example.com/",
				},
				activeStandby: false,
			},
			want: "test-resources/result-custom-ingress5.yaml",
		},
		{
			name: "test6 - invalid domain",
			args: args{
				route: lagoon.RouteV2{
					Domain:         "fail@.extra-long-name.a-really-long-name-that-should-truncate.www.example.com",
					LagoonService:  "nginx",
					MonitoringPath: "/",
					Insecure:       helpers.StrPtr("Redirect"),
					TLSAcme:        helpers.BoolPtr(true),
					Migrate:        helpers.BoolPtr(false),
					Annotations: map[string]string{
						"custom-annotation": "custom annotation value",
					},
					Fastly: lagoon.Fastly{
						Watch: false,
					},
					IngressClass: "nginx",
				},
				values: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-with-really-really-reall-3fdb",
					EnvironmentType: "development",
					Namespace:       "myexample-project-environment-with-really-really-reall-3fdb",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "lagoon.local",
					Branch:          "environment-with-really-really-reall-3fdb",
					Monitoring: generator.MonitoringConfig{
						AlertContact: "abcdefg",
						StatusPageID: "12345",
						Enabled:      true,
					},
				},
				activeStandby: false,
			},
			wantErr: true,
		},
		{
			name: "test7 - invalid annotation",
			args: args{
				route: lagoon.RouteV2{
					Domain:         "extra-long-name.a-really-long-name-that-should-truncate.www.example.com",
					LagoonService:  "nginx",
					MonitoringPath: "/",
					Insecure:       helpers.StrPtr("Redirect"),
					TLSAcme:        helpers.BoolPtr(true),
					Migrate:        helpers.BoolPtr(false),
					Annotations: map[string]string{
						"custom-annotation": "custom annotation value",
						"@invalid":          "this is an invalid annotation",
					},
					Fastly: lagoon.Fastly{
						Watch: false,
					},
					IngressClass: "nginx",
				},
				values: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-with-really-really-reall-3fdb",
					EnvironmentType: "development",
					Namespace:       "myexample-project-environment-with-really-really-reall-3fdb",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "lagoon.local",
					Branch:          "environment-with-really-really-reall-3fdb",
					Monitoring: generator.MonitoringConfig{
						AlertContact: "abcdefg",
						StatusPageID: "12345",
						Enabled:      true,
					},
				},
				activeStandby: false,
			},
			wantErr: true,
		},
		{
			name: "test8 - invalid label",
			args: args{
				route: lagoon.RouteV2{
					Domain:         "extra-long-name.a-really-long-name-that-should-truncate.www.example.com",
					LagoonService:  "nginx",
					MonitoringPath: "/",
					Insecure:       helpers.StrPtr("Redirect"),
					TLSAcme:        helpers.BoolPtr(true),
					Migrate:        helpers.BoolPtr(false),
					Annotations: map[string]string{
						"custom-annotation": "custom annotation value",
					},
					Labels: map[string]string{
						"@invalid": "this is an invalid annotation",
					},
					Fastly: lagoon.Fastly{
						Watch: false,
					},
					IngressClass: "nginx",
				},
				values: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-with-really-really-reall-3fdb",
					EnvironmentType: "development",
					Namespace:       "myexample-project-environment-with-really-really-reall-3fdb",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "lagoon.local",
					Branch:          "environment-with-really-really-reall-3fdb",
					Monitoring: generator.MonitoringConfig{
						AlertContact: "abcdefg",
						StatusPageID: "12345",
						Enabled:      true,
					},
				},
				activeStandby: false,
			},
			wantErr: true,
		},
		{
			name: "test9 - too long domain",
			args: args{
				route: lagoon.RouteV2{
					Domain:         "extra-long-name.a-really-long-name-that-should-truncate.extra-long-name.a-really-long-name-that-should-truncate.extra-long-name.a-really-long-name-that-should-truncate.extra-long-name.a-really-long-name-that-should-truncate.extra-long-name.a-really-long-name-that-should-truncate.www.example.com",
					LagoonService:  "nginx",
					MonitoringPath: "/",
					Insecure:       helpers.StrPtr("Redirect"),
					TLSAcme:        helpers.BoolPtr(true),
					Migrate:        helpers.BoolPtr(false),
					Annotations: map[string]string{
						"custom-annotation": "custom annotation value",
					},
					Fastly: lagoon.Fastly{
						Watch: false,
					},
					IngressClass: "nginx",
				},
				values: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-with-really-really-reall-3fdb",
					EnvironmentType: "development",
					Namespace:       "myexample-project-environment-with-really-really-reall-3fdb",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "lagoon.local",
					Branch:          "environment-with-really-really-reall-3fdb",
					Monitoring: generator.MonitoringConfig{
						AlertContact: "abcdefg",
						StatusPageID: "12345",
						Enabled:      true,
					},
				},
				activeStandby: false,
			},
			wantErr: true,
		},
		{
			name: "test10 - custom ingress with exceptionally long subdomain",
			args: args{
				route: lagoon.RouteV2{
					Domain:         "hmm-this-is-a-really-long-branch-name-designed-to-test-a-specific-feature.www.example.com",
					LagoonService:  "nginx",
					MonitoringPath: "/",
					Insecure:       helpers.StrPtr("Redirect"),
					TLSAcme:        helpers.BoolPtr(true),
					Migrate:        helpers.BoolPtr(false),
					Annotations: map[string]string{
						"custom-annotation": "custom annotation value",
					},
					Fastly: lagoon.Fastly{
						Watch: false,
					},
					IngressClass: "nginx",
				},
				values: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-with-really-really-reall-3fdb",
					EnvironmentType: "development",
					Namespace:       "myexample-project-environment-with-really-really-reall-3fdb",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "lagoon.local",
					Branch:          "environment-with-really-really-reall-3fdb",
					Monitoring: generator.MonitoringConfig{
						AlertContact: "abcdefg",
						StatusPageID: "12345",
						Enabled:      true,
					},
				},
				activeStandby: false,
			},
			want: "test-resources/result-custom-ingress6.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// add dbaasclient overrides for tests
			tt.args.values.DBaaSClient = dbaasclient.NewClient(dbaasclient.Client{
				RetryMax:     5,
				RetryWaitMin: time.Duration(10) * time.Millisecond,
				RetryWaitMax: time.Duration(50) * time.Millisecond,
			})
			got, err := GenerateIngressTemplate(tt.args.route, tt.args.values)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("couldn't generate template %v: %v", tt.want, err)
				}
			}
			if got != nil && tt.wantErr {
				t.Errorf("wanted an error, but didn't get one")
			}
			if !tt.wantErr {
				r1, err := os.ReadFile(tt.want)
				if err != nil {
					t.Errorf("couldn't read file %v: %v", tt.want, err)
				}
				if !reflect.DeepEqual(string(got), string(r1)) {
					t.Errorf("GenerateIngressTemplate() = %v, want %v", string(got), string(r1))
				}
			}
		})
	}
}
