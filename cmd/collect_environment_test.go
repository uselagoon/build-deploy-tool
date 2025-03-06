package cmd

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/uselagoon/build-deploy-tool/internal/collector"
	"github.com/uselagoon/build-deploy-tool/internal/k8s"

	// changes the testing to source from root so paths to test resources must be defined from repo root
	_ "github.com/uselagoon/build-deploy-tool/internal/testing"
)

func TestCollectEnvironment(t *testing.T) {
	type args struct {
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
			name: "list-environment",
			args: args{
				namespace: "example-project-main",
			},
			seedDir: "internal/collector/testdata/seed/seed-1",
			want:    "internal/collector/testdata/json-result/result-1.json",
			wantErr: false,
		},

		{
			name: "list-environment-with-pvcs",
			args: args{
				namespace: "example-project-main",
			},
			seedDir: "internal/collector/testdata/seed/seed-2",
			want:    "internal/collector/testdata/json-result/result-2.json",
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
			col := collector.NewCollector(client)
			got, err := CollectEnvironment(col, tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("CollectEnvironment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			results, err := os.ReadFile(tt.want)
			if err != nil {
				t.Errorf("couldn't read file %v: %v", tt.want, err)
			}
			env, err := json.MarshalIndent(got, "", "  ")
			if err != nil {
				t.Errorf("couldn't read file %v: %v", tt.want, err)
			}
			if !reflect.DeepEqual(string(results), string(env)) {
				t.Errorf("Collect() = \n%v", diff.LineDiff(string(env), string(results)))
			}
		})
	}
}
