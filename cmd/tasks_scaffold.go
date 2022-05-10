package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
)

var runPreRollout, runPostRollout bool
var namespace string

var tasksRun = &cobra.Command{
	Use:     "tasks",
	Aliases: []string{"tr"},
	Short:   "Will run pre and/or post rollout tasks defined in .lagoon.yml",
	RunE: func(cmd *cobra.Command, args []string) error {

		if !runPostRollout && !runPostRollout {
			return fmt.Errorf("Neither pre nor post rollout tasks have been selected - exiting")
		}

		if namespace == "" {
			return fmt.Errorf("A target namespace is required to run pre/post-rollout tasks")
		}

		// get the project and environment variables
		projectVariables = helpers.GetEnv("LAGOON_PROJECT_VARIABLES", projectVariables, true)
		environmentVariables = helpers.GetEnv("LAGOON_ENVIRONMENT_VARIABLES", environmentVariables, true)

		// unmarshal and then merge the two so there is only 1 set of variables to iterate over
		projectVars := []lagoon.EnvironmentVariable{}
		envVars := []lagoon.EnvironmentVariable{}
		json.Unmarshal([]byte(projectVariables), &projectVars)
		json.Unmarshal([]byte(environmentVariables), &envVars)
		lagoonEnvVars := lagoon.MergeVariables(projectVars, envVars)

		// Give context in the logs to how the tasks execution is being evaluated
		if len(envVars) > 0 {
			fmt.Println("Evaluating tasks with the following variables")
			for _, envVar := range lagoonEnvVars {
				//Trim the value of the env var so it doesn't persist in the logs
				fmt.Printf("%v:%v\n", envVar.Name, envVar.Value[0:1]+"...")
			}
		} else {
			fmt.Println("There were no Environment or Projects values found for evaluation - continuing.")
		}

		// read the .lagoon.yml file
		var lYAML lagoon.YAML
		lPolysite := make(map[string]interface{})
		if err := lagoon.UnmarshalLagoonYAML(lagoonYml, &lYAML, &lPolysite); err != nil {
			return fmt.Errorf("couldn't read provided file `%v`: %v", lagoonYml, err)
		}

		//TODO: Answer question - is the only diff between pre and post that if they fail, the process exits?

		if runPreRollout {
			fmt.Println("Executing Pre-rollout Tasks")
			for _, run := range lYAML.Tasks.Prerollout {
				task := lagoon.NewTask()
				task.Command = run.Run.Command
				task.Namespace = namespace
				task.Service = run.Run.Service
				task.Shell = run.Run.Shell
				task.Container = run.Run.Container
				err := lagoon.ExecuteTaskInEnvironment(task)
				if err != nil {
					return err
				}
			}
			fmt.Println("Pre-rollout Tasks Complete")
		} else {
			fmt.Println("Skipping pre-rollout tasks")
		}

		if runPostRollout {
			for _, run := range lYAML.Tasks.Postrollout {
				fmt.Println("Executing Post-rollout Tasks")
				task := lagoon.NewTask()
				task.Command = run.Run.Command
				task.Namespace = namespace
				task.Service = run.Run.Service
				task.Shell = run.Run.Shell
				task.Container = run.Run.Container
				err := lagoon.ExecuteTaskInEnvironment(task)
				if err != nil {
					return err
				}
				fmt.Println("Post-rollout Tasks Complete")
			}
		} else {
			fmt.Println("Skipping post-rollout tasks")
		}

		return nil
	},
}

func init() {
	configCmd.AddCommand(tasksRun)
	tasksRun.Flags().StringVarP(&projectVariables, "project-variables", "v", "",
		"The projects environment variables JSON payload")
	tasksRun.Flags().StringVarP(&environmentVariables, "environment-variables", "V", "",
		"The environments environment variables JSON payload")
	tasksRun.Flags().StringVarP(&lagoonYml, "lagoon-yml", "l", ".lagoon.yml",
		"The .lagoon.yml file to read")
	tasksRun.Flags().BoolVarP(&runPreRollout, "pre-rollout", "", false,
		"Will run pre-rollout tasks if true")
	tasksRun.Flags().BoolVarP(&runPostRollout, "post-rollout", "", false,
		"Will run post-rollout tasks if true")
	tasksRun.Flags().StringVarP(&namespace, "namespace", "n", "",
		"The environments environment variables JSON payload")
}
