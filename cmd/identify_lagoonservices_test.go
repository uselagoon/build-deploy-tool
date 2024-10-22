package cmd

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/andreyvit/diff"
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
		want        *identifyServices
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
			want: &identifyServices{
				Deployments: []string{
					"node",
				},
				Services: []string{
					"node",
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
			want: &identifyServices{
				Deployments: []string{
					"cli",
					"redis",
					"varnish",
					"nginx-php",
				},
				Volumes: []string{
					"nginx-php",
				},
				Services: []string{
					"redis",
					"varnish",
					"nginx-php",
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
			want: &identifyServices{
				Deployments: []string{
					"cli",
					"redis",
					"varnish",
					"nginx-php",
				},
				Volumes: []string{
					"nginx-php",
				},
				Services: []string{
					"redis",
					"varnish",
					"nginx-php",
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
			want: &identifyServices{
				Deployments: []string{
					"lnd",
					"thunderhub",
					"tor",
				},
				Volumes: []string{
					"lnd",
				},
				Services: []string{
					"lnd",
					"thunderhub",
					"tor",
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
			want: &identifyServices{
				Deployments: []string{
					"lnd",
					"tor",
				},
				Volumes: []string{
					"lnd",
				},
				Services: []string{
					"lnd",
				},
			},
		},

		{
			name: "test5-complex-custom-volumes",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					BuildType:       "branch",
					LagoonYAML:      "internal/testdata/complex/lagoon.multiple-volumes.yml",
					ImageReferences: map[string]string{
						"nginx":   "harbor.example/example-project/main/nginx@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"php":     "harbor.example/example-project/main/php@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"cli":     "harbor.example/example-project/main/cli@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"nginx2":  "harbor.example/example-project/main/nginx2@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"php2":    "harbor.example/example-project/main/php2@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"mariadb": "harbor.example/example-project/main/mariadb@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
				}, true),
			want: &identifyServices{
				Deployments: []string{
					"cli",
					"mariadb",
					"nginx",
				},
				Volumes: []string{
					"mariadb",
					"nginx",
					"custom-files",
				},
				Services: []string{
					"mariadb",
					"nginx",
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
			if !reflect.DeepEqual(out, tt.want) {
				r1, _ := json.MarshalIndent(out, "", " ")
				s1, _ := json.MarshalIndent(tt.want, "", " ")
				t.Errorf("LagoonServiceTemplateIdentification() = \n%v", diff.LineDiff(string(r1), string(s1)))
			}
			t.Cleanup(func() {
				helpers.UnsetEnvVars(nil)
				helpers.UnsetEnvVars(tt.args.BuildPodVariables)
			})
		})
	}
}
