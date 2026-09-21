package lagoon

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
)

func TestGenerateRouteStructure(t *testing.T) {
	type args struct {
		yamlRoutes          *RoutesV2
		yamlRouteMap        map[string][]Route
		variables           []EnvironmentVariable
		defaultIngressClass string
		activeStandby       bool
	}
	tests := []struct {
		name    string
		args    args
		want    *RoutesV2
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				yamlRoutes: &RoutesV2{},
				yamlRouteMap: map[string][]Route{
					"nginx": {
						{
							Name: "example.com",
						},
						{
							Name: "www.example.com",
						},
					},
				},
				activeStandby: false,
			},
			want: &RoutesV2{
				Routes: []RouteV2{
					{
						Domain:         "example.com",
						LagoonService:  "nginx",
						MonitoringPath: "/",
						Insecure:       helpers.StrPtr("Redirect"),
						TLSAcme:        helpers.BoolPtr(true),
						Annotations:    map[string]string{},
						Fastly: Fastly{
							Watch: false,
						},
						AlternativeNames:    []string{},
						IngressName:         "example.com",
						RequestVerification: helpers.BoolPtr(false),
					},
					{
						Domain:         "www.example.com",
						LagoonService:  "nginx",
						MonitoringPath: "/",
						Insecure:       helpers.StrPtr("Redirect"),
						TLSAcme:        helpers.BoolPtr(true),
						Annotations:    map[string]string{},
						Fastly: Fastly{
							Watch: false,
						},
						AlternativeNames:    []string{},
						IngressName:         "www.example.com",
						RequestVerification: helpers.BoolPtr(false),
					},
				},
			},
		},
		{
			name: "test2",
			args: args{
				yamlRoutes: &RoutesV2{},
				yamlRouteMap: map[string][]Route{
					"nginx": {
						{
							Name: "example.com",
						},
						{
							Ingresses: map[string]Ingress{
								"www.example.com": {
									Fastly: Fastly{
										Watch:     true,
										ServiceID: "12345",
									},
								},
							},
						},
					},
				},
				activeStandby: false,
			},
			want: &RoutesV2{
				Routes: []RouteV2{
					{
						Domain:         "example.com",
						LagoonService:  "nginx",
						MonitoringPath: "/",
						Insecure:       helpers.StrPtr("Redirect"),
						TLSAcme:        helpers.BoolPtr(true),
						Annotations:    map[string]string{},
						Fastly: Fastly{
							Watch: false,
						},
						AlternativeNames:    []string{},
						IngressName:         "example.com",
						RequestVerification: helpers.BoolPtr(false),
					},
					{
						Domain:         "www.example.com",
						LagoonService:  "nginx",
						MonitoringPath: "/",
						Insecure:       helpers.StrPtr("Redirect"),
						TLSAcme:        helpers.BoolPtr(true),
						Annotations:    map[string]string{},
						Fastly: Fastly{
							Watch:     true,
							ServiceID: "12345",
						},
						AlternativeNames:    []string{},
						IngressName:         "www.example.com",
						RequestVerification: helpers.BoolPtr(false),
					},
				},
			},
		},
		{
			name: "test3",
			args: args{
				yamlRoutes: &RoutesV2{},
				yamlRouteMap: map[string][]Route{
					"nginx": {
						{
							Ingresses: map[string]Ingress{
								"example.com": {
									Fastly: Fastly{
										Watch:     true,
										ServiceID: "12345",
									},
									AlternativeNames: []string{
										"www.example.com",
										"en.example.com",
									},
								},
							},
						},
					},
				},
				activeStandby: false,
			},
			want: &RoutesV2{
				Routes: []RouteV2{
					{
						Domain:         "example.com",
						LagoonService:  "nginx",
						MonitoringPath: "/",
						Insecure:       helpers.StrPtr("Redirect"),
						TLSAcme:        helpers.BoolPtr(true),
						Annotations:    map[string]string{},
						Fastly: Fastly{
							Watch:     true,
							ServiceID: "12345",
						},
						AlternativeNames: []string{
							"www.example.com",
							"en.example.com",
						},
						IngressName:         "example.com",
						RequestVerification: helpers.BoolPtr(false),
					},
				},
			},
		},
		{
			name: "test4 - ingress class",
			args: args{
				yamlRoutes: &RoutesV2{},
				yamlRouteMap: map[string][]Route{
					"nginx": {
						{
							Name: "example.com",
						},
						{
							Name: "www.example.com",
						},
					},
				},
				defaultIngressClass: "nginx",
				activeStandby:       false,
			},
			want: &RoutesV2{
				Routes: []RouteV2{
					{
						Domain:         "example.com",
						LagoonService:  "nginx",
						IngressClass:   "nginx",
						MonitoringPath: "/",
						Insecure:       helpers.StrPtr("Redirect"),
						TLSAcme:        helpers.BoolPtr(true),
						Annotations:    map[string]string{},
						Fastly: Fastly{
							Watch: false,
						},
						AlternativeNames:    []string{},
						IngressName:         "example.com",
						RequestVerification: helpers.BoolPtr(false),
					},
					{
						Domain:         "www.example.com",
						LagoonService:  "nginx",
						IngressClass:   "nginx",
						MonitoringPath: "/",
						Insecure:       helpers.StrPtr("Redirect"),
						TLSAcme:        helpers.BoolPtr(true),
						Annotations:    map[string]string{},
						Fastly: Fastly{
							Watch: false,
						},
						AlternativeNames:    []string{},
						IngressName:         "www.example.com",
						RequestVerification: helpers.BoolPtr(false),
					},
				},
			},
		},
		{
			name: "test5 - custom ingress class on one route",
			args: args{
				yamlRoutes: &RoutesV2{},
				yamlRouteMap: map[string][]Route{
					"nginx": {
						{
							Name: "example.com",
						},
						{
							Ingresses: map[string]Ingress{
								"www.example.com": {
									Fastly: Fastly{
										Watch:     true,
										ServiceID: "12345",
									},
									IngressClass: "custom-ingress",
								},
							},
						},
					},
				},
				defaultIngressClass: "nginx",
				activeStandby:       false,
			},
			want: &RoutesV2{
				Routes: []RouteV2{
					{
						Domain:         "example.com",
						LagoonService:  "nginx",
						IngressClass:   "nginx",
						MonitoringPath: "/",
						Insecure:       helpers.StrPtr("Redirect"),
						TLSAcme:        helpers.BoolPtr(true),
						Annotations:    map[string]string{},
						Fastly: Fastly{
							Watch: false,
						},
						AlternativeNames:    []string{},
						IngressName:         "example.com",
						RequestVerification: helpers.BoolPtr(false),
					},
					{
						Domain:         "www.example.com",
						LagoonService:  "nginx",
						IngressClass:   "custom-ingress",
						MonitoringPath: "/",
						Insecure:       helpers.StrPtr("Redirect"),
						TLSAcme:        helpers.BoolPtr(true),
						Annotations:    map[string]string{},
						Fastly: Fastly{
							Watch:     true,
							ServiceID: "12345",
						},
						AlternativeNames:    []string{},
						IngressName:         "www.example.com",
						RequestVerification: helpers.BoolPtr(false),
					},
				},
			},
		},
		{
			name: "test6 - hsts",
			args: args{
				yamlRoutes: &RoutesV2{},
				yamlRouteMap: map[string][]Route{
					"nginx": {
						{
							Name: "example.com",
						},
						{
							Ingresses: map[string]Ingress{
								"www.example.com": {
									Fastly: Fastly{
										Watch:     true,
										ServiceID: "12345",
									},
									HSTSEnabled: helpers.BoolPtr(true),
									HSTSMaxAge:  10000,
								},
							},
						},
					},
				},
				activeStandby: false,
			},
			want: &RoutesV2{
				Routes: []RouteV2{
					{
						Domain:         "example.com",
						LagoonService:  "nginx",
						MonitoringPath: "/",
						Insecure:       helpers.StrPtr("Redirect"),
						TLSAcme:        helpers.BoolPtr(true),
						Annotations:    map[string]string{},
						Fastly: Fastly{
							Watch: false,
						},
						AlternativeNames:    []string{},
						IngressName:         "example.com",
						RequestVerification: helpers.BoolPtr(false),
					},
					{
						Domain:         "www.example.com",
						LagoonService:  "nginx",
						MonitoringPath: "/",
						Insecure:       helpers.StrPtr("Redirect"),
						TLSAcme:        helpers.BoolPtr(true),
						Annotations:    map[string]string{},
						Fastly: Fastly{
							Watch:     true,
							ServiceID: "12345",
						},
						HSTSEnabled:         helpers.BoolPtr(true),
						HSTSMaxAge:          10000,
						AlternativeNames:    []string{},
						IngressName:         "www.example.com",
						RequestVerification: helpers.BoolPtr(false),
					},
				},
			},
		},
		{
			name: "test7 - wildcard with tls-acme true (should error)",
			args: args{
				yamlRoutes: &RoutesV2{},
				yamlRouteMap: map[string][]Route{
					"nginx": {
						{
							Ingresses: map[string]Ingress{
								"www.example.com": {
									TLSAcme:  helpers.BoolPtr(true),
									Wildcard: helpers.BoolPtr(true),
								},
							},
						},
					},
				},
				activeStandby: false,
			},
			wantErr: true,
			want: &RoutesV2{
				Routes: nil,
			},
		},
		{
			name: "test7 - wildcard with tls-acme false",
			args: args{
				yamlRoutes: &RoutesV2{},
				yamlRouteMap: map[string][]Route{
					"nginx": {
						{
							Ingresses: map[string]Ingress{
								"www.example.com": {
									TLSAcme:  helpers.BoolPtr(false),
									Wildcard: helpers.BoolPtr(true),
								},
							},
						},
					},
				},
				activeStandby: false,
			},
			want: &RoutesV2{
				Routes: []RouteV2{
					{
						Domain:              "www.example.com",
						LagoonService:       "nginx",
						MonitoringPath:      "/",
						Insecure:            helpers.StrPtr("Redirect"),
						TLSAcme:             helpers.BoolPtr(false),
						Annotations:         map[string]string{},
						AlternativeNames:    []string{},
						Wildcard:            helpers.BoolPtr(true),
						WildcardApex:        helpers.BoolPtr(true),
						IngressName:         "wildcard-www.example.com",
						RequestVerification: helpers.BoolPtr(false),
					},
				},
			},
		},
		{
			name: "test8 - wildcard with tls-acme false wildcard apex disabled",
			args: args{
				yamlRoutes: &RoutesV2{},
				yamlRouteMap: map[string][]Route{
					"nginx": {
						{
							Ingresses: map[string]Ingress{
								"www.example.com": {
									TLSAcme:      helpers.BoolPtr(false),
									Wildcard:     helpers.BoolPtr(true),
									WildcardApex: helpers.BoolPtr(false),
								},
							},
						},
					},
				},
				activeStandby: false,
			},
			want: &RoutesV2{
				Routes: []RouteV2{
					{
						Domain:              "www.example.com",
						LagoonService:       "nginx",
						MonitoringPath:      "/",
						Insecure:            helpers.StrPtr("Redirect"),
						TLSAcme:             helpers.BoolPtr(false),
						Annotations:         map[string]string{},
						AlternativeNames:    []string{},
						Wildcard:            helpers.BoolPtr(true),
						WildcardApex:        helpers.BoolPtr(false),
						IngressName:         "wildcard-www.example.com",
						RequestVerification: helpers.BoolPtr(false),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := GenerateRoutesV2(tt.args.yamlRoutes, tt.args.yamlRouteMap, tt.args.variables, tt.args.defaultIngressClass, tt.args.activeStandby)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateRouteStructure() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(tt.args.yamlRoutes, tt.want) && !tt.wantErr {
				stra, _ := json.Marshal(tt.args.yamlRoutes)
				strb, _ := json.Marshal(tt.want)
				t.Errorf("GenerateRouteStructure() = %v, want %v", string(stra), string(strb))
			}
		})
	}
}

