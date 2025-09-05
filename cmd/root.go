package cmd

/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/uselagoon/build-deploy-tool/internal/dbaasclient"
	"github.com/uselagoon/build-deploy-tool/internal/generator"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "lagoon-build",
	Short: "A tool to help with generating Lagoon resources for Lagoon builds",
	Long: `A tool to help with generating Lagoon resources for Lagoon builds
This tool will read a .lagoon.yml file and also all the required environment variables from
within a Lagoon build to help with generating the resources`,
}

var templateCmd = &cobra.Command{
	Use:     "template",
	Aliases: []string{"t"},
	Short:   "Generate templates",
	Long:    `Generate any templates for Lagoon builds`,
}

var configCmd = &cobra.Command{
	Use:     "configuration",
	Aliases: []string{"config", "c"},
	Short:   "Generate configurations",
	Long:    `Generate any configurations for Lagoon builds`,
}

var identifyCmd = &cobra.Command{
	Use:     "identify",
	Aliases: []string{"id", "i"},
	Short:   "Identify resources",
	Long:    `Identify resources for Lagoon builds`,
}

var validateCmd = &cobra.Command{
	Use:     "validate",
	Aliases: []string{"valid", "v"},
	Short:   "Validate resources",
	Long:    `Validate resources for Lagoon builds`,
}

var collectCmd = &cobra.Command{
	Use:     "collect",
	Aliases: []string{"col", "c"},
	Short:   "Collect resource information",
	Long:    `Collect resource information for Lagoon builds`,
}

var runCmd = &cobra.Command{
	Use:     "run",
	Aliases: []string{"r"},
	Short:   "Run a process",
	Long:    `Run a process`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// version/build information (populated at build time by make file)
var (
	bdtName    = "build-deploy-tool"
	bdtVersion = "0.x.x"
	bdtBuild   = ""
	goVersion  = ""
)

// version/build information command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Version information",
	Run: func(cmd *cobra.Command, args []string) {
		displayVersionInfo()
	},
}

