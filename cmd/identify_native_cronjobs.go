package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	generator "github.com/uselagoon/build-deploy-tool/internal/generator"
)

var nativeCronjobsIdentify = &cobra.Command{
	Use:     "native-cronjobs",
	Aliases: []string{"nc"},
	Short:   "Identify any native cronjobs for a specific environment",
	RunE: func(cmd *cobra.Command, args []string) error {
		generator, err := generator.GenerateInput(*rootCmd, false)
		if err != nil {
			return err
		}
		cronjobs, err := IdentifyNativeCronjobs(generator)
		if err != nil {
			return err
		}
		fmt.Println(cronjobs)
		return nil
	},
}

// IdentifyNativeCronjobs .
func IdentifyNativeCronjobs(g generator.GeneratorInput) (string, error) {
	lagoonBuild, err := generator.NewGenerator(
		g,
	)
	if err != nil {
		return "", err
	}

	nativeCronjobs := []string{}
	for _, service := range lagoonBuild.BuildValues.Services {
		for _, nc := range service.NativeCronjobs {
			nativeCronjobs = append(nativeCronjobs, nc.Name)
		}
	}
	nativeCronjobsBytes, _ := json.Marshal(nativeCronjobs)

	return string(nativeCronjobsBytes), nil
}

func init() {
	identifyCmd.AddCommand(nativeCronjobsIdentify)
	identifyCmd.AddCommand(ingressIdentify)
}
