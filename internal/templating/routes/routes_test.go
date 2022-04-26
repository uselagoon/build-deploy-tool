package routes

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/uselagoon/lagoon-routegen/internal/helpers"
	"github.com/uselagoon/lagoon-routegen/internal/lagoon"
)

func TestGenerateRouteStructure(t *testing.T) {
	type args struct {
		genRoutes     *lagoon.RoutesV2
		routeMap      map[string][]lagoon.Route
		variables     []lagoon.EnvironmentVariable
		activeStandby bool
	}
	tests := []struct {
		name string
		args args
		want *lagoon.RoutesV2
	}{
		{
			name: "generate routes",
			args: args{
				genRoutes: &lagoon.RoutesV2{},
				routeMap: map[string][]lagoon.Route{
					"nginx": []lagoon.Route{
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
			want: &lagoon.RoutesV2{
				Routes: []lagoon.RouteV2{
					{
						Domain:         "example.com",
						Service:        "nginx",
						MonitoringPath: "/",
						Insecure:       helpers.StrPtr("Redirect"),
						HSTS:           helpers.StrPtr("null"),
						TLSAcme:        helpers.BoolPtr(true),
						Annotations:    map[string]string{},
						Fastly: lagoon.Fastly{
							Watch: false,
						},
					},
					{
						Domain:         "www.example.com",
						Service:        "nginx",
						MonitoringPath: "/",
						Insecure:       helpers.StrPtr("Redirect"),
						HSTS:           helpers.StrPtr("null"),
						TLSAcme:        helpers.BoolPtr(true),
						Annotations:    map[string]string{},
						Fastly: lagoon.Fastly{
							Watch: false,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			GenerateRouteStructure(tt.args.genRoutes, tt.args.routeMap, tt.args.variables, tt.args.activeStandby)
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
		genRoutes lagoon.RoutesV2
		apiRoutes lagoon.RoutesV2
	}
	tests := []struct {
		name string
		args args
		want lagoon.RoutesV2
	}{
		{
			name: "generate routes",
			args: args{
				genRoutes: lagoon.RoutesV2{
					Routes: []lagoon.RouteV2{
						{
							Domain:         "example.com",
							Service:        "nginx",
							MonitoringPath: "/",
							Insecure:       helpers.StrPtr("Redirect"),
							HSTS:           helpers.StrPtr("null"),
							TLSAcme:        helpers.BoolPtr(true),
							Annotations:    map[string]string{},
							Fastly: lagoon.Fastly{
								Watch: true,
							},
						},
						{
							Domain:         "www.example.com",
							Service:        "nginx",
							MonitoringPath: "/",
							Insecure:       helpers.StrPtr("Redirect"),
							HSTS:           helpers.StrPtr("null"),
							TLSAcme:        helpers.BoolPtr(true),
							Annotations:    map[string]string{},
						},
					},
				},
				apiRoutes: lagoon.RoutesV2{
					Routes: []lagoon.RouteV2{
						{
							Domain:         "www.example.com",
							Service:        "nginx",
							MonitoringPath: "/",
							Insecure:       helpers.StrPtr("Redirect"),
							HSTS:           helpers.StrPtr("null"),
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
							HSTS:           helpers.StrPtr("null"),
							TLSAcme:        helpers.BoolPtr(true),
							Annotations:    map[string]string{},
						},
					},
				},
			},
			want: lagoon.RoutesV2{
				Routes: []lagoon.RouteV2{
					{
						Domain:         "example.com",
						Service:        "nginx",
						MonitoringPath: "/",
						Insecure:       helpers.StrPtr("Redirect"),
						HSTS:           helpers.StrPtr("null"),
						TLSAcme:        helpers.BoolPtr(true),
						Annotations:    map[string]string{},
						Fastly: lagoon.Fastly{
							Watch: true,
						},
					},
					{
						Domain:         "www.example.com",
						Service:        "nginx",
						MonitoringPath: "/",
						Insecure:       helpers.StrPtr("Redirect"),
						HSTS:           helpers.StrPtr("null"),
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
						HSTS:           helpers.StrPtr("null"),
						TLSAcme:        helpers.BoolPtr(true),
						Annotations:    map[string]string{},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MergeRouteStructures(tt.args.genRoutes, tt.args.apiRoutes); !reflect.DeepEqual(got, tt.want) {
				stra, _ := json.Marshal(got)
				strb, _ := json.Marshal(tt.want)
				t.Errorf("MergeRouteStructures() = %v, want %v", string(stra), string(strb))
			}
		})
	}
}
