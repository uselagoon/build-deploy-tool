package services

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/compose-spec/compose-go/types"
	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	"sigs.k8s.io/yaml"
)

func TestGenerateDeploymentTemplate(t *testing.T) {
	type args struct {
		buildValues generator.BuildValues
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test1 - basic",
			args: args{
				buildValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-name",
					EnvironmentType: "production",
					Namespace:       "myexample-project-environment-name",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "environment-name",
					GitSHA:          "0",
					ConfigMapSha:    "32bf1359ac92178c8909f0ef938257b477708aa0d78a5a15ad7c2d7919adf273",
					ImageReferences: map[string]string{
						"myservice":                "harbor.example.com/example-project/environment-name/myservice@latest",
						"myservice-po":             "harbor.example.com/example-project/environment-name/myservice-po@latest",
						"myservice-persist":        "harbor.example.com/example-project/environment-name/myservice-persist@latest",
						"myservice-persist-po":     "harbor.example.com/example-project/environment-name/myservice-persist-po@latest",
						"myservice-persist-posize": "harbor.example.com/example-project/environment-name/myservice-persist-posize@latest",
					},
					Services: []generator.ServiceValues{
						{
							Name:             "myservice",
							OverrideName:     "myservice",
							Type:             "basic",
							DBaaSEnvironment: "production",
							InPodCronjobs: []lagoon.Cronjob{
								{
									Name:     "cron - inpod",
									Schedule: "M/5 * * * *",
									Command:  "drush cron",
									Service:  "basic",
								},
								{
									Name:     "cron2 - inpod",
									Schedule: "M/15 * * * *",
									Command:  "other cronjob",
									Service:  "basic",
								},
							},
							NativeCronjobs: []lagoon.Cronjob{
								{
									Name:     "cron3 - native",
									Schedule: "35 * * * *",
									Command:  "drush cron",
									Service:  "basic",
								},
							},
						},
						{
							Name:             "myservice-po",
							OverrideName:     "myservice-po",
							Type:             "basic",
							DBaaSEnvironment: "production",
							ServicePort:      8080, // template should have port changed to 8080 from 3000
							UseSpotInstances: true, // template should have spot instance label and toleration/selector/affinity
							Replicas:         2,
						},
						{
							Name:                 "myservice-persist",
							OverrideName:         "myservice-persist",
							Type:                 "basic-persistent",
							DBaaSEnvironment:     "production",
							PersistentVolumePath: "/storage/data",
							PersistentVolumeName: "basic",
						},
						{
							Name:                 "myservice-persist-po",
							OverrideName:         "myservice-persist-po",
							Type:                 "basic-persistent",
							DBaaSEnvironment:     "production",
							ServicePort:          8080,
							PersistentVolumePath: "/storage/data",
							PersistentVolumeName: "basic",
						},
						{
							Name:                 "myservice-persist-posize",
							OverrideName:         "myservice-persist-posize",
							Type:                 "basic-persistent",
							DBaaSEnvironment:     "production",
							ServicePort:          8080,
							PersistentVolumeSize: "100Gi",
							PersistentVolumePath: "/storage/data",
							PersistentVolumeName: "basic",
						},
					},
				},
			},
			want: "test-resources/deployment/result-basic-1.yaml",
		},
		{
			name: "test2 - nginx-php",
			args: args{
				buildValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-name",
					EnvironmentType: "production",
					Namespace:       "myexample-project-environment-name",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "environment-name",
					FeatureFlags: map[string]bool{
						"rootlessworkloads": true,
					},
					PodSecurityContext: generator.PodSecurityContext{
						RunAsGroup: 0,
						RunAsUser:  10000,
						FsGroup:    10001,
					},
					GitSHA:       "0",
					ConfigMapSha: "32bf1359ac92178c8909f0ef938257b477708aa0d78a5a15ad7c2d7919adf273",
					ImageReferences: map[string]string{
						"nginx":  "harbor.example.com/example-project/environment-name/nginx@latest",
						"php":    "harbor.example.com/example-project/environment-name/php@latest",
						"nginx2": "harbor.example.com/example-project/environment-name/nginx2@latest",
						"php2":   "harbor.example.com/example-project/environment-name/php2@latest",
					},
					Services: []generator.ServiceValues{
						{
							Name:             "nginx",
							OverrideName:     "nginx",
							Type:             "nginx-php",
							DBaaSEnvironment: "production",
						},
						{
							Name:             "php",
							OverrideName:     "nginx",
							Type:             "nginx-php",
							DBaaSEnvironment: "production",
						},
						{
							Name:                 "nginx2",
							OverrideName:         "nginx-2",
							Type:                 "nginx-php-persistent",
							DBaaSEnvironment:     "production",
							PersistentVolumePath: "/storage/data",
							PersistentVolumeName: "nginx2",
						},
						{
							Name:                 "php2",
							OverrideName:         "nginx-2",
							Type:                 "nginx-php-persistent",
							DBaaSEnvironment:     "production",
							PersistentVolumeSize: "10Gi",
							PersistentVolumePath: "/storage/data",
							PersistentVolumeName: "nginx2",
						},
					},
				},
			},
			want: "test-resources/deployment/result-nginx-1.yaml",
		},
		{
			name: "test3 - cli",
			args: args{
				buildValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-name",
					EnvironmentType: "production",
					Namespace:       "myexample-project-environment-name",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "environment-name",
					GitSHA:          "0",
					ConfigMapSha:    "32bf1359ac92178c8909f0ef938257b477708aa0d78a5a15ad7c2d7919adf273",
					ImageReferences: map[string]string{
						"myservice":         "harbor.example.com/example-project/environment-name/myservice@latest",
						"myservice-persist": "harbor.example.com/example-project/environment-name/myservice-persist@latest",
					},
					Services: []generator.ServiceValues{
						{
							Name:             "myservice",
							OverrideName:     "myservice",
							Type:             "cli",
							DBaaSEnvironment: "production",
						},
						{
							Name:                 "myservice-persist",
							OverrideName:         "myservice-persist",
							Type:                 "cli-persistent",
							DBaaSEnvironment:     "production",
							PersistentVolumePath: "/storage/data",
							PersistentVolumeName: "nginx-php",
						},
					},
				},
			},
			want: "test-resources/deployment/result-cli-1.yaml",
		},
		{
			name: "test4 - postgres-single",
			args: args{
				buildValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-name",
					EnvironmentType: "production",
					Namespace:       "myexample-project-environment-name",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					PodSecurityContext: generator.PodSecurityContext{
						OnRootMismatch: true,
					},
					Branch:       "environment-name",
					GitSHA:       "0",
					ConfigMapSha: "32bf1359ac92178c8909f0ef938257b477708aa0d78a5a15ad7c2d7919adf273",
					ImageReferences: map[string]string{
						"myservice": "harbor.example.com/example-project/environment-name/myservice@latest",
					},
					Services: []generator.ServiceValues{
						{
							Name:             "myservice",
							OverrideName:     "myservice",
							Type:             "postgres-single",
							DBaaSEnvironment: "development",
						},
					},
				},
			},
			want: "test-resources/deployment/result-postgres-single-1.yaml",
		},
		{
			name: "test5 - elasticsearch",
			args: args{
				buildValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-name",
					EnvironmentType: "production",
					Namespace:       "myexample-project-environment-name",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					PodSecurityContext: generator.PodSecurityContext{
						OnRootMismatch: true,
					},
					Branch:       "environment-name",
					GitSHA:       "0",
					ConfigMapSha: "32bf1359ac92178c8909f0ef938257b477708aa0d78a5a15ad7c2d7919adf273",
					ImageReferences: map[string]string{
						"myservice":      "harbor.example.com/example-project/environment-name/myservice@latest",
						"myservice-size": "harbor.example.com/example-project/environment-name/myservice-size@latest",
					},
					Services: []generator.ServiceValues{
						{
							Name:                 "myservice",
							OverrideName:         "myservice",
							Type:                 "elasticsearch",
							DBaaSEnvironment:     "production",
							PersistentVolumeSize: "5Gi",
						},
						{
							Name:                 "myservice-size",
							OverrideName:         "myservice-size",
							Type:                 "elasticsearch",
							DBaaSEnvironment:     "production",
							PersistentVolumeSize: "100Gi",
						},
					},
				},
			},
			want: "test-resources/deployment/result-elasticsearch-1.yaml",
		},
		{
			name: "test6 - opensearch",
			args: args{
				buildValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-name",
					EnvironmentType: "production",
					Namespace:       "myexample-project-environment-name",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					PodSecurityContext: generator.PodSecurityContext{
						OnRootMismatch: true,
					},
					Branch:       "environment-name",
					GitSHA:       "0",
					ConfigMapSha: "32bf1359ac92178c8909f0ef938257b477708aa0d78a5a15ad7c2d7919adf273",
					ImageReferences: map[string]string{
						"myservice":      "harbor.example.com/example-project/environment-name/myservice@latest",
						"myservice-size": "harbor.example.com/example-project/environment-name/myservice-size@latest",
					},
					Services: []generator.ServiceValues{
						{
							Name:                 "myservice",
							OverrideName:         "myservice",
							Type:                 "opensearch",
							DBaaSEnvironment:     "production",
							PersistentVolumeSize: "5Gi",
						},
						{
							Name:                 "myservice-size",
							OverrideName:         "myservice-size",
							Type:                 "opensearch",
							DBaaSEnvironment:     "production",
							PersistentVolumeSize: "100Gi",
						},
					},
				},
			},
			want: "test-resources/deployment/result-opensearch-1.yaml",
		},
		{
			name: "test7 - basic compose ports",
			args: args{
				buildValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-name",
					EnvironmentType: "production",
					Namespace:       "myexample-project-environment-name",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "environment-name",
					GitSHA:          "0",
					ConfigMapSha:    "32bf1359ac92178c8909f0ef938257b477708aa0d78a5a15ad7c2d7919adf273",
					ImageReferences: map[string]string{
						"myservice-po": "harbor.example.com/example-project/environment-name/myservice-po@latest",
					},
					Services: []generator.ServiceValues{
						{
							Name:             "myservice-po",
							OverrideName:     "myservice-po",
							Type:             "basic",
							DBaaSEnvironment: "production",
							AdditionalServicePorts: []generator.AdditionalServicePort{
								{
									ServiceName: "myservice-po-8191",
									ServicePort: types.ServicePortConfig{
										Target:   8191,
										Protocol: "tcp",
									},
								},
								{
									ServiceName: "myservice-po-8211",
									ServicePort: types.ServicePortConfig{
										Target:   8211,
										Protocol: "tcp",
									},
								},
							},
							UseSpotInstances: true, // template should have spot instance label and toleration/selector/affinity
							Replicas:         2,
						},
					},
				},
			},
			want: "test-resources/deployment/result-basic-2.yaml",
		},
		{
			name: "test8 - solr",
			args: args{
				buildValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-name",
					EnvironmentType: "production",
					Namespace:       "myexample-project-environment-name",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					PodSecurityContext: generator.PodSecurityContext{
						OnRootMismatch: true,
					},
					Branch:       "environment-name",
					GitSHA:       "0",
					ConfigMapSha: "32bf1359ac92178c8909f0ef938257b477708aa0d78a5a15ad7c2d7919adf273",
					ImageReferences: map[string]string{
						"solr": "harbor.example.com/example-project/environment-name/solr@latest",
					},
					Services: []generator.ServiceValues{
						{
							Name:             "solr",
							OverrideName:     "solr",
							Type:             "solr",
							DBaaSEnvironment: "development",
						},
					},
				},
			},
			want: "test-resources/deployment/result-solr-1.yaml",
		},
		{
			name: "test9 - basic resources",
			args: args{
				buildValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-name",
					EnvironmentType: "production",
					Namespace:       "myexample-project-environment-name",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "environment-name",
					Resources: generator.Resources{
						Limits: generator.ResourceLimits{
							Memory:           "16Gi",
							EphemeralStorage: "160Gi",
						},
						Requests: generator.ResourceRequests{
							EphemeralStorage: "1Gi",
						},
					},
					GitSHA:       "0",
					ConfigMapSha: "32bf1359ac92178c8909f0ef938257b477708aa0d78a5a15ad7c2d7919adf273",
					ImageReferences: map[string]string{
						"myservice-po": "harbor.example.com/example-project/environment-name/myservice-po@latest",
					},
					Services: []generator.ServiceValues{
						{
							Name:             "myservice-po",
							OverrideName:     "myservice-po",
							Type:             "basic",
							DBaaSEnvironment: "production",
							AdditionalServicePorts: []generator.AdditionalServicePort{
								{
									ServiceName: "myservice-po-8191",
									ServicePort: types.ServicePortConfig{
										Target:   8191,
										Protocol: "tcp",
									},
								},
								{
									ServiceName: "myservice-po-8211",
									ServicePort: types.ServicePortConfig{
										Target:   8211,
										Protocol: "tcp",
									},
								},
							},
							UseSpotInstances: true, // template should have spot instance label and toleration/selector/affinity
							Replicas:         2,
						},
					},
				},
			},
			want: "test-resources/deployment/result-basic-3.yaml",
		},
		{
			name: "test10 - nginx-php with imagecache override",
			args: args{
				buildValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-name",
					EnvironmentType: "production",
					Namespace:       "myexample-project-environment-name",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "environment-name",
					ImageCache:      "imagecache.example.com/",
					FeatureFlags: map[string]bool{
						"rootlessworkloads": true,
					},
					PodSecurityContext: generator.PodSecurityContext{
						RunAsGroup:     0,
						RunAsUser:      10000,
						FsGroup:        10001,
						OnRootMismatch: true,
					},
					GitSHA:       "0",
					ConfigMapSha: "32bf1359ac92178c8909f0ef938257b477708aa0d78a5a15ad7c2d7919adf273",
					ImageReferences: map[string]string{
						"nginx":  "harbor.example.com/example-project/environment-name/nginx@latest",
						"php":    "harbor.example.com/example-project/environment-name/php@latest",
						"nginx2": "harbor.example.com/example-project/environment-name/nginx2@latest",
						"php2":   "harbor.example.com/example-project/environment-name/php2@latest",
					},
					Services: []generator.ServiceValues{
						{
							Name:             "nginx",
							OverrideName:     "nginx",
							Type:             "nginx-php",
							DBaaSEnvironment: "production",
						},
						{
							Name:             "php",
							OverrideName:     "nginx",
							Type:             "nginx-php",
							DBaaSEnvironment: "production",
						},
						{
							Name:                 "nginx2",
							OverrideName:         "nginx-2",
							Type:                 "nginx-php-persistent",
							DBaaSEnvironment:     "production",
							PersistentVolumePath: "/storage/data",
							PersistentVolumeName: "nginx2",
						},
						{
							Name:                 "php2",
							OverrideName:         "nginx-2",
							Type:                 "nginx-php-persistent",
							DBaaSEnvironment:     "production",
							PersistentVolumeSize: "10Gi",
							PersistentVolumePath: "/storage/data",
							PersistentVolumeName: "nginx2",
						},
					},
				},
			},
			want: "test-resources/deployment/result-nginx-2.yaml",
		},
		{
			name: "test12 - worker",
			args: args{
				buildValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-name",
					EnvironmentType: "production",
					Namespace:       "example-project-environment-name",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "environment-name",
					PodSecurityContext: generator.PodSecurityContext{
						RunAsGroup:     0,
						RunAsUser:      10000,
						FsGroup:        10001,
						OnRootMismatch: true,
					},
					GitSHA:       "0",
					ConfigMapSha: "32bf1359ac92178c8909f0ef938257b477708aa0d78a5a15ad7c2d7919adf273",
					ImageReferences: map[string]string{
						"worker":         "harbor.example.com/example-project/environment-name/worker@latest",
						"worker-persist": "harbor.example.com/example-project/environment-name/worker-persist@latest",
					},
					Services: []generator.ServiceValues{
						{
							Name:         "worker",
							OverrideName: "worker",
							Type:         "worker",
						},
						{
							Name:         "worker-persist",
							OverrideName: "worker-persist",
							Type:         "worker-persistent",
						},
					},
				},
			},
			want: "test-resources/deployment/result-worker-1.yaml",
		},
		{
			name: "test13 - python",
			args: args{
				buildValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-name",
					EnvironmentType: "production",
					Namespace:       "example-project-environment-name",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "environment-name",
					PodSecurityContext: generator.PodSecurityContext{
						RunAsGroup:     0,
						RunAsUser:      10000,
						FsGroup:        10001,
						OnRootMismatch: true,
					},
					GitSHA:       "0",
					ConfigMapSha: "32bf1359ac92178c8909f0ef938257b477708aa0d78a5a15ad7c2d7919adf273",
					ImageReferences: map[string]string{
						"python":         "harbor.example.com/example-project/environment-name/python@latest",
						"python-persist": "harbor.example.com/example-project/environment-name/python-persist@latest",
					},
					Services: []generator.ServiceValues{
						{
							Name:         "python",
							OverrideName: "python",
							Type:         "python",
						},
						{
							Name:         "python-persist",
							OverrideName: "python-persist",
							Type:         "python-persistent",
						},
					},
				},
			},
			want: "test-resources/deployment/result-python-1.yaml",
		},
		{
			name: "test14 - node",
			args: args{
				buildValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-name",
					EnvironmentType: "production",
					Namespace:       "example-project-environment-name",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "environment-name",
					PodSecurityContext: generator.PodSecurityContext{
						RunAsGroup:     0,
						RunAsUser:      10000,
						FsGroup:        10001,
						OnRootMismatch: true,
					},
					GitSHA:       "0",
					ConfigMapSha: "32bf1359ac92178c8909f0ef938257b477708aa0d78a5a15ad7c2d7919adf273",
					ImageReferences: map[string]string{
						"node":         "harbor.example.com/example-project/environment-name/node@latest",
						"node-persist": "harbor.example.com/example-project/environment-name/node-persist@latest",
					},
					Services: []generator.ServiceValues{
						{
							Name:         "node",
							OverrideName: "node",
							Type:         "node",
						},
						{
							Name:         "node-persist",
							OverrideName: "node-persist",
							Type:         "node-persistent",
						},
					},
				},
			},
			want: "test-resources/deployment/result-node-1.yaml",
		},
		{
			name: "test15 - varnish",
			args: args{
				buildValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-name",
					EnvironmentType: "production",
					Namespace:       "example-project-environment-name",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "environment-name",
					PodSecurityContext: generator.PodSecurityContext{
						RunAsGroup:     0,
						RunAsUser:      10000,
						FsGroup:        10001,
						OnRootMismatch: true,
					},
					GitSHA:       "0",
					ConfigMapSha: "32bf1359ac92178c8909f0ef938257b477708aa0d78a5a15ad7c2d7919adf273",
					ImageReferences: map[string]string{
						"varnish":         "harbor.example.com/example-project/environment-name/varnish@latest",
						"varnish-persist": "harbor.example.com/example-project/environment-name/varnish-persist@latest",
					},
					Services: []generator.ServiceValues{
						{
							Name:         "varnish",
							OverrideName: "varnish",
							Type:         "varnish",
						},
						{
							Name:         "varnish-persist",
							OverrideName: "varnish-persist",
							Type:         "varnish-persistent",
						},
					},
				},
			},
			want: "test-resources/deployment/result-varnish-1.yaml",
		},
		{
			name: "test16 - redis",
			args: args{
				buildValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-name",
					EnvironmentType: "production",
					Namespace:       "example-project-environment-name",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "environment-name",
					PodSecurityContext: generator.PodSecurityContext{
						RunAsGroup:     0,
						RunAsUser:      10000,
						FsGroup:        10001,
						OnRootMismatch: true,
					},
					GitSHA:       "0",
					ConfigMapSha: "32bf1359ac92178c8909f0ef938257b477708aa0d78a5a15ad7c2d7919adf273",
					ImageReferences: map[string]string{
						"redis":         "harbor.example.com/example-project/environment-name/redis@latest",
						"redis-persist": "harbor.example.com/example-project/environment-name/redis-persist@latest",
					},
					Services: []generator.ServiceValues{
						{
							Name:         "redis",
							OverrideName: "redis",
							Type:         "redis",
						},
						{
							Name:         "redis-persist",
							OverrideName: "redis-persist",
							Type:         "redis-persistent",
						},
					},
				},
			},
			want: "test-resources/deployment/result-redis-1.yaml",
		},
		{
			name: "test17a - mariadb",
			args: args{
				buildValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-name",
					EnvironmentType: "production",
					Namespace:       "example-project-environment-name",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "environment-name",
					PodSecurityContext: generator.PodSecurityContext{
						OnRootMismatch: true,
					},
					GitSHA:       "0",
					ConfigMapSha: "32bf1359ac92178c8909f0ef938257b477708aa0d78a5a15ad7c2d7919adf273",
					ImageReferences: map[string]string{
						"mariadb": "harbor.example.com/example-project/environment-name/mariadb@latest",
					},
					Services: []generator.ServiceValues{
						{
							Name:         "mariadb",
							OverrideName: "mariadb",
							Type:         "mariadb-single",
						},
					},
				},
			},
			want: "test-resources/deployment/result-mariadb-1.yaml",
		},
		{
			name: "test17b - mariadb k8upv2",
			args: args{
				buildValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-name",
					EnvironmentType: "production",
					Namespace:       "example-project-environment-name",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "environment-name",
					PodSecurityContext: generator.PodSecurityContext{
						OnRootMismatch: true,
					},
					GitSHA:       "0",
					ConfigMapSha: "32bf1359ac92178c8909f0ef938257b477708aa0d78a5a15ad7c2d7919adf273",
					ImageReferences: map[string]string{
						"mariadb": "harbor.example.com/example-project/environment-name/mariadb@latest",
					},
					Backup: generator.BackupConfiguration{
						K8upVersion: "v2",
					},
					Services: []generator.ServiceValues{
						{
							Name:         "mariadb",
							OverrideName: "mariadb",
							Type:         "mariadb-single",
						},
					},
				},
			},
			want: "test-resources/deployment/result-mariadb-2.yaml",
		},
		{
			name: "test18 - mongodb",
			args: args{
				buildValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-name",
					EnvironmentType: "production",
					Namespace:       "example-project-environment-name",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "environment-name",
					PodSecurityContext: generator.PodSecurityContext{
						OnRootMismatch: true,
					},
					GitSHA:       "0",
					ConfigMapSha: "32bf1359ac92178c8909f0ef938257b477708aa0d78a5a15ad7c2d7919adf273",
					ImageReferences: map[string]string{
						"mongodb": "harbor.example.com/example-project/environment-name/mongodb@latest",
					},
					Services: []generator.ServiceValues{
						{
							Name:         "mongodb",
							OverrideName: "mongodb",
							Type:         "mongodb-single",
						},
					},
				},
			},
			want: "test-resources/deployment/result-mongodb-1.yaml",
		},
		{
			name: "test19 - postgres",
			args: args{
				buildValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-name",
					EnvironmentType: "production",
					Namespace:       "example-project-environment-name",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "environment-name",
					PodSecurityContext: generator.PodSecurityContext{
						OnRootMismatch: true,
					},
					GitSHA:       "0",
					ConfigMapSha: "32bf1359ac92178c8909f0ef938257b477708aa0d78a5a15ad7c2d7919adf273",
					ImageReferences: map[string]string{
						"postgres": "harbor.example.com/example-project/environment-name/postgres@latest",
					},
					Services: []generator.ServiceValues{
						{
							Name:         "postgres",
							OverrideName: "postgres",
							Type:         "postgres-single",
						},
					},
				},
			},
			want: "test-resources/deployment/result-postgres-1.yaml",
		},
		{
			name: "test-basic-antiaffinity",
			args: args{
				buildValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-name",
					EnvironmentType: "production",
					Namespace:       "myexample-project-environment-name",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "environment-name",
					Resources: generator.Resources{
						Limits: generator.ResourceLimits{
							Memory:           "16Gi",
							EphemeralStorage: "160Gi",
						},
						Requests: generator.ResourceRequests{
							EphemeralStorage: "1Gi",
						},
					},
					GitSHA:       "0",
					ConfigMapSha: "32bf1359ac92178c8909f0ef938257b477708aa0d78a5a15ad7c2d7919adf273",
					ImageReferences: map[string]string{
						"myservice": "harbor.example.com/example-project/environment-name/myservice@latest",
					},
					PodAntiAffinity: true,
					Services: []generator.ServiceValues{
						{
							Name:             "myservice",
							OverrideName:     "myservice",
							Type:             "basic",
							DBaaSEnvironment: "production",
							AdditionalServicePorts: []generator.AdditionalServicePort{
								{
									ServiceName: "myservice-8191",
									ServicePort: types.ServicePortConfig{
										Target:   8191,
										Protocol: "tcp",
									},
								},
								{
									ServiceName: "myservice-8211",
									ServicePort: types.ServicePortConfig{
										Target:   8211,
										Protocol: "tcp",
									},
								},
							},
							Replicas: 2,
						},
					},
				},
			},
			want: "test-resources/deployment/result-basic-4.yaml",
		},
		{
			name: "test20 - nginx-php ServiceValues Resource Override",
			args: args{
				buildValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "environment-name",
					EnvironmentType: "production",
					Namespace:       "myexample-project-environment-name",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "environment-name",
					FeatureFlags: map[string]bool{
						"rootlessworkloads": true,
					},
					PodSecurityContext: generator.PodSecurityContext{
						RunAsGroup: 0,
						RunAsUser:  10000,
						FsGroup:    10001,
					},
					GitSHA:       "0",
					ConfigMapSha: "32bf1359ac92178c8909f0ef938257b477708aa0d78a5a15ad7c2d7919adf273",
					ImageReferences: map[string]string{
						"nginx": "harbor.example.com/example-project/environment-name/nginx@latest",
						"php":   "harbor.example.com/example-project/environment-name/php@latest",
					},
					// BuildValues.Resouces not expected to be used as is overriden by ServiceValues resource
					// (Included to verify ServiceValues is final word)
					Resources: generator.Resources{Limits: generator.ResourceLimits{Memory: "2Gi"}},
					Services: []generator.ServiceValues{
						{
							Name:             "nginx",
							OverrideName:     "nginx",
							Type:             "nginx-php",
							DBaaSEnvironment: "production",
							// Leave primary mem request as default, ensure that default from ServiceType is still templated
							Resources: generator.Resources{
								Requests: generator.ResourceRequests{Cpu: "500m"},
								Limits: generator.ResourceLimits{
									Cpu:    "2",
									Memory: "1Gi",
								},
							},
						},
						{
							Name:             "php",
							OverrideName:     "nginx",
							Type:             "nginx-php",
							DBaaSEnvironment: "production",
							Resources: generator.Resources{
								Requests: generator.ResourceRequests{
									Cpu:    "500m",
									Memory: "200Mi",
								},
								Limits: generator.ResourceLimits{
									Cpu:    "500m",
									Memory: "1Gi",
								},
							},
						},
					},
				},
			},
			want: "test-resources/deployment/result-nginx-php-resources-1.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateDeploymentTemplate(tt.args.buildValues)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateDeploymentTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			r1, err := os.ReadFile(tt.want)
			if err != nil {
				t.Errorf("couldn't read file %v: %v", tt.want, err)
			}
			separator := []byte("---\n")
			var result []byte
			for _, d := range got {
				deploymentBytes, err := yaml.Marshal(d)
				if err != nil {
					t.Errorf("couldn't generate template  %v", err)
				}
				restoreResult := append(separator[:], deploymentBytes[:]...)
				result = append(result, restoreResult[:]...)
			}
			if !reflect.DeepEqual(string(result), string(r1)) {
				t.Errorf("GenerateDeploymentTemplate() = \n%v", diff.LineDiff(string(r1), string(result)))
			}
		})
	}
}

