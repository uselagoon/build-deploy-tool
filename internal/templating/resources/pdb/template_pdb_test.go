package pdb

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/uselagoon/build-deploy-tool/internal/dbaasclient"
	"github.com/uselagoon/build-deploy-tool/internal/generator"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestGeneratePDBTemplate(t *testing.T) {
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
			name: "test1 - nginx pdb",
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
					},
					ResourceWorkloads: map[string]generator.ResourceWorkloads{
						"nginx": {
							ServiceType: "nginx",
							PDB: &generator.PDBSpec{
								Spec: policyv1.PodDisruptionBudgetSpec{
									MinAvailable: &intstr.IntOrString{
										IntVal: 1,
										Type:   intstr.Int,
									},
								},
							},
						},
						"nginx-php-performance": {
							ServiceType: "nginx-php-persistent",

							PDB: &generator.PDBSpec{
								Spec: policyv1.PodDisruptionBudgetSpec{
									MinAvailable: &intstr.IntOrString{
										IntVal: 3,
										Type:   intstr.Int,
									},
								},
							},
						},
					},
				},
			},
			want: "test-resources/result-nginx.yaml",
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
			got, err := GeneratePDBTemplate(tt.args.lValues)
			if (err != nil) != tt.wantErr {
				t.Errorf("GeneratePDBTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			r1, err := os.ReadFile(tt.want)
			if err != nil {
				t.Errorf("couldn't read file %v: %v", tt.want, err)
			}
			if !reflect.DeepEqual(string(got), string(r1)) {
				t.Errorf("GeneratePDBTemplate() = %v, want %v", string(got), string(r1))
			}
		})
	}
}
