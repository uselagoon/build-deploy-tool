package cmd

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/uselagoon/build-deploy-tool/internal/dbaasclient"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	"github.com/uselagoon/build-deploy-tool/internal/testdata"
	"github.com/uselagoon/machinery/api/schema"

	// changes the testing to source from root so paths to test resources must be defined from repo root
	_ "github.com/uselagoon/build-deploy-tool/internal/testing"
)

func TestIdentifyLagoonServices(t *testing.T) {
	tests := []struct {
		name        string
		description string
		args        testdata.TestData
		want        []schema.EnvironmentService
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
						"node": "harbor.example/example-project/main/node@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
				}, true),
			want: []schema.EnvironmentService{
				{
					Name: "node",
					Type: "basic",
					Containers: []schema.ServiceContainer{
						{
							Name: "basic",
							Ports: []schema.ServiceContainerPort{
								{Port: 1234, Protocol: "TCP"},
								{Port: 8191, Protocol: "TCP"},
								{Port: 9001, Protocol: "UDP"},
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
						"nginx":   "harbor.example/example-project/main/nginx@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"php":     "harbor.example/example-project/main/php@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"cli":     "harbor.example/example-project/main/cli@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"redis":   "harbor.example/example-project/main/redis@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"varnish": "harbor.example/example-project/main/varnish@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
				}, true),
			want: []schema.EnvironmentService{
				{
					Name: "cli",
					Type: "cli-persistent",
					Containers: []schema.ServiceContainer{
						{
							Name: "cli",
							Volumes: []schema.ServiceContainerVolume{
								{
									Name: "nginx-php",
									Path: "/app/docroot/sites/default/files/",
								},
							},
						},
					},
				},
				{
					Name: "redis",
					Type: "redis",
					Containers: []schema.ServiceContainer{
						{
							Name: "redis",
							Ports: []schema.ServiceContainerPort{
								{Port: 6379, Protocol: "TCP"},
							},
						},
					},
				},
				{
					Name: "varnish",
					Type: "varnish",
					Containers: []schema.ServiceContainer{
						{
							Name: "varnish",
							Ports: []schema.ServiceContainerPort{
								{Port: 8080, Protocol: "TCP"},
								{Port: 6082, Protocol: "TCP"},
							},
						},
					},
				},
				{
					Name: "nginx-php",
					Type: "nginx-php-persistent",
					Containers: []schema.ServiceContainer{
						{
							Name: "nginx",
							Ports: []schema.ServiceContainerPort{
								{Port: 8080, Protocol: "TCP"},
							},
							Volumes: []schema.ServiceContainerVolume{
								{
									Name: "nginx-php",
									Path: "/app/docroot/sites/default/files/",
								},
							},
						},
						{
							Name: "php",
							Ports: []schema.ServiceContainerPort{
								{Port: 9000, Protocol: "TCP"},
							},
							Volumes: []schema.ServiceContainerVolume{
								{
									Name: "nginx-php",
									Path: "/app/docroot/sites/default/files/",
								},
							},
						},
					},
				},
				{
					Name: "mariadb",
					Type: "mariadb-dbaas",
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
						"nginx":   "harbor.example/example-project/main/nginx@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"php":     "harbor.example/example-project/main/php@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"cli":     "harbor.example/example-project/main/cli@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"redis":   "harbor.example/example-project/main/redis@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"varnish": "harbor.example/example-project/main/varnish@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_ROOTLESS_WORKLOAD",
							Value: "enabled",
							Scope: "build",
						},
					},
				}, true),
			want: []schema.EnvironmentService{
				{
					Name: "cli",
					Type: "cli-persistent",
					Containers: []schema.ServiceContainer{
						{
							Name: "cli",
							Volumes: []schema.ServiceContainerVolume{
								{
									Name: "nginx-php",
									Path: "/app/docroot/sites/default/files/",
								},
							},
						},
					},
				},
				{
					Name: "redis",
					Type: "redis",
					Containers: []schema.ServiceContainer{
						{
							Name: "redis",
							Ports: []schema.ServiceContainerPort{
								{Port: 6379, Protocol: "TCP"},
							},
						},
					},
				},
				{
					Name: "varnish",
					Type: "varnish",
					Containers: []schema.ServiceContainer{
						{
							Name: "varnish",
							Ports: []schema.ServiceContainerPort{
								{Port: 8080, Protocol: "TCP"},
								{Port: 6082, Protocol: "TCP"},
							},
						},
					},
				},
				{
					Name: "nginx-php",
					Type: "nginx-php-persistent",
					Containers: []schema.ServiceContainer{
						{
							Name: "nginx",
							Ports: []schema.ServiceContainerPort{
								{Port: 8080, Protocol: "TCP"},
							},
							Volumes: []schema.ServiceContainerVolume{
								{
									Name: "nginx-php",
									Path: "/app/docroot/sites/default/files/",
								},
							},
						},
						{
							Name: "php",
							Ports: []schema.ServiceContainerPort{
								{Port: 9000, Protocol: "TCP"},
							},
							Volumes: []schema.ServiceContainerVolume{
								{
									Name: "nginx-php",
									Path: "/app/docroot/sites/default/files/",
								},
							},
						},
					},
				},
				{
					Name: "mariadb",
					Type: "mariadb-dbaas",
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
						"lnd":        "harbor.example/example-project/main/lnd@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"thunderhub": "harbor.example/example-project/main/thunderhub@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"tor":        "harbor.example/example-project/main/tor@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_ROOTLESS_WORKLOAD",
							Value: "enabled",
							Scope: "build",
						},
					},
				}, true),
			want: []schema.EnvironmentService{
				{
					Name: "lnd",
					Type: "basic-persistent",
					Containers: []schema.ServiceContainer{
						{
							Name: "basic",
							Ports: []schema.ServiceContainerPort{
								{Port: 8080, Protocol: "TCP"},
								{Port: 10009, Protocol: "TCP"},
							},
							Volumes: []schema.ServiceContainerVolume{
								{
									Name: "lnd",
									Path: "/app/storage",
								},
							},
						},
					},
				},
				{
					Name: "thunderhub",
					Type: "basic-persistent",
					Containers: []schema.ServiceContainer{
						{
							Name: "basic",
							Ports: []schema.ServiceContainerPort{
								{Port: 3000, Protocol: "TCP"},
							},
							Volumes: []schema.ServiceContainerVolume{
								{
									Name: "lnd",
									Path: "/data",
								},
							},
						},
					},
				},
				{
					Name: "tor",
					Type: "basic",
					Containers: []schema.ServiceContainer{
						{
							Name: "basic",
							Ports: []schema.ServiceContainerPort{
								{Port: 9050, Protocol: "TCP"},
								{Port: 9051, Protocol: "TCP"},
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
						"lnd": "harbor.example/example-project/main/lnd@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"tor": "harbor.example/example-project/main/tor@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_ROOTLESS_WORKLOAD",
							Value: "enabled",
							Scope: "build",
						},
					},
				}, true),
			want: []schema.EnvironmentService{
				{
					Name: "lnd",
					Type: "basic-persistent",
					Containers: []schema.ServiceContainer{
						{
							Name: "basic",
							Ports: []schema.ServiceContainerPort{
								{Port: 8080, Protocol: "TCP"},
								{Port: 10009, Protocol: "TCP"},
							},
							Volumes: []schema.ServiceContainerVolume{
								{
									Name: "lnd",
									Path: "/app/storage",
								},
							},
						},
					},
				},
				{
					Name: "tor",
					Type: "worker-persistent",
					Containers: []schema.ServiceContainer{
						{
							Name: "worker",
							Volumes: []schema.ServiceContainerVolume{
								{
									Name: "lnd",
									Path: "/data",
								},
							},
						},
					},
				},
			},
		},
		{
			name:        "basic-custom-volumes",
			description: "create a basic with custom volumes",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					BuildType:       "branch",
					LagoonYAML:      "internal/testdata/basic/lagoon.multiple-volumes-2.yml",
					ImageReferences: map[string]string{
						"node": "harbor.example/example-project/main/node@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
				}, true),
			want: []schema.EnvironmentService{
				{
					Name: "node",
					Type: "basic",
					Containers: []schema.ServiceContainer{
						{Name: "basic",
							Ports: []schema.ServiceContainerPort{
								{Port: 3000, Protocol: "TCP"},
							},
							Volumes: []schema.ServiceContainerVolume{
								{
									Name: "custom-node",
									Path: "/data",
								},
								{
									Name: "custom-config",
									Path: "/config",
								},
								{
									Name: "custom-files",
									Path: "/app/files/",
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.UnsetEnvVars(nil) //unset variables before running tests
			// set the environment variables from args
			savedTemplates := "testoutput"
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
			oJ, _ := json.MarshalIndent(out, "", "  ")
			wJ, _ := json.MarshalIndent(tt.want, "", "  ")
			if string(oJ) != string(wJ) {
				t.Errorf("LagoonServiceTemplateIdentification() = \n%v", diff.LineDiff(string(oJ), string(wJ)))
			}
			t.Cleanup(func() {
				helpers.UnsetEnvVars(nil)
				helpers.UnsetEnvVars(tt.args.BuildPodVariables)
			})
		})
	}
}
