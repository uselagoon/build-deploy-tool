package cmd

import generator "github.com/uselagoon/build-deploy-tool/internal/generator"

var lagoonYml, lagoonYmlOverride, environmentName, projectName, activeEnvironment, standbyEnvironment, environmentType string
var buildType, lagoonVersion, branch, prTitle, prNumber, prHeadBranch, prBaseBranch string
var projectVariables, environmentVariables, monitoringStatusPageID, monitoringContact string
var templateValues, savedTemplates, fastlyCacheNoCahce, fastlyServiceID, fastlyAPISecretPrefix string
var monitoringEnabled bool

func generatorInput(debug bool) generator.GeneratorInput {
	return generator.GeneratorInput{
		Debug:                    debug,
		LagoonYAML:               lagoonYml,
		LagoonYAMLOverride:       lagoonYmlOverride,
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
		IgnoreMissingEnvFiles:    ignoreMissingEnvFiles,
		IgnoreNonStringKeyErrors: ignoreNonStringKeyErrors,
	}
}
