package tasklib

import (
	"github.com/PaesslerAG/gval"
)

//Note - the structure of a task environment is going to mirror what gval uses for now
// it makes sense - but we can wrap this in convenience functions to improve the total safety

type TaskEnvironment map[string]interface{}

func EvaluateExpressionsInTaskEnvironment(expression string, env TaskEnvironment) (interface{}, error) {
	value, err := gval.Evaluate(expression, env)
	if err != nil {
		return nil, err
	}
	return value, nil
}
