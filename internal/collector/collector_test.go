package collector

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/uselagoon/build-deploy-tool/internal/k8s"
	"sigs.k8s.io/yaml"
)

func TestCollector_Collect(t *testing.T) {
	type args struct {
		ctx       context.Context
		namespace string
	}
	tests := []struct {
		name string
		args args
		// want    *LagoonEnvState
		seedDir string
		want    string
		wantErr bool
	}{
		{
			name: "list-environment",
			args: args{
				ctx:       context.Background(),
				namespace: "example-project-main",
			},
			seedDir: "testdata/seed/seed-1",
			want:    "testdata/result/result-1",
			wantErr: false,
		},
		{
			name: "list-environment-pvc",
			args: args{
				ctx:       context.Background(),
				namespace: "example-project-main",
			},
			seedDir: "testdata/seed/seed-2",
			want:    "testdata/result/result-2",
			wantErr: false,
		},
		{
			name: "list-environment-netpol",
			args: args{
				ctx:       context.Background(),
				namespace: "example-project-main",
			},
			seedDir: "testdata/seed/seed-3",
			want:    "testdata/result/result-3",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := k8s.NewFakeClient(tt.args.namespace)
			if err != nil {
				t.Errorf("error creating fake client")
			}
			err = k8s.SeedFakeData(client, tt.args.namespace, tt.seedDir)
			if err != nil {
				t.Errorf("error seeding fake data: %v", err)
			}
			c := &Collector{
				Client: client,
			}
			got, err := c.Collect(tt.args.ctx, tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("Collector.Collect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got.Deployments.Items) > 0 {
				checkResult(t, fmt.Sprintf("%s/%s", tt.want, "lagoon-deployments.yaml"), got.Deployments)
			}
			if len(got.Cronjobs.Items) > 0 {
				checkResult(t, fmt.Sprintf("%s/%s", tt.want, "lagoon-cronjobs.yaml"), got.Cronjobs)
			}
			if len(got.Ingress.Items) > 0 {
				checkResult(t, fmt.Sprintf("%s/%s", tt.want, "lagoon-ingress.yaml"), got.Ingress)
			}
			if len(got.Services.Items) > 0 {
				checkResult(t, fmt.Sprintf("%s/%s", tt.want, "lagoon-services.yaml"), got.Services)
			}
			if len(got.Secrets.Items) > 0 {
				checkResult(t, fmt.Sprintf("%s/%s", tt.want, "lagoon-secrets.yaml"), got.Secrets)
			}
			if len(got.MariaDBConsumers.Items) > 0 {
				checkResult(t, fmt.Sprintf("%s/%s", tt.want, "lagoon-mariadb-consumers.yaml"), got.MariaDBConsumers)
			}
			if len(got.MongoDBConsumers.Items) > 0 {
				checkResult(t, fmt.Sprintf("%s/%s", tt.want, "lagoon-mongodb-consumers.yaml"), got.MongoDBConsumers)
			}
			if len(got.PostgreSQLConsumers.Items) > 0 {
				checkResult(t, fmt.Sprintf("%s/%s", tt.want, "lagoon-postgres-consumers.yaml"), got.PostgreSQLConsumers)
			}
			if len(got.SchedulesV1.Items) > 0 {
				checkResult(t, fmt.Sprintf("%s/%s", tt.want, "lagoon-schedules-v1.yaml"), got.SchedulesV1)
			}
			if len(got.SchedulesV1Alpha1.Items) > 0 {
				checkResult(t, fmt.Sprintf("%s/%s", tt.want, "lagoon-schedules-v1alpha1.yaml"), got.SchedulesV1Alpha1)
			}
			if len(got.PreBackupPodsV1.Items) > 0 {
				checkResult(t, fmt.Sprintf("%s/%s", tt.want, "lagoon-prebackuppods-v1.yaml"), got.PreBackupPodsV1)
			}
			if len(got.PreBackupPodsV1Alpha1.Items) > 0 {
				checkResult(t, fmt.Sprintf("%s/%s", tt.want, "lagoon-prebackuppods-v1alpha1.yaml"), got.PreBackupPodsV1Alpha1)
			}
			if len(got.NetworkPolicies.Items) > 0 {
				checkResult(t, fmt.Sprintf("%s/%s", tt.want, "lagoon-networkpolicies.yaml"), got.NetworkPolicies)
			}
		})
	}
}

func checkResult(t *testing.T, want string, got interface{}) {
	rBytes, err := yaml.Marshal(got)
	if err != nil {
		t.Errorf("couldn't marshal deployments: %v", err)
	}
	r1, err := os.ReadFile(want)
	if err != nil {
		// try create the file if it doesn't exist
		err := os.WriteFile(want, rBytes, 0644)
		if err != nil {
			t.Errorf("couldn't write file %v: %v", want, err)
		} else {
			t.Errorf("couldn't read file %v: %v", want, err)
		}
	}
	if !reflect.DeepEqual(rBytes, r1) {
		t.Errorf("Collect() = \n%v", diff.LineDiff(string(r1), string(rBytes)))
	}
}
