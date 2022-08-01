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
		genRoutes           *RoutesV2
		routeMap            map[string][]Route
		variables           []EnvironmentVariable
		defaultIngressClass string
		secretPrefix        string
		activeStandby       bool
	}
	tests := []struct {
		name string
		args args
		want *RoutesV2
	}{
		{
			name: "test1",
			args: args{
				genRoutes: &RoutesV2{},
				routeMap: map[string][]Route{
					"nginx": {
						{
							Name: "example.com",
						},
						{
							Name: "www.example.com",
						},
					},
				},
				secretPrefix:  "",
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
					},
				},
			},
		},
		{
			name: "test2",
			args: args{
				genRoutes: &RoutesV2{},
				routeMap: map[string][]Route{
					"nginx": {
						{
							Name: "example.com",
						},
						{
							Ingresses: map[string]Ingress{
								"www.example.com": {
									Fastly: Fastly{
										APISecretName: "annotationscom",
										Watch:         true,
										ServiceID:     "12345",
									},
								},
							},
						},
					},
				},
				secretPrefix:  "fastly-api-",
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
					},
					{
						Domain:         "www.example.com",
						LagoonService:  "nginx",
						MonitoringPath: "/",
						Insecure:       helpers.StrPtr("Redirect"),
						TLSAcme:        helpers.BoolPtr(true),
						Annotations:    map[string]string{},
						Fastly: Fastly{
							APISecretName: "fastly-api-annotationscom",
							Watch:         true,
							ServiceID:     "12345",
						},
					},
				},
			},
		},
		{
			name: "test3 - ingress class",
			args: args{
				genRoutes: &RoutesV2{},
				routeMap: map[string][]Route{
					"nginx": {
						{
							Name: "example.com",
						},
						{
							Name: "www.example.com",
						},
					},
				},
				secretPrefix:        "",
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
					},
				},
			},
		},
		{
			name: "test4 - custom ingress class on one route",
			args: args{
				genRoutes: &RoutesV2{},
				routeMap: map[string][]Route{
					"nginx": {
						{
							Name: "example.com",
						},
						{
							Ingresses: map[string]Ingress{
								"www.example.com": {
									Fastly: Fastly{
										APISecretName: "annotationscom",
										Watch:         true,
										ServiceID:     "12345",
									},
									IngressClass: "custom-ingress",
								},
							},
						},
					},
				},
				secretPrefix:        "fastly-api-",
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
							APISecretName: "fastly-api-annotationscom",
							Watch:         true,
							ServiceID:     "12345",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			GenerateRoutesV2(tt.args.genRoutes, tt.args.routeMap, tt.args.variables, tt.args.defaultIngressClass, tt.args.secretPrefix, tt.args.activeStandby)
			if !cmp.Equal(tt.args.genRoutes, tt.want) {
				stra, _ := json.Marshal(tt.args.genRoutes)
				strb, _ := json.Marshal(tt.want)
				t.Errorf("GenerateRouteStructure() = %v, want %v", string(stra), string(strb))
			}
		})
	}
}

func TestMergeRouteStructures(t *testing.T) {
	type args struct {
		genRoutes           RoutesV2
		apiRoutes           RoutesV2
		variables           []EnvironmentVariable
		defaultIngressClass string
		secretPrefix        string
	}
	tests := []struct {
		name string
		args args
		want RoutesV2
	}{
		{
			name: "test1",
			args: args{
				genRoutes: RoutesV2{
					Routes: []RouteV2{
						{
							Domain:         "example.com",
							LagoonService:  "nginx",
							MonitoringPath: "/",
							Insecure:       helpers.StrPtr("Redirect"),
							TLSAcme:        helpers.BoolPtr(true),
							Annotations:    map[string]string{},
							Fastly: Fastly{
								Watch:         true,
								ServiceID:     "12345",
								APISecretName: "annotationscom",
							},
						},
						{
							Domain:         "www.example.com",
							LagoonService:  "nginx",
							MonitoringPath: "/",
							Insecure:       helpers.StrPtr("Redirect"),
							TLSAcme:        helpers.BoolPtr(true),
							Annotations:    map[string]string{},
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
						},
						{
							Domain:         "another.example.com",
							LagoonService:  "nginx",
							MonitoringPath: "/",
							Insecure:       helpers.StrPtr("Redirect"),
							TLSAcme:        helpers.BoolPtr(true),
							Annotations:    map[string]string{},
						},
					},
				},
				secretPrefix: "fastly-api-",
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
							Watch:         true,
							ServiceID:     "12345",
							APISecretName: "fastly-api-annotationscom",
						},
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
					},
					{
						Domain:         "another.example.com",
						LagoonService:  "nginx",
						MonitoringPath: "/",
						Insecure:       helpers.StrPtr("Redirect"),
						TLSAcme:        helpers.BoolPtr(true),
						Annotations:    map[string]string{},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MergeRoutesV2(tt.args.genRoutes, tt.args.apiRoutes, tt.args.variables, tt.args.defaultIngressClass, tt.args.secretPrefix); !reflect.DeepEqual(got, tt.want) {
				stra, _ := json.Marshal(got)
				strb, _ := json.Marshal(tt.want)
				t.Errorf("MergeRouteStructures() = %v, want %v", string(stra), string(strb))
			}
		})
	}
}
