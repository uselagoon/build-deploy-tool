package generator

import "fmt"

// MergeVariables merges lagoon environment variables.
func MergeVariables(project, environment []LagoonEnvironmentVariable) []LagoonEnvironmentVariable {
	allVars := []LagoonEnvironmentVariable{}
	existsInEnvironment := false
	for _, pVar := range project {
		add := LagoonEnvironmentVariable{}
		for _, eVar := range environment {
			if pVar.Name == eVar.Name {
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
	return allVars
}

func getLagoonVariable(name string, variables []LagoonEnvironmentVariable) (LagoonEnvironmentVariable, error) {
	for _, v := range variables {
		if v.Name == name {
			return v, nil
		}
	}
	return LagoonEnvironmentVariable{}, fmt.Errorf("variable not found")
}
