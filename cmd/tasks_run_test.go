package cmd

import (
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	"github.com/uselagoon/build-deploy-tool/internal/tasklib"
	"os"
	"reflect"
	"testing"
)

func Test_evaluateWhenConditionsForTaskInEnvironment(t *testing.T) {
	type args struct {
		environment tasklib.TaskEnvironment
		task        lagoon.Task
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Successful evaluation",
			args: args{
				environment: map[string]interface{}{},
				task: lagoon.Task{
					When: "true",
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Returns non-bool int",
			args: args{
				environment: map[string]interface{}{},
				task: lagoon.Task{
					When: "5",
				},
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "Syntax error",
			args: args{
				environment: map[string]interface{}{},
				task: lagoon.Task{
					When: "7+)3==2",
				},
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := evaluateWhenConditionsForTaskInEnvironment(tt.args.environment, tt.args.task)
			if (err != nil) != tt.wantErr {
				t.Errorf("evaluateWhenConditionsForTaskInEnvironment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("evaluateWhenConditionsForTaskInEnvironment() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getEnvironmentVariablesForConditionalEvaluation(t *testing.T) {
	tests := []struct {
		name        string
		projectVars string
		envVars     string
		want        tasklib.TaskEnvironment
		wantErr     bool
	}{
		{
			name:        "Test unmarshalling of project vars",
			projectVars: `[{"name":"LAGOON_FASTLY_SERVICE_ID","value":"service-id:true","scope":"global"}]`,
			envVars:     ``,
			want: tasklib.TaskEnvironment{
				"LAGOON_FASTLY_SERVICE_ID": "service-id:true",
			},
			wantErr: false,
		},
		{
			name:        "Test unmarshalling of environment vars",
			projectVars: ``,
			envVars:     `[{"name":"LAGOON_FASTLY_SERVICE_ID","value":"env","scope":"global"}]`,
			want: tasklib.TaskEnvironment{
				"LAGOON_FASTLY_SERVICE_ID": "env",
			},
			wantErr: false,
		},
		{
			name:        "Test overwriting of project vars by environment vars",
			projectVars: `[{"name":"LAGOON_FASTLY_SERVICE_ID","value":"proj","scope":"global"}]`,
			envVars:     `[{"name":"LAGOON_FASTLY_SERVICE_ID","value":"env","scope":"global"}]`,
			want: tasklib.TaskEnvironment{
				"LAGOON_FASTLY_SERVICE_ID": "env",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldProjVars := os.Getenv("LAGOON_PROJECT_VARIABLES")
			oldEnvVars := os.Getenv("LAGOON_ENVIRONMENT_VARIABLES")
			os.Setenv("LAGOON_PROJECT_VARIABLES", tt.projectVars)
			os.Setenv("LAGOON_ENVIRONMENT_VARIABLES", tt.envVars)
			got, err := getEnvironmentVariablesForConditionalEvaluation()
			if (err != nil) != tt.wantErr {
				t.Errorf("getEnvironmentVariablesForConditionalEvaluation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getEnvironmentVariablesForConditionalEvaluation() got = %v, want %v", got, tt.want)
			}
			os.Setenv("LAGOON_PROJECT_VARIABLES", oldProjVars)
			os.Setenv("LAGOON_ENVIRONMENT_VARIABLES", oldEnvVars)
		})
	}
}
