package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
)

var validateDockerCompose = &cobra.Command{
	Use:     "docker-compose",
	Aliases: []string{"compose", "dc"},
	Short:   "Verify docker-compose file for compatability with this tool",
	Run: func(cmd *cobra.Command, args []string) {
		// @TODO: ignoreNonStringKeyErrors is `true` by default because Lagoon doesn't enforce
		// docker-compose compliance yet
		ignoreMissingEnvFiles, err := rootCmd.PersistentFlags().GetBool("ignore-missing-env-files")
		if err != nil {
			fmt.Println(fmt.Errorf("error reading ignore-missing-env-files flag: %v", err))
			os.Exit(1)
		}
		ignoreNonStringKeyErrors, err := rootCmd.PersistentFlags().GetBool("ignore-non-string-key-errors")
		if err != nil {
			fmt.Println(fmt.Errorf("error reading ignore-non-string-key-errors flag: %v", err))
			os.Exit(1)
		}
		dockerComposeFile, err := cmd.Flags().GetString("docker-compose")
		if err != nil {
			fmt.Println(fmt.Errorf("error reading docker-compose flag: %v", err))
			os.Exit(1)
		}

		err = ValidateDockerCompose(dockerComposeFile, ignoreNonStringKeyErrors, ignoreMissingEnvFiles)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	},
}

var validateDockerComposeWithErrors = &cobra.Command{
	Use:     "docker-compose-with-errors",
	Aliases: []string{"dcwe"},
	Short:   "Verify docker-compose file for compatability with this tool with next versions of compose-go library",
	Run: func(cmd *cobra.Command, args []string) {
		dockerComposeFile, err := cmd.Flags().GetString("docker-compose")
		if err != nil {
			fmt.Println(fmt.Errorf("error reading docker-compose flag: %v", err))
			os.Exit(1)
		}

		err = validateDockerComposeWithError(dockerComposeFile)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	},
}

// ValidateDockerCompose validate a docker-compose file
func ValidateDockerCompose(file string, ignoreErrors, ignoreMisEnvFiles bool) error {
	_, _, _, err := lagoon.UnmarshaDockerComposeYAML(file, ignoreErrors, ignoreMisEnvFiles, map[string]string{})
	if err != nil {
		return err
	}
	return nil
}

// validateDockerComposeWithErrors validate a docker-compose file yaml structure properly
func validateDockerComposeWithError(file string) error {
	err := lagoon.ValidateUnmarshalDockerComposeYAML(file)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	validateCmd.AddCommand(validateDockerCompose)
	validateCmd.AddCommand(validateDockerComposeWithErrors)
	validateDockerCompose.Flags().StringP("docker-compose", "", "docker-compose.yml",
		"The docker-compose.yml file to read.")
	validateDockerComposeWithErrors.Flags().StringP("docker-compose", "", "docker-compose.yml",
		"The docker-compose.yml file to read.")
}
