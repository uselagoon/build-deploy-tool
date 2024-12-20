package lagoon

import (
	"fmt"
	"slices"

	"github.com/uselagoon/build-deploy-tool/internal/helpers"
)

// EnvironmentVariable is used to define Lagoon environment variables.
type EnvironmentVariable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Scope string `json:"scope"`
}

// MergeVariables merges lagoon environment variables.
func MergeVariables(organization, project, environment, config []EnvironmentVariable) []EnvironmentVariable {

	// Helper function to compare environment variable names.
	findByName := func(name string) func(EnvironmentVariable) bool {
		return func(eVar EnvironmentVariable) bool { return eVar.Name == name }
	}

	// Start with config variables since they are most specific.
	allVars := make([]EnvironmentVariable, len(config))
	copy(allVars, config)

	for _, eVar := range environment {
		idx := slices.IndexFunc(allVars, findByName(eVar.Name))

		// Append environment variables that are distinct.
		if idx == -1 {
			allVars = append(allVars, eVar)
		}
	}

	for _, pVar := range project {
		idx := slices.IndexFunc(allVars, findByName(pVar.Name))

		// Append project variables that are distinct.
		if idx == -1 {
			allVars = append(allVars, pVar)
			continue
		}

		// Overwrite environment variables if they are suppossed to be internally
		// scoped.
		if pVar.Scope == "internal_system" {
			allVars[idx] = pVar
		}
	}

	for _, oVar := range organization {
		idx := slices.IndexFunc(allVars, findByName(oVar.Name))

		// Append organization variables that are distinct.
		if idx == -1 {
			allVars = append(allVars, oVar)
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
	return nil, fmt.Errorf("variable %s not found", name)
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
