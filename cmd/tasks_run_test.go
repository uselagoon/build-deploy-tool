package cmd

import (
	"fmt"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	"github.com/uselagoon/build-deploy-tool/internal/tasklib"
	"os"
	"reflect"
	"strings"
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
			//We want to clear out all env vars
			prevEnv := os.Environ()
			for _, entry := range prevEnv {
				parts := strings.SplitN(entry, "=", 2)
				os.Unsetenv(parts[0])
			}
			oldProjVars := os.Getenv("LAGOON_PROJECT_VARIABLES")
			oldEnvVars := os.Getenv("LAGOON_ENVIRONMENT_VARIABLES")

			os.Setenv("LAGOON_PROJECT_VARIABLES", tt.projectVars)
			os.Setenv("LAGOON_ENVIRONMENT_VARIABLES", tt.envVars)
			got, err := getEnvironmentVariablesForConditionalEvaluation(false)
			if (err != nil) != tt.wantErr {
				t.Errorf("getEnvironmentVariablesForConditionalEvaluation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getEnvironmentVariablesForConditionalEvaluation() got = %v, want %v", got, tt.want)
			}
			os.Setenv("LAGOON_PROJECT_VARIABLES", oldProjVars)
			os.Setenv("LAGOON_ENVIRONMENT_VARIABLES", oldEnvVars)
			t.Cleanup(func() {
				for _, entry := range prevEnv {
					parts := strings.SplitN(entry, "=", 2)
					os.Setenv(parts[0], parts[1])
				}
			})
		})
	}
}

func Test_runTasks(t *testing.T) {
	type args struct {
		taskType                               int
		taskRunner                             iterateTaskFuncType
		lYAML                                  lagoon.YAML
		lagoonConditionalEvaluationEnvironment tasklib.TaskEnvironment
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Basic test",
			args: args{
				taskType: PRE_ROLLOUT_TASKS,
				lYAML: lagoon.YAML{
					Tasks: lagoon.Tasks{
						Prerollout: []lagoon.TaskRun{
							{
								Run: lagoon.Task{
									Command: "",
									When:    "",
								},
							},
						},
						Postrollout: nil,
					},
				},
				lagoonConditionalEvaluationEnvironment: tasklib.TaskEnvironment{
					"KEY1": "KEY2",
				},
				taskRunner: func(lagoonConditionalEvaluationEnvironment tasklib.TaskEnvironment, tasks []lagoon.Task) (error, bool) {
					if _, ok := lagoonConditionalEvaluationEnvironment["KEY1"]; !ok {
						return fmt.Errorf("Unable to find Key 1"), false
					}
					return nil, true
				},
			},
			wantErr: false,
		},
		{
			name: "Condition should fail",
			args: args{
				taskType: PRE_ROLLOUT_TASKS,
				lYAML: lagoon.YAML{
					Tasks: lagoon.Tasks{
						Prerollout: []lagoon.TaskRun{
							{
								Run: lagoon.Task{
									Command: "",
									When:    "NONEXISTANT == true",
								},
							},
						},
						Postrollout: nil,
					},
				},
				lagoonConditionalEvaluationEnvironment: tasklib.TaskEnvironment{
					"KEY1": "KEY2",
				},
				taskRunner: func(lagoonConditionalEvaluationEnvironment tasklib.TaskEnvironment, tasks []lagoon.Task) (error, bool) {
					for _, task := range tasks {
						_, err := evaluateWhenConditionsForTaskInEnvironment(lagoonConditionalEvaluationEnvironment, task)
						if err != nil {
							return err, true
						}
					}
					return nil, false
				},
			},
			wantErr: true,
		},
	}

	oldNamespace := namespace
	namespace = "default"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := runTasks(tt.args.taskType, tt.args.taskRunner, tt.args.lYAML, tt.args.lagoonConditionalEvaluationEnvironment); (err != nil) != tt.wantErr {
				t.Errorf("runTasks() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}
	namespace = oldNamespace
}
