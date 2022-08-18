package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/uselagoon/build-deploy-tool/internal/compat"
	generator "github.com/uselagoon/build-deploy-tool/internal/generator"
)

var validateLagoonVersion = &cobra.Command{
	Use:     "lagoon-version",
	Aliases: []string{"lagoon-ver", "lag-ver", "lver", "lv"},
	Short:   "Check if the Lagoon version provided is supported by this tool",
	Run: func(cmd *cobra.Command, args []string) {
		// check for the LAGOON_SYSTEM_CORE_VERSION
		version, err := ValidateLagoonVersion(false)
		if err != nil {
			fmt.Println(fmt.Sprintf("Unable to validate lagoon version; %v", err))
			os.Exit(1)
		}
		supported := compat.CheckVersion(version)
		if !supported {
			fmt.Println(fmt.Sprintf("Lagoon version %s is not supported by this build-deploy-tool, you will need to upgrade your lagoon-core to at least version %s", version, compat.SupportedMinVersion()))
			os.Exit(1)
		}
	},
}

// ValidateLagoonVersion .
func ValidateLagoonVersion(debug bool) (string, error) {
	lagoonBuild, err := generator.NewGenerator(
		lagoonYml,
		lagoonYmlOverride,
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
		return "", err
	}

	return lagoonBuild.BuildValues.LagoonCoreVersion, nil
}

func init() {
	validateCmd.AddCommand(validateLagoonVersion)
}
