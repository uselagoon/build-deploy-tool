package routes

import (
	"os"
	"reflect"
	"testing"

	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
)

func TestGenerateKubeTemplate(t *testing.T) {
	type args struct {
		route                  lagoon.RouteV2
		values                 lagoon.BuildValues
		monitoringContact      string
		monitoringStatusPageID string
		monitoringEnabled      bool
		activeStandby          bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "active-standby1",
			args: args{
				route: lagoon.RouteV2{
					Domain:         "extra-long-name.a-really-long-name-that-should-truncate.www.example.com",
					Service:        "nginx",
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
				values: lagoon.BuildValues{
					Project:         "example-project",
					Environment:     "environment-with-really-really-reall-3fdb",
					EnvironmentType: "production",
					Namespace:       "myexample-project-environment-with-really-really-reall-3fdb",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "lagoon.local",
					Branch:          "environment-with-really-really-reall-3fdb",
				},
				monitoringContact:      "abcdefg",
				monitoringStatusPageID: "12345",
				monitoringEnabled:      true,
				activeStandby:          true,
			},
			want: "test-resources/result-active-standby1.yaml",
		},
		{
			name: "custom-ingress1",
			args: args{
				route: lagoon.RouteV2{
					Domain:         "extra-long-name.a-really-long-name-that-should-truncate.www.example.com",
					Service:        "nginx",
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
				values: lagoon.BuildValues{
					Project:         "example-project",
					Environment:     "environment-with-really-really-reall-3fdb",
					EnvironmentType: "production",
					Namespace:       "myexample-project-environment-with-really-really-reall-3fdb",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "lagoon.local",
					Branch:          "environment-with-really-really-reall-3fdb",
				},
				monitoringContact:      "abcdefg",
				monitoringStatusPageID: "12345",
				monitoringEnabled:      true,
				activeStandby:          false,
			},
			want: "test-resources/result-custom-ingress1.yaml",
		},
		{
			name: "custom-ingress2",
			args: args{
				route: lagoon.RouteV2{
					Domain:         "extra-long-name.a-really-long-name-that-should-truncate.www.example.com",
					Service:        "nginx",
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
				values: lagoon.BuildValues{
					Project:         "example-project",
					Environment:     "environment-with-really-really-reall-3fdb",
					EnvironmentType: "development",
					Namespace:       "myexample-project-environment-with-really-really-reall-3fdb",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "lagoon.local",
					Branch:          "environment-with-really-really-reall-3fdb",
				},
				monitoringContact:      "abcdefg",
				monitoringStatusPageID: "12345",
				monitoringEnabled:      true,
				activeStandby:          false,
			},
			want: "test-resources/result-custom-ingress2.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateIngressTemplate(tt.args.route, tt.args.values, tt.args.monitoringContact, tt.args.monitoringStatusPageID, tt.args.monitoringEnabled)
			r1, err := os.ReadFile(tt.want)
			if err != nil {
				t.Errorf("couldn't read file %v: %v", tt.want, err)
			}
			if !reflect.DeepEqual(string(got), string(r1)) {
				t.Errorf("GenerateIngressTemplate() = %v, want %v", string(got), string(r1))
			}
		})
	}
}
