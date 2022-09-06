package routes

import (
	"os"
	"reflect"
	"testing"

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
				},
				activeStandby: false,
			},
			want: "test-resources/result-custom-ingress3.yaml",
		},
		{
			name: "test4 - invalid domain",
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
			name: "test5 - invalid annotation",
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
			name: "test6 - invalid label",
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
			name: "test7 - too long domain",
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
			name: "test8 - custom ingress with exceptionally log subdomain",
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
			want: "test-resources/result-custom-ingress4.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
