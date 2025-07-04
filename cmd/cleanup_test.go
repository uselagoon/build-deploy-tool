package cmd

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/uselagoon/build-deploy-tool/internal/collector"
	"github.com/uselagoon/build-deploy-tool/internal/dbaasclient"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/k8s"
	"github.com/uselagoon/build-deploy-tool/internal/testdata"

	// changes the testing to source from root so paths to test resources must be defined from repo root
	_ "github.com/uselagoon/build-deploy-tool/internal/testing"
)

func TestRunCleanup(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		args           testdata.TestData
		deleteServices bool
		wantErr        bool
		seedDir        string
		wantMariaDB    []string
		wantPsqlDB     []string
		wantMongoDB    []string
		wantDep        []string
		wantVol        []string
		wantServ       []string
	}{
		{
			name: "basic deployment",
			args: testdata.GetSeedData(
				testdata.TestData{
					ProjectName:     "example-project",
					EnvironmentName: "main",
					Branch:          "main",
					LagoonYAML:      "internal/testdata/basic/lagoon.yml",
					ImageReferences: map[string]string{
						"node": "harbor.example/example-project/main/node@sha256:b2001babafaa8128fe89aa8fd11832cade59931d14c3de5b3ca32e2a010fbaa8",
					},
				}, true),
			deleteServices: true,
			namespace:      "example-project-main",
			seedDir:        "internal/testdata/basic/cleanup-seed/basic-deployment",
			wantMariaDB:    []string{"mariadb"},
			wantDep:        []string{"basic"},
			wantServ:       []string{"basic"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.UnsetEnvVars(nil) //unset variables before running tests
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

			client, err := k8s.NewFakeClient(tt.namespace)
			if err != nil {
				t.Errorf("error creating fake client")
			}
			err = k8s.SeedFakeData(client, tt.namespace, tt.seedDir)
			if err != nil {
				t.Errorf("error seeding fake data: %v", err)
			}
			col := collector.NewCollector(client)
			beforeState, _ := col.Collect(context.Background(), tt.namespace)
			want := false
			for _, i2 := range tt.wantDep {
				for _, i1 := range beforeState.Deployments.Items {
					if i1.Name == i2 {
						want = true
					}
				}
				if !want {
					t.Errorf("RunCleanup() deployment %v should exist", i2)
				}
				want = false
			}
			for _, i2 := range tt.wantVol {
				for _, i1 := range beforeState.PVCs.Items {
					if i1.Name == i2 {
						want = true
					}
				}
				if !want {
					t.Errorf("RunCleanup() pvc %v should exist", i2)
				}
				want = false
			}
			for _, i2 := range tt.wantServ {
				for _, i1 := range beforeState.Services.Items {
					if i1.Name == i2 {
						want = true
					}
				}
				if !want {
					t.Errorf("RunCleanup() service %v should exist", i2)
				}
				want = false
			}
			for _, i2 := range tt.wantMariaDB {
				for _, i1 := range beforeState.MariaDBConsumers.Items {
					if i1.Name == i2 {
						want = true
					}
				}
				if !want {
					t.Errorf("RunCleanup() mariadb consumer %v should exist", i2)
				}
				want = false
			}
			for _, i2 := range tt.wantMongoDB {
				for _, i1 := range beforeState.MongoDBConsumers.Items {
					if i1.Name == i2 {
						want = true
					}
				}
				if !want {
					t.Errorf("RunCleanup() mongodb consumer %v should exist", i2)
				}
				want = false
			}
			for _, i2 := range tt.wantPsqlDB {
				for _, i1 := range beforeState.PostgreSQLConsumers.Items {
					if i1.Name == i2 {
						want = true
					}
				}
				if !want {
					t.Errorf("RunCleanup() mariadb consumer %v should exist", i2)
				}
				want = false
			}
			mdb, mongdb, psqdb, dep, vol, serv, err := RunCleanup(col, generator, tt.deleteServices)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunCleanup() error = %v, wantErr %v", err, tt.wantErr)
			}
			if mdb != nil && tt.wantMariaDB != nil && !reflect.DeepEqual(tt.wantMariaDB, mdb) {
				t.Errorf("RunCleanup() %v, wantMariaDB %v", mdb, tt.wantMariaDB)
			}
			if mongdb != nil && tt.wantMongoDB != nil && !reflect.DeepEqual(tt.wantMongoDB, mongdb) {
				t.Errorf("RunCleanup() %v, wantMongoDB %v", mongdb, tt.wantMongoDB)
			}
			if psqdb != nil && tt.wantPsqlDB != nil && !reflect.DeepEqual(tt.wantPsqlDB, psqdb) {
				t.Errorf("RunCleanup() %v, wantPsqlDB %v", psqdb, tt.wantPsqlDB)
			}
			if dep != nil && tt.wantDep != nil && !reflect.DeepEqual(tt.wantDep, dep) {
				t.Errorf("RunCleanup() %v, wantDep %v", dep, tt.wantDep)
			}
			if serv != nil && tt.wantServ != nil && !reflect.DeepEqual(tt.wantServ, serv) {
				t.Errorf("RunCleanup()%v, wantServ %v", serv, tt.wantServ)
			}
			if vol != nil && tt.wantVol != nil && !reflect.DeepEqual(tt.wantVol, vol) {
				t.Errorf("RunCleanup() %v, wantVol %v", vol, tt.wantVol)
			}
			afterState, _ := col.Collect(context.Background(), tt.namespace)
			for _, i1 := range afterState.Deployments.Items {
				for _, i2 := range dep {
					if i1.Name == i2 {
						t.Errorf("RunCleanup() deployment %v shouldn't exist", i2)
					}
				}
			}
			for _, i1 := range afterState.PVCs.Items {
				for _, i2 := range dep {
					if i1.Name == i2 {
						t.Errorf("RunCleanup() pvc %v shouldn't exist", i2)
					}
				}
			}
			for _, i1 := range afterState.Services.Items {
				for _, i2 := range dep {
					if i1.Name == i2 {
						t.Errorf("RunCleanup() service %v shouldn't exist", i2)
					}
				}
			}
			for _, i1 := range afterState.MariaDBConsumers.Items {
				for _, i2 := range dep {
					if i1.Name == i2 {
						t.Errorf("RunCleanup() mariadb consumer %v shouldn't exist", i2)
					}
				}
			}
			for _, i1 := range afterState.MongoDBConsumers.Items {
				for _, i2 := range dep {
					if i1.Name == i2 {
						t.Errorf("RunCleanup() mongodb consumer %v shouldn't exist", i2)
					}
				}
			}
			for _, i1 := range afterState.PostgreSQLConsumers.Items {
				for _, i2 := range dep {
					if i1.Name == i2 {
						t.Errorf("RunCleanup() postgres consumer %v shouldn't exist", i2)
					}
				}
			}
		})
	}
}
