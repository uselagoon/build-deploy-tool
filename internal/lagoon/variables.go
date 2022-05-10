package lagoon

import (
	"fmt"

	"github.com/uselagoon/build-deploy-tool/internal/helpers"
)

// EnvironmentVariable is used to define Lagoon environment variables.
type EnvironmentVariable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Scope string `json:"scope"`
}

// MergeVariables merges lagoon environment variables.
func MergeVariables(project, environment []EnvironmentVariable) []EnvironmentVariable {
	allVars := []EnvironmentVariable{}
	existsInEnvironment := false
	// replace any variables from the project with ones from the environment
	// this only modifies ones that exist in project
	for _, pVar := range project {
		add := EnvironmentVariable{}
		for _, eVar := range environment {
			if eVar.Name == pVar.Name {
				existsInEnvironment = true
				add = eVar
			}
		}
		if existsInEnvironment {
			allVars = append(allVars, add)
			existsInEnvironment = false
		} else {
			allVars = append(allVars, pVar)
		}
	}
	// add any that exist in the environment only to the final variables list
	for _, eVar := range environment {
		add := EnvironmentVariable{}
		for _, aVar := range allVars {
			add = eVar
			if eVar.Name == aVar.Name {
				existsInEnvironment = true
			}
		}
		if existsInEnvironment {
			existsInEnvironment = false
		} else {
			allVars = append(allVars, add)
		}
	}
	return allVars
}

// GetLagoonVariable returns a given environment variable
func GetLagoonVariable(name string, scope []string, variables []EnvironmentVariable) (*EnvironmentVariable, error) {
	for _, v := range variables {
		scoped := true
		if scope != nil {
			scoped = false
			if helpers.Contains(scope, v.Scope) {
				scoped = true
			}
		}
		if v.Name == name && scoped {
			return &v, nil
		}
	}
	return nil, fmt.Errorf("variable not found")
}

// VariableExists checks if a variable exists in a slice of environment variables
func VariableExists(vars *[]EnvironmentVariable, name, value string) bool {
	exists := false
	for _, v := range *vars {
		if v.Name == name && v.Value == value {
			exists = true
		}
	}
	return exists
}