func TestLinkedServiceCalculator(t *testing.T) {
	type args struct {
		services []generator.ServiceValues
	}
	tests := []struct {
		name string
		args args
		want []generator.ServiceValues
	}{
		{
			name: "test1 - standard nginx-php",
			args: args{
				services: []generator.ServiceValues{
					{
						Name:             "nginx",
						OverrideName:     "nginx",
						Type:             "nginx-php",
						DBaaSEnvironment: "production",
					},
					{
						Name:             "php",
						OverrideName:     "nginx",
						Type:             "nginx-php",
						DBaaSEnvironment: "production",
					},
				},
			},
			want: []generator.ServiceValues{
				{
					Name:             "nginx",
					OverrideName:     "nginx",
					Type:             "nginx-php",
					DBaaSEnvironment: "production",
					LinkedService: &generator.ServiceValues{
						Name:             "php",
						OverrideName:     "nginx",
						Type:             "nginx-php",
						DBaaSEnvironment: "production",
					},
				},
			},
		},
		{
			name: "test2 - multiple linked services (2 separate nginx-php)",
			args: args{
				services: []generator.ServiceValues{
					{
						Name:             "nginx",
						OverrideName:     "nginx",
						Type:             "nginx-php",
						DBaaSEnvironment: "production",
					},
					{
						Name:             "php",
						OverrideName:     "nginx",
						Type:             "nginx-php",
						DBaaSEnvironment: "production",
					},
					{
						Name:             "nginx2",
						OverrideName:     "nginx2",
						Type:             "nginx-php-persistent",
						DBaaSEnvironment: "production",
					},
					{
						Name:             "php2",
						OverrideName:     "nginx2",
						Type:             "nginx-php-persistent",
						DBaaSEnvironment: "production",
					},
				},
			},
			want: []generator.ServiceValues{
				{
					Name:             "nginx",
					OverrideName:     "nginx",
					Type:             "nginx-php",
					DBaaSEnvironment: "production",
					LinkedService: &generator.ServiceValues{
						Name:             "php",
						OverrideName:     "nginx",
						Type:             "nginx-php",
						DBaaSEnvironment: "production",
					},
				},
				{
					Name:             "nginx2",
					OverrideName:     "nginx2",
					Type:             "nginx-php-persistent",
					DBaaSEnvironment: "production",
					LinkedService: &generator.ServiceValues{
						Name:             "php2",
						OverrideName:     "nginx2",
						Type:             "nginx-php-persistent",
						DBaaSEnvironment: "production",
					},
				},
			},
		},
		{
			name: "test3 - single nginx-php and a single nginx standalone",
			args: args{
				services: []generator.ServiceValues{
					{
						Name:             "nginx",
						OverrideName:     "nginx",
						Type:             "nginx-php",
						DBaaSEnvironment: "production",
					},
					{
						Name:             "php",
						OverrideName:     "nginx",
						Type:             "nginx-php",
						DBaaSEnvironment: "production",
					},
					{
						Name:             "normalnginx",
						OverrideName:     "normalnginx",
						Type:             "nginx",
						DBaaSEnvironment: "production",
					},
				},
			},
			want: []generator.ServiceValues{
				{
					Name:             "normalnginx",
					OverrideName:     "normalnginx",
					Type:             "nginx",
					DBaaSEnvironment: "production",
				},
				{
					Name:             "nginx",
					OverrideName:     "nginx",
					Type:             "nginx-php",
					DBaaSEnvironment: "production",
					LinkedService: &generator.ServiceValues{
						Name:             "php",
						OverrideName:     "nginx",
						Type:             "nginx-php",
						DBaaSEnvironment: "production",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LinkedServiceCalculator(tt.args.services)
			lValues, _ := json.Marshal(got)
			wValues, _ := json.Marshal(tt.want)
			if !reflect.DeepEqual(string(lValues), string(wValues)) {
				t.Errorf("LinkedServiceCalculator() = %v, want %v", string(lValues), string(wValues))
			}
		})
	}
}
