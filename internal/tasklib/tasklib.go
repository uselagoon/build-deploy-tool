package tasklib

import (
	"fmt"
	"github.com/PaesslerAG/gval"
)

//Note - the structure of a task environment is going to mirror what gval uses for now
// it makes sense - but we can wrap this in convenience functions to improve the total safety

type TaskEnvironment map[string]interface{}

func EvaluateExpressionsInTaskEnvironment(expression string, env TaskEnvironment) (interface{}, error) {
	value, err := gval.Evaluate(expression, env,
		gval.Function("env", func(args ...interface{}) (interface{}, error) {
			name := args[0].(string)
			var theDefault interface{}
			if len(args) == 2 {
				theDefault := args[1]
			}

			if val, ok := env[name]; ok != nil

			return false, nil
		}))
	if err != nil {
		return nil, err
	}
	return value, nil
}

type environmentTools struct {
	Env TaskEnvironment
}

//func (t *environmentTools) exists(environmentElement string) bool {
//	if _, ok := t.Env[environmentElement]
//}