func TestMergeRouteStructures(t *testing.T) {
	type args struct {
		yamlRoutes          RoutesV2
		apiRoutes           RoutesV2
		variables           []EnvironmentVariable
		defaultIngressClass string
	}
	tests := []struct {
		name    string
		args    args
		want    RoutesV2
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				yamlRoutes: RoutesV2{
					Routes: []RouteV2{
						{
							Domain:         "example.com",
							LagoonService:  "nginx",
							MonitoringPath: "/",
							Insecure:       helpers.StrPtr("Redirect"),
							TLSAcme:        helpers.BoolPtr(true),
							Annotations:    map[string]string{},
							Fastly: Fastly{
								Watch:     true,
								ServiceID: "12345",
							},
							IngressName:         "example.com",
							RequestVerification: helpers.BoolPtr(false),
						},
						{
							Domain:              "www.example.com",
							LagoonService:       "nginx",
							MonitoringPath:      "/",
							Insecure:            helpers.StrPtr("Redirect"),
							TLSAcme:             helpers.BoolPtr(true),
							Annotations:         map[string]string{},
							IngressName:         "www.example.com",
							RequestVerification: helpers.BoolPtr(false),
						},
						{
							Domain:              "hsts.example.com",
							LagoonService:       "nginx",
							MonitoringPath:      "/",
							Insecure:            helpers.StrPtr("Redirect"),
							TLSAcme:             helpers.BoolPtr(true),
							Annotations:         map[string]string{},
							HSTSEnabled:         helpers.BoolPtr(true),
							HSTSMaxAge:          20000,
							IngressName:         "hsts.example.com",
							RequestVerification: helpers.BoolPtr(false),
						},
					},
				},
				apiRoutes: RoutesV2{
					Routes: []RouteV2{
						{
							Domain:         "www.example.com",
							LagoonService:  "nginx",
							MonitoringPath: "/",
							Insecure:       helpers.StrPtr("Redirect"),
							TLSAcme:        helpers.BoolPtr(true),
							Annotations: map[string]string{
								"nginx": "nginx",
							},
							RequestVerification: helpers.BoolPtr(false),
						},
						{
							Domain:              "another.example.com",
							LagoonService:       "nginx",
							MonitoringPath:      "/",
							Insecure:            helpers.StrPtr("Redirect"),
							TLSAcme:             helpers.BoolPtr(true),
							Annotations:         map[string]string{},
							RequestVerification: helpers.BoolPtr(false),
						},
						{
							Domain:              "hsts.example.com",
							LagoonService:       "nginx",
							MonitoringPath:      "/",
							Insecure:            helpers.StrPtr("Redirect"),
							TLSAcme:             helpers.BoolPtr(true),
							Annotations:         map[string]string{},
							HSTSEnabled:         helpers.BoolPtr(true),
							HSTSMaxAge:          10000,
							RequestVerification: helpers.BoolPtr(false),
						},
					},
				},
			},
			want: RoutesV2{
				Routes: []RouteV2{
					{
						Domain:         "example.com",
						LagoonService:  "nginx",
						MonitoringPath: "/",
						Insecure:       helpers.StrPtr("Redirect"),
						TLSAcme:        helpers.BoolPtr(true),
						Annotations:    map[string]string{},
						Fastly: Fastly{
							Watch:     true,
							ServiceID: "12345",
						},
						AlternativeNames:    []string{},
						IngressName:         "example.com",
						RequestVerification: helpers.BoolPtr(false),
					},
					{
						Domain:         "www.example.com",
						LagoonService:  "nginx",
						MonitoringPath: "/",
						Insecure:       helpers.StrPtr("Redirect"),
						TLSAcme:        helpers.BoolPtr(true),
						Annotations: map[string]string{
							"nginx": "nginx",
						},
						AlternativeNames:    []string{},
						IngressName:         "www.example.com",
						RequestVerification: helpers.BoolPtr(false),
						Source:              "API",
					},
					{
						Domain:              "hsts.example.com",
						LagoonService:       "nginx",
						MonitoringPath:      "/",
						Insecure:            helpers.StrPtr("Redirect"),
						TLSAcme:             helpers.BoolPtr(true),
						Annotations:         map[string]string{},
						HSTSEnabled:         helpers.BoolPtr(true),
						HSTSMaxAge:          10000,
						AlternativeNames:    []string{},
						IngressName:         "hsts.example.com",
						RequestVerification: helpers.BoolPtr(false),
						Source:              "API",
					},
					{
						Domain:              "another.example.com",
						LagoonService:       "nginx",
						MonitoringPath:      "/",
						Insecure:            helpers.StrPtr("Redirect"),
						TLSAcme:             helpers.BoolPtr(true),
						Annotations:         map[string]string{},
						AlternativeNames:    []string{},
						IngressName:         "another.example.com",
						RequestVerification: helpers.BoolPtr(false),
						Source:              "API",
					},
				},
			},
		},
		{
			name: "test2 - wildcard with tls-acme changed to false",
			args: args{
				yamlRoutes: RoutesV2{
					Routes: []RouteV2{
						{
							Domain:         "example.com",
							LagoonService:  "nginx",
							MonitoringPath: "/",
							Insecure:       helpers.StrPtr("Redirect"),
							TLSAcme:        helpers.BoolPtr(true),
							Annotations:    map[string]string{},
							Wildcard:       helpers.BoolPtr(true),
						},
						{
							Domain:         "a.example.com",
							LagoonService:  "nginx",
							MonitoringPath: "/",
							Insecure:       helpers.StrPtr("Redirect"),
							TLSAcme:        helpers.BoolPtr(false),
							Annotations:    map[string]string{},
							Wildcard:       helpers.BoolPtr(true),
						},
					},
				},
				apiRoutes: RoutesV2{
					Routes: []RouteV2{
						{
							Domain:         "example.com",
							LagoonService:  "nginx",
							MonitoringPath: "/",
							Insecure:       helpers.StrPtr("Redirect"),
							TLSAcme:        helpers.BoolPtr(false),
							Annotations:    map[string]string{},
							Wildcard:       helpers.BoolPtr(true),
						},
						{
							Domain:         "a.example.com",
							LagoonService:  "nginx",
							MonitoringPath: "/",
							Insecure:       helpers.StrPtr("Redirect"),
							TLSAcme:        helpers.BoolPtr(false),
							Annotations:    map[string]string{},
							Wildcard:       helpers.BoolPtr(true),
						},
					},
				},
			},
			want: RoutesV2{
				Routes: []RouteV2{
					{
						Domain:              "example.com",
						LagoonService:       "nginx",
						MonitoringPath:      "/",
						Insecure:            helpers.StrPtr("Redirect"),
						TLSAcme:             helpers.BoolPtr(false),
						Annotations:         map[string]string{},
						AlternativeNames:    []string{},
						Wildcard:            helpers.BoolPtr(true),
						WildcardApex:        helpers.BoolPtr(true),
						IngressName:         "wildcard-example.com",
						RequestVerification: helpers.BoolPtr(false),
						Source:              "API",
					},
					{
						Domain:              "a.example.com",
						LagoonService:       "nginx",
						MonitoringPath:      "/",
						Insecure:            helpers.StrPtr("Redirect"),
						TLSAcme:             helpers.BoolPtr(false),
						Annotations:         map[string]string{},
						AlternativeNames:    []string{},
						Wildcard:            helpers.BoolPtr(true),
						WildcardApex:        helpers.BoolPtr(true),
						IngressName:         "wildcard-a.example.com",
						RequestVerification: helpers.BoolPtr(false),
						Source:              "API",
					},
				},
			},
		},
		{
			name: "test3 - wildcard with tls-acme true (should error)",
			args: args{
				yamlRoutes: RoutesV2{
					Routes: []RouteV2{
						{
							Domain:         "example.com",
							LagoonService:  "nginx",
							MonitoringPath: "/",
							Insecure:       helpers.StrPtr("Redirect"),
							TLSAcme:        helpers.BoolPtr(true),
							Annotations:    map[string]string{},
							Wildcard:       helpers.BoolPtr(true),
						},
					},
				},
				apiRoutes: RoutesV2{
					Routes: []RouteV2{
						{
							Domain:         "example.com",
							LagoonService:  "nginx",
							MonitoringPath: "/",
							Insecure:       helpers.StrPtr("Redirect"),
							TLSAcme:        helpers.BoolPtr(true),
							Annotations:    map[string]string{},
							Wildcard:       helpers.BoolPtr(true),
						},
					},
				},
			},
			wantErr: true,
			want: RoutesV2{
				Routes: nil,
			},
		},
		{
			name: "test4 - invalid yaml route",
			args: args{
				yamlRoutes: RoutesV2{
					Routes: []RouteV2{
						{
							Domain:         "*._re/f#3safasF*.was_-..asfexample.com",
							LagoonService:  "nginx",
							MonitoringPath: "/",
							Insecure:       helpers.StrPtr("Redirect"),
							TLSAcme:        helpers.BoolPtr(true),
							Annotations:    map[string]string{},
							Wildcard:       helpers.BoolPtr(true),
						},
					},
				},
				apiRoutes: RoutesV2{
					Routes: []RouteV2{
						{
							Domain:         "fail@example.com",
							LagoonService:  "nginx",
							MonitoringPath: "/",
							Insecure:       helpers.StrPtr("Redirect"),
							TLSAcme:        helpers.BoolPtr(true),
							Annotations:    map[string]string{},
							Wildcard:       helpers.BoolPtr(true),
						},
					},
				},
			},
			wantErr: true,
			want: RoutesV2{
				Routes: nil,
			},
		},
		{
			name: "test5 - invalid api route",
			args: args{
				yamlRoutes: RoutesV2{
					Routes: []RouteV2{
						{
							Domain:         "example.com",
							LagoonService:  "nginx",
							MonitoringPath: "/",
							Insecure:       helpers.StrPtr("Redirect"),
							TLSAcme:        helpers.BoolPtr(true),
							Annotations:    map[string]string{},
							Wildcard:       helpers.BoolPtr(true),
						},
					},
				},
				apiRoutes: RoutesV2{
					Routes: []RouteV2{
						{
							Domain:         "fail@example.com",
							LagoonService:  "nginx",
							MonitoringPath: "/",
							Insecure:       helpers.StrPtr("Redirect"),
							TLSAcme:        helpers.BoolPtr(true),
							Annotations:    map[string]string{},
							Wildcard:       helpers.BoolPtr(true),
						},
					},
				},
			},
			wantErr: true,
			want: RoutesV2{
				Routes: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MergeRoutesV2(tt.args.yamlRoutes, tt.args.apiRoutes, tt.args.variables, tt.args.defaultIngressClass)
			if (err != nil) != tt.wantErr {
				t.Errorf("MergeRouteStructures() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) && !tt.wantErr {
				stra, _ := json.Marshal(got)
				strb, _ := json.Marshal(tt.want)
				t.Errorf("MergeRouteStructures() = %v, want %v", string(stra), string(strb))
			}
		})
	}
}

