package cmd

import (
	generator "github.com/uselagoon/build-deploy-tool/internal/generator"
)

func generatorInput(debug bool) generator.GeneratorInput {
	lagoonYAML, _ := rootCmd.PersistentFlags().GetString("lagoon-yml")
	lagoonYAMLOverride, _ := rootCmd.PersistentFlags().GetString("lagoon-yml-override")
	lagoonVersion, _ := rootCmd.PersistentFlags().GetString("lagoon-version")
	projectName, _ := rootCmd.PersistentFlags().GetString("project-name")
	environmentName, _ := rootCmd.PersistentFlags().GetString("environment-name")
	environmentType, _ := rootCmd.PersistentFlags().GetString("environment-type")
	activeEnvironment, _ := rootCmd.PersistentFlags().GetString("active-environment")
	standbyEnvironment, _ := rootCmd.PersistentFlags().GetString("stabdby-environment")
	projectVariables, _ := rootCmd.PersistentFlags().GetString("project-variables")
	environmentVariables, _ := rootCmd.PersistentFlags().GetString("environment-variables")
	buildType, _ := rootCmd.PersistentFlags().GetString("build-type")
	branch, _ := rootCmd.PersistentFlags().GetString("branch")
	prNumber, _ := rootCmd.PersistentFlags().GetString("pullrequest-number")
	prTitle, _ := rootCmd.PersistentFlags().GetString("pullrequest-title")
	prHeadBranch, _ := rootCmd.PersistentFlags().GetString("pullrequest-head-branch")
	prBaseBranch, _ := rootCmd.PersistentFlags().GetString("pullrequest-base-branch")
	monitoringContact, _ := rootCmd.PersistentFlags().GetString("monitoring-config")
	monitoringStatusPageID, _ := rootCmd.PersistentFlags().GetString("monitoring-status-page-id")
	fastlyCacheNoCahce, _ := rootCmd.PersistentFlags().GetString("fastly-cache-no-cache-id")
	fastlyAPISecretPrefix, _ := rootCmd.PersistentFlags().GetString("fastly-api-secret-prefix")
	ignoreMissingEnvFiles, _ := rootCmd.PersistentFlags().GetBool("ignore-missing-env-files")
	ignoreNonStringKeyErrors, _ := rootCmd.PersistentFlags().GetBool("ignore-non-string-key-errors")
	savedTemplates, _ := rootCmd.PersistentFlags().GetString("saved-templates-path")
	return generator.GeneratorInput{
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
	}
}
