package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/spf13/cobra"
	generator "github.com/uselagoon/build-deploy-tool/internal/generator"
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

		lYAML, lagoonConditionalEvaluationEnvironment, err := getEnvironmentInfo(true)
		if err != nil {
			return err
		}
		fmt.Println("Executing Pre-rollout Tasks")
		err = runTasks(iterateTaskGenerator(true, runCleanTaskInEnvironment), lYAML.Tasks.Prerollout, lagoonConditionalEvaluationEnvironment)
		if err != nil {
			return err
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

		lYAML, lagoonConditionalEvaluationEnvironment, err := getEnvironmentInfo(true)
		if err != nil {
			return err
		}

		fmt.Println("Executing Post-rollout Tasks")
		err = runTasks(iterateTaskGenerator(false, runCleanTaskInEnvironment), lYAML.Tasks.Postrollout, lagoonConditionalEvaluationEnvironment)
		if err != nil {
			return err
		}
		fmt.Println("Post-rollout Tasks Complete")
		return nil
	},
}

func getEnvironmentInfo(debug bool) (lagoon.YAML, tasklib.TaskEnvironment, error) {
	// read the .lagoon.yml file
	lagoonBuild, err := generator.NewGenerator(
		lagoonYml,
		projectVariables,
		environmentVariables,
		projectName,
		environmentName,
		environmentType,
		activeEnvironment,
		standbyEnvironment,
		buildType,
		branch,
		prNumber,
		prTitle,
		prHeadBranch,
		prBaseBranch,
		lagoonVersion,
		defaultBackupSchedule,
		hourlyDefaultBackupRetention,
		dailyDefaultBackupRetention,
		weeklyDefaultBackupRetention,
		monthlyDefaultBackupRetention,
		monitoringContact,
		monitoringStatusPageID,
		fastlyCacheNoCahce,
		fastlyAPISecretPrefix,
		fastlyServiceID,
		ignoreNonStringKeyErrors,
		ignoreMissingEnvFiles,
		debug,
	)
	if err != nil {
		return lagoon.YAML{}, nil, err
	}

	lagoonConditionalEvaluationEnvironment := tasklib.TaskEnvironment{}
	if len(*lagoonBuild.LagoonEnvironmentVariables) > 0 {
		for _, envVar := range *lagoonBuild.LagoonEnvironmentVariables {
			lagoonConditionalEvaluationEnvironment[envVar.Name] = envVar.Value
		}
	}
	return *lagoonBuild.LagoonYAML, lagoonConditionalEvaluationEnvironment, nil
}

func runTasks(taskRunner iterateTaskFuncType, tasks []lagoon.TaskRun, lagoonConditionalEvaluationEnvironment tasklib.TaskEnvironment) error {

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

	done, err := taskRunner(lagoonConditionalEvaluationEnvironment, unwindTaskRun(tasks))
	if done {
		return err
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

func iterateTaskGenerator(allowDeployMissingErrors bool, taskRunner runTaskInEnvironmentFuncType) iterateTaskFuncType {
	return func(lagoonConditionalEvaluationEnvironment tasklib.TaskEnvironment, tasks []lagoon.Task) (bool, error) {
		for _, task := range tasks {
			runTask, err := evaluateWhenConditionsForTaskInEnvironment(lagoonConditionalEvaluationEnvironment, task)
			if err != nil {
				return true, err
			}
			if runTask {
				err := taskRunner(task)
				if err != nil {
					switch e := err.(type) {
					case *lagoon.DeploymentMissingError:
						if allowDeployMissingErrors {
							fmt.Println("No running deployment found, skipping")
						} else {
							return true, e
						}
					default:
						return true, e
					}
				}
			} else {
				fmt.Printf("Conditional '%v' for task: \n '%v' \n evaluated to false, skipping\n", task.When, task.Command)
			}
		}
		return false, nil
	}
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

type runTaskInEnvironmentFuncType func(incoming lagoon.Task) error

// implements runTaskInEnvironmentFuncType
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
