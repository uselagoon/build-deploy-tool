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
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(templateCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(identifyCmd)

	rootCmd.Flags().StringVarP(&lagoonYml, "lagoon-yml", "l", ".lagoon.yml",
		"The .lagoon.yml file to read")
	rootCmd.Flags().StringVarP(&projectName, "project-name", "p", "",
		"The project name")
	rootCmd.Flags().StringVarP(&environmentName, "environment-name", "e", "",
		"The environment name to check")
	rootCmd.Flags().StringVarP(&environmentType, "environment-type", "E", "",
		"The type of environment (development or production)")
	rootCmd.Flags().StringVarP(&buildType, "build-type", "d", "",
		"The type of build (branch, pullrequest, promote)")
	rootCmd.Flags().StringVarP(&branch, "branch", "b", "",
		"The name of the branch")
	rootCmd.Flags().StringVarP(&prNumber, "pullrequest-number", "P", "",
		"The pullrequest number")
	rootCmd.Flags().StringVarP(&prHeadBranch, "pullrequest-head-branch", "H", "",
		"The pullrequest head branch")
	rootCmd.Flags().StringVarP(&prBaseBranch, "pullrequest-base-branch", "B", "",
		"The pullrequest base branch")
	rootCmd.Flags().StringVarP(&lagoonVersion, "lagoon-version", "L", "",
		"The lagoon version")
	rootCmd.Flags().StringVarP(&activeEnvironment, "active-environment", "a", "",
		"Name of the active environment if known")
	rootCmd.Flags().StringVarP(&standbyEnvironment, "standby-environment", "s", "",
		"Name of the standby environment if known")
	rootCmd.Flags().StringVarP(&templateValues, "template-path", "t", "/kubectl-build-deploy/",
		"Path to the template on disk")
	rootCmd.Flags().StringVarP(&savedTemplates, "saved-templates-path", "T", "/kubectl-build-deploy/lagoon/services-routes",
		"Path to where the resulting templates are saved")
	rootCmd.Flags().StringVarP(&monitoringContact, "monitoring-config", "M", "",
		"The monitoring contact config if known")
	rootCmd.Flags().StringVarP(&monitoringStatusPageID, "monitoring-status-page-id", "m", "",
		"The monitoring status page ID if known")
	rootCmd.Flags().StringVarP(&fastlyCacheNoCahce, "fastly-cache-no-cache-id", "F", "",
		"The fastly cache no cache service ID to use")
	rootCmd.Flags().StringVarP(&fastlyServiceID, "fastly-service-id", "f", "",
		"The fastly service ID to use")
	rootCmd.Flags().StringVarP(&fastlyAPISecretPrefix, "fastly-api-secret-prefix", "A", "fastly-api-",
		"The fastly secret prefix to use")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
}
