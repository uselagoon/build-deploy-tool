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
			wantErr: true, //error = parsing error: )1==1 :1:1 - 1:2 unexpected ")" while scanning extensions, wantErr false
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
