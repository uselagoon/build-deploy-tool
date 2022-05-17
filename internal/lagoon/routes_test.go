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
		genRoutes     *RoutesV2
		routeMap      map[string][]Route
		variables     []EnvironmentVariable
		secretPrefix  string
		activeStandby bool
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
						Service:        "nginx",
						MonitoringPath: "/",
						Insecure:       helpers.StrPtr("Redirect"),
						TLSAcme:        helpers.BoolPtr(true),
						Annotations:    map[string]string{},
						Fastly: Fastly{
							Watch: false,
						},
						AlternativeNames: []string{},
					},
					{
						Domain:         "www.example.com",
						Service:        "nginx",
						MonitoringPath: "/",
						Insecure:       helpers.StrPtr("Redirect"),
						TLSAcme:        helpers.BoolPtr(true),
						Annotations:    map[string]string{},
						Fastly: Fastly{
							Watch: false,
						},
						AlternativeNames: []string{},
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
						Service:        "nginx",
						MonitoringPath: "/",
						Insecure:       helpers.StrPtr("Redirect"),
						TLSAcme:        helpers.BoolPtr(true),
						Annotations:    map[string]string{},
						Fastly: Fastly{
							Watch: false,
						},
						AlternativeNames: []string{},
					},
					{
						Domain:         "www.example.com",
						Service:        "nginx",
						MonitoringPath: "/",
						Insecure:       helpers.StrPtr("Redirect"),
						TLSAcme:        helpers.BoolPtr(true),
						Annotations:    map[string]string{},
						Fastly: Fastly{
							APISecretName: "fastly-api-annotationscom",
							Watch:         true,
							ServiceID:     "12345",
						},
						AlternativeNames: []string{},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			GenerateRoutesV2(tt.args.genRoutes, tt.args.routeMap, tt.args.variables, tt.args.secretPrefix, tt.args.activeStandby)
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
		genRoutes    RoutesV2
		apiRoutes    RoutesV2
		variables    []EnvironmentVariable
		secretPrefix string
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
							Service:        "nginx",
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
							Service:        "nginx",
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
							Service:        "nginx",
							MonitoringPath: "/",
							Insecure:       helpers.StrPtr("Redirect"),
							TLSAcme:        helpers.BoolPtr(true),
							Annotations: map[string]string{
								"nginx": "nginx",
							},
						},
						{
							Domain:         "another.example.com",
							Service:        "nginx",
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
						Service:        "nginx",
						MonitoringPath: "/",
						Insecure:       helpers.StrPtr("Redirect"),
						TLSAcme:        helpers.BoolPtr(true),
						Annotations:    map[string]string{},
						Fastly: Fastly{
							Watch:         true,
							ServiceID:     "12345",
							APISecretName: "fastly-api-annotationscom",
						},
						AlternativeNames: []string{},
					},
					{
						Domain:         "www.example.com",
						Service:        "nginx",
						MonitoringPath: "/",
						Insecure:       helpers.StrPtr("Redirect"),
						TLSAcme:        helpers.BoolPtr(true),
						Annotations: map[string]string{
							"nginx": "nginx",
						},
						AlternativeNames: []string{},
					},
					{
						Domain:           "another.example.com",
						Service:          "nginx",
						MonitoringPath:   "/",
						Insecure:         helpers.StrPtr("Redirect"),
						TLSAcme:          helpers.BoolPtr(true),
						Annotations:      map[string]string{},
						AlternativeNames: []string{},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MergeRoutesV2(tt.args.genRoutes, tt.args.apiRoutes, tt.args.variables, tt.args.secretPrefix); !reflect.DeepEqual(got, tt.want) {
				stra, _ := json.Marshal(got)
				strb, _ := json.Marshal(tt.want)
				t.Errorf("MergeRouteStructures() = %v, want %v", string(stra), string(strb))
			}
		})
	}
}
