package hpa

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/uselagoon/build-deploy-tool/internal/dbaasclient"
	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
)

func TestGenerateHPATemplate(t *testing.T) {
	type args struct {
		lValues generator.BuildValues
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test1 - nginx hpa",
			args: args{
				lValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "brancha",
					EnvironmentType: "production",
					Namespace:       "myexample-project-brancha",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "brancha",
					Services: []generator.ServiceValues{
						{
							Name:             "nginx",
							OverrideName:     "nginx",
							Type:             "nginx-php-persistent",
							ResourceWorkload: "nginx-php-performance",
						},
						{
							Name:         "php",
							OverrideName: "nginx",
							Type:         "nginx-php-persistent",
						},
					},
					ResourceWorkloads: map[string]generator.ResourceWorkloads{
						"nginx": {
							ServiceType: "nginx",
							HPA: &generator.HPASpec{
								Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
									MinReplicas: helpers.Int32Ptr(2),
									MaxReplicas: *helpers.Int32Ptr(5),
									Metrics: []autoscalingv2.MetricSpec{
										{
											Type: autoscalingv2.ResourceMetricSourceType,
											Resource: &autoscalingv2.ResourceMetricSource{
												Name: corev1.ResourceCPU,
												Target: autoscalingv2.MetricTarget{
													Type:               autoscalingv2.UtilizationMetricType,
													AverageUtilization: helpers.Int32Ptr(3000),
												},
											},
										},
									}},
							},
						},
						"nginx-php-performance": {
							ServiceType: "nginx-php-persistent",
							HPA: &generator.HPASpec{
								Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
									MinReplicas: helpers.Int32Ptr(8),
									MaxReplicas: *helpers.Int32Ptr(16),
									Metrics: []autoscalingv2.MetricSpec{
										{
											Type: autoscalingv2.ResourceMetricSourceType,
											Resource: &autoscalingv2.ResourceMetricSource{
												Name: corev1.ResourceCPU,
												Target: autoscalingv2.MetricTarget{
													Type:               autoscalingv2.UtilizationMetricType,
													AverageUtilization: helpers.Int32Ptr(1500),
												},
											},
										},
									}},
							},
						},
					},
				},
			},
			want: "test-resources/result-nginx.yaml",
		},
		{
			name: "test2 - no resources",
			args: args{
				lValues: generator.BuildValues{
					Project:         "example-project",
					Environment:     "brancha",
					EnvironmentType: "production",
					Namespace:       "myexample-project-brancha",
					BuildType:       "branch",
					LagoonVersion:   "v2.x.x",
					Kubernetes:      "generator.local",
					Branch:          "brancha",
					Services: []generator.ServiceValues{
						{
							Name:         "nginx",
							OverrideName: "nginx",
							Type:         "nginx-php-persistent",
						},
						{
							Name:         "php",
							OverrideName: "nginx",
							Type:         "nginx-php-persistent",
						},
					},
					ResourceWorkloads: map[string]generator.ResourceWorkloads{},
				},
			},
			want: "test-resources/result-no-resources.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// add dbaasclient overrides for tests
			tt.args.lValues.DBaaSClient = dbaasclient.NewClient(dbaasclient.Client{
				RetryMax:     5,
				RetryWaitMin: time.Duration(10) * time.Millisecond,
				RetryWaitMax: time.Duration(50) * time.Millisecond,
			})
			got, err := GenerateHPATemplate(tt.args.lValues)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateHPATemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			r1, err := os.ReadFile(tt.want)
			if err != nil {
				t.Errorf("couldn't read file %v: %v", tt.want, err)
			}
			if !reflect.DeepEqual(string(got), string(r1)) {
				t.Errorf("GenerateHPATemplate() = %v, want %v", string(got), string(r1))
			}
		})
	}
}