func displayVersionInfo() {
	fmt.Printf("%s %s (built: %s / go %s)\n", bdtName, bdtVersion, bdtBuild, goVersion)
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(templateCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(taskCmd)
	rootCmd.AddCommand(identifyCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(collectCmd)
	rootCmd.AddCommand(runCmd)

	rootCmd.PersistentFlags().StringP("lagoon-yml", "l", ".lagoon.yml",
		"The .lagoon.yml file to read")
	rootCmd.PersistentFlags().StringP("lagoon-yml-override", "", ".lagoon.override.yml",
		"The .lagoon.yml override file to read for merging values into target lagoon.yml")
	rootCmd.PersistentFlags().StringP("project-name", "p", "",
		"The project name")
	rootCmd.PersistentFlags().StringP("environment-name", "e", "",
		"The environment name to check")
	rootCmd.PersistentFlags().StringP("environment-type", "E", "",
		"The type of environment (development or production)")
	rootCmd.PersistentFlags().StringP("build-type", "d", "",
		"The type of build (branch, pullrequest, promote)")
	rootCmd.PersistentFlags().StringP("branch", "b", "",
		"The name of the branch")
	rootCmd.PersistentFlags().StringP("pullrequest-number", "P", "",
		"The pullrequest number")
	rootCmd.PersistentFlags().StringP("pullrequest-title", "", "",
		"The pullrequest title")
	rootCmd.PersistentFlags().StringP("pullrequest-head-branch", "H", "",
		"The pullrequest head branch")
	rootCmd.PersistentFlags().StringP("pullrequest-base-branch", "B", "",
		"The pullrequest base branch")
	rootCmd.PersistentFlags().StringP("lagoon-version", "L", "",
		"The lagoon version")
	rootCmd.PersistentFlags().StringP("project-variables", "", "",
		"The JSON payload for project scope variables")
	rootCmd.PersistentFlags().StringP("environment-variables", "", "",
		"The JSON payload for environment scope variables")
	rootCmd.PersistentFlags().StringP("active-environment", "a", "",
		"Name of the active environment if known")
	rootCmd.PersistentFlags().StringP("standby-environment", "s", "",
		"Name of the standby environment if known")
	rootCmd.PersistentFlags().StringP("template-path", "t", "/kubectl-build-deploy/",
		"Path to the template on disk")
	rootCmd.PersistentFlags().StringP("saved-templates-path", "T", "/kubectl-build-deploy/lagoon/services-routes",
		"Path to where the resulting templates are saved")
	rootCmd.PersistentFlags().String("default-backup-schedule", "", "The default backup schedule to use")
	rootCmd.PersistentFlags().StringP("monitoring-config", "M", "",
		"The monitoring contact config if known")
	rootCmd.PersistentFlags().StringP("monitoring-status-page-id", "m", "",
		"The monitoring status page ID if known")
	rootCmd.PersistentFlags().StringP("fastly-cache-no-cache-id", "F", "",
		"The fastly cache no cache service ID to use")
	rootCmd.PersistentFlags().StringP("fastly-service-id", "f", "",
		"The fastly service ID to use")
	rootCmd.PersistentFlags().StringP("fastly-api-secret-prefix", "A", "fastly-api-",
		"The fastly secret prefix to use")
	rootCmd.PersistentFlags().BoolP("ignore-non-string-key-errors", "", true,
		"Ignore non-string-key docker-compose errors (true by default, subject to change).")
	rootCmd.PersistentFlags().BoolP("ignore-missing-env-files", "", true,
		"Ignore missing env_file files (true by default, subject to change).")
	rootCmd.PersistentFlags().StringP("images", "", "",
		"JSON representation of service:image reference")
	rootCmd.PersistentFlags().StringP("dbaas-creds", "", "",
		"JSON representation of dbaas credential references")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
}

// helper function that reads flag overrides and retruns a generated input dataset
// this is called from within the main environment setup helper function
func GenerateInput(rootCmd cobra.Command, debug bool) (generator.GeneratorInput, error) {
	lagoonYAML, err := rootCmd.PersistentFlags().GetString("lagoon-yml")
	if err != nil {
		return generator.GeneratorInput{}, fmt.Errorf("error reading lagoon-yml flag: %v", err)
	}
	lagoonYAMLOverride, err := rootCmd.PersistentFlags().GetString("lagoon-yml-override")
	if err != nil {
		return generator.GeneratorInput{}, fmt.Errorf("error reading lagoon-yml-override flag: %v", err)
	}
	lagoonVersion, err := rootCmd.PersistentFlags().GetString("lagoon-version")
	if err != nil {
		return generator.GeneratorInput{}, fmt.Errorf("error reading lagoon-version flag: %v", err)
	}
	projectName, err := rootCmd.PersistentFlags().GetString("project-name")
	if err != nil {
		return generator.GeneratorInput{}, fmt.Errorf("error reading project-name flag: %v", err)
	}
	environmentName, err := rootCmd.PersistentFlags().GetString("environment-name")
	if err != nil {
		return generator.GeneratorInput{}, fmt.Errorf("error reading environment-name flag: %v", err)
	}
	environmentType, err := rootCmd.PersistentFlags().GetString("environment-type")
	if err != nil {
		return generator.GeneratorInput{}, fmt.Errorf("error reading environment-type flag: %v", err)
	}
	activeEnvironment, err := rootCmd.PersistentFlags().GetString("active-environment")
	if err != nil {
		return generator.GeneratorInput{}, fmt.Errorf("error reading active-environment flag: %v", err)
	}
	standbyEnvironment, err := rootCmd.PersistentFlags().GetString("standby-environment")
	if err != nil {
		return generator.GeneratorInput{}, fmt.Errorf("error reading standby-environment flag: %v", err)
	}
	projectVariables, err := rootCmd.PersistentFlags().GetString("project-variables")
	if err != nil {
		return generator.GeneratorInput{}, fmt.Errorf("error reading project-variables flag: %v", err)
	}
	environmentVariables, err := rootCmd.PersistentFlags().GetString("environment-variables")
	if err != nil {
		return generator.GeneratorInput{}, fmt.Errorf("error reading environment-variables flag: %v", err)
	}
	buildType, err := rootCmd.PersistentFlags().GetString("build-type")
	if err != nil {
		return generator.GeneratorInput{}, fmt.Errorf("error reading build-type flag: %v", err)
	}
	branch, err := rootCmd.PersistentFlags().GetString("branch")
	if err != nil {
		return generator.GeneratorInput{}, fmt.Errorf("error reading branch flag: %v", err)
	}
	prNumber, err := rootCmd.PersistentFlags().GetString("pullrequest-number")
	if err != nil {
		return generator.GeneratorInput{}, fmt.Errorf("error reading pullrequest-number flag: %v", err)
	}
	prTitle, err := rootCmd.PersistentFlags().GetString("pullrequest-title")
	if err != nil {
		return generator.GeneratorInput{}, fmt.Errorf("error reading pullrequest-title flag: %v", err)
	}
	prHeadBranch, err := rootCmd.PersistentFlags().GetString("pullrequest-head-branch")
	if err != nil {
		return generator.GeneratorInput{}, fmt.Errorf("error reading pullrequest-head-branch flag: %v", err)
	}
	prBaseBranch, err := rootCmd.PersistentFlags().GetString("pullrequest-base-branch")
	if err != nil {
		return generator.GeneratorInput{}, fmt.Errorf("error reading pullrequest-base-branch flag: %v", err)
	}
	monitoringContact, err := rootCmd.PersistentFlags().GetString("monitoring-config")
	if err != nil {
		return generator.GeneratorInput{}, fmt.Errorf("error reading monitoring-config flag: %v", err)
	}
	monitoringStatusPageID, err := rootCmd.PersistentFlags().GetString("monitoring-status-page-id")
	if err != nil {
		return generator.GeneratorInput{}, fmt.Errorf("error reading monitoring-status-page-id flag: %v", err)
	}
	fastlyCacheNoCache, err := rootCmd.PersistentFlags().GetString("fastly-cache-no-cache-id")
	if err != nil {
		return generator.GeneratorInput{}, fmt.Errorf("error reading fastly-cache-no-cache-id flag: %v", err)
	}
	ignoreMissingEnvFiles, err := rootCmd.PersistentFlags().GetBool("ignore-missing-env-files")
	if err != nil {
		return generator.GeneratorInput{}, fmt.Errorf("error reading ignore-missing-env-files flag: %v", err)
	}
	ignoreNonStringKeyErrors, err := rootCmd.PersistentFlags().GetBool("ignore-non-string-key-errors")
	if err != nil {
		return generator.GeneratorInput{}, fmt.Errorf("error reading ignore-non-string-key-errors flag: %v", err)
	}
	savedTemplates, err := rootCmd.PersistentFlags().GetString("saved-templates-path")
	if err != nil {
		return generator.GeneratorInput{}, fmt.Errorf("error reading saved-templates-path flag: %v", err)
	}
	defaultBackupSchedule, err := rootCmd.PersistentFlags().GetString("default-backup-schedule")
	if err != nil {
		return generator.GeneratorInput{}, fmt.Errorf("error reading default-backup-schedule flag: %v", err)
	}
	// create a dbaas client with the default configuration
	dbaas := dbaasclient.NewClient(dbaasclient.Client{})
	return generator.GeneratorInput{
		Debug:                    debug,
		LagoonYAML:               lagoonYAML,
		LagoonYAMLOverride:       lagoonYAMLOverride,
		LagoonVersion:            lagoonVersion,
		ProjectName:              projectName,
		EnvironmentName:          environmentName,
		EnvironmentType:          environmentType,
		ActiveEnvironment:        activeEnvironment,
		StandbyEnvironment:       standbyEnvironment,
		ProjectVariables:         projectVariables,
		EnvironmentVariables:     environmentVariables,
		BuildType:                buildType,
		Branch:                   branch,
		PRNumber:                 prNumber,
		PRTitle:                  prTitle,
		PRHeadBranch:             prHeadBranch,
		PRBaseBranch:             prBaseBranch,
		MonitoringContact:        monitoringContact,
		MonitoringStatusPageID:   monitoringStatusPageID,
		FastlyCacheNoCache:       fastlyCacheNoCache,
		SavedTemplatesPath:       savedTemplates,
		IgnoreMissingEnvFiles:    ignoreMissingEnvFiles,
		IgnoreNonStringKeyErrors: ignoreNonStringKeyErrors,
		DBaaSClient:              dbaas,
		DefaultBackupSchedule:    defaultBackupSchedule,
	}, nil
}
