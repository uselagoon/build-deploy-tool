package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
)

var tasksScaffold = &cobra.Command{
	Use:     "taskScaffold",
	Aliases: []string{"ts"},
	Short:   "This is just a scaffolding command to get me started writing tasks stuff",
	RunE: func(cmd *cobra.Command, args []string) error {
		// get the project and environment variables
		projectVariables = helpers.GetEnv("LAGOON_PROJECT_VARIABLES", projectVariables, true)
		environmentVariables = helpers.GetEnv("LAGOON_ENVIRONMENT_VARIABLES", environmentVariables, true)

		// unmarshal and then merge the two so there is only 1 set of variables to iterate over
		projectVars := []lagoon.EnvironmentVariable{}
		envVars := []lagoon.EnvironmentVariable{}
		json.Unmarshal([]byte(projectVariables), &projectVars)
		json.Unmarshal([]byte(environmentVariables), &envVars)
		lagoonEnvVars := lagoon.MergeVariables(projectVars, envVars)

		if len(envVars) > 0 {
			for i, envVar := range lagoonEnvVars {
				fmt.Sprintf("lagoonEnvVars[%i] is %v:%v\n", i, envVar.Name, envVar.Value)
			}
		} else {
			fmt.Println("No Project or Environment Variables!")
		}

		//For now, we actually just want to run some arbitrary command in a running container, perhaps?

		task := lagoon.NewTask()
		task.Command = "env"
		lagoon.ExecuteTaskInEnvironment(task)

		command := []string{
			"sh",
			"-c",
			"env",
		}
		stdout, stdin, error := lagoon.ExecPod("nginx-deployment", "default", command, false, "ubuntu")

		if error != nil {
			panic(error.Error())
		}
		fmt.Println(stdout)
		fmt.Println(stdin)

		return nil
	},
}

func init() {
	configCmd.AddCommand(tasksScaffold)
	tasksScaffold.Flags().StringVarP(&domainName, "domain", "D", "",
		"The .lagoon.yml file to read")
	tasksScaffold.Flags().StringVarP(&projectVariables, "project-variables", "v", "",
		"The projects environment variables JSON payload")
	tasksScaffold.Flags().StringVarP(&environmentVariables, "environment-variables", "V", "",
		"The environments environment variables JSON payload")
}
