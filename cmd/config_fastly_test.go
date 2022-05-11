package cmd

import (
	"os"
	"reflect"
	"testing"

	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
)

func TestGenerateFastlyConfig(t *testing.T) {
	type args struct {
		projectVars  string
		envVars      string
		cacheNoCache string
		serviceID    string
		domain       string
		secretPrefix string
	}
	tests := []struct {
		name string
		args args
		want lagoon.Fastly
	}{
		{
			name: "test1 check LAGOON_FASTLY_SERVICE_ID no secret",
			args: args{
				projectVars:  `[{"name":"LAGOON_FASTLY_SERVICE_ID","value":"service-id:true","scope":"global"}]`,
				envVars:      `[]`,
				cacheNoCache: "",
				serviceID:    "",
				domain:       "example.com",
				secretPrefix: "faslty-api-",
			},
			want: lagoon.Fastly{
				ServiceID:     "service-id",
				APISecretName: "",
				Watch:         true,
			},
		},
		{
			name: "test2 check LAGOON_FASTLY_SERVICE_IDS no secret",
			args: args{
				projectVars:  `[{"name":"LAGOON_FASTLY_SERVICE_IDS","value":"example.com:service-id:true","scope":"global"}]`,
				envVars:      `[]`,
				cacheNoCache: "",
				serviceID:    "",
				domain:       "example.com",
				secretPrefix: "faslty-api-",
			},
			want: lagoon.Fastly{
				ServiceID:     "service-id",
				APISecretName: "",
				Watch:         true,
			},
		},
		{
			name: "test3 check LAGOON_FASTLY_SERVICE_ID with secret",
			args: args{
				projectVars:  `[{"name":"LAGOON_FASTLY_SERVICE_ID","value":"service-id:true:secret","scope":"global"}]`,
				envVars:      `[]`,
				cacheNoCache: "",
				serviceID:    "",
				domain:       "example.com",
				secretPrefix: "faslty-api-",
			},
			want: lagoon.Fastly{
				ServiceID:     "service-id",
				APISecretName: "faslty-api-secret",
				Watch:         true,
			},
		},
		{
			name: "test4 check LAGOON_FASTLY_SERVICE_IDS with secret",
			args: args{
				projectVars:  `[{"name":"LAGOON_FASTLY_SERVICE_IDS","value":"example.com:service-id:true:secret","scope":"global"}]`,
				envVars:      `[]`,
				cacheNoCache: "",
				serviceID:    "",
				domain:       "example.com",
				secretPrefix: "faslty-api-",
			},
			want: lagoon.Fastly{
				ServiceID:     "service-id",
				APISecretName: "faslty-api-secret",
				Watch:         true,
			},
		},
		{
			name: "test5 check no LAGOON_FASTLY_SERVICE_ID with service id found from ROUTE_FASTLY_SERVICE_ID",
			args: args{
				projectVars:  `[]`,
				envVars:      `[]`,
				cacheNoCache: "",
				serviceID:    "dedicated-service-id",
				domain:       "example.com",
				secretPrefix: "faslty-api-",
			},
			want: lagoon.Fastly{
				ServiceID: "dedicated-service-id",
				Watch:     false,
			},
		},
		{
			name: "test6 check LAGOON_FASTLY_SERVICE_ID with service id found from ROUTE_FASTLY_SERVICE_ID (should use one from LAGOON_FASTLY_SERVICE_ID)",
			args: args{
				projectVars:  `[{"name":"LAGOON_FASTLY_SERVICE_ID","value":"service-id:true","scope":"global"}]`,
				envVars:      `[]`,
				cacheNoCache: "",
				serviceID:    "dedicated-service-id",
				domain:       "example.com",
				secretPrefix: "faslty-api-",
			},
			want: lagoon.Fastly{
				ServiceID: "service-id",
				Watch:     true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// set the environment variables from args
			err := os.Setenv("LAGOON_FASTLY_NOCACHE_SERVICE_ID", tt.args.cacheNoCache)
			if err != nil {
				t.Errorf("%v", err)
			}
			err = os.Setenv("ROUTE_FASTLY_SERVICE_ID", tt.args.serviceID)
			if err != nil {
				t.Errorf("%v", err)
			}
			err = os.Setenv("FASTLY_API_SECRET_PREFIX", tt.args.secretPrefix)
			if err != nil {
				t.Errorf("%v", err)
			}
			err = os.Setenv("LAGOON_PROJECT_VARIABLES", tt.args.projectVars)
			if err != nil {
				t.Errorf("%v", err)
			}
			err = os.Setenv("LAGOON_ENVIRONMENT_VARIABLES", tt.args.envVars)
			if err != nil {
				t.Errorf("%v", err)
			}

			// generate the fastly configuration from the provided flags/variables
			got, err := FastlyConfigGeneration(false, tt.args.domain)
			if err != nil {
				t.Errorf("%v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fastlyConfigGeneration() = %v, want %v", got, tt.want)
			}
		})
	}
}
