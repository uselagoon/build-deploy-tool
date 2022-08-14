package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	"os"
)

var validateLagoonYml = &cobra.Command{
	Use:   "lagoon-yml",
	Short: "Verify .lagoon.yml and environment for compatability with this tool",
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		lYAML := &lagoon.YAML{}
		err = ValidateLagoonYml(lagoonYml, lYAML, projectName, false)
		if err != nil {
			fmt.Println("Could not validate your .lagoon.yml - ", err.Error())
			os.Exit(1)
		}
	},
}

func ValidateLagoonYml(lagoonYml string, lYAML *lagoon.YAML, projectName string, debug bool) error {
	if err := generator.LoadAndUnmarshallLagoonYml(lagoonYml, lagoonYmlOverride, lagoonYmlEnvVar, lYAML, projectName, debug); err != nil {
		return err
	}
	return nil
}

func init() {
	validateCmd.AddCommand(validateLagoonYml)
}
