package identify

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/uselagoon/build-deploy-tool/internal/collector"
	"github.com/uselagoon/build-deploy-tool/internal/dbaasclient"
	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/k8s"
	"github.com/uselagoon/build-deploy-tool/internal/testdata"

	// changes the testing to source from root so paths to test resources must be defined from repo root
	_ "github.com/uselagoon/build-deploy-tool/internal/testing"
)

func TestGetCurrentState(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		args           testdata.TestData
		deleteServices bool
		wantErr        bool
		seedDir        string
		wantServices   LagoonServices
	}{
		{
			name: "basic-deployment",
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
			deleteServices: true,
			namespace:      "example-project-main",
			seedDir:        "internal/testdata/basic/cleanup-seed/basic-deployment",
			wantServices: LagoonServices{
				Services: []EnvironmentService{
					{
						Name:      "mariadb",
						Type:      "mariadb-dbaas",
						Abandoned: true,
					},
					{
						Name:      "basic",
						Type:      "basic",
						Abandoned: true,
						Containers: []ServiceContainer{
							{
								Name: "basic",
								Ports: []ContainerPort{
									{
										Name: "tcp-1234",
										Port: 1234,
									},
									{
										Name: "tcp-8191",
										Port: 8191,
									},
									{
										Name: "udp-9001",
										Port: 9001,
									},
								},
							},
						},
					},
					{
						Name: "node",
						Type: "basic",
						Containers: []ServiceContainer{
							{
								Name: "basic",
								Ports: []ContainerPort{
									{
										Name: "tcp-1234",
										Port: 1234,
									},
									{
										Name: "tcp-8191",
										Port: 8191,
									},
									{
										Name: "udp-9001",
										Port: 9001,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "multivolumes",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/basic/lagoon.multiple-volumes.yml",
					ImageReferences: map[string]string{
						"node": "harbor.example/example-project/main/node@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
				}, true),
			deleteServices: false,
			namespace:      "example-project-main",
			seedDir:        "internal/testdata/basic/service-templates/test12-basic-persistent-custom-volumes",
			wantServices: LagoonServices{
				Services: []EnvironmentService{
					{
						Name: "node",
						Type: "basic-persistent",
						Containers: []ServiceContainer{
							{
								Name: "basic",
								Volumes: []VolumeMount{
									{
										Name: "custom-config",
										Path: "/config",
									},
									{
										Name: "custom-files",
										Path: "/app/files/",
									},
									{
										Name: "node",
										Path: "/data",
									},
								},
								Ports: []ContainerPort{
									{
										Name: "http",
										Port: 3000,
									},
								},
							},
						},
					},
				},
				Volumes: []EnvironmentVolume{
					{
						Name:        "custom-config",
						StorageType: "bulk",
						Type:        "additional-volume",
						Size:        "5Gi",
					},
					{
						Name:        "custom-files",
						StorageType: "bulk",
						Type:        "additional-volume",
						Size:        "10Gi",
					},
					{
						Name:        "node",
						StorageType: "bulk",
						Type:        "basic-persistent",
						Size:        "5Gi",
					},
				},
			},
		},
		{
			name: "complex-singles",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/complex/lagoon.services.yml",
					ImageReferences: map[string]string{
						"web":          "harbor.example/example-project/main/web@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"mariadb-10-5": "harbor.example/example-project/main/mariadb-10-5@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"postgres-11":  "harbor.example/example-project/main/postgres-11@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"opensearch-2": "harbor.example/example-project/main/opensearch-2@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"redis-6":      "harbor.example/example-project/main/redis-6@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"redis-7":      "harbor.example/example-project/main/redis-7@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"solr-8":       "harbor.example/example-project/main/solr-8@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
				}, true),
			deleteServices: false,
			namespace:      "example-project-main",
			seedDir:        "internal/testdata/complex/service-templates/test8-multiple-services",
			wantServices: LagoonServices{
				Services: []EnvironmentService{
					{
						Name: "mariadb-10-5",
						Type: "mariadb-single",
						Containers: []ServiceContainer{
							{
								Name: "mariadb-single",
								Volumes: []VolumeMount{
									{
										Name: "mariadb-10-5",
										Path: "/var/lib/mysql",
									},
								},
								Ports: []ContainerPort{
									{
										Name: "3306-tcp",
										Port: 3306,
									},
								},
							},
						},
					},
					{
						Name: "opensearch-2",
						Type: "opensearch-persistent",
						Containers: []ServiceContainer{
							{
								Name: "opensearch",
								Volumes: []VolumeMount{
									{
										Name: "opensearch-2",
										Path: "/usr/share/opensearch/data",
									},
								},
								Ports: []ContainerPort{
									{
										Name: "9200-tcp",
										Port: 9200,
									},
								},
							},
						},
					},
					{
						Name: "postgres-11",
						Type: "postgres-single",
						Containers: []ServiceContainer{
							{
								Name: "postgres-single",
								Volumes: []VolumeMount{
									{
										Name: "postgres-11",
										Path: "/var/lib/postgresql/data",
									},
								},
								Ports: []ContainerPort{
									{
										Name: "5432-tcp",
										Port: 5432,
									},
								},
							},
						},
					},
					{
						Name: "redis-6",
						Type: "redis",
						Containers: []ServiceContainer{
							{
								Name: "redis",
								Ports: []ContainerPort{
									{
										Name: "6379-tcp",
										Port: 6379,
									},
								},
							},
						},
					},
					{
						Name: "redis-7",
						Type: "redis",
						Containers: []ServiceContainer{
							{
								Name: "redis",
								Ports: []ContainerPort{
									{
										Name: "6379-tcp",
										Port: 6379,
									},
								},
							},
						},
					},
					{
						Name: "solr-8",
						Type: "solr-php-persistent",
						Containers: []ServiceContainer{
							{
								Name: "solr",
								Volumes: []VolumeMount{
									{
										Name: "solr-8",
										Path: "/var/solr",
									},
								},
								Ports: []ContainerPort{
									{
										Name: "8983-tcp",
										Port: 8983,
									},
								},
							},
						},
					},
					{
						Name: "web",
						Type: "basic-persistent",
						Containers: []ServiceContainer{
							{
								Name: "basic",
								Volumes: []VolumeMount{
									{
										Name: "web",
										Path: "/app/files",
									},
								},
								Ports: []ContainerPort{
									{
										Name: "http",
										Port: 3000,
									},
								},
							},
						},
					},
					{
						Name: "mariadb-10-11",
						Type: "mariadb-dbaas",
					},
					{
						Name: "postgres-15",
						Type: "postgres-dbaas",
					},
					{
						Name: "mongo-4",
						Type: "mongodb-dbaas",
					},
				},
				Volumes: []EnvironmentVolume{
					{
						Name:        "mariadb-10-5",
						StorageType: "block",
						Type:        "mariadb-single",
						Size:        "100Mi",
					},
					{
						Name:        "opensearch-2",
						StorageType: "block",
						Type:        "opensearch-persistent",
						Size:        "100Mi",
					},
					{
						Name:        "postgres-11",
						StorageType: "block",
						Type:        "postgres-single",
						Size:        "100Mi",
					},
					{
						Name:        "solr-8",
						StorageType: "block",
						Type:        "solr-php-persistent",
						Size:        "100Mi",
					},
					{
						Name:        "web",
						StorageType: "bulk",
						Type:        "basic-persistent",
						Size:        "10Mi",
					},
				},
			},
		},
		{
			name: "complex-nginx",
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
			deleteServices: false,
			namespace:      "example-project-main",
			seedDir:        "internal/testdata/complex/service-templates/test2-nginx-php",
			wantServices: LagoonServices{
				Services: []EnvironmentService{
					{
						Name: "cli",
						Type: "cli-persistent",
						Containers: []ServiceContainer{
							{
								Name: "cli",
								Volumes: []VolumeMount{
									{
										Name: "nginx-php",
										Path: "/app/docroot/sites/default/files/",
									},
								},
							},
						},
					},
					{
						Name: "nginx-php",
						Type: "nginx-php-persistent",
						Containers: []ServiceContainer{
							{
								Name: "nginx",
								Volumes: []VolumeMount{
									{
										Name: "nginx-php",
										Path: "/app/docroot/sites/default/files/",
									},
								},
								Ports: []ContainerPort{
									{
										Name: "http",
										Port: 8080,
									},
								},
							},
							{
								Name: "php",
								Volumes: []VolumeMount{
									{
										Name: "nginx-php",
										Path: "/app/docroot/sites/default/files/",
									},
								},
								Ports: []ContainerPort{
									{
										Name: "php",
										Port: 9000,
									},
								},
							},
						},
					},
					{
						Name: "redis",
						Type: "redis",
						Containers: []ServiceContainer{
							{
								Name: "redis",
								Ports: []ContainerPort{
									{
										Name: "6379-tcp",
										Port: 6379,
									},
								},
							},
						},
					},
					{
						Name: "varnish",
						Type: "varnish",
						Containers: []ServiceContainer{
							{
								Name: "varnish",
								Ports: []ContainerPort{
									{
										Name: "http",
										Port: 8080,
									},
									{
										Name: "controlport",
										Port: 6082,
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
				Volumes: []EnvironmentVolume{
					{
						Name:        "nginx-php",
						StorageType: "bulk",
						Type:        "nginx-php-persistent",
						Size:        "5Gi",
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
			generator, err := testdata.SetupEnvironment(generator.GeneratorInput{}, savedTemplates, tt.args)
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

			client, err := k8s.NewFakeClient(tt.namespace)
			if err != nil {
				t.Errorf("error creating fake client")
			}
			err = k8s.SeedFakeData(client, tt.namespace, tt.seedDir)
			if err != nil {
				t.Errorf("error seeding fake data: %v", err)
			}
			col := collector.NewCollector(client)
			lagoonServices, _, _, _, _, _, _, _, err := GetCurrentState(col, generator)
			if err != nil {
				t.Errorf("GetCurrentState() %v ", err)
			}
			r1, _ := json.MarshalIndent(lagoonServices, "", " ")
			s1, _ := json.MarshalIndent(tt.wantServices, "", " ")
			if !reflect.DeepEqual(r1, s1) {
				t.Errorf("GetCurrentState() = \n%v", diff.LineDiff(string(s1), string(r1)))
			}
		})
	}
}
