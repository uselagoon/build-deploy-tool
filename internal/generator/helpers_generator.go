package generator

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/uselagoon/build-deploy-tool/internal/dbaasclient"
)

// helper function that reads flag overrides and retruns a generated input dataset
// this is called from within the main environment setup helper function
func GenerateInput(rootCmd cobra.Command, debug bool) (GeneratorInput, error) {
	lagoonYAML, err := rootCmd.PersistentFlags().GetString("lagoon-yml")
	if err != nil {
		return GeneratorInput{}, fmt.Errorf("error reading lagoon-yml flag: %v", err)
	}
	lagoonYAMLOverride, err := rootCmd.PersistentFlags().GetString("lagoon-yml-override")
	if err != nil {
		return GeneratorInput{}, fmt.Errorf("error reading lagoon-yml-override flag: %v", err)
	}
	lagoonVersion, err := rootCmd.PersistentFlags().GetString("lagoon-version")
	if err != nil {
		return GeneratorInput{}, fmt.Errorf("error reading lagoon-version flag: %v", err)
	}
	projectName, err := rootCmd.PersistentFlags().GetString("project-name")
	if err != nil {
		return GeneratorInput{}, fmt.Errorf("error reading project-name flag: %v", err)
	}
	environmentName, err := rootCmd.PersistentFlags().GetString("environment-name")
	if err != nil {
		return GeneratorInput{}, fmt.Errorf("error reading environment-name flag: %v", err)
	}
	environmentType, err := rootCmd.PersistentFlags().GetString("environment-type")
	if err != nil {
		return GeneratorInput{}, fmt.Errorf("error reading environment-type flag: %v", err)
	}
	activeEnvironment, err := rootCmd.PersistentFlags().GetString("active-environment")
	if err != nil {
		return GeneratorInput{}, fmt.Errorf("error reading active-environment flag: %v", err)
	}
	standbyEnvironment, err := rootCmd.PersistentFlags().GetString("standby-environment")
	if err != nil {
		return GeneratorInput{}, fmt.Errorf("error reading standby-environment flag: %v", err)
	}
	projectVariables, err := rootCmd.PersistentFlags().GetString("project-variables")
	if err != nil {
		return GeneratorInput{}, fmt.Errorf("error reading project-variables flag: %v", err)
	}
	environmentVariables, err := rootCmd.PersistentFlags().GetString("environment-variables")
	if err != nil {
		return GeneratorInput{}, fmt.Errorf("error reading environment-variables flag: %v", err)
	}
	buildType, err := rootCmd.PersistentFlags().GetString("build-type")
	if err != nil {
		return GeneratorInput{}, fmt.Errorf("error reading build-type flag: %v", err)
	}
	branch, err := rootCmd.PersistentFlags().GetString("branch")
	if err != nil {
		return GeneratorInput{}, fmt.Errorf("error reading branch flag: %v", err)
	}
	prNumber, err := rootCmd.PersistentFlags().GetString("pullrequest-number")
	if err != nil {
		return GeneratorInput{}, fmt.Errorf("error reading pullrequest-number flag: %v", err)
	}
	prTitle, err := rootCmd.PersistentFlags().GetString("pullrequest-title")
	if err != nil {
		return GeneratorInput{}, fmt.Errorf("error reading pullrequest-title flag: %v", err)
	}
	prHeadBranch, err := rootCmd.PersistentFlags().GetString("pullrequest-head-branch")
	if err != nil {
		return GeneratorInput{}, fmt.Errorf("error reading pullrequest-head-branch flag: %v", err)
	}
	prBaseBranch, err := rootCmd.PersistentFlags().GetString("pullrequest-base-branch")
	if err != nil {
		return GeneratorInput{}, fmt.Errorf("error reading pullrequest-base-branch flag: %v", err)
	}
	monitoringContact, err := rootCmd.PersistentFlags().GetString("monitoring-config")
	if err != nil {
		return GeneratorInput{}, fmt.Errorf("error reading monitoring-config flag: %v", err)
	}
	monitoringStatusPageID, err := rootCmd.PersistentFlags().GetString("monitoring-status-page-id")
	if err != nil {
		return GeneratorInput{}, fmt.Errorf("error reading monitoring-status-page-id flag: %v", err)
	}
	fastlyCacheNoCahce, err := rootCmd.PersistentFlags().GetString("fastly-cache-no-cache-id")
	if err != nil {
		return GeneratorInput{}, fmt.Errorf("error reading fastly-cache-no-cache-id flag: %v", err)
	}
	fastlyAPISecretPrefix, err := rootCmd.PersistentFlags().GetString("fastly-api-secret-prefix")
	if err != nil {
		return GeneratorInput{}, fmt.Errorf("error reading fastly-api-secret-prefix flag: %v", err)
	}
	ignoreMissingEnvFiles, err := rootCmd.PersistentFlags().GetBool("ignore-missing-env-files")
	if err != nil {
		return GeneratorInput{}, fmt.Errorf("error reading ignore-missing-env-files flag: %v", err)
	}
	ignoreNonStringKeyErrors, err := rootCmd.PersistentFlags().GetBool("ignore-non-string-key-errors")
	if err != nil {
		return GeneratorInput{}, fmt.Errorf("error reading ignore-non-string-key-errors flag: %v", err)
	}
	savedTemplates, err := rootCmd.PersistentFlags().GetString("saved-templates-path")
	if err != nil {
		return GeneratorInput{}, fmt.Errorf("error reading saved-templates-path flag: %v", err)
	}
	defaultBackupSchedule, err := rootCmd.PersistentFlags().GetString("default-backup-schedule")
	if err != nil {
		return GeneratorInput{}, fmt.Errorf("error reading default-backup-schedule flag: %v", err)
	}
	// create a dbaas client with the default configuration
	dbaas := dbaasclient.NewClient(dbaasclient.Client{})
	return GeneratorInput{
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
		FastlyCacheNoCahce:       fastlyCacheNoCahce,
		FastlyAPISecretPrefix:    fastlyAPISecretPrefix,
		SavedTemplatesPath:       savedTemplates,
		IgnoreMissingEnvFiles:    ignoreMissingEnvFiles,
		IgnoreNonStringKeyErrors: ignoreNonStringKeyErrors,
		DBaaSClient:              dbaas,
		DefaultBackupSchedule:    defaultBackupSchedule,
	}, nil
}
