package cmd

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/uselagoon/build-deploy-tool/internal/dbaasclient"
	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	"github.com/uselagoon/build-deploy-tool/internal/testdata"

	// changes the testing to source from root so paths to test resources must be defined from repo root
	_ "github.com/uselagoon/build-deploy-tool/internal/testing"
)

func TestTemplateLagoonServices(t *testing.T) {
	tests := []struct {
		name        string
		description string
		args        testdata.TestData
		want        string
		imageData   string
		vars        []helpers.EnvironmentVariable
	}{
		{
			name:        "test1-basic-deployment",
			description: "tests a basic deployment",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/basic/lagoon.container-registry-deep.yml",
					ImageReferences: map[string]string{
						"node": "harbor.example/example-project/main/node@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "REGISTRY_PASSWORD",
							Value: "myenvvarregistrypassword",
							Scope: "container_registry",
						},
						{
							Name:  "REGISTRY_DOCKERHUB_USERNAME",
							Value: "dockerhubusername",
							Scope: "container_registry",
						},
						{
							Name:  "REGISTRY_DOCKERHUB_PASSWORD",
							Value: "dockerhubpassword",
							Scope: "container_registry",
						},
						{
							Name:  "REGISTRY_MY_OTHER_REGISTRY_USERNAME",
							Value: "otherusername",
							Scope: "container_registry",
						},
						{
							Name:  "REGISTRY_MY_OTHER_REGISTRY_PASSWORD",
							Value: "otherpassword",
							Scope: "container_registry",
						},
					},
				}, true),
			want: "internal/testdata/basic/service-templates/test1-basic-deployment",
		},
		{
			name:        "test2-nginx-php",
			description: "tests an nginx-php deployment",
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
			want: "internal/testdata/complex/service-templates/test2-nginx-php",
		},
		{
			name:        "test2a-nginx-php",
			description: "tests an nginx-php deployment using images from images.yaml (same result as test2)",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/complex/lagoon.varnish.yml",
				}, true),
			imageData: "internal/testdata/complex/images-service1.yaml",
			want:      "internal/testdata/complex/service-templates/test2-nginx-php",
		},
		{
			name:        "test2b-nginx-php",
			description: "tests an nginx-php deployment with rootless workloads enabled",
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
			want: "internal/testdata/complex/service-templates/test2b-nginx-php",
		},
		{
			name:        "test2c-nginx-php",
			description: "tests an nginx-php deployment with spot workloads enabled",
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
						{
							Name:  "LAGOON_FEATURE_FLAG_SPOT_INSTANCE_PRODUCTION",
							Value: "enabled",
							Scope: "global",
						},
						{
							Name:  "LAGOON_FEATURE_FLAG_SPOT_INSTANCE_PRODUCTION_TYPES",
							Value: "nginx,nginx-persistent,nginx-php,nginx-php-persistent",
							Scope: "global",
						},
						{
							Name:  "LAGOON_FEATURE_FLAG_SPOT_INSTANCE_PRODUCTION_CRONJOB_TYPES",
							Value: "cli,cli-persistent",
							Scope: "global",
						},
					},
				}, true),
			want: "internal/testdata/complex/service-templates/test2c-nginx-php",
		},
		{
			name:        "test2d-nginx-php",
			description: "tests an nginx-php deployment with admin resource overrides",
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
					BuildPodVariables: []helpers.EnvironmentVariable{
						{
							Name:  "ADMIN_LAGOON_FEATURE_FLAG_CONTAINER_MEMORY_LIMIT",
							Value: "16Gi",
						},
						{
							Name:  "ADMIN_LAGOON_FEATURE_FLAG_EPHEMERAL_STORAGE_LIMIT",
							Value: "160Gi",
						},
						{
							Name:  "ADMIN_LAGOON_FEATURE_FLAG_EPHEMERAL_STORAGE_REQUESTS",
							Value: "1Gi",
						},
					},
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_ROOTLESS_WORKLOAD",
							Value: "enabled",
							Scope: "build",
						},
					},
				}, true),
			want: "internal/testdata/complex/service-templates/test2d-nginx-php",
		},
		{
			name:        "test3-funky-pvcs",
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
			want: "internal/testdata/basic/service-templates/test3-funky-pvcs",
		},
		{
			name:        "test4-basic-worker",
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
			want: "internal/testdata/basic/service-templates/test4-basic-worker",
		},
		{
			name:        "test5-basic-promote",
			description: "create a basic deployment of the promote build type",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					BuildType:       "promote",
					LagoonYAML:      "internal/testdata/basic/lagoon.yml",
					ImageReferences: map[string]string{
						"node": "harbor.example/example-project/main/node@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
				}, true),
			want: "internal/testdata/basic/service-templates/test5-basic-promote",
		},
		{
			name:        "test6-basic-networkpolicy",
			description: "create basic deployment pullrequest with isolation network policy",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "pr-123",
					BuildType:       "pullrequest",
					PRNumber:        "123",
					PRHeadBranch:    "pr-head",
					PRBaseBranch:    "pr-base",
					PRHeadSHA:       "123456",
					PRBaseSHA:       "abcdef",
					LagoonYAML:      "internal/testdata/basic/lagoon.yml",
					ImageReferences: map[string]string{
						"node": "harbor.example/example-project/pr-123/node@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_ISOLATION_NETWORK_POLICY",
							Value: "enabled",
							Scope: "build",
						},
					},
				}, true),
			want: "internal/testdata/basic/service-templates/test6-basic-networkpolicy",
		},
		{
			name:        "test7-basic-dynamic-secrets",
			description: "create a basic deployment with dynamic secrets support",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/basic/lagoon.container-registry.yml",
					ImageReferences: map[string]string{
						"node": "harbor.example/example-project/main/node@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
					DynamicSecrets:      []string{"insights-token"},
					DynamicDBaaSSecrets: []string{"mariadb-dbaas-a4hs12h3"},
				}, true),
			want: "internal/testdata/basic/service-templates/test7-basic-dynamic-secrets",
		},
		{
			name:        "test8-multiple-services",
			description: "create a deployment with multiple services of various types",
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
			want: "internal/testdata/complex/service-templates/test8-multiple-services",
		},
		{
			name:        "test9-meta-dbaas-types",
			description: "create a deployment with meta dbaas types",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/complex/lagoon.compact-services.yml",
					ImageReferences: map[string]string{
						"mariadb-10-5":  "harbor.example/example-project/main/mariadb-10-5@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"mariadb-10-11": "harbor.example/example-project/main/mariadb-10-11@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"postgres-11":   "harbor.example/example-project/main/postgres-11@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"postgres-15":   "harbor.example/example-project/main/postgres-15@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"mongo-4":       "harbor.example/example-project/main/mongo-4@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_DBAAS_ENVIRONMENT_TYPES",
							Value: "postgres-15:production-postgres,mongo-4:production-mongo,mariadb-10-11:production-mariadb",
							Scope: "build"},
						{
							Name:  "LAGOON_SYSTEM_CORE_VERSION",
							Value: "v2.19.0",
							Scope: "internal_system",
						},
					},
				}, true),
			want: "internal/testdata/complex/service-templates/test9-meta-dbaas-types",
		},
		{
			name:        "test10-basic-no-native-cronjobs",
			description: "create a basic deployment with native cronjobs disabled",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/basic/lagoon-cronjob-native-disable.yml",
					ImageReferences: map[string]string{
						"node": "harbor.example/example-project/main/node@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
				}, true),
			want: "internal/testdata/basic/service-templates/test10-basic-no-native-cronjobs",
		},
		{
			name:        "test11-basic-polysite-cronjobs",
			description: "create a basic deployment polysite with cronjobs",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/basic/lagoon.polysite-cronjobs.yml",
					ImageReferences: map[string]string{
						"node": "harbor.example/example-project/main/node@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
				}, true),
			want: "internal/testdata/basic/service-templates/test11-basic-polysite-cronjobs",
		},
		{
			name:        "test12-basic-persistent-custom-volumes",
			description: "create a basic persistent with the seed volume and other custom volumes",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					BuildType:       "branch",
					LagoonYAML:      "internal/testdata/basic/lagoon.multiple-volumes.yml",
					ImageReferences: map[string]string{
						"node": "harbor.example/example-project/main/node@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
				}, true),
			want: "internal/testdata/basic/service-templates/test12-basic-persistent-custom-volumes",
		},
		{
			name:        "test13-basic-custom-volumes",
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
			want: "internal/testdata/basic/service-templates/test13-basic-custom-volumes",
		},
		{
			name: "test14-complex-custom-volumes",
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
			want: "internal/testdata/complex/service-templates/test14-complex-custom-volumes",
		},
		{
			name:        "test15-basic-custom-volume-no-backup",
			description: "create a basic with custom volumes with one volume flagged to not be backed up",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					BuildType:       "branch",
					LagoonYAML:      "internal/testdata/basic/lagoon.multiple-volumes-3.yml",
					ImageReferences: map[string]string{
						"node": "harbor.example/example-project/main/node@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
				}, true),
			want: "internal/testdata/basic/service-templates/test15-basic-custom-volume-no-backup",
		},
		{
			name: "test-basic-spot-affinity",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					EnvironmentType: "development",
					LagoonYAML:      "internal/testdata/basic/lagoon.yml",
					ImageReferences: map[string]string{
						"node": "harbor.example/example-project/main/node@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
					BuildPodVariables: []helpers.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_DEFAULT_POD_SPREADCONSTRAINTS",
							Value: "enabled",
						},
						{
							Name:  "LAGOON_FEATURE_FLAG_DEFAULT_ROOTLESS_WORKLOAD",
							Value: "enabled",
						},
						{
							Name:  "LAGOON_FEATURE_FLAG_DEFAULT_SPOT_INSTANCE_DEVELOPMENT",
							Value: "enabled",
						},
						{
							Name:  "LAGOON_FEATURE_FLAG_DEFAULT_SPOT_INSTANCE_DEVELOPMENT_TYPES",
							Value: "basic,basic-persistent",
						},
						{
							// `ADMIN_` are only configurable by the remote-controller
							Name:  "ADMIN_LAGOON_FEATURE_FLAG_SPOT_TYPE_REPLICAS_DEVELOPMENT",
							Value: "basic,basic-persistent",
						},
					},
				}, true),
			want: "internal/testdata/basic/service-templates/test-basic-spot-affinity",
		},
		{
			name:        "test-basic-persistent-name",
			description: "tests a basic deployment with a persistent name",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/basic/lagoon.persistent-name.yml",
					ImageReferences: map[string]string{
						"basic": "harbor.example/example-project/main/basic@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
				}, true),
			want: "internal/testdata/basic/service-templates/test-basic-persistent-name",
		},
		{
			name:        "test-basic-persistent-names",
			description: "tests a basic deployment with two services with persistent name",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/basic/lagoon.persistent-name-2.yml",
					ImageReferences: map[string]string{
						"basic":  "harbor.example/example-project/main/basic@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"basic2": "harbor.example/example-project/main/basic2@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
				}, true),
			want: "internal/testdata/basic/service-templates/test-basic-persistent-names",
		},
		{
			name:        "test-complex-persistent-names",
			description: "tests a complex deployment with two services with persistent name",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/complex/lagoon.persistent-name.yml",
					ImageReferences: map[string]string{
						"cli":                           "harbor.example/example-project/main/cli@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"gotenberg":                     "harbor.example/example-project/main/gotenberg@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"nginx":                         "harbor.example/example-project/main/nginx@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"opensearch":                    "harbor.example/example-project/main/opensearch@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"php":                           "harbor.example/example-project/main/php@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"rabbitmq":                      "harbor.example/example-project/main/rabbitmq@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"redis":                         "harbor.example/example-project/main/redis@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"redis-persist":                 "harbor.example/example-project/main/redis-persist@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"redis-session":                 "harbor.example/example-project/main/redis-session@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"queue-worker-entity-index":     "harbor.example/example-project/main/queue-worker-entity-index@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"queue-worker-priority-high":    "harbor.example/example-project/main/queue-worker-priority-high@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"queue-worker-priority-instant": "harbor.example/example-project/main/queue-worker-priority-instant@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"queue-worker-priority-low":     "harbor.example/example-project/main/queue-worker-priority-low@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
						"queue-worker-priority-medium":  "harbor.example/example-project/main/queue-worker-priority-medium@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
				}, true),
			want: "internal/testdata/complex/service-templates/test-complex-persistent-names",
		},
		{
			name:        "test-redis-persistent-k8upv2",
			description: "tests a basic deployment with a redis-persistent with k8up v2 featureflag",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/complex/lagoon.redis-persistent.yml",
					ImageReferences: map[string]string{
						"redis": "harbor.example/example-project/main/redis@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
					BuildPodVariables: []helpers.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_DEFAULT_K8UP_V2",
							Value: "enabled",
						},
					},
				}, true),
			want: "internal/testdata/complex/service-templates/test-redis-persistent-k8upv2",
		},
		{
			name:        "test-basic-single-k8upv2",
			description: "tests a basic deployment with a basic-single with k8up v2 featureflag",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/basic/lagoon.basic-single.yml",
					ImageReferences: map[string]string{
						"basic": "harbor.example/example-project/main/basic@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
					BuildPodVariables: []helpers.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_DEFAULT_K8UP_V2",
							Value: "enabled",
						},
					},
				}, true),
			want: "internal/testdata/basic/service-templates/test-basic-single-k8upv2",
		},
		{
			name:        "test-basic-external-service",
			description: "tests a basic deployment with a basic and an external service and network policy to share a service",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/basic/lagoon.externalservice.yml",
					ImageReferences: map[string]string{
						"basic1": "harbor.example/example-project/main/basic1@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
				}, true),
			want: "internal/testdata/basic/service-templates/test-basic-external-service",
		},
		{
			name:        "test-basic-external-service-environment",
			description: "tests a basic deployment with a basic and an external service and network policy to share when defined under the environment",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "stage",
					Branch:          "stage",
					LagoonYAML:      "internal/testdata/basic/lagoon.externalservice.yml",
					ImageReferences: map[string]string{
						"basic1": "harbor.example/example-project/stage/basic1@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
				}, true),
			want: "internal/testdata/basic/service-templates/test-basic-external-service-environment",
		},
		{
			name:        "test-basic-deployment-revision-history",
			description: "tests a basic deployment with revision history set",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/basic/lagoon.yml",
					ImageReferences: map[string]string{
						"node": "harbor.example/example-project/main/node@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
					BuildPodVariables: []helpers.EnvironmentVariable{
						{
							Name:  "ADMIN_LAGOON_FEATURE_FLAG_DEPLOYMENT_REVISION_HISTORY",
							Value: "5",
						},
					},
				}, true),
			templatePath: "testoutput",
			want:         "internal/testdata/basic/service-templates/test-basic-deployment-revision-history",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.UnsetEnvVars(tt.vars) //unset variables before running tests
			for _, envVar := range tt.vars {
				err := os.Setenv(envVar.Name, envVar.Value)
				if err != nil {
					t.Errorf("%v", err)
				}
			}
			// set the environment variables from args
			savedTemplates, err := os.MkdirTemp("", "testoutput")
			if err != nil {
				t.Errorf("%v", err)
			}
			generator, err := testdata.SetupEnvironment(generator.GeneratorInput{}, savedTemplates, tt.args)
			if err != nil {
				t.Errorf("%v", err)
			}
			defer os.RemoveAll(savedTemplates)

			ts := dbaasclient.TestDBaaSHTTPServer()
			defer ts.Close()
			err = os.Setenv("DBAAS_OPERATOR_HTTP", ts.URL)
			if err != nil {
				t.Errorf("%v", err)
			}

			if tt.imageData != "" {
				imageRefs, err := loadImagesFromFile(tt.imageData)
				if err != nil {
					t.Errorf("%v", err)
				}
				generator.ImageReferences = imageRefs.Images
			}
			err = LagoonServiceTemplateGeneration(generator)
			if err != nil {
				t.Errorf("%v", err)
			}

			files, err := os.ReadDir(savedTemplates)
			if err != nil {
				t.Errorf("couldn't read directory %v: %v", savedTemplates, err)
			}
			results, err := os.ReadDir(tt.want)
			if err != nil {
				t.Errorf("couldn't read directory %v: %v", tt.want, err)
			}
			if len(files) != len(results) {
				for _, f := range files {
					f1, err := os.ReadFile(fmt.Sprintf("%s/%s", savedTemplates, f.Name()))
					if err != nil {
						t.Errorf("couldn't read file %v: %v", savedTemplates, err)
					}
					fmt.Println(string(f1))
				}
				t.Errorf("number of generated templates doesn't match results %v/%v: %v", len(files), len(results), err)
			}
			fCount := 0
			for _, f := range files {
				for _, r := range results {
					if f.Name() == r.Name() {
						fCount++
						f1, err := os.ReadFile(fmt.Sprintf("%s/%s", savedTemplates, f.Name()))
						if err != nil {
							t.Errorf("couldn't read file %v: %v", savedTemplates, err)
						}
						r1, err := os.ReadFile(fmt.Sprintf("%s/%s", tt.want, f.Name()))
						if err != nil {
							t.Errorf("couldn't read file %v: %v", tt.want, err)
						}
						if !reflect.DeepEqual(f1, r1) {
							t.Errorf("TemplateLagoonServices() = \n%v", diff.LineDiff(string(r1), string(f1)))
						}
					}
				}
			}
			if fCount != len(files) {
				for _, f := range files {
					f1, err := os.ReadFile(fmt.Sprintf("%s/%s", savedTemplates, f.Name()))
					if err != nil {
						t.Errorf("couldn't read file %v: %v", savedTemplates, err)
					}
					fmt.Println(string(f1))
				}
				t.Errorf("resulting templates do not match")
			}
			t.Cleanup(func() {
				helpers.UnsetEnvVars(tt.vars)
				helpers.UnsetEnvVars(tt.args.BuildPodVariables)
			})
		})
	}
}
