package tasklib

import (
	"github.com/PaesslerAG/gval"
)

//Note - the structure of a task environment is going to mirror what gval uses for now
// it makes sense - but we can wrap this in convenience functions to improve the total safety

type TaskEnvironment map[string]interface{}

func EvaluateExpressionsInTaskEnvironment(expression string, env TaskEnvironment) (interface{}, error) {
	value, err := gval.Evaluate(expression, env,
		gval.Function("withDefault", func(args ...interface{}) (interface{}, error) {
			name := args[0].(string)
			var val, theDefault interface{}
			val, ok := env[name]
			if len(args) == 2 {
				theDefault = args[1]
			}
			if !ok {
				return theDefault, nil
			}

			return val, nil
		}),
		gval.Function("exists", func(args ...interface{}) bool {
			name := args[0].(string)
			_, ok := env[name]
			if !ok {
				return false
			}
			return true
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
