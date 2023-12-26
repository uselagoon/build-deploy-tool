package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	"github.com/uselagoon/build-deploy-tool/internal/tasklib"
)

var runPreRollout, runPostRollout, outOfClusterConfig bool

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

// unidleThenRun is a wrapper around 'runCleanTaskInEnvironment' used for pre-rollout tasks
// We actually want to unidle the namespace before running pre-rollout tasks,
// so we wrap the usual task runner before calling it.
func unidleThenRun(namespace string, prePost string, incoming lagoon.Task) error {
	fmt.Printf("Unidling namespace with RequiresEnvironment: %v, ScaleMaxIterations:%v and ScaleWaitTime:%v\n", incoming.RequiresEnvironment, incoming.ScaleMaxIterations, incoming.ScaleWaitTime)
	err := lagoon.UnidleNamespace(context.TODO(), namespace, incoming.ScaleMaxIterations, incoming.ScaleWaitTime)
	if err != nil {
		switch {
		case errors.Is(err, lagoon.NamespaceUnidlingTimeoutError):
			if !incoming.RequiresEnvironment { // we don't have to kill this build if we can't bring the services up, so we just note the issue and continue
				fmt.Println("Namespace unidling is taking longer than expected - this might affect pre-rollout tasks that rely on multiple services")
			} else {
				return fmt.Errorf("Unable to unidle the environment for pre-rollout tasks in time (waited %v seconds, retried %v times) - exiting as the task is defined as requiring the environment to be up.",
					incoming.ScaleWaitTime, incoming.ScaleMaxIterations)
			}
		default:
			return fmt.Errorf("There was a problem when unidling the environment for pre-rollout tasks: %v", err.Error())
		}
	}
	return runCleanTaskInEnvironment(namespace, prePost, incoming)
}

var tasksPreRun = &cobra.Command{
	Use:     "pre-rollout",
	Aliases: []string{"pre"},
	Short:   "Will run pre rollout tasks defined in .lagoon.yml",
	RunE: func(cmd *cobra.Command, args []string) error {
		generator, err := generator.GenerateInput(*rootCmd, true)
		if err != nil {
			return err
		}
		lYAML, lagoonConditionalEvaluationEnvironment, buildValues, err := getEnvironmentInfo(generator)
		if err != nil {
			return err
		}
		fmt.Println("Executing Pre-rollout Tasks")

		taskIterator, err := iterateTaskGenerator(true, unidleThenRun, buildValues, "Pre-Rollout", true)
		if err != nil {
			fmt.Println("Pre-rollout Tasks Failed with the following error: ", err.Error())
			os.Exit(1)
		}

		err = runTasks(taskIterator, lYAML.Tasks.Prerollout, lagoonConditionalEvaluationEnvironment)
		if err != nil {
			fmt.Println("Pre-rollout Tasks Failed with the following error: ", err.Error())
			os.Exit(1)
		}
		fmt.Println("Pre-rollout Tasks Complete")
		return nil
	},
}

var tasksPostRun = &cobra.Command{
	Use:     "post-rollout",
	Aliases: []string{"post"},
	Short:   "Will run post rollout tasks defined in .lagoon.yml",
	RunE: func(cmd *cobra.Command, args []string) error {
		generator, err := generator.GenerateInput(*rootCmd, true)
		if err != nil {
			return err
		}
		lYAML, lagoonConditionalEvaluationEnvironment, buildValues, err := getEnvironmentInfo(generator)
		if err != nil {
			return err
		}

		fmt.Println("Executing Post-rollout Tasks")

		taskIterator, err := iterateTaskGenerator(false, runCleanTaskInEnvironment, buildValues, "Post-Rollout", true)
		if err != nil {
			fmt.Println("Pre-rollout Tasks Failed with the following error: ", err.Error())
			os.Exit(1)
		}
		err = runTasks(taskIterator, lYAML.Tasks.Postrollout, lagoonConditionalEvaluationEnvironment)
		if err != nil {
			fmt.Println("Post-rollout Tasks Failed with the following error: ", err.Error())
			os.Exit(1)
		}
		fmt.Println("Post-rollout Tasks Complete")
		return nil
	},
}

func getEnvironmentInfo(g generator.GeneratorInput) (lagoon.YAML, tasklib.TaskEnvironment, generator.BuildValues, error) {
	// read the .lagoon.yml file
	lagoonBuild, err := generator.NewGenerator(
		g,
	)
	if err != nil {
		return lagoon.YAML{}, nil, generator.BuildValues{}, err
	}

	lagoonConditionalEvaluationEnvironment := tasklib.TaskEnvironment{}
	if len(*lagoonBuild.LagoonEnvironmentVariables) > 0 {
		for _, envVar := range *lagoonBuild.LagoonEnvironmentVariables {
			lagoonConditionalEvaluationEnvironment[envVar.Name] = envVar.Value
		}
	}
	return *lagoonBuild.LagoonYAML, lagoonConditionalEvaluationEnvironment, *lagoonBuild.BuildValues, nil
}

