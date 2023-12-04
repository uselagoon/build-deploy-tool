package generator

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
)

func Test_getRoutesFromAPIEnvVar(t *testing.T) {
	type args struct {
		envVars []lagoon.EnvironmentVariable
		debug   bool
	}
	tests := []struct {
		name    string
		args    args
		want    *lagoon.RoutesV2
		wantErr bool
	}{
		{
			name: "test1 - check that route in API is converted to RoutesV2",
			args: args{
				envVars: []lagoon.EnvironmentVariable{
					{Name: "LAGOON_ROUTES_JSON", Value: "eyJyb3V0ZXMiOlt7ImRvbWFpbiI6InRlc3QxLmV4YW1wbGUuY29tIiwic2VydmljZSI6Im5naW54IiwidGxzLWFjbWUiOmZhbHNlLCJtb25pdG9yaW5nLXBhdGgiOiIvYnlwYXNzLWNhY2hlIn1dfQo=", Scope: "build"},
				},
			},
			want: &lagoon.RoutesV2{
				Routes: []lagoon.RouteV2{
					{
						Domain:         "test1.example.com",
						LagoonService:  "nginx",
						TLSAcme:        helpers.BoolPtr(false),
						MonitoringPath: "/bypass-cache",
					},
				},
			},
		},
		{
			name: "test2 - check that route in API is converted to RoutesV2 with ingress class name",
			args: args{
				envVars: []lagoon.EnvironmentVariable{
					{Name: "LAGOON_ROUTES_JSON", Value: "eyJyb3V0ZXMiOlt7ImRvbWFpbiI6InRlc3QxLmV4YW1wbGUuY29tIiwic2VydmljZSI6Im5naW54IiwiaW5ncmVzc0NsYXNzIjoiY3VzdG9tLW5naW54IiwidGxzLWFjbWUiOmZhbHNlLCJtb25pdG9yaW5nLXBhdGgiOiIvYnlwYXNzLWNhY2hlIn1dfQ==", Scope: "build"},
				},
			},
			want: &lagoon.RoutesV2{
				Routes: []lagoon.RouteV2{
					{
						Domain:         "test1.example.com",
						LagoonService:  "nginx",
						TLSAcme:        helpers.BoolPtr(false),
						MonitoringPath: "/bypass-cache",
						IngressClass:   "custom-nginx",
					},
				},
			},
		},
		{
			name: "test3 - check that route in API is converted to RoutesV2 with ingress class name and hsts",
			args: args{
				envVars: []lagoon.EnvironmentVariable{
					{Name: "LAGOON_ROUTES_JSON", Value: "eyJyb3V0ZXMiOlt7ImRvbWFpbiI6InRlc3QxLmV4YW1wbGUuY29tIiwic2VydmljZSI6Im5naW54IiwiaW5ncmVzc0NsYXNzIjoiY3VzdG9tLW5naW54IiwidGxzLWFjbWUiOmZhbHNlLCJtb25pdG9yaW5nLXBhdGgiOiIvYnlwYXNzLWNhY2hlIiwiaHN0c0VuYWJsZWQiOnRydWUsImhzdHNNYXhBZ2UiOjM2MDAwfV19", Scope: "build"},
				},
			},
			want: &lagoon.RoutesV2{
				Routes: []lagoon.RouteV2{
					{
						Domain:         "test1.example.com",
						LagoonService:  "nginx",
						TLSAcme:        helpers.BoolPtr(false),
						MonitoringPath: "/bypass-cache",
						IngressClass:   "custom-nginx",
						HSTSEnabled:    helpers.BoolPtr(true),
						HSTSMaxAge:     36000,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getRoutesFromAPIEnvVar(tt.args.envVars, tt.args.debug)
			if (err != nil) != tt.wantErr {
				t.Errorf("getRoutesFromAPIEnvVar() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			lValues, _ := json.Marshal(got)
			wValues, _ := json.Marshal(tt.want)
			if !reflect.DeepEqual(string(lValues), string(wValues)) {
				t.Errorf("getRoutesFromAPIEnvVar() = %v, want %v", string(lValues), string(wValues))
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("getRoutesFromAPIEnvVar() = %v, want %v", got, tt.want)
			// }
		})
	}
}

func Test_generateAndMerge(t *testing.T) {
	type args struct {
		api         lagoon.RoutesV2
		envVars     []lagoon.EnvironmentVariable
		lagoonYAML  lagoon.YAML
		buildValues BuildValues
	}
	tests := []struct {
		name    string
		args    args
		want    lagoon.RoutesV2
		wantErr bool
	}{
		{
			name: "test1 - generate routes from lagoon yaml and merge ones from api onto them",
			args: args{
				buildValues: BuildValues{
					Branch: "main",
				},
				lagoonYAML: lagoon.YAML{
					Environments: lagoon.Environments{
						"main": lagoon.Environment{
							Routes: []map[string][]lagoon.Route{
								{
									"nginx": {
										{
											Ingresses: map[string]lagoon.Ingress{
												"a.example.com": {
													TLSAcme: helpers.BoolPtr(true),
												},
											},
										},
										{
											Name: "b.example.com",
										},
										{
											Name: "c.example.com",
										},
									},
								},
							},
						},
					},
				},
				api: lagoon.RoutesV2{
					Routes: []lagoon.RouteV2{
						{
							Domain:         "test1.example.com",
							LagoonService:  "nginx",
							TLSAcme:        helpers.BoolPtr(false),
							MonitoringPath: "/bypass-cache",
						},
						{
							Domain:         "a.example.com",
							LagoonService:  "nginx",
							TLSAcme:        helpers.BoolPtr(false),
							MonitoringPath: "/bypass-cache",
						},
					},
				},
			},
			want: lagoon.RoutesV2{
				Routes: []lagoon.RouteV2{
					{
						Domain:           "a.example.com",
						LagoonService:    "nginx",
						TLSAcme:          helpers.BoolPtr(false),
						Annotations:      map[string]string{},
						Insecure:         helpers.StrPtr("Redirect"),
						MonitoringPath:   "/bypass-cache",
						AlternativeNames: []string{},
						IngressName:      "a.example.com",
					},
					{
						Domain:           "b.example.com",
						LagoonService:    "nginx",
						TLSAcme:          helpers.BoolPtr(true),
						Annotations:      map[string]string{},
						Insecure:         helpers.StrPtr("Redirect"),
						MonitoringPath:   "/",
						AlternativeNames: []string{},
						IngressName:      "b.example.com",
					},
					{
						Domain:           "c.example.com",
						LagoonService:    "nginx",
						TLSAcme:          helpers.BoolPtr(true),
						Annotations:      map[string]string{},
						Insecure:         helpers.StrPtr("Redirect"),
						MonitoringPath:   "/",
						AlternativeNames: []string{},
						IngressName:      "c.example.com",
					},
					{
						Domain:           "test1.example.com",
						LagoonService:    "nginx",
						TLSAcme:          helpers.BoolPtr(false),
						MonitoringPath:   "/bypass-cache",
						Insecure:         helpers.StrPtr("Redirect"),
						Annotations:      map[string]string{},
						AlternativeNames: []string{},
						IngressName:      "test1.example.com",
					},
				},
			},
		},
		{
			name: "test2 - don't generate routes from lagoon yaml and only merge ones from api onto them",
			args: args{
				buildValues: BuildValues{
					Branch: "main",
				},
				lagoonYAML: lagoon.YAML{},
				api: lagoon.RoutesV2{
					Routes: []lagoon.RouteV2{
						{
							Domain:         "test1.example.com",
							LagoonService:  "nginx",
							TLSAcme:        helpers.BoolPtr(false),
							MonitoringPath: "/bypass-cache",
						},
						{
							Domain:         "a.example.com",
							LagoonService:  "nginx",
							TLSAcme:        helpers.BoolPtr(false),
							MonitoringPath: "/bypass-cache",
						},
					},
				},
			},
			want: lagoon.RoutesV2{
				Routes: []lagoon.RouteV2{
					{
						Domain:           "test1.example.com",
						LagoonService:    "nginx",
						TLSAcme:          helpers.BoolPtr(false),
						MonitoringPath:   "/bypass-cache",
						Insecure:         helpers.StrPtr("Redirect"),
						Annotations:      map[string]string{},
						AlternativeNames: []string{},
						IngressName:      "test1.example.com",
					},
					{
						Domain:           "a.example.com",
						LagoonService:    "nginx",
						TLSAcme:          helpers.BoolPtr(false),
						Annotations:      map[string]string{},
						Insecure:         helpers.StrPtr("Redirect"),
						MonitoringPath:   "/bypass-cache",
						AlternativeNames: []string{},
						IngressName:      "a.example.com",
					},
				},
			},
		},
		{
			name: "test3 - generate routes from lagoon yaml and merge ones from api onto them with ingress class",
			args: args{
				buildValues: BuildValues{
					Branch:       "main",
					IngressClass: "nginx",
				},
				lagoonYAML: lagoon.YAML{
					Environments: lagoon.Environments{
						"main": lagoon.Environment{
							Routes: []map[string][]lagoon.Route{
								{
									"nginx": {
										{
											Ingresses: map[string]lagoon.Ingress{
												"a.example.com": {
													TLSAcme: helpers.BoolPtr(true),
												},
											},
										},
										{
											Name: "b.example.com",
										},
										{
											Name: "c.example.com",
										},
									},
								},
							},
						},
					},
				},
				api: lagoon.RoutesV2{
					Routes: []lagoon.RouteV2{
						{
							Domain:         "test1.example.com",
							LagoonService:  "nginx",
							TLSAcme:        helpers.BoolPtr(false),
							MonitoringPath: "/bypass-cache",
						},
						{
							Domain:         "a.example.com",
							LagoonService:  "nginx",
							TLSAcme:        helpers.BoolPtr(false),
							MonitoringPath: "/bypass-cache",
						},
					},
				},
			},
			want: lagoon.RoutesV2{
				Routes: []lagoon.RouteV2{
					{
						Domain:           "a.example.com",
						LagoonService:    "nginx",
						TLSAcme:          helpers.BoolPtr(false),
						Annotations:      map[string]string{},
						Insecure:         helpers.StrPtr("Redirect"),
						MonitoringPath:   "/bypass-cache",
						IngressClass:     "nginx",
						AlternativeNames: []string{},
						IngressName:      "a.example.com",
					},
					{
						Domain:           "b.example.com",
						LagoonService:    "nginx",
						TLSAcme:          helpers.BoolPtr(true),
						Annotations:      map[string]string{},
						Insecure:         helpers.StrPtr("Redirect"),
						MonitoringPath:   "/",
						IngressClass:     "nginx",
						AlternativeNames: []string{},
						IngressName:      "b.example.com",
					},
					{
						Domain:           "c.example.com",
						LagoonService:    "nginx",
						TLSAcme:          helpers.BoolPtr(true),
						Annotations:      map[string]string{},
						Insecure:         helpers.StrPtr("Redirect"),
						MonitoringPath:   "/",
						IngressClass:     "nginx",
						AlternativeNames: []string{},
						IngressName:      "c.example.com",
					},
					{
						Domain:           "test1.example.com",
						LagoonService:    "nginx",
						TLSAcme:          helpers.BoolPtr(false),
						MonitoringPath:   "/bypass-cache",
						Insecure:         helpers.StrPtr("Redirect"),
						Annotations:      map[string]string{},
						IngressClass:     "nginx",
						AlternativeNames: []string{},
						IngressName:      "test1.example.com",
					},
				},
			},
		},
		{
			name: "test4 - generate routes from lagoon yaml and merge ones from api with hsts",
			args: args{
				buildValues: BuildValues{
					Branch:       "main",
					IngressClass: "nginx",
				},
				lagoonYAML: lagoon.YAML{
					Environments: lagoon.Environments{
						"main": lagoon.Environment{
							Routes: []map[string][]lagoon.Route{
								{
									"nginx": {
										{
											Ingresses: map[string]lagoon.Ingress{
												"a.example.com": {
													TLSAcme: helpers.BoolPtr(true),
												},
											},
										},
									},
								},
							},
						},
					},
				},
				api: lagoon.RoutesV2{
					Routes: []lagoon.RouteV2{
						{
							Domain:         "a.example.com",
							LagoonService:  "nginx",
							TLSAcme:        helpers.BoolPtr(false),
							MonitoringPath: "/bypass-cache",
							HSTSEnabled:    helpers.BoolPtr(true),
							HSTSMaxAge:     36000,
						},
					},
				},
			},
			want: lagoon.RoutesV2{
				Routes: []lagoon.RouteV2{
					{
						Domain:           "a.example.com",
						LagoonService:    "nginx",
						TLSAcme:          helpers.BoolPtr(false),
						Annotations:      map[string]string{},
						Insecure:         helpers.StrPtr("Redirect"),
						MonitoringPath:   "/bypass-cache",
						IngressClass:     "nginx",
						HSTSEnabled:      helpers.BoolPtr(true),
						HSTSMaxAge:       36000,
						AlternativeNames: []string{},
						IngressName:      "a.example.com",
					},
				},
			},
		},
		{
			name: "test5 - wildcard with tls-acme false",
			args: args{
				buildValues: BuildValues{
					Branch:       "main",
					IngressClass: "nginx",
				},
				lagoonYAML: lagoon.YAML{
					Environments: lagoon.Environments{
						"main": lagoon.Environment{
							Routes: []map[string][]lagoon.Route{
								{
									"nginx": {
										{
											Ingresses: map[string]lagoon.Ingress{
												"a.example.com": {
													TLSAcme:  helpers.BoolPtr(false),
													Wildcard: helpers.BoolPtr(true),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: lagoon.RoutesV2{
				Routes: []lagoon.RouteV2{
					{
						Domain:           "a.example.com",
						LagoonService:    "nginx",
						TLSAcme:          helpers.BoolPtr(false),
						Annotations:      map[string]string{},
						MonitoringPath:   "/",
						Insecure:         helpers.StrPtr("Redirect"),
						IngressClass:     "nginx",
						AlternativeNames: []string{},
						Wildcard:         helpers.BoolPtr(true),
						IngressName:      "wildcard-a.example.com",
					},
				},
			},
		},
		{
			name: "test6 - wildcard with tls-acme true (should error)",
			args: args{
				buildValues: BuildValues{
					Branch:       "main",
					IngressClass: "nginx",
				},
				lagoonYAML: lagoon.YAML{
					Environments: lagoon.Environments{
						"main": lagoon.Environment{
							Routes: []map[string][]lagoon.Route{
								{
									"nginx": {
										{
											Ingresses: map[string]lagoon.Ingress{
												"a.example.com": {
													TLSAcme:  helpers.BoolPtr(true),
													Wildcard: helpers.BoolPtr(true),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
			want: lagoon.RoutesV2{
				Routes: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateAndMerge(tt.args.api, tt.args.envVars, tt.args.lagoonYAML, tt.args.buildValues)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateAndMerge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			lValues, _ := json.Marshal(got)
			wValues, _ := json.Marshal(tt.want)
			if !reflect.DeepEqual(string(lValues), string(wValues)) && !tt.wantErr {
				t.Errorf("generateAndMerge() = %v, want %v", string(lValues), string(wValues))
			}
		})
	}
}

func Test_generateActiveStandbyRoutes(t *testing.T) {
	type args struct {
		envVars     []lagoon.EnvironmentVariable
		lagoonYAML  lagoon.YAML
		buildValues BuildValues
	}
	tests := []struct {
		name    string
		args    args
		want    lagoon.RoutesV2
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				buildValues: BuildValues{
					IsActiveEnvironment: true,
				},
				lagoonYAML: lagoon.YAML{
					ProductionRoutes: &lagoon.ProductionRoutes{
						Active: &lagoon.Environment{
							Routes: []map[string][]lagoon.Route{
								{
									"nginx": {
										{
											Name: "active.example.com",
										},
									},
								},
							},
						},
					},
				},
				envVars: []lagoon.EnvironmentVariable{},
			},
			want: lagoon.RoutesV2{
				Routes: []lagoon.RouteV2{
					{
						Domain:           "active.example.com",
						LagoonService:    "nginx",
						TLSAcme:          helpers.BoolPtr(true),
						Annotations:      map[string]string{},
						Migrate:          helpers.BoolPtr(true),
						Insecure:         helpers.StrPtr("Redirect"),
						MonitoringPath:   "/",
						AlternativeNames: []string{},
						IngressName:      "active.example.com",
					},
				},
			},
		},
		{
			name: "test2 - with ingress class defined",
			args: args{
				buildValues: BuildValues{
					IsActiveEnvironment: true,
					IngressClass:        "nginx",
				},
				lagoonYAML: lagoon.YAML{
					ProductionRoutes: &lagoon.ProductionRoutes{
						Active: &lagoon.Environment{
							Routes: []map[string][]lagoon.Route{
								{
									"nginx": {
										{
											Ingresses: map[string]lagoon.Ingress{
												"active.example.com": {
													TLSAcme: helpers.BoolPtr(true),
												},
											},
										},
									},
								},
							},
						},
					},
				},
				envVars: []lagoon.EnvironmentVariable{},
			},
			want: lagoon.RoutesV2{
				Routes: []lagoon.RouteV2{
					{
						Domain:           "active.example.com",
						LagoonService:    "nginx",
						TLSAcme:          helpers.BoolPtr(true),
						Annotations:      map[string]string{},
						Migrate:          helpers.BoolPtr(true),
						Insecure:         helpers.StrPtr("Redirect"),
						MonitoringPath:   "/",
						IngressClass:     "nginx",
						AlternativeNames: []string{},
						IngressName:      "active.example.com",
					},
				},
			},
		},
		{
			name: "test3 - with custom ingress class defined",
			args: args{
				buildValues: BuildValues{
					IsActiveEnvironment: true,
					IngressClass:        "nginx",
				},
				lagoonYAML: lagoon.YAML{
					ProductionRoutes: &lagoon.ProductionRoutes{
						Active: &lagoon.Environment{
							Routes: []map[string][]lagoon.Route{
								{
									"nginx": {
										{
											Ingresses: map[string]lagoon.Ingress{
												"active.example.com": {
													TLSAcme:      helpers.BoolPtr(true),
													IngressClass: "custom-nginx",
												},
											},
										},
									},
								},
							},
						},
					},
				},
				envVars: []lagoon.EnvironmentVariable{},
			},
			want: lagoon.RoutesV2{
				Routes: []lagoon.RouteV2{
					{
						Domain:           "active.example.com",
						LagoonService:    "nginx",
						TLSAcme:          helpers.BoolPtr(true),
						Annotations:      map[string]string{},
						Migrate:          helpers.BoolPtr(true),
						Insecure:         helpers.StrPtr("Redirect"),
						MonitoringPath:   "/",
						IngressClass:     "custom-nginx",
						AlternativeNames: []string{},
						IngressName:      "active.example.com",
					},
				},
			},
		},
		{
			name: "test4 - with wildcard and tls-acme true (should error)",
			args: args{
				buildValues: BuildValues{
					IngressClass:        "nginx",
					IsActiveEnvironment: true,
				},
				lagoonYAML: lagoon.YAML{
					ProductionRoutes: &lagoon.ProductionRoutes{
						Active: &lagoon.Environment{
							Routes: []map[string][]lagoon.Route{
								{
									"nginx": {
										{
											Ingresses: map[string]lagoon.Ingress{
												"active.example.com": {
													TLSAcme:  helpers.BoolPtr(true),
													Wildcard: helpers.BoolPtr(true),
												},
											},
										},
									},
								},
							},
						},
					},
				},
				envVars: []lagoon.EnvironmentVariable{},
			},
			wantErr: true,
			want: lagoon.RoutesV2{
				Routes: nil,
			},
		},
		{
			name: "test5 - with wildcard and tls-acme false",
			args: args{
				buildValues: BuildValues{
					IngressClass:        "nginx",
					IsActiveEnvironment: true,
				},
				lagoonYAML: lagoon.YAML{
					ProductionRoutes: &lagoon.ProductionRoutes{
						Active: &lagoon.Environment{
							Routes: []map[string][]lagoon.Route{
								{
									"nginx": {
										{
											Ingresses: map[string]lagoon.Ingress{
												"active.example.com": {
													TLSAcme:  helpers.BoolPtr(false),
													Wildcard: helpers.BoolPtr(true),
												},
											},
										},
									},
								},
							},
						},
					},
				},
				envVars: []lagoon.EnvironmentVariable{},
			},
			want: lagoon.RoutesV2{
				Routes: []lagoon.RouteV2{
					{
						Domain:           "active.example.com",
						LagoonService:    "nginx",
						TLSAcme:          helpers.BoolPtr(false),
						Annotations:      map[string]string{},
						Migrate:          helpers.BoolPtr(true),
						Insecure:         helpers.StrPtr("Redirect"),
						MonitoringPath:   "/",
						IngressClass:     "nginx",
						AlternativeNames: []string{},
						Wildcard:         helpers.BoolPtr(true),
						IngressName:      "wildcard-active.example.com",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateActiveStandbyRoutes(tt.args.envVars, tt.args.lagoonYAML, tt.args.buildValues)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateAndMerge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			lValues, _ := json.Marshal(got)
			wValues, _ := json.Marshal(tt.want)
			if !reflect.DeepEqual(string(lValues), string(wValues)) && !tt.wantErr {
				t.Errorf("generateAndMerge() = %v, want %v", string(lValues), string(wValues))
			}
		})
	}
}

func Test_autogeneratedDomainFromPattern(t *testing.T) {
	type args struct {
		pattern         string
		service         string
		projectName     string
		environmentName string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		{
			name: "test1",
			args: args{
				pattern:         "${service}-${environment}-${project}.example.com",
				service:         "nginx",
				projectName:     "example-com",
				environmentName: "main",
			},
			want:  "nginx-main-example-com.example.com",
			want1: "nginx-bvxea6pd-wjscrqcw.example.com",
		},
		{
			name: "test2",
			args: args{
				pattern:         "${service}.${environment}-${project}.example.com",
				service:         "nginx",
				projectName:     "example-com",
				environmentName: "main",
			},
			want:  "nginx.main-example-com.example.com",
			want1: "nginx.bvxea6pd-wjscrqcw.example.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := autogeneratedDomainFromPattern(tt.args.pattern, tt.args.service, tt.args.projectName, tt.args.environmentName)
			if got != tt.want {
				t.Errorf("autogeneratedDomainFromPattern() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("autogeneratedDomainFromPattern() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_generateAutogenRoutes(t *testing.T) {
	type args struct {
		envVars       []lagoon.EnvironmentVariable
		lagoonYAML    *lagoon.YAML
		buildValues   *BuildValues
		autogenRoutes *lagoon.RoutesV2
	}
	tests := []struct {
		name    string
		args    args
		want    lagoon.RoutesV2
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				envVars: []lagoon.EnvironmentVariable{
					{
						Name:  "LAGOON_SYSTEM_ROUTER_PATTERN",
						Value: "${service}-${project}-${environment}.example.com",
						Scope: "internal_system",
					},
				},
				lagoonYAML: &lagoon.YAML{},
				buildValues: &BuildValues{
					Project:         "example-com",
					BuildType:       "branch",
					Environment:     "main",
					EnvironmentType: "development",
					Namespace:       "example-com-main",
					Services: []ServiceValues{
						{
							Name:                       "nginx",
							Type:                       "nginx",
							AutogeneratedRoutesEnabled: true,
							AutogeneratedRoutesTLSAcme: true,
						},
					},
				},
				autogenRoutes: &lagoon.RoutesV2{},
			},
			want: lagoon.RoutesV2{
				Routes: []lagoon.RouteV2{
					{
						Domain:           "nginx-example-com-main.example.com",
						LagoonService:    "nginx",
						ComposeService:   "nginx",
						Autogenerated:    true,
						TLSAcme:          helpers.BoolPtr(true),
						Insecure:         helpers.StrPtr("Allow"),
						AlternativeNames: []string{},
						Labels: map[string]string{
							"app.kubernetes.io/instance": "nginx",
							"app.kubernetes.io/name":     "autogenerated-ingress",
							"helm.sh/chart":              "autogenerated-ingress-0.1.0",
							"lagoon.sh/autogenerated":    "true",
							"lagoon.sh/service":          "nginx",
							"lagoon.sh/service-type":     "nginx",
						},
						IngressName: "nginx",
					},
				},
			},
		},
		{
			name: "test2 - default ingress class",
			args: args{
				envVars: []lagoon.EnvironmentVariable{
					{
						Name:  "LAGOON_SYSTEM_ROUTER_PATTERN",
						Value: "${service}-${project}-${environment}.example.com",
						Scope: "internal_system",
					},
				},
				lagoonYAML: &lagoon.YAML{},
				buildValues: &BuildValues{
					Project:         "example-com",
					BuildType:       "branch",
					Environment:     "main",
					EnvironmentType: "development",
					Namespace:       "example-com-main",
					Services: []ServiceValues{
						{
							Name:                       "nginx",
							Type:                       "nginx",
							AutogeneratedRoutesEnabled: true,
							AutogeneratedRoutesTLSAcme: true,
						},
					},
					IngressClass: "nginx",
				},
				autogenRoutes: &lagoon.RoutesV2{},
			},
			want: lagoon.RoutesV2{
				Routes: []lagoon.RouteV2{
					{
						Domain:           "nginx-example-com-main.example.com",
						LagoonService:    "nginx",
						ComposeService:   "nginx",
						Autogenerated:    true,
						TLSAcme:          helpers.BoolPtr(true),
						IngressClass:     "nginx",
						Insecure:         helpers.StrPtr("Allow"),
						AlternativeNames: []string{},
						Labels: map[string]string{
							"app.kubernetes.io/instance": "nginx",
							"app.kubernetes.io/name":     "autogenerated-ingress",
							"helm.sh/chart":              "autogenerated-ingress-0.1.0",
							"lagoon.sh/autogenerated":    "true",
							"lagoon.sh/service":          "nginx",
							"lagoon.sh/service-type":     "nginx",
						},
						IngressName: "nginx",
					},
				},
			},
		},
		{
			name: "test2 - autogenerated routes ingress class",
			args: args{
				envVars: []lagoon.EnvironmentVariable{
					{
						Name:  "LAGOON_SYSTEM_ROUTER_PATTERN",
						Value: "${service}-${project}-${environment}.example.com",
						Scope: "internal_system",
					},
				},
				lagoonYAML: &lagoon.YAML{
					Routes: lagoon.Routes{
						Autogenerate: lagoon.Autogenerate{
							IngressClass: "custom-nginx",
						},
					},
				},
				buildValues: &BuildValues{
					Project:         "example-com",
					BuildType:       "branch",
					Environment:     "main",
					EnvironmentType: "development",
					Namespace:       "example-com-main",
					Services: []ServiceValues{
						{
							Name:                       "nginx",
							Type:                       "nginx",
							AutogeneratedRoutesEnabled: true,
							AutogeneratedRoutesTLSAcme: true,
						},
					},
					IngressClass: "nginx",
				},
				autogenRoutes: &lagoon.RoutesV2{},
			},
			want: lagoon.RoutesV2{
				Routes: []lagoon.RouteV2{
					{
						Domain:           "nginx-example-com-main.example.com",
						LagoonService:    "nginx",
						ComposeService:   "nginx",
						Autogenerated:    true,
						TLSAcme:          helpers.BoolPtr(true),
						IngressClass:     "custom-nginx",
						Insecure:         helpers.StrPtr("Allow"),
						AlternativeNames: []string{},
						Labels: map[string]string{
							"app.kubernetes.io/instance": "nginx",
							"app.kubernetes.io/name":     "autogenerated-ingress",
							"helm.sh/chart":              "autogenerated-ingress-0.1.0",
							"lagoon.sh/autogenerated":    "true",
							"lagoon.sh/service":          "nginx",
							"lagoon.sh/service-type":     "nginx",
						},
						IngressName: "nginx",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := generateAutogenRoutes(tt.args.envVars, tt.args.lagoonYAML, tt.args.buildValues, tt.args.autogenRoutes); (err != nil) != tt.wantErr {
				t.Errorf("generateAutogenRoutes() error = %v, wantErr %v", err, tt.wantErr)
			}
			lValues, _ := json.Marshal(tt.args.autogenRoutes)
			wValues, _ := json.Marshal(tt.want)
			if !reflect.DeepEqual(string(lValues), string(wValues)) {
				t.Errorf("generateAutogenRoutes() = %v, want %v", string(lValues), string(wValues))
			}
		})
	}
}

func Test_generateRoutes(t *testing.T) {
	type args struct {
		envVars            []lagoon.EnvironmentVariable
		buildValues        BuildValues
		lagoonYAML         lagoon.YAML
		autogenRoutes      *lagoon.RoutesV2
		mainRoutes         *lagoon.RoutesV2
		activeStanbyRoutes *lagoon.RoutesV2
		debug              bool
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   []string
		want2   []string
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				envVars: []lagoon.EnvironmentVariable{
					{
						Name:  "LAGOON_SYSTEM_ROUTER_PATTERN",
						Value: "${service}-${project}-${environment}.example.com",
						Scope: "internal_system",
					},
				},
				lagoonYAML: lagoon.YAML{},
				buildValues: BuildValues{
					Project:         "example-com",
					BuildType:       "branch",
					Environment:     "main",
					EnvironmentType: "development",
					Namespace:       "example-com-main",
					Services: []ServiceValues{
						{
							Name:                       "nginx",
							Type:                       "nginx",
							AutogeneratedRoutesEnabled: true,
							AutogeneratedRoutesTLSAcme: true,
						},
					},
				},
				autogenRoutes:      &lagoon.RoutesV2{},
				mainRoutes:         &lagoon.RoutesV2{},
				activeStanbyRoutes: &lagoon.RoutesV2{},
			},
			want:  "https://nginx-example-com-main.example.com",
			want1: []string{"https://nginx-example-com-main.example.com"},
			want2: []string{"https://nginx-example-com-main.example.com"},
		},
		{
			name: "test2",
			args: args{
				envVars: []lagoon.EnvironmentVariable{
					{
						Name:  "LAGOON_SYSTEM_ROUTER_PATTERN",
						Value: "${service}-${project}-${environment}.example.com",
						Scope: "internal_system",
					},
				},
				lagoonYAML: lagoon.YAML{
					Environments: lagoon.Environments{
						"main": lagoon.Environment{
							Routes: []map[string][]lagoon.Route{
								{
									"nginx": {
										{
											Ingresses: map[string]lagoon.Ingress{
												"a.example.com": {
													TLSAcme: helpers.BoolPtr(true),
												},
											},
										},
										{
											Name: "b.example.com",
										},
										{
											Name: "c.example.com",
										},
									},
								},
							},
						},
					},
				},
				buildValues: BuildValues{
					Project:         "example-com",
					BuildType:       "branch",
					Environment:     "main",
					Branch:          "main",
					EnvironmentType: "development",
					Namespace:       "example-com-main",
					Services: []ServiceValues{
						{
							Name:                       "nginx",
							Type:                       "nginx",
							AutogeneratedRoutesEnabled: true,
							AutogeneratedRoutesTLSAcme: true,
						},
					},
				},
				autogenRoutes:      &lagoon.RoutesV2{},
				mainRoutes:         &lagoon.RoutesV2{},
				activeStanbyRoutes: &lagoon.RoutesV2{},
			},
			want:  "https://a.example.com",
			want1: []string{"https://nginx-example-com-main.example.com", "https://a.example.com", "https://b.example.com", "https://c.example.com"},
			want2: []string{"https://nginx-example-com-main.example.com"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, err := generateRoutes(tt.args.envVars, tt.args.buildValues, tt.args.lagoonYAML, tt.args.autogenRoutes, tt.args.mainRoutes, tt.args.activeStanbyRoutes, tt.args.debug)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateRoutes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("generateRoutes() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("generateRoutes() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("generateRoutes() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}
