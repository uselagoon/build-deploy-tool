package collector

import (
	"context"
	"os"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/uselagoon/build-deploy-tool/internal/k8s"
	"sigs.k8s.io/yaml"
)

func TestCollector_CollectMariaDBConsumers(t *testing.T) {
	type args struct {
		ctx       context.Context
		namespace string
	}
	tests := []struct {
		name    string
		args    args
		seedDir string
		want    string
		wantErr bool
	}{
		{
			name: "new-environment",
			args: args{
				ctx:       context.Background(),
				namespace: "example-project-main",
			},
			seedDir: "testdata/seed/seed-empty",
			want:    "testdata/result/result-empty/lagoon-mariadb-consumers.yaml",
			wantErr: false,
		},
		{
			name: "list-mariadb-consumers",
			args: args{
				ctx:       context.Background(),
				namespace: "example-project-main",
			},
			seedDir: "testdata/seed/seed-1",
			want:    "testdata/result/result-1/lagoon-mariadb-consumers.yaml",
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
			got, err := c.CollectMariaDBConsumers(tt.args.ctx, tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("Collector.CollectMariaDBConsumers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			oJ, _ := yaml.Marshal(got)
			results, err := os.ReadFile(tt.want)
			if err != nil {
				// try create the file if it doesn't exist
				err := os.WriteFile(tt.want, oJ, 0644)
				if err != nil {
					t.Errorf("couldn't write file %v: %v", tt.want, err)
				} else {
					t.Errorf("couldn't read file %v: %v", tt.want, err)
				}
			}
			if string(oJ) != string(results) {
				t.Errorf("Collector.CollectMariaDBConsumers() = \n%v", diff.LineDiff(string(results), string(oJ)))
			}
		})
	}
}