// runTasks is essentially an interpreter. It takes in a runner function (that does the interpreting), the task list (a series of instructions)
// and the environment in which conditional statements are going to be run (i.e. a list of variables available to "where" clauses) and runs them.
func runTasks(taskRunner iterateTaskFuncType, tasks []lagoon.TaskRun, lagoonConditionalEvaluationEnvironment tasklib.TaskEnvironment) error {

	done, err := taskRunner(lagoonConditionalEvaluationEnvironment, unwindTaskRun(tasks))
	if done {
		return err
	}

	return nil
}

// unwindTaskRun simply reformats a []lagoon.TaskRun structure. It gets rid of the nested "run" field so that the array is flatter.
func unwindTaskRun(taskRun []lagoon.TaskRun) []lagoon.Task {
	var tasks []lagoon.Task
	for _, taskrun := range taskRun {
		tasks = append(tasks, taskrun.Run)
	}
	return tasks
}

// iterateTaskFuncType defines what a function that runs tasks looks like. There's an environment to evaluate a task,
// as well as the task definition itself.
type iterateTaskFuncType func(tasklib.TaskEnvironment, []lagoon.Task) (bool, error)

// iterateTaskGenerator is probably a little trickier than it should be, but it's essentially a factory for iterateTaskFuncTypes
// that lets the resulting function reference values as part of the closure, thereby cleaning up the definition a bit.
// so, the variables passed into the factor (eg. allowDeployMissingErrors, etc.) determine the way the function behaves,
// without needing to pass those into the call to the returned function itself.
func iterateTaskGenerator(allowDeployMissingErrors bool, taskRunner runTaskInEnvironmentFuncType, buildValues generator.BuildValues, prePost string, debug bool) (iterateTaskFuncType, error) {
	var retErr error
	return func(lagoonConditionalEvaluationEnvironment tasklib.TaskEnvironment, tasks []lagoon.Task) (bool, error) {
		for _, task := range tasks {
			// set the iterations and wait times here
			if task.ScaleMaxIterations == 0 {
				task.ScaleMaxIterations = buildValues.TaskScaleMaxIterations
			}
			if task.ScaleWaitTime == 0 {
				task.ScaleWaitTime = buildValues.TaskScaleWaitTime
			}
			runTask, err := evaluateWhenConditionsForTaskInEnvironment(lagoonConditionalEvaluationEnvironment, task, debug)
			if err != nil {
				return true, err
			}
			if runTask {
				err := taskRunner(buildValues.Namespace, prePost, task)
				if err != nil {
					switch e := err.(type) {
					case *lagoon.DeploymentMissingError:
						if allowDeployMissingErrors {
							if debug {
								fmt.Println("No running deployment found, skipping")
							}
						} else {
							return true, e
						}
					default:
						return true, e
					}
				}
			} else {
				if debug {
					fmt.Printf("Conditional '%v' for task: \n '%v' \n evaluated to false, skipping\n", task.When, task.Command)
				}
			}
		}
		return false, nil
	}, retErr
}

// evaluateWhenConditionsForTaskInEnvironment will take a task, check if it has a "when" field, and if it does, will evaluate it,
// in the environment given. It will return 'true' if the "when" condition evaluates to "true" (false otherwise), indicating
// that the task should be run (i.e. we execute the task in a running container).
func evaluateWhenConditionsForTaskInEnvironment(environment tasklib.TaskEnvironment, task lagoon.Task, debug bool) (bool, error) {

	if len(task.When) == 0 { //no condition, so we run ...
		return true, nil
	}
	if debug {
		fmt.Println("Evaluating task condition - ", task.When)
	}
	ret, err := tasklib.EvaluateExpressionsInTaskEnvironment(task.When, environment)
	if err != nil {
		if debug {
			fmt.Println("Error evaluating condition: ", err.Error())
		}
		return false, err
	}
	retBool, okay := ret.(bool)
	if !okay {
		err := fmt.Errorf("Expression doesn't evaluate to a boolean")
		if debug {
			fmt.Println(err.Error())
		}
		return false, err
	}
	return retBool, nil
}

type runTaskInEnvironmentFuncType func(namespace string, prePost string, incoming lagoon.Task) error

// runCleanTaskInEnvironment implements runTaskInEnvironmentFuncType and will
// 1. make sure the task we pass to the execution environment is free of any data we don't want (hence the new task)
// 2. will actually execute the task in the environment.
func runCleanTaskInEnvironment(namespace string, prePost string, incoming lagoon.Task) error {
	task := lagoon.NewTask()
	task.Command = incoming.Command
	task.Namespace = namespace
	task.Service = incoming.Service
	task.Shell = incoming.Shell
	task.Container = incoming.Container
	task.Name = incoming.Name
	task.ScaleMaxIterations = incoming.ScaleMaxIterations
	task.ScaleWaitTime = incoming.ScaleWaitTime
	err := lagoon.ExecuteTaskInEnvironment(task, prePost)
	return err
}

func init() {
	taskCmd.AddCommand(tasksPreRun)
	taskCmd.AddCommand(tasksPostRun)

	addArgs := func(command *cobra.Command) {
		command.Flags().StringP("namespace", "n", "",
			"The environments environment variables JSON payload")
	}
	addArgs(tasksPreRun)
	addArgs(tasksPostRun)
}
