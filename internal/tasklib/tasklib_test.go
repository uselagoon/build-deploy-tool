package tasklib

import (
	"reflect"
	"testing"
)

func TestEvaluateExpressionsInTaskEnvironment(t *testing.T) {
	type args struct {
		expression string
		env        TaskEnvironment
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "Simple Text Expression Eval",
			args: args{
				expression: `"Hello " + name + "!"`,
				env: TaskEnvironment{
					"name": "world",
				},
			},
			want: "Hello world!",
		},
		{
			name: "Particular expression value == is true",
			args: args{
				expression: `value=="test"`,
				env: TaskEnvironment{
					"value": "test",
				},
			},
			want: true,
		},
		{
			name: "invalid expression",
			args: args{
				expression: `)1==1`,
				env: TaskEnvironment{
					"value": "test",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Uses env with no need for default",
			args: args{
				expression: `withDefault("value", "adefault") == "test"`,
				env: TaskEnvironment{
					"value": "test",
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Uses env with default",
			args: args{
				expression: `withDefault("valuethatdoesntexist'", "the_default") == "the_default"`,
				env: TaskEnvironment{
					"value": "test",
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Uses regex match",
			args: args{
				expression: `LAGOON_PR_VALUE=~"PR.*"`,
				env: TaskEnvironment{
					"LAGOON_PR_VALUE": "PR-9345",
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Variable exists",
			args: args{
				expression: `exists("EXISTENT_VALUE")`,
				env: TaskEnvironment{
					"EXISTENT_VALUE": "I exist!",
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Variable exists",
			args: args{
				expression: `exists("NON_EXISTENT_VALUE")`,
				env: TaskEnvironment{
					"EXISTENT_VALUE": "I exist!",
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Variable doesn't exist - will throw error",
			args: args{
				expression: `NON_EXISTENT_VALUE==1`,
				env:        TaskEnvironment{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EvaluateExpressionsInTaskEnvironment(tt.args.expression, tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("EvaluateExpressionsInTaskEnvironment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EvaluateExpressionsInTaskEnvironment() got = %v, want %v", got, tt.want)
			}
		})
	}
}
