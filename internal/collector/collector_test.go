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
			want:    "testdata/result/result-1/",
			wantErr: false,
		},
		{
			name: "list-environment-pvc",
			args: args{
				ctx:       context.Background(),
				namespace: "example-project-main",
			},
			seedDir: "testdata/seed/seed-2",
			want:    "testdata/result/result-2/",
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
			results, err := os.ReadDir(tt.want)
			if err != nil {
				t.Errorf("couldn't read directory %v: %v", tt.want, err)
			}
			for _, r := range results {
				var rBytes []byte
				switch r.Name() {
				case "lagoon-deployments.yaml":
					rBytes, err = yaml.Marshal(got.Deployments)
					if err != nil {
						t.Errorf("couldn't marshal deployments: %v", err)
					}
				case "lagoon-cronjobs.yaml":
					rBytes, err = yaml.Marshal(got.Cronjobs)
					if err != nil {
						t.Errorf("couldn't marshal cronjobs: %v", err)
					}
				case "lagoon-ingress.yaml":
					rBytes, err = yaml.Marshal(got.Ingress)
					if err != nil {
						t.Errorf("couldn't marshal ingress: %v", err)
					}
				case "lagoon-services.yaml":
					rBytes, err = yaml.Marshal(got.Services)
					if err != nil {
						t.Errorf("couldn't marshal services: %v", err)
					}
				case "lagoon-secrets.yaml":
					rBytes, err = yaml.Marshal(got.Secrets)
					if err != nil {
						t.Errorf("couldn't marshal secrets: %v", err)
					}
				case "lagoon-mariadb-consumers.yaml":
					rBytes, err = yaml.Marshal(got.MariaDBConsumers)
					if err != nil {
						t.Errorf("couldn't marshal mariadb-consumers: %v", err)
					}
				case "lagoon-mongodb-consumers.yaml":
					rBytes, err = yaml.Marshal(got.MongoDBConsumers)
					if err != nil {
						t.Errorf("couldn't marshal mongodb-consumers: %v", err)
					}
				case "lagoon-postgres-consumers.yaml":
					rBytes, err = yaml.Marshal(got.PostgreSQLConsumers)
					if err != nil {
						t.Errorf("couldn't marshal postgres-consumers: %v", err)
					}
				case "lagoon-schedules-v1.yaml":
					rBytes, err = yaml.Marshal(got.SchedulesV1)
					if err != nil {
						t.Errorf("couldn't marshal schedules-v1: %v", err)
					}
				case "lagoon-schedules-v1alpha1.yaml":
					rBytes, err = yaml.Marshal(got.SchedulesV1Alpha1)
					if err != nil {
						t.Errorf("couldn't marshal schedules-v1alpha1: %v", err)
					}
				case "lagoon-prebackuppods-v1.yaml":
					rBytes, err = yaml.Marshal(got.PreBackupPodsV1)
					if err != nil {
						t.Errorf("couldn't marshal prebackuppods-v1: %v", err)
					}
				case "lagoon-prebackuppods-v1alpha1.yaml":
					rBytes, err = yaml.Marshal(got.PreBackupPodsV1Alpha1)
					if err != nil {
						t.Errorf("couldn't marshal prebackuppods-v1alpha1: %v", err)
					}
				default:
					continue
				}
				r1, err := os.ReadFile(fmt.Sprintf("%s/%s", tt.want, r.Name()))
				if err != nil {
					t.Errorf("couldn't read file %v: %v", fmt.Sprintf("%s/%s", tt.want, r.Name()), err)
				}
				if !reflect.DeepEqual(rBytes, r1) {
					t.Errorf("Collect() = \n%v", diff.LineDiff(string(r1), string(rBytes)))
				}
			}
		})
	}
}