func TestMergeApexAlternativeNames(t *testing.T) {
	tests := []struct {
		name  string
		input []RouteV2
		want  []RouteV2
	}{
		{
			name: "apex and www with identical config are merged",
			input: []RouteV2{
				{
					Domain:         "example.com",
					LagoonService:  "node",
					IngressName:    "example.com",
					TLSAcme:        helpers.BoolPtr(true),
					Insecure:       helpers.StrPtr("Redirect"),
					IngressClass:   "nginx",
					MonitoringPath: "/",
					Annotations:    map[string]string{},
				},
				{
					Domain:         "www.example.com",
					LagoonService:  "node",
					IngressName:    "www.example.com",
					TLSAcme:        helpers.BoolPtr(true),
					Insecure:       helpers.StrPtr("Redirect"),
					IngressClass:   "nginx",
					MonitoringPath: "/",
					Annotations:    map[string]string{},
				},
			},
			want: []RouteV2{
				{
					Domain:           "example.com",
					LagoonService:    "node",
					IngressName:      "example.com",
					TLSAcme:          helpers.BoolPtr(true),
					Insecure:         helpers.StrPtr("Redirect"),
					IngressClass:     "nginx",
					MonitoringPath:   "/",
					Annotations:      map[string]string{},
					AlternativeNames: []string{"www.example.com"},
				},
			},
		},
		{
			name: "www defined before apex retains www as the primary host",
			input: []RouteV2{
				{Domain: "www.example.com", LagoonService: "node"},
				{Domain: "example.com", LagoonService: "node"},
			},
			want: []RouteV2{
				{Domain: "www.example.com", LagoonService: "node", AlternativeNames: []string{"example.com"}},
			},
		},
		{
			name: "apex defined before www retains apex as the primary host",
			input: []RouteV2{
				{Domain: "example.com", LagoonService: "node"},
				{Domain: "www.example.com", LagoonService: "node"},
			},
			want: []RouteV2{
				{Domain: "example.com", LagoonService: "node", AlternativeNames: []string{"www.example.com"}},
			},
		},
		{
			name: "order preserved among unrelated routes when one pair merges",
			input: []RouteV2{
				{Domain: "other.com", LagoonService: "node"},
				{Domain: "www.example.com", LagoonService: "node"},
				{Domain: "another.com", LagoonService: "node"},
				{Domain: "example.com", LagoonService: "node"},
			},
			want: []RouteV2{
				{Domain: "other.com", LagoonService: "node"},
				{Domain: "www.example.com", LagoonService: "node", AlternativeNames: []string{"example.com"}},
				{Domain: "another.com", LagoonService: "node"},
			},
		},
		{
			name: "different backend service is not merged",
			input: []RouteV2{
				{Domain: "example.com", LagoonService: "node"},
				{Domain: "www.example.com", LagoonService: "cli"},
			},
			want: []RouteV2{
				{Domain: "example.com", LagoonService: "node"},
				{Domain: "www.example.com", LagoonService: "cli"},
			},
		},
		{
			name: "mismatched tls-acme is not merged",
			input: []RouteV2{
				{Domain: "example.com", LagoonService: "node", TLSAcme: helpers.BoolPtr(true)},
				{Domain: "www.example.com", LagoonService: "node", TLSAcme: helpers.BoolPtr(false)},
			},
			want: []RouteV2{
				{Domain: "example.com", LagoonService: "node", TLSAcme: helpers.BoolPtr(true)},
				{Domain: "www.example.com", LagoonService: "node", TLSAcme: helpers.BoolPtr(false)},
			},
		},
		{
			name: "www route with its own alternative names is not merged",
			input: []RouteV2{
				{Domain: "example.com", LagoonService: "node"},
				{Domain: "www.example.com", LagoonService: "node", AlternativeNames: []string{"othername.com"}},
			},
			want: []RouteV2{
				{Domain: "example.com", LagoonService: "node"},
				{Domain: "www.example.com", LagoonService: "node", AlternativeNames: []string{"othername.com"}},
			},
		},
		{
			name: "wildcard or autogenerated routes are never merged",
			input: []RouteV2{
				{Domain: "example.com", LagoonService: "node", Autogenerated: true},
				{Domain: "www.example.com", LagoonService: "node"},
			},
			want: []RouteV2{
				{Domain: "example.com", LagoonService: "node", Autogenerated: true},
				{Domain: "www.example.com", LagoonService: "node"},
			},
		},
		{
			name: "no matching www route is a no-op",
			input: []RouteV2{
				{Domain: "example.com", LagoonService: "node"},
				{Domain: "other.com", LagoonService: "node"},
			},
			want: []RouteV2{
				{Domain: "example.com", LagoonService: "node"},
				{Domain: "other.com", LagoonService: "node"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeApexAlternativeNames(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				stra, _ := json.Marshal(got)
				strb, _ := json.Marshal(tt.want)
				t.Errorf("MergeApexAlternativeNames() = %v, want %v", string(stra), string(strb))
			}
		})
	}
}
