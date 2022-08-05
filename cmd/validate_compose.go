package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
)

var (
	dockerComposeFile        string
	ignoreNonStringKeyErrors bool
	ignoreMissingEnvFiles    bool
)

var validateDockerCompose = &cobra.Command{
	Use:     "docker-compose",
	Aliases: []string{"compose", "dc"},
	Short:   "Verify docker-compose file for compatability with this tool",
	Run: func(cmd *cobra.Command, args []string) {
		// @TODO: ignoreNonStringKeyErrors is `true` by default because Lagoon doesn't enforce
		// docker-compose compliance yet
		err := ValidateDockerCompose(dockerComposeFile, ignoreNonStringKeyErrors, ignoreMissingEnvFiles)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	},
}

// ValidateDockerCompose validate a docker-compose file
func ValidateDockerCompose(file string, ignoreErrors, ignoreMisEnvFiles bool) error {
	_, _, err := lagoon.UnmarshaDockerComposeYAML(file, ignoreErrors, ignoreMisEnvFiles, map[string]string{})
	if err != nil {
		return err
	}
	return nil
}

func init() {
	validateCmd.AddCommand(validateDockerCompose)
	validateDockerCompose.Flags().StringVarP(&dockerComposeFile, "docker-compose", "", "docker-compose.yml",
		"The docker-compose.yml file to read.")
}
