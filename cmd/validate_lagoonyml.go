package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	"sigs.k8s.io/yaml"
)

var validateLagoonYml = &cobra.Command{
	Use:   "lagoon-yml",
	Short: "Verify .lagoon.yml and environment for compatability with this tool",
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		lagoonYAML, _ := rootCmd.PersistentFlags().GetString("lagoon-yml")
		lagoonYAMLOverride, _ := rootCmd.PersistentFlags().GetString("lagoon-yml-override")
		projectName, _ := rootCmd.PersistentFlags().GetString("project-name")
		printOutput, _ := cmd.Flags().GetBool("print-resulting-lagoonyml")

		lYAML := &lagoon.YAML{}
		err = ValidateLagoonYml(lagoonYAML, lagoonYAMLOverride, "LAGOON_YAML_OVERRIDE", lYAML, projectName, false)
		if err != nil {
			fmt.Println("Could not validate your .lagoon.yml - ", err.Error())
			os.Exit(1)
		}

		if printOutput {
			resultingBS, err := yaml.Marshal(lYAML)
			if err != nil {
				fmt.Println("Unable to unmarshal resulting yml for printing: ", err)
				os.Exit(1)
			}
			fmt.Println(string(resultingBS))
		}
	},
}

func ValidateLagoonYml(lagoonYml string, lagoonYmlOverride string, lagoonYmlEnvVar string, lYAML *lagoon.YAML, projectName string, debug bool) error {
	if err := generator.LoadAndUnmarshalLagoonYml(lagoonYml, lagoonYmlOverride, lagoonYmlEnvVar, lYAML, projectName, debug); err != nil {
		return err
	}
	return nil
}

func init() {
	validateCmd.PersistentFlags().BoolP("print-resulting-lagoonyml", "", false,
		"Display the resulting, post merging, lagoon.yml file.")
	validateCmd.AddCommand(validateLagoonYml)
}
