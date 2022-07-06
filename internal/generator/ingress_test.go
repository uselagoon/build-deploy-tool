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
						Service:        "nginx",
						TLSAcme:        helpers.BoolPtr(false),
						MonitoringPath: "/bypass-cache",
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
		api          lagoon.RoutesV2
		envVars      []lagoon.EnvironmentVariable
		lagoonYAML   lagoon.YAML
		lagoonValues BuildValues
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
				lagoonValues: BuildValues{
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
							Service:        "nginx",
							TLSAcme:        helpers.BoolPtr(false),
							MonitoringPath: "/bypass-cache",
						},
						{
							Domain:         "a.example.com",
							Service:        "nginx",
							TLSAcme:        helpers.BoolPtr(false),
							MonitoringPath: "/bypass-cache",
						},
					},
				},
			},
			want: lagoon.RoutesV2{
				Routes: []lagoon.RouteV2{
					{
						Domain:         "a.example.com",
						Service:        "nginx",
						TLSAcme:        helpers.BoolPtr(false),
						Annotations:    map[string]string{},
						Insecure:       helpers.StrPtr("Redirect"),
						MonitoringPath: "/bypass-cache",
					},
					{
						Domain:         "b.example.com",
						Service:        "nginx",
						TLSAcme:        helpers.BoolPtr(true),
						Annotations:    map[string]string{},
						Insecure:       helpers.StrPtr("Redirect"),
						MonitoringPath: "/",
					},
					{
						Domain:         "c.example.com",
						Service:        "nginx",
						TLSAcme:        helpers.BoolPtr(true),
						Annotations:    map[string]string{},
						Insecure:       helpers.StrPtr("Redirect"),
						MonitoringPath: "/",
					},
					{
						Domain:         "test1.example.com",
						Service:        "nginx",
						TLSAcme:        helpers.BoolPtr(false),
						MonitoringPath: "/bypass-cache",
						Insecure:       helpers.StrPtr("Redirect"),
						Annotations:    map[string]string{},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateAndMerge(tt.args.api, tt.args.envVars, tt.args.lagoonYAML, tt.args.lagoonValues)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateAndMerge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			lValues, _ := json.Marshal(got)
			wValues, _ := json.Marshal(tt.want)
			if !reflect.DeepEqual(string(lValues), string(wValues)) {
				t.Errorf("generateAndMerge() = %v, want %v", string(lValues), string(wValues))
			}
		})
	}
}

func Test_generateActiveStandbyRoutes(t *testing.T) {
	type args struct {
		active       bool
		standby      bool
		envVars      []lagoon.EnvironmentVariable
		lagoonYAML   lagoon.YAML
		lagoonValues BuildValues
	}
	tests := []struct {
		name string
		args args
		want lagoon.RoutesV2
	}{
		{
			name: "test1",
			args: args{
				active:       true,
				lagoonValues: BuildValues{},
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
						Domain:         "active.example.com",
						Service:        "nginx",
						TLSAcme:        helpers.BoolPtr(true),
						Annotations:    map[string]string{},
						Migrate:        helpers.BoolPtr(true),
						Insecure:       helpers.StrPtr("Redirect"),
						MonitoringPath: "/",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateActiveStandbyRoutes(tt.args.active, tt.args.standby, tt.args.envVars, tt.args.lagoonYAML, tt.args.lagoonValues)
			lValues, _ := json.Marshal(got)
			wValues, _ := json.Marshal(tt.want)
			if !reflect.DeepEqual(string(lValues), string(wValues)) {
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
		lagoonValues  *BuildValues
		autogenRoutes *lagoon.RoutesV2
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := generateAutogenRoutes(tt.args.envVars, tt.args.lagoonYAML, tt.args.lagoonValues, tt.args.autogenRoutes); (err != nil) != tt.wantErr {
				t.Errorf("generateAutogenRoutes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_generateRoutes(t *testing.T) {
	type args struct {
		lagoonEnvVars      []lagoon.EnvironmentVariable
		lagoonValues       BuildValues
		lYAML              lagoon.YAML
		autogenRoutes      *lagoon.RoutesV2
		mainRoutes         *lagoon.RoutesV2
		activeStanbyRoutes *lagoon.RoutesV2
		activeEnv          bool
		standbyEnv         bool
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, err := generateRoutes(tt.args.lagoonEnvVars, tt.args.lagoonValues, tt.args.lYAML, tt.args.autogenRoutes, tt.args.mainRoutes, tt.args.activeStanbyRoutes, tt.args.activeEnv, tt.args.standbyEnv, tt.args.debug)
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
