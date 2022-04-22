package generator

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGenerateRouteStructure(t *testing.T) {
	type args struct {
		genRoutes     *RoutesV2
		routeMap      map[string][]LagoonRoute
		activeStandby bool
	}
	tests := []struct {
		name string
		args args
		want *RoutesV2
	}{
		{
			name: "generate routes",
			args: args{
				genRoutes: &RoutesV2{},
				routeMap: map[string][]LagoonRoute{
					"nginx": []LagoonRoute{
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
						Service:        "nginx",
						MonitoringPath: "/",
						Insecure:       strPtr("Redirect"),
						HSTS:           strPtr("null"),
						TLSAcme:        strPtr("true"),
						Annotations:    map[string]string{},
						Fastly: Fastly{
							Watch: false,
						},
					},
					{
						Domain:         "www.example.com",
						Service:        "nginx",
						MonitoringPath: "/",
						Insecure:       strPtr("Redirect"),
						HSTS:           strPtr("null"),
						TLSAcme:        strPtr("true"),
						Annotations:    map[string]string{},
						Fastly: Fastly{
							Watch: false,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			GenerateRouteStructure(tt.args.genRoutes, tt.args.routeMap, tt.args.activeStandby)
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
		genRoutes RoutesV2
		apiRoutes RoutesV2
	}
	tests := []struct {
		name string
		args args
		want RoutesV2
	}{
		{
			name: "generate routes",
			args: args{
				genRoutes: RoutesV2{
					Routes: []RouteV2{
						{
							Domain:         "example.com",
							Service:        "nginx",
							MonitoringPath: "/",
							Insecure:       strPtr("Redirect"),
							HSTS:           strPtr("null"),
							TLSAcme:        strPtr("true"),
							Annotations:    map[string]string{},
							Fastly: Fastly{
								Watch: true,
							},
						},
						{
							Domain:         "www.example.com",
							Service:        "nginx",
							MonitoringPath: "/",
							Insecure:       strPtr("Redirect"),
							HSTS:           strPtr("null"),
							TLSAcme:        strPtr("true"),
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
							Insecure:       strPtr("Redirect"),
							HSTS:           strPtr("null"),
							TLSAcme:        strPtr("true"),
							Annotations: map[string]string{
								"nginx": "nginx",
							},
						},
						{
							Domain:         "another.example.com",
							Service:        "nginx",
							MonitoringPath: "/",
							Insecure:       strPtr("Redirect"),
							HSTS:           strPtr("null"),
							TLSAcme:        strPtr("true"),
							Annotations:    map[string]string{},
						},
					},
				},
			},
			want: RoutesV2{
				Routes: []RouteV2{
					{
						Domain:         "example.com",
						Service:        "nginx",
						MonitoringPath: "/",
						Insecure:       strPtr("Redirect"),
						HSTS:           strPtr("null"),
						TLSAcme:        strPtr("true"),
						Annotations:    map[string]string{},
						Fastly: Fastly{
							Watch: true,
						},
					},
					{
						Domain:         "www.example.com",
						Service:        "nginx",
						MonitoringPath: "/",
						Insecure:       strPtr("Redirect"),
						HSTS:           strPtr("null"),
						TLSAcme:        strPtr("true"),
						Annotations: map[string]string{
							"nginx": "nginx",
						},
					},
					{
						Domain:         "another.example.com",
						Service:        "nginx",
						MonitoringPath: "/",
						Insecure:       strPtr("Redirect"),
						HSTS:           strPtr("null"),
						TLSAcme:        strPtr("true"),
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
