package templating

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/andreyvit/diff"
	"github.com/compose-spec/compose-go/types"
	"github.com/uselagoon/build-deploy-tool/internal/dbaasclient"
	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
)

func TestGenerateIngressTemplate(t *testing.T) {
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
					IngressName: "extra-long-name.a-really-long-name-that-should-truncate.www.example.com",
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
					Services: []generator.ServiceValues{
						{
							Name:         "nginx",
							OverrideName: "nginx",
							Type:         "nginx-php",
						},
					},
					Route: "https://extra-long-name.a-really-long-name-that-should-truncate.www.example.com/",
				},
				activeStandby: true,
			},
			want: "test-resources/ingress/result-active-standby1.yaml",
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
					IngressName: "extra-long-name.a-really-long-name-that-should-truncate.www.example.com",
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
					Services: []generator.ServiceValues{
						{
							Name:         "nginx",
							OverrideName: "nginx",
							Type:         "nginx-php",
						},
					},
					Route: "https://extra-long-name.a-really-long-name-that-should-truncate.www.example.com/",
				},
				activeStandby: false,
			},
			want: "test-resources/ingress/result-custom-ingress1.yaml",
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
					IngressName: "extra-long-name.a-really-long-name-that-should-truncate.www.example.com",
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
					Services: []generator.ServiceValues{
						{
							Name:         "nginx",
							OverrideName: "nginx",
							Type:         "nginx-php",
						},
					},
					Route: "https://extra-long-name.a-really-long-name-that-should-truncate.www.example.com/",
				},
				activeStandby: false,
			},
			want: "test-resources/ingress/result-custom-ingress2.yaml",
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
					IngressName:  "extra-long-name.a-really-long-name-that-should-truncate.www.example.com",
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
					Services: []generator.ServiceValues{
						{
							Name:         "nginx",
							OverrideName: "nginx",
							Type:         "nginx-php",
						},
					},
					Route: "https://extra-long-name.a-really-long-name-that-should-truncate.www.example.com/",
				},
				activeStandby: false,
			},
			want: "test-resources/ingress/result-custom-ingress3.yaml",
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
					IngressName:  "extra-long-name.a-really-long-name-that-should-truncate.www.example.com",
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
					Services: []generator.ServiceValues{
						{
							Name:         "nginx",
							OverrideName: "nginx",
							Type:         "nginx-php",
						},
					},
					Route: "https://extra-long-name.a-really-long-name-that-should-truncate.www.example.com/",
				},
				activeStandby: false,
			},
			want: "test-resources/ingress/result-custom-ingress4.yaml",
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
					IngressName:           "extra-long-name.a-really-long-name-that-should-truncate.www.example.com",
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
					Services: []generator.ServiceValues{
						{
							Name:         "nginx",
							OverrideName: "nginx",
							Type:         "nginx-php",
						},
					},
					Route: "https://extra-long-name.a-really-long-name-that-should-truncate.www.example.com/",
				},
				activeStandby: false,
			},
			want: "test-resources/ingress/result-custom-ingress5.yaml",
		},
		{
			name: "test6 - invalid annotation",
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
					IngressName:  "extra-long-name.a-really-long-name-that-should-truncate.www.example.com",
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
					Services: []generator.ServiceValues{
						{
							Name:         "nginx",
							OverrideName: "nginx",
							Type:         "nginx-php",
						},
					},
				},
				activeStandby: false,
			},
			wantErr: true,
		},
		{
			name: "test7 - invalid label",
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
					IngressName:  "extra-long-name.a-really-long-name-that-should-truncate.www.example.com",
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
			name: "test8 - custom ingress with exceptionally long subdomain",
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
					IngressName:  "hmm-this-is-a-really-long-branch-name-designed-to-test-a-specific-feature.www.example.com",
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
					Services: []generator.ServiceValues{
						{
							Name:         "nginx",
							OverrideName: "nginx",
							Type:         "nginx-php",
						},
					},
				},
				activeStandby: false,
			},
			want: "test-resources/ingress/result-custom-ingress6.yaml",
		},
		{
			name: "test9 - wildcard ingress",
			args: args{
				route: lagoon.RouteV2{
					Domain:         "www.example.com",
					LagoonService:  "nginx",
					MonitoringPath: "/",
					Insecure:       helpers.StrPtr("Redirect"),
					TLSAcme:        helpers.BoolPtr(false),
					Migrate:        helpers.BoolPtr(false),
					Annotations: map[string]string{
						"custom-annotation": "custom annotation value",
					},
					Fastly: lagoon.Fastly{
						Watch: false,
					},
					IngressClass: "nginx",
					Wildcard:     helpers.BoolPtr(true),
					WildcardApex: helpers.BoolPtr(true),
					IngressName:  "wildcard-www.example.com",
				},
				values: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment",
					EnvironmentType: "development",
					Namespace:       "myexample-project-environment",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "lagoon.local",
					Branch:          "environment",
					Monitoring: generator.MonitoringConfig{
						AlertContact: "abcdefg",
						StatusPageID: "12345",
						Enabled:      true,
					},
					Services: []generator.ServiceValues{
						{
							Name:         "nginx",
							OverrideName: "nginx",
							Type:         "nginx-php",
						},
					},
				},
				activeStandby: false,
			},
			want: "test-resources/ingress/result-wildcard-ingress1.yaml",
		},
		{
			name: "test10 - wildcard ingress",
			args: args{
				route: lagoon.RouteV2{
					Domain:         "this-truncate.extra-long-name.a-really-long-name-that-should-truncate.extra-long-name.a-really-long-name-that-should-truncate.extra-long-name.a-really-long-name-that-should-truncate.extra-long-name.a-really-long-name-that-should-truncate.www.example.com",
					LagoonService:  "nginx",
					MonitoringPath: "/",
					Insecure:       helpers.StrPtr("Redirect"),
					TLSAcme:        helpers.BoolPtr(false),
					Migrate:        helpers.BoolPtr(false),
					Annotations: map[string]string{
						"custom-annotation": "custom annotation value",
					},
					Fastly: lagoon.Fastly{
						Watch: false,
					},
					IngressClass: "nginx",
					Wildcard:     helpers.BoolPtr(true),
					WildcardApex: helpers.BoolPtr(true),
					IngressName:  "wildcard-this-truncate.extra-long-name.a-really-long-name-that-should-truncate.extra-long-name.a-really-long-name-that-should-truncate.extra-long-name.a-really-long-name-that-should-truncate.extra-long-name.a-really-long-name-that-should-truncate.www.e-f1945",
				},
				values: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment",
					EnvironmentType: "development",
					Namespace:       "myexample-project-environment",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "lagoon.local",
					Branch:          "environment",
					Monitoring: generator.MonitoringConfig{
						AlertContact: "abcdefg",
						StatusPageID: "12345",
						Enabled:      true,
					},
					Services: []generator.ServiceValues{
						{
							Name:         "nginx",
							OverrideName: "nginx",
							Type:         "nginx-php",
						},
					},
				},
				activeStandby: false,
			},
			want: "test-resources/ingress/result-wildcard-ingress2.yaml",
		},
		{
			name: "custom-ingress1 with specific port service",
			args: args{
				route: lagoon.RouteV2{
					Domain:         "extra-long-name.a-really-long-name-that-should-truncate.www.example.com",
					LagoonService:  "myservice-po",
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
					IngressName:         "extra-long-name.a-really-long-name-that-should-truncate.www.example.com",
					RequestVerification: helpers.BoolPtr(true),
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
					Services: []generator.ServiceValues{
						{
							Name:         "myservice-po",
							OverrideName: "myservice-po",
							Type:         "basic",
							AdditionalServicePorts: []generator.AdditionalServicePort{
								{
									ServiceName: "myservice-po-8192",
									ServicePort: types.ServicePortConfig{
										Target:   8192,
										Protocol: "tcp",
									},
								},
								{
									ServiceName: "myservice-po-8211",
									ServicePort: types.ServicePortConfig{
										Target:   8211,
										Protocol: "tcp",
									},
								},
							},
						},
					},
				},
				activeStandby: false,
			},
			want: "test-resources/ingress/result-custom-ingress7.yaml",
		},
		{
			name: "custom-ingress1 with specific port service 2",
			args: args{
				route: lagoon.RouteV2{
					Domain:         "extra-long-name.a-really-long-name-that-should-truncate.www.example.com",
					LagoonService:  "myservice-po-8192",
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
					IngressName:         "extra-long-name.a-really-long-name-that-should-truncate.www.example.com",
					RequestVerification: helpers.BoolPtr(true),
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
					Services: []generator.ServiceValues{
						{
							Name:         "myservice-po",
							OverrideName: "myservice-po",
							Type:         "basic",
							AdditionalServicePorts: []generator.AdditionalServicePort{
								{
									ServiceName: "myservice-po-8192",
									ServicePort: types.ServicePortConfig{
										Target:   8192,
										Protocol: "tcp",
									},
								},
								{
									ServiceName: "myservice-po-8211",
									ServicePort: types.ServicePortConfig{
										Target:   8211,
										Protocol: "tcp",
									},
								},
							},
						},
					},
				},
				activeStandby: false,
			},
			want: "test-resources/ingress/result-custom-ingress8.yaml",
		},
		{
			name: "custom-ingress9",
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
					RequestVerification: helpers.BoolPtr(true),
					IngressName:         "extra-long-name.a-really-long-name-that-should-truncate.www.example.com",
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
					Services: []generator.ServiceValues{
						{
							Name:         "nginx",
							OverrideName: "nginx",
							Type:         "basic",
						},
					},
					Route: "https://extra-long-name.a-really-long-name-that-should-truncate.www.example.com/",
				},
				activeStandby: false,
			},
			want: "test-resources/ingress/result-custom-ingress9.yaml",
		},
		{
			name: "wildcard ingress no apex",
			args: args{
				route: lagoon.RouteV2{
					Domain:         "www.example.com",
					LagoonService:  "nginx",
					MonitoringPath: "/",
					Insecure:       helpers.StrPtr("Redirect"),
					TLSAcme:        helpers.BoolPtr(false),
					Migrate:        helpers.BoolPtr(false),
					Annotations: map[string]string{
						"custom-annotation": "custom annotation value",
					},
					Fastly: lagoon.Fastly{
						Watch: false,
					},
					IngressClass: "nginx",
					Wildcard:     helpers.BoolPtr(true),
					WildcardApex: helpers.BoolPtr(false),
					IngressName:  "wildcard-www.example.com",
				},
				values: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment",
					EnvironmentType: "development",
					Namespace:       "myexample-project-environment",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "lagoon.local",
					Branch:          "environment",
					Monitoring: generator.MonitoringConfig{
						AlertContact: "abcdefg",
						StatusPageID: "12345",
						Enabled:      true,
					},
					Services: []generator.ServiceValues{
						{
							Name:         "nginx",
							OverrideName: "nginx",
							Type:         "nginx-php",
						},
					},
				},
				activeStandby: false,
			},
			want: "test-resources/ingress/result-wildcard-ingress3.yaml",
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
				gotR, err := TemplateIngress(got)
				if err != nil {
					t.Errorf("couldn't generate template  %v", err)
				}
				if !reflect.DeepEqual(string(gotR), string(r1)) {
					t.Errorf("GenerateIngressTemplate() = \n%v", diff.LineDiff(string(r1), string(gotR)))
				}
			}
		})
	}
}
