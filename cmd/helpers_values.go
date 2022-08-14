package cmd

var lagoonYml, lagoonYmlOverride, lagoonYmlEnvVar, environmentName, projectName, activeEnvironment, standbyEnvironment, environmentType string
var buildType, lagoonVersion, branch, prTitle, prNumber, prHeadBranch, prBaseBranch string
var projectVariables, environmentVariables, monitoringStatusPageID, monitoringContact string
var templateValues, savedTemplates, fastlyCacheNoCahce, fastlyServiceID, fastlyAPISecretPrefix string
var monitoringEnabled bool

var defaultBackupSchedule, hourlyDefaultBackupRetention, dailyDefaultBackupRetention, weeklyDefaultBackupRetention, monthlyDefaultBackupRetention string
