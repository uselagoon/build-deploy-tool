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
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
}
