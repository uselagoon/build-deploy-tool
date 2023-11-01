package generator

import (
	"encoding/json"
	"reflect"
	"testing"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
)

func Test_getResourcesFromAPIEnvVar(t *testing.T) {
	type args struct {
		envVars []lagoon.EnvironmentVariable
		debug   bool
	}
	tests := []struct {
		name    string
		args    args
		want    *map[string]ResourceWorkloads
		wantErr bool
	}{
		{
			name: "test1 - check that a scaling parameters are correctly defined",
			args: args{
				envVars: []lagoon.EnvironmentVariable{
					{
						Name:  "LAGOON_FEATURE_FLAG_WORKLOAD_RESOURCES",
						Value: helpers.ReadFileBase64Encode("test-resources/resources/test1-workload.json"),
						Scope: "global",
					},
				},
			},
			want: &map[string]ResourceWorkloads{
				"nginx": {
					ServiceType: "nginx",
					HPA: &HPASpec{
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
											AverageUtilization: helpers.Int32Ptr(3000),
										},
									},
								},
							},
						},
					},
					PDB: &PDBSpec{
						Spec: policyv1.PodDisruptionBudgetSpec{
							MinAvailable: &intstr.IntOrString{
								IntVal: 1,
								Type:   intstr.Int,
							},
						},
					},
					Resources: []Resource{
						{
							Name: "php",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("10m"),
									corev1.ResourceMemory: resource.MustParse("10Mi"),
								},
							},
						},
						{
							Name: "nginx",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("10m"),
									corev1.ResourceMemory: resource.MustParse("10Mi"),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "test2 - check that a scaling parameters are correctly defined for multiple services",
			args: args{
				envVars: []lagoon.EnvironmentVariable{
					{
						Name:  "LAGOON_FEATURE_FLAG_WORKLOAD_RESOURCES",
						Value: helpers.ReadFileBase64Encode("test-resources/resources/test2-workload.json"),
						Scope: "global",
					},
				},
			},
			want: &map[string]ResourceWorkloads{
				"nginx": {
					ServiceType: "nginx",
					HPA: &HPASpec{
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
											AverageUtilization: helpers.Int32Ptr(3000),
										},
									},
								},
							},
						},
					},
				},
				"node": {
					ServiceType: "node",
					HPA: &HPASpec{
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
											AverageUtilization: helpers.Int32Ptr(3000),
										},
									},
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
			got, err := getResourcesFromAPIEnvVar(tt.args.envVars, tt.args.debug)
			if (err != nil) != tt.wantErr {
				t.Errorf("getResourcesFromAPIEnvVar() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			lValues, _ := json.Marshal(got)
			wValues, _ := json.Marshal(tt.want)
			if !reflect.DeepEqual(string(lValues), string(wValues)) {
				t.Errorf("getResourcesFromAPIEnvVar() = %v, want %v", string(lValues), string(wValues))
			}
		})
	}
}
