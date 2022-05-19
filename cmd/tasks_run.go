package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	"github.com/uselagoon/build-deploy-tool/internal/tasklib"
	"io/ioutil"
	"os"
	"strings"
)

var runPreRollout, runPostRollout, outOfClusterConfig bool
var namespace string

const (
	PRE_ROLLOUT_TASKS = iota
	POST_ROLLOUT_TASKS
)

var taskCmd = &cobra.Command{
	Use:     "tasks",
	Aliases: []string{"tsk"},
	Short:   "Run tasks",
	Long:    `Will run Pre/Post/etc. tasks defined in a .lagoon.yml`,
}

var tasksPreRun = &cobra.Command{
	Use:     "pre-rollout",
	Aliases: []string{"pre"},
	Short:   "Will run pre rollout tasks defined in .lagoon.yml",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTasks(PRE_ROLLOUT_TASKS)
	},
}

var tasksPostRun = &cobra.Command{
	Use:     "post-rollout",
	Aliases: []string{"post"},
	Short:   "Will run post rollout tasks defined in .lagoon.yml",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTasks(POST_ROLLOUT_TASKS)
	},
}

func runTasks(taskType int) error {

	if namespace == "" {
		//Try load from file
		const filename = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
		if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("A target namespace is required to run pre/post-rollout tasks")
		}
		nsb, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}
		namespace = strings.Trim(string(nsb), "\n ")
	}

	// read the .lagoon.yml file
	var lYAML lagoon.YAML
	lPolysite := make(map[string]interface{})
	if err := lagoon.UnmarshalLagoonYAML(lagoonYml, &lYAML, &lPolysite); err != nil {
		return fmt.Errorf("couldn't read provided file `%v`: %v", lagoonYml, err)
	}

	lagoonConditionalEvaluationEnvironment, err := getEnvironmentVariablesForConditionalEvaluation(true)
	if err != nil {
		return err
	}

	if taskType == PRE_ROLLOUT_TASKS {
		fmt.Println("Executing Pre-rollout Tasks")
		err2, done := iterateTasks(lagoonConditionalEvaluationEnvironment, unwindTaskRun(lYAML.Tasks.Prerollout))
		if done {
			return err2
		}
		fmt.Println("Pre-rollout Tasks Complete")
	} else {
		fmt.Println("Skipping pre-rollout tasks")
	}

	if taskType == POST_ROLLOUT_TASKS {
		fmt.Println("Executing Post-rollout Tasks")
		err2, done := iterateTasks(lagoonConditionalEvaluationEnvironment, unwindTaskRun(lYAML.Tasks.Postrollout))
		if done {
			return err2
		}
		fmt.Println("Post-rollout Tasks Complete")
	} else {
		fmt.Println("Skipping post-rollout tasks")
	}
	return nil
}

func unwindTaskRun(taskRun []lagoon.TaskRun) []lagoon.Task {
	var tasks []lagoon.Task
	for _, taskrun := range taskRun {
		tasks = append(tasks, taskrun.Run)
	}
	return tasks
}

func iterateTasks(lagoonConditionalEvaluationEnvironment tasklib.TaskEnvironment, tasks []lagoon.Task) (error, bool) {
	for _, task := range tasks {
		runTask, err := evaluateWhenConditionsForTaskInEnvironment(lagoonConditionalEvaluationEnvironment, task)
		if err != nil {
			return err, true
		}
		if runTask {
			err := runCleanTaskInEnvironment(task)
			if err != nil {
				return err, true
			}
			if err != nil {
				return err, true
			}
		} else {
			fmt.Printf("Conditional '%v' for task: \n '%v' \n evaluated to false, skipping\n", task.When, task.Command)
		}
	}
	return nil, false
}

func getEnvironmentVariablesForConditionalEvaluation(pullWiderEnvironment bool) (tasklib.TaskEnvironment, error) {

	//TODO: a lot of this will likely be replacable by library functions
	lagoonConditionalEvaluationEnvironment := tasklib.TaskEnvironment{}
	//pull all pod env vars
	if pullWiderEnvironment {
		allEnvVarNames := os.Environ()
		for _, n := range allEnvVarNames {
			kv := strings.Split(n, "=")
			lagoonConditionalEvaluationEnvironment[kv[0]] = kv[1]
		}
	}

	projectVars := []lagoon.EnvironmentVariable{}
	envVars := []lagoon.EnvironmentVariable{}
	// get the project and environment variables
	projectVariables = helpers.GetEnv("LAGOON_PROJECT_VARIABLES", projectVariables, true)
	environmentVariables = helpers.GetEnv("LAGOON_ENVIRONMENT_VARIABLES", environmentVariables, true)
	json.Unmarshal([]byte(projectVariables), &projectVars)
	json.Unmarshal([]byte(environmentVariables), &envVars)
	lagoonEnvVars := lagoon.MergeVariables(projectVars, envVars)

	// Give context in the logs to how the tasks execution is being evaluated
	if len(lagoonEnvVars) > 0 {
		for _, envVar := range lagoonEnvVars {
			lagoonConditionalEvaluationEnvironment[envVar.Name] = envVar.Value
		}
	}
	blockList := []string{
		"LAGOON_PROJECT_VARIABLES",
		"LAGOON_ENVIRONMENT_VARIABLES",
	}
	for _, blockItem := range blockList {
		delete(lagoonConditionalEvaluationEnvironment, blockItem)
	}

	return lagoonConditionalEvaluationEnvironment, nil
}

func evaluateWhenConditionsForTaskInEnvironment(environment tasklib.TaskEnvironment, task lagoon.Task) (bool, error) {

	if len(task.When) == 0 { //no condition, so we run ...
		return true, nil
	}
	fmt.Println("Evaluating task condition - ", task.When)
	ret, err := tasklib.EvaluateExpressionsInTaskEnvironment(task.When, environment)
	if err != nil {
		fmt.Println("Error evaluating condition: ", err.Error())
		return false, err
	}
	retBool, okay := ret.(bool)
	if !okay {
		err := fmt.Errorf("Expression doesn't evaluate to a boolean")
		fmt.Println(err.Error())
		return false, err
	}
	return retBool, nil
}

func runCleanTaskInEnvironment(incoming lagoon.Task) error {
	task := lagoon.NewTask()
	task.Command = incoming.Command
	task.Namespace = namespace
	task.Service = incoming.Service
	task.Shell = incoming.Shell
	task.Container = incoming.Container
	err := lagoon.ExecuteTaskInEnvironment(task)
	return err
}

func init() {
	taskCmd.AddCommand(tasksPreRun)
	taskCmd.AddCommand(tasksPostRun)
	//tasksPreRun.Flags().StringVarP(&lagoonYml, "lagoon-yml", "l", ".lagoon.yml",
	//	"The .lagoon.yml file to read")

	addArgs := func(command *cobra.Command) {
		command.Flags().StringVarP(&namespace, "namespace", "n", "",
			"The environments environment variables JSON payload")
		//	"Will attempt to use KUBECONFIG to connect to cluster, defaults to in-cluster")
		command.Flags().StringVarP(&lagoonYml, "lagoon-yml", "l", ".lagoon.yml",
			"The .lagoon.yml file to read")
	}
	addArgs(tasksPreRun)
	addArgs(tasksPostRun)
}
