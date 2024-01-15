package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	generator "github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
)

var featureFlagIdentify = &cobra.Command{
	Use:     "feature",
	Aliases: []string{"f"},
	Short:   "Identify if a feature flag has been enabled",
	RunE: func(cmd *cobra.Command, args []string) error {
		generator, err := generator.GenerateInput(*rootCmd, false)
		if err != nil {
			return err
		}
		flagValue, err := IdentifyFeatureFlag(generator, "")
		if err != nil {
			return err
		}
		fmt.Println(flagValue)
		return nil
	},
}

// IdentifyFeatureFlag checks if a feature flag of given name has been set or not in a build
func IdentifyFeatureFlag(g generator.GeneratorInput, name string) (string, error) {
	lagoonBuild, err := generator.NewGenerator(
		g,
	)
	if err != nil {
		return "", err
	}

	forceFlagVar := helpers.GetEnv(fmt.Sprintf("%s%s", "LAGOON_FEATURE_FLAG_FORCE_", name), "", g.Debug)
	if forceFlagVar != "" {
		return forceFlagVar, nil
	}
	featureFlagVar, _ := lagoon.GetLagoonVariable(fmt.Sprintf("%s%s", "LAGOON_FEATURE_FLAG_", name), []string{"build", "global"}, *lagoonBuild.LagoonEnvironmentVariables)
	if featureFlagVar != nil {
		return featureFlagVar.Value, nil
	}
	defaultFlagVar := helpers.GetEnv(fmt.Sprintf("%s%s", "LAGOON_FEATURE_FLAG_DEFAULT_", name), "", g.Debug)
	return defaultFlagVar, nil
}

func init() {
	identifyCmd.AddCommand(featureFlagIdentify)
}
