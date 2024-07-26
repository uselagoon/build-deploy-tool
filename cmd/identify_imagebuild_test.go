package cmd

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/uselagoon/build-deploy-tool/internal/dbaasclient"
	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	"github.com/uselagoon/build-deploy-tool/internal/testdata"

	// changes the testing to source from root so paths to test resources must be defined from repo root
	_ "github.com/uselagoon/build-deploy-tool/internal/testing"
)

func TestImageBuildConfigurationIdentification(t *testing.T) {
	tests := []struct {
		name        string
		description string
		args        testdata.TestData
		want        imageBuild
		vars        []helpers.EnvironmentVariable
	}{
		{
			name: "test1 basic deployment",
			args: testdata.GetSeedData(
				testdata.TestData{
					Namespace:       "example-project-main",
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/basic/lagoon.yml",
				}, true),
			want: imageBuild{
				BuildKit: helpers.BoolPtr(true),
				BuildArguments: map[string]string{
					"LAGOON_BUILD_NAME":            "lagoon-build-abcdefg",
					"LAGOON_PROJECT":               "example-project",
					"LAGOON_ENVIRONMENT":           "main",
					"LAGOON_ENVIRONMENT_TYPE":      "production",
					"LAGOON_BUILD_TYPE":            "branch",
					"LAGOON_GIT_SOURCE_REPOSITORY": "ssh://git@example.com/lagoon-demo.git",
					"LAGOON_KUBERNETES":            "remote-cluster1",
					"LAGOON_GIT_SHA":               "abcdefg123456",
					"LAGOON_GIT_BRANCH":            "main",
					"NODE_IMAGE":                   "example-project-main-node",
				},
				Images: []imageBuilds{
					{
						Name: "node",
						ImageBuild: generator.ImageBuild{
							BuildImage:     "harbor.example/example-project/main/node:latest",
							Context:        "internal/testdata/basic/docker",
							DockerFile:     "basic.dockerfile",
							TemporaryImage: "example-project-main-node",
						},
					},
				},
			},
		},
		{
			name: "test2a nginx-php deployment",
			args: testdata.GetSeedData(
				testdata.TestData{
					Namespace:       "example-project-main",
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/complex/lagoon.varnish.yml",
				}, true),
			want: imageBuild{
				BuildKit: helpers.BoolPtr(true),
				BuildArguments: map[string]string{
					"LAGOON_BUILD_NAME":            "lagoon-build-abcdefg",
					"LAGOON_PROJECT":               "example-project",
					"LAGOON_ENVIRONMENT":           "main",
					"LAGOON_ENVIRONMENT_TYPE":      "production",
					"LAGOON_BUILD_TYPE":            "branch",
					"LAGOON_GIT_SOURCE_REPOSITORY": "ssh://git@example.com/lagoon-demo.git",
					"LAGOON_KUBERNETES":            "remote-cluster1",
					"LAGOON_GIT_SHA":               "0000000000000000000000000000000000000000",
					"LAGOON_GIT_BRANCH":            "main",
					"CLI_IMAGE":                    "example-project-main-cli",
					"NGINX_IMAGE":                  "example-project-main-nginx",
					"PHP_IMAGE":                    "example-project-main-php",
				},
				Images: []imageBuilds{
					{
						Name: "cli",
						ImageBuild: generator.ImageBuild{
							BuildImage:     "harbor.example/example-project/main/cli:latest",
							Context:        "internal/testdata/complex/docker",
							DockerFile:     ".docker/Dockerfile.cli",
							TemporaryImage: "example-project-main-cli",
						},
					}, {
						Name: "nginx",
						ImageBuild: generator.ImageBuild{
							BuildImage:     "harbor.example/example-project/main/nginx:latest",
							Context:        "internal/testdata/complex/docker",
							DockerFile:     ".docker/Dockerfile.nginx-drupal",
							TemporaryImage: "example-project-main-nginx",
						},
					}, {
						Name: "php",
						ImageBuild: generator.ImageBuild{
							BuildImage:     "harbor.example/example-project/main/php:latest",
							Context:        "internal/testdata/complex/docker",
							DockerFile:     ".docker/Dockerfile.php",
							TemporaryImage: "example-project-main-php",
						},
					}, {
						Name: "redis",
						ImageBuild: generator.ImageBuild{
							BuildImage: "harbor.example/example-project/main/redis:latest",
							PullImage:  "quay.io/notlagoon/redis",
						},
					}, {
						Name: "varnish",
						ImageBuild: generator.ImageBuild{
							BuildImage: "harbor.example/example-project/main/varnish:latest",
							PullImage:  "uselagoon/varnish-5-drupal:latest",
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
							Name:  "LAGOON_FEATURE_FLAG_IMAGECACHE_REGISTRY",
							Value: "imagecache.example.com",
							Scope: "build",
						},
					},
				}, true),
			want: imageBuild{
				BuildKit: helpers.BoolPtr(true),
				BuildArguments: map[string]string{
					"LAGOON_FEATURE_FLAG_ROOTLESS_WORKLOAD":   "enabled",
					"LAGOON_BUILD_NAME":                       "lagoon-build-abcdefg",
					"LAGOON_PROJECT":                          "example-project",
					"LAGOON_ENVIRONMENT":                      "main",
					"LAGOON_ENVIRONMENT_TYPE":                 "production",
					"LAGOON_BUILD_TYPE":                       "branch",
					"LAGOON_GIT_SOURCE_REPOSITORY":            "ssh://git@example.com/lagoon-demo.git",
					"LAGOON_KUBERNETES":                       "remote-cluster1",
					"LAGOON_GIT_SHA":                          "0000000000000000000000000000000000000000",
					"LAGOON_GIT_BRANCH":                       "main",
					"CLI_IMAGE":                               "example-project-main-cli",
					"NGINX_IMAGE":                             "example-project-main-nginx",
					"PHP_IMAGE":                               "example-project-main-php",
					"LAGOON_FEATURE_FLAG_IMAGECACHE_REGISTRY": "imagecache.example.com",
				},
				Images: []imageBuilds{
					{
						Name: "cli",
						ImageBuild: generator.ImageBuild{
							BuildImage:     "harbor.example/example-project/main/cli:latest",
							Context:        "internal/testdata/complex/docker",
							DockerFile:     ".docker/Dockerfile.cli",
							TemporaryImage: "example-project-main-cli",
						},
					}, {
						Name: "nginx",
						ImageBuild: generator.ImageBuild{
							BuildImage:     "harbor.example/example-project/main/nginx:latest",
							Context:        "internal/testdata/complex/docker",
							DockerFile:     ".docker/Dockerfile.nginx-drupal",
							TemporaryImage: "example-project-main-nginx",
						},
					}, {
						Name: "php",
						ImageBuild: generator.ImageBuild{
							BuildImage:     "harbor.example/example-project/main/php:latest",
							Context:        "internal/testdata/complex/docker",
							DockerFile:     ".docker/Dockerfile.php",
							TemporaryImage: "example-project-main-php",
						},
					}, {
						Name: "redis",
						ImageBuild: generator.ImageBuild{
							BuildImage: "harbor.example/example-project/main/redis:latest",
							PullImage:  "quay.io/notlagoon/redis",
						},
					}, {
						Name: "varnish",
						ImageBuild: generator.ImageBuild{
							BuildImage: "harbor.example/example-project/main/varnish:latest",
							PullImage:  "imagecache.example.com/uselagoon/varnish-5-drupal:latest",
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
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_ROOTLESS_WORKLOAD",
							Value: "enabled",
							Scope: "build",
						},
					},
				}, true),
			want: imageBuild{
				BuildKit: helpers.BoolPtr(true),
				BuildArguments: map[string]string{
					"LAGOON_FEATURE_FLAG_ROOTLESS_WORKLOAD": "enabled",
					"LAGOON_BUILD_NAME":                     "lagoon-build-abcdefg",
					"LAGOON_PROJECT":                        "example-project",
					"LAGOON_ENVIRONMENT":                    "main",
					"LAGOON_ENVIRONMENT_TYPE":               "production",
					"LAGOON_BUILD_TYPE":                     "branch",
					"LAGOON_GIT_SOURCE_REPOSITORY":          "ssh://git@example.com/lagoon-demo.git",
					"LAGOON_KUBERNETES":                     "remote-cluster1",
					"LAGOON_GIT_SHA":                        "0000000000000000000000000000000000000000",
					"LAGOON_GIT_BRANCH":                     "main",
					"LND_IMAGE":                             "example-project-main-lnd",
					"THUNDERHUB_IMAGE":                      "example-project-main-thunderhub",
					"TOR_IMAGE":                             "example-project-main-tor",
				},
				Images: []imageBuilds{
					{
						Name: "lnd",
						ImageBuild: generator.ImageBuild{
							BuildImage:     "harbor.example/example-project/main/lnd:latest",
							Context:        "internal/testdata/basic/docker",
							DockerFile:     "Dockerfile",
							TemporaryImage: "example-project-main-lnd",
						},
					}, {
						Name: "thunderhub",
						ImageBuild: generator.ImageBuild{
							BuildImage:     "harbor.example/example-project/main/thunderhub:latest",
							Context:        "internal/testdata/basic/docker",
							DockerFile:     "Dockerfile",
							TemporaryImage: "example-project-main-thunderhub",
						},
					}, {
						Name: "tor",
						ImageBuild: generator.ImageBuild{
							BuildImage:     "harbor.example/example-project/main/tor:latest",
							Context:        "internal/testdata/basic/docker",
							DockerFile:     "Dockerfile",
							TemporaryImage: "example-project-main-tor",
						},
					},
				},
			},
		},
		{
			name:        "test4 - basic-persistent with worker-persistent with buildkit disabled",
			description: "create a basic-persistent that gets a pvc and mount that volume on a worker-persistent type",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/basic/lagoon.thunderhub-2.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_ROOTLESS_WORKLOAD",
							Value: "enabled",
							Scope: "build",
						},
						{
							Name:  "DOCKER_BUILDKIT",
							Value: "false",
							Scope: "build",
						},
					},
				}, true),
			want: imageBuild{
				BuildKit: helpers.BoolPtr(false),
				BuildArguments: map[string]string{
					"LAGOON_FEATURE_FLAG_ROOTLESS_WORKLOAD": "enabled",
					"DOCKER_BUILDKIT":                       "false",
					"LAGOON_BUILD_NAME":                     "lagoon-build-abcdefg",
					"LAGOON_PROJECT":                        "example-project",
					"LAGOON_ENVIRONMENT":                    "main",
					"LAGOON_ENVIRONMENT_TYPE":               "production",
					"LAGOON_BUILD_TYPE":                     "branch",
					"LAGOON_GIT_SOURCE_REPOSITORY":          "ssh://git@example.com/lagoon-demo.git",
					"LAGOON_KUBERNETES":                     "remote-cluster1",
					"LAGOON_GIT_SHA":                        "0000000000000000000000000000000000000000",
					"LAGOON_GIT_BRANCH":                     "main",
					"LND_IMAGE":                             "example-project-main-lnd",
					"TOR_IMAGE":                             "example-project-main-tor",
				},
				Images: []imageBuilds{
					{
						Name: "lnd",
						ImageBuild: generator.ImageBuild{
							BuildImage:     "harbor.example/example-project/main/lnd:latest",
							Context:        "internal/testdata/basic/docker",
							DockerFile:     "Dockerfile",
							TemporaryImage: "example-project-main-lnd",
						},
					}, {
						Name: "tor",
						ImageBuild: generator.ImageBuild{
							BuildImage:     "harbor.example/example-project/main/tor:latest",
							Context:        "internal/testdata/basic/docker",
							DockerFile:     "Dockerfile",
							TemporaryImage: "example-project-main-tor",
						},
					},
				},
			},
		},
		{
			name: "test5 basic deployment promote",
			args: testdata.GetSeedData(
				testdata.TestData{
					Namespace:       "example-project-main",
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					BuildType:       "promote",
					LagoonYAML:      "internal/testdata/basic/lagoon.yml",
				}, true),
			want: imageBuild{
				BuildKit: helpers.BoolPtr(true),
				BuildArguments: map[string]string{
					"LAGOON_BUILD_NAME":            "lagoon-build-abcdefg",
					"LAGOON_PROJECT":               "example-project",
					"LAGOON_ENVIRONMENT":           "main",
					"LAGOON_ENVIRONMENT_TYPE":      "production",
					"LAGOON_BUILD_TYPE":            "promote",
					"LAGOON_GIT_SOURCE_REPOSITORY": "ssh://git@example.com/lagoon-demo.git",
					"LAGOON_KUBERNETES":            "remote-cluster1",
				},
				Images: []imageBuilds{
					{
						Name: "node",
						ImageBuild: generator.ImageBuild{
							BuildImage:   "harbor.example/example-project/main/node:latest",
							PromoteImage: "harbor.example/example-project/promote-main/node:latest",
						},
					},
				},
			},
		},
		{
			name: "test6 basic deployment pr",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "pr-123",
					BuildType:       "pullrequest",
					PRNumber:        "123",
					PRTitle:         "My PullRequest",
					PRHeadBranch:    "pr-head",
					PRBaseBranch:    "pr-base",
					PRHeadSHA:       "123456",
					PRBaseSHA:       "abcdef",
					LagoonYAML:      "internal/testdata/basic/lagoon.yml",
					ImageReferences: map[string]string{
						"node": "harbor.example/example-project/pr-123/node@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
				}, true),
			want: imageBuild{
				BuildKit: helpers.BoolPtr(true),
				BuildArguments: map[string]string{
					"LAGOON_BUILD_NAME":            "lagoon-build-abcdefg",
					"LAGOON_PROJECT":               "example-project",
					"LAGOON_ENVIRONMENT":           "pr-123",
					"LAGOON_ENVIRONMENT_TYPE":      "production",
					"LAGOON_BUILD_TYPE":            "pullrequest",
					"LAGOON_PR_BASE_BRANCH":        "pr-base",
					"LAGOON_PR_BASE_SHA":           "abcdef",
					"LAGOON_PR_HEAD_BRANCH":        "pr-head",
					"LAGOON_PR_HEAD_SHA":           "123456",
					"LAGOON_PR_NUMBER":             "123",
					"LAGOON_PR_TITLE":              "My PullRequest",
					"LAGOON_GIT_SHA":               "abcdefg123456",
					"LAGOON_GIT_SOURCE_REPOSITORY": "ssh://git@example.com/lagoon-demo.git",
					"LAGOON_KUBERNETES":            "remote-cluster1",
					"NODE_IMAGE":                   "example-project-pr-123-node",
				},
				Images: []imageBuilds{
					{
						Name: "node",
						ImageBuild: generator.ImageBuild{
							BuildImage:     "harbor.example/example-project/pr-123/node:latest",
							Context:        "internal/testdata/basic/docker",
							DockerFile:     "basic.dockerfile",
							TemporaryImage: "example-project-pr-123-node",
						},
					},
				},
			},
		},
		{
			name: "test7 nginx-php deployment promote",
			args: testdata.GetSeedData(
				testdata.TestData{
					Namespace:       "example-project-main",
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					BuildType:       "promote",
					LagoonYAML:      "internal/testdata/complex/lagoon.varnish.yml",
				}, true),
			want: imageBuild{
				BuildKit: helpers.BoolPtr(true),
				BuildArguments: map[string]string{
					"LAGOON_BUILD_NAME":            "lagoon-build-abcdefg",
					"LAGOON_PROJECT":               "example-project",
					"LAGOON_ENVIRONMENT":           "main",
					"LAGOON_ENVIRONMENT_TYPE":      "production",
					"LAGOON_BUILD_TYPE":            "promote",
					"LAGOON_GIT_SOURCE_REPOSITORY": "ssh://git@example.com/lagoon-demo.git",
					"LAGOON_KUBERNETES":            "remote-cluster1",
				},
				Images: []imageBuilds{
					{
						Name: "cli",
						ImageBuild: generator.ImageBuild{
							BuildImage:   "harbor.example/example-project/main/cli:latest",
							PromoteImage: "harbor.example/example-project/promote-main/cli:latest",
						},
					}, {
						Name: "nginx",
						ImageBuild: generator.ImageBuild{
							BuildImage:   "harbor.example/example-project/main/nginx:latest",
							PromoteImage: "harbor.example/example-project/promote-main/nginx:latest",
						},
					}, {
						Name: "php",
						ImageBuild: generator.ImageBuild{
							BuildImage:   "harbor.example/example-project/main/php:latest",
							PromoteImage: "harbor.example/example-project/promote-main/php:latest",
						},
					}, {
						Name: "redis",
						ImageBuild: generator.ImageBuild{
							BuildImage:   "harbor.example/example-project/main/redis:latest",
							PromoteImage: "harbor.example/example-project/promote-main/redis:latest",
						},
					}, {
						Name: "varnish",
						ImageBuild: generator.ImageBuild{
							BuildImage:   "harbor.example/example-project/main/varnish:latest",
							PromoteImage: "harbor.example/example-project/promote-main/varnish:latest",
						},
					},
				},
			},
		},
		{
			name: "test8 nginx-php external pull images",
			args: testdata.GetSeedData(
				testdata.TestData{
					Namespace:       "example-project-main",
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/complex/lagoon.varnish2.yml",
					ProjectVariables: []lagoon.EnvironmentVariable{
						{
							Name:  "LAGOON_FEATURE_FLAG_IMAGECACHE_REGISTRY",
							Value: "imagecache.example.com",
							Scope: "build",
						},
					},
				}, true),
			want: imageBuild{
				BuildKit: helpers.BoolPtr(true),
				BuildArguments: map[string]string{
					"LAGOON_BUILD_NAME":                       "lagoon-build-abcdefg",
					"LAGOON_PROJECT":                          "example-project",
					"LAGOON_ENVIRONMENT":                      "main",
					"LAGOON_ENVIRONMENT_TYPE":                 "production",
					"LAGOON_BUILD_TYPE":                       "branch",
					"LAGOON_GIT_SOURCE_REPOSITORY":            "ssh://git@example.com/lagoon-demo.git",
					"LAGOON_KUBERNETES":                       "remote-cluster1",
					"LAGOON_GIT_SHA":                          "0000000000000000000000000000000000000000",
					"LAGOON_GIT_BRANCH":                       "main",
					"CLI_IMAGE":                               "example-project-main-cli",
					"NGINX_IMAGE":                             "example-project-main-nginx",
					"PHP_IMAGE":                               "example-project-main-php",
					"LAGOON_FEATURE_FLAG_IMAGECACHE_REGISTRY": "imagecache.example.com",
				},
				ContainerRegistries: []generator.ContainerRegistry{
					{
						Name:           "my-custom-registry",
						Username:       "registry_user",
						Password:       "REGISTRY_PASSWORD",
						SecretName:     "lagoon-private-registry-my-custom-registry",
						URL:            "registry1.example.com",
						UsernameSource: ".lagoon.yml",
						PasswordSource: ".lagoon.yml (we recommend using an environment variable, see the docs on container-registries for more information)",
					},
				},
				Images: []imageBuilds{
					{
						Name: "cli",
						ImageBuild: generator.ImageBuild{
							BuildImage:     "harbor.example/example-project/main/cli:latest",
							Context:        "internal/testdata/complex/docker",
							DockerFile:     ".docker/Dockerfile.cli",
							TemporaryImage: "example-project-main-cli",
						},
					}, {
						Name: "nginx",
						ImageBuild: generator.ImageBuild{
							BuildImage:     "harbor.example/example-project/main/nginx:latest",
							Context:        "internal/testdata/complex/docker",
							DockerFile:     ".docker/Dockerfile.nginx-drupal",
							TemporaryImage: "example-project-main-nginx",
						},
					}, {
						Name: "php",
						ImageBuild: generator.ImageBuild{
							BuildImage:     "harbor.example/example-project/main/php:latest",
							Context:        "internal/testdata/complex/docker",
							DockerFile:     ".docker/Dockerfile.php",
							TemporaryImage: "example-project-main-php",
						},
					}, {
						Name: "redis",
						ImageBuild: generator.ImageBuild{
							BuildImage: "harbor.example/example-project/main/redis:latest",
							PullImage:  "registry1.example.com/amazeeio/redis",
						},
					}, {
						Name: "varnish",
						ImageBuild: generator.ImageBuild{
							BuildImage: "harbor.example/example-project/main/varnish:latest",
							PullImage:  "imagecache.example.com/uselagoon/varnish-5-drupal:latest",
						},
					},
				},
			},
		},
		{
			name: "test9 basic deployment with cache args",
			args: testdata.GetSeedData(
				testdata.TestData{
					Namespace:               "example-project-main",
					ProjectName:             "example-project",
					EnvironmentName:         "main",
					Branch:                  "main",
					LagoonYAML:              "internal/testdata/basic/lagoon.yml",
					ImageCacheBuildArgsJSON: `[{"image":"harbor.example/example-project/main/node@sha256:e90daba405cbf33bab23fe8a021146811b2c258df5f2afe7dadc92c0778eef45","name":"node"}]`,
				}, true),
			want: imageBuild{
				BuildKit: helpers.BoolPtr(true),
				BuildArguments: map[string]string{
					"LAGOON_BUILD_NAME":            "lagoon-build-abcdefg",
					"LAGOON_PROJECT":               "example-project",
					"LAGOON_ENVIRONMENT":           "main",
					"LAGOON_ENVIRONMENT_TYPE":      "production",
					"LAGOON_BUILD_TYPE":            "branch",
					"LAGOON_GIT_SOURCE_REPOSITORY": "ssh://git@example.com/lagoon-demo.git",
					"LAGOON_KUBERNETES":            "remote-cluster1",
					"LAGOON_GIT_SHA":               "abcdefg123456",
					"LAGOON_GIT_BRANCH":            "main",
					"NODE_IMAGE":                   "example-project-main-node",
					"LAGOON_CACHE_node":            "harbor.example/example-project/main/node@sha256:e90daba405cbf33bab23fe8a021146811b2c258df5f2afe7dadc92c0778eef45",
				},
				Images: []imageBuilds{
					{
						Name: "node",
						ImageBuild: generator.ImageBuild{
							BuildImage:     "harbor.example/example-project/main/node:latest",
							Context:        "internal/testdata/basic/docker",
							DockerFile:     "basic.dockerfile",
							TemporaryImage: "example-project-main-node",
						},
					},
				},
			},
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

			out, err := ImageBuildConfigurationIdentification(generator)
			if err != nil {
				t.Errorf("%v", err)
			}

			oJ, _ := json.Marshal(out)
			wJ, _ := json.Marshal(tt.want)
			if string(oJ) != string(wJ) {
				t.Errorf("returned output %v doesn't match want %v", string(oJ), string(wJ))
			}
			t.Cleanup(func() {
				helpers.UnsetEnvVars(tt.vars)
			})
		})
	}
}
