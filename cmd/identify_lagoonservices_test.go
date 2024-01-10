package cmd

import (
	"os"
	"reflect"
	"testing"

	"github.com/uselagoon/build-deploy-tool/internal/dbaasclient"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	"github.com/uselagoon/build-deploy-tool/internal/testdata"

	// changes the testing to source from root so paths to test resources must be defined from repo root
	_ "github.com/uselagoon/build-deploy-tool/internal/testing"
)

func TestIdentifyLagoonServices(t *testing.T) {
	tests := []struct {
		name        string
		description string
		args        testdata.TestData
		want        []identifyServices
	}{
		{
			name: "test1 basic deployment",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/basic/lagoon.yml",
					ImageReferences: map[string]string{
						"node": "harbor.example/example-project/main/node:latest",
					},
				}, true),
			want: []identifyServices{
				{
					Name: "node",
					Type: "basic",
					Containers: []containers{
						{
							Name: "basic",
							Ports: []ports{
								{Port: 1234},
								{Port: 8191},
								{Port: 9001},
							},
						},
					},
				},
			},
		},
		{
			name: "test2a nginx-php deployment",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/complex/lagoon.varnish.yml",
					ImageReferences: map[string]string{
						"nginx":   "harbor.example/example-project/main/nginx:latest",
						"php":     "harbor.example/example-project/main/php:latest",
						"cli":     "harbor.example/example-project/main/cli:latest",
						"redis":   "harbor.example/example-project/main/redis:latest",
						"varnish": "harbor.example/example-project/main/varnish:latest",
					},
				}, true),
			want: []identifyServices{
				{
					Name: "cli",
					Type: "cli-persistent",
					Containers: []containers{
						{
							Name:  "cli",
							Ports: []ports{},
						},
					},
				},
				{
					Name: "redis",
					Type: "redis",
					Containers: []containers{
						{
							Name: "redis",
							Ports: []ports{
								{Port: 6379},
							},
						},
					},
				},
				{
					Name: "varnish",
					Type: "varnish",
					Containers: []containers{
						{
							Name: "varnish",
							Ports: []ports{
								{Port: 8080},
								{Port: 6082},
							},
						},
					},
				},
				{
					Name: "nginx-php",
					Type: "nginx-php-persistent",
					Containers: []containers{
						{
							Name: "nginx",
							Ports: []ports{
								{Port: 8080},
							},
						},
						{
							Name: "php",
							Ports: []ports{
								{Port: 9000},
							},
						},
					},
				},
			},
		},
		{
			name: "test2b nginx-php deployment - rootless",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/complex/lagoon.varnish.yml",
					ImageReferences: map[string]string{
						"nginx":   "harbor.example/example-project/main/nginx:latest",
						"php":     "harbor.example/example-project/main/php:latest",
						"cli":     "harbor.example/example-project/main/cli:latest",
						"redis":   "harbor.example/example-project/main/redis:latest",
						"varnish": "harbor.example/example-project/main/varnish:latest",
					},
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_ROOTLESS_WORKLOAD",
							Value: "enabled",
							Scope: "build",
						},
					},
				}, true),
			want: []identifyServices{
				{
					Name: "cli",
					Type: "cli-persistent",
					Containers: []containers{
						{
							Name:  "cli",
							Ports: []ports{},
						},
					},
				},
				{
					Name: "redis",
					Type: "redis",
					Containers: []containers{
						{
							Name: "redis",
							Ports: []ports{
								{Port: 6379},
							},
						},
					},
				},
				{
					Name: "varnish",
					Type: "varnish",
					Containers: []containers{
						{
							Name: "varnish",
							Ports: []ports{
								{Port: 8080},
								{Port: 6082},
							},
						},
					},
				},
				{
					Name: "nginx-php",
					Type: "nginx-php-persistent",
					Containers: []containers{
						{
							Name: "nginx",
							Ports: []ports{
								{Port: 8080},
							},
						},
						{
							Name: "php",
							Ports: []ports{
								{Port: 9000},
							},
						},
					},
				},
			},
		},
		{
			name:        "test3 - funky pvcs",
			description: "only create pvcs of the requested persistent-name in the docker-compose file",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/basic/lagoon.thunderhub.yml",
					ImageReferences: map[string]string{
						"lnd":        "harbor.example/example-project/main/lnd:latest",
						"thunderhub": "harbor.example/example-project/main/thunderhub:latest",
						"tor":        "harbor.example/example-project/main/tor:latest",
					},
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_ROOTLESS_WORKLOAD",
							Value: "enabled",
							Scope: "build",
						},
					},
				}, true),
			want: []identifyServices{
				{
					Name: "lnd",
					Type: "basic-persistent",
					Containers: []containers{
						{
							Name: "basic",
							Ports: []ports{
								{Port: 8080},
								{Port: 10009},
							},
						},
					},
				},
				{
					Name: "thunderhub",
					Type: "basic-persistent",
					Containers: []containers{
						{
							Name: "basic",
							Ports: []ports{
								{Port: 3000},
							},
						},
					},
				},
				{
					Name: "tor",
					Type: "basic",
					Containers: []containers{
						{
							Name: "basic",
							Ports: []ports{
								{Port: 9050},
								{Port: 9051},
							},
						},
					},
				},
			},
		},
		{
			name:        "test4 - basic-persistent with worker-persistent",
			description: "create a basic-persistent that gets a pvc and mount that volume on a worker-persistent type",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/basic/lagoon.thunderhub-2.yml",
					ImageReferences: map[string]string{
						"lnd": "harbor.example/example-project/main/lnd:latest",
						"tor": "harbor.example/example-project/main/tor:latest",
					},
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_ROOTLESS_WORKLOAD",
							Value: "enabled",
							Scope: "build",
						},
					},
				}, true),
			want: []identifyServices{
				{
					Name: "lnd",
					Type: "basic-persistent",
					Containers: []containers{
						{Name: "basic",
							Ports: []ports{
								{Port: 8080},
								{Port: 10009},
							}},
					},
				},
				{
					Name: "tor",
					Type: "worker-persistent",
					Containers: []containers{
						{Name: "worker",
							Ports: []ports{}},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// set the environment variables from args
			savedTemplates := "test-resources/output"
			generator, err := testdata.SetupEnvironment(*rootCmd, savedTemplates, tt.args)
			if err != nil {
				t.Errorf("%v", err)
			}

			err = os.MkdirAll(savedTemplates, 0755)
			if err != nil {
				t.Errorf("couldn't create directory %v: %v", savedTemplates, err)
			}

			defer os.RemoveAll(savedTemplates)

			ts := dbaasclient.TestDBaaSHTTPServer()
			defer ts.Close()
			err = os.Setenv("DBAAS_OPERATOR_HTTP", ts.URL)
			if err != nil {
				t.Errorf("%v", err)
			}

			out, err := LagoonServiceTemplateIdentification(generator)
			if err != nil {
				t.Errorf("%v", err)
			}
			if !reflect.DeepEqual(out, tt.want) {
				t.Errorf("returned output %v doesn't match want %v", out, tt.want)
			}
			t.Cleanup(func() {
				helpers.UnsetEnvVars(nil)
			})
		})
	}
}
