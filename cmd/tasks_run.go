package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	"github.com/uselagoon/build-deploy-tool/internal/tasklib"
)

var runPreRollout, runPostRollout, outOfClusterConfig bool
var namespace string

const (
	preRolloutTasks = iota
	postRolloutTasks
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

		lYAML, lagoonConditionalEvaluationEnvironment, err := getEnvironmentInfo()
		if err != nil {
			return err
		}
		return runTasks(preRolloutTasks, iterateTasks, lYAML, lagoonConditionalEvaluationEnvironment)
	},
}

var tasksPostRun = &cobra.Command{
	Use:     "post-rollout",
	Aliases: []string{"post"},
	Short:   "Will run post rollout tasks defined in .lagoon.yml",
	RunE: func(cmd *cobra.Command, args []string) error {

		lYAML, lagoonConditionalEvaluationEnvironment, err := getEnvironmentInfo()
		if err != nil {
			return err
		}

		return runTasks(postRolloutTasks, iterateTasks, lYAML, lagoonConditionalEvaluationEnvironment)
	},
}

func getEnvironmentInfo() (lagoon.YAML, tasklib.TaskEnvironment, error) {
	// read the .lagoon.yml file
	activeEnv := false
	standbyEnv := false

	lagoonEnvVars := []lagoon.EnvironmentVariable{}
	lagoonValues := lagoon.BuildValues{}
	lYAML := lagoon.YAML{}
	autogenRoutes := new(lagoon.RoutesV2)
	mainRoutes := new(lagoon.RoutesV2)
	activeStandbyRoutes := new(lagoon.RoutesV2)

	err := collectBuildValues(false, &activeEnv, &standbyEnv, &lagoonEnvVars, &lagoonValues, &lYAML, autogenRoutes, mainRoutes, activeStandbyRoutes, ignoreNonStringKeyErrors)
	if err != nil {
		return lagoon.YAML{}, nil, err
	}

	lagoonConditionalEvaluationEnvironment := tasklib.TaskEnvironment{}
	if len(lagoonEnvVars) > 0 {
		for _, envVar := range lagoonEnvVars {
			lagoonConditionalEvaluationEnvironment[envVar.Name] = envVar.Value
		}
	}
	return lYAML, lagoonConditionalEvaluationEnvironment, nil
}

func runTasks(taskType int, taskRunner iterateTaskFuncType, lYAML lagoon.YAML, lagoonConditionalEvaluationEnvironment tasklib.TaskEnvironment) error {

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

	if taskType == preRolloutTasks {
		fmt.Println("Executing Pre-rollout Tasks")
		done, err := taskRunner(lagoonConditionalEvaluationEnvironment, unwindTaskRun(lYAML.Tasks.Prerollout))
		if done {
			return err
		}
		fmt.Println("Pre-rollout Tasks Complete")
	} else {
		fmt.Println("Skipping pre-rollout tasks")
	}

	if taskType == postRolloutTasks {
		fmt.Println("Executing Post-rollout Tasks")
		fmt.Println(lYAML.Tasks.Postrollout)
		done, err := taskRunner(lagoonConditionalEvaluationEnvironment, unwindTaskRun(lYAML.Tasks.Postrollout))
		if done {
			return err
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

type iterateTaskFuncType func(tasklib.TaskEnvironment, []lagoon.Task) (bool, error)

func iterateTasks(lagoonConditionalEvaluationEnvironment tasklib.TaskEnvironment, tasks []lagoon.Task) (bool, error) {
	for _, task := range tasks {
		runTask, err := evaluateWhenConditionsForTaskInEnvironment(lagoonConditionalEvaluationEnvironment, task)
		if err != nil {
			return true, err
		}
		if runTask {
			err := runCleanTaskInEnvironment(task)
			if err != nil {
				return true, err
			}
			if err != nil {
				return true, err
			}
		} else {
			fmt.Printf("Conditional '%v' for task: \n '%v' \n evaluated to false, skipping\n", task.When, task.Command)
		}
	}
	return false, nil
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
