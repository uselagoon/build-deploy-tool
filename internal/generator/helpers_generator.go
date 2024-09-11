package generator

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/uselagoon/build-deploy-tool/internal/dbaasclient"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
	"k8s.io/apimachinery/pkg/api/resource"
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

// checks the provided environment variables looking for feature flag based variables
func CheckFeatureFlag(key string, envVariables []lagoon.EnvironmentVariable, debug bool) string {
	// check for force value
	if value, ok := os.LookupEnv(fmt.Sprintf("LAGOON_FEATURE_FLAG_FORCE_%s", key)); ok {
		if debug {
			fmt.Printf("Using forced flag value from build variable %s\n", fmt.Sprintf("LAGOON_FEATURE_FLAG_FORCE_%s", key))
		}
		return value
	}
	// check lagoon environment variables
	for _, lVar := range envVariables {
		if strings.Contains(lVar.Name, fmt.Sprintf("LAGOON_FEATURE_FLAG_%s", key)) {
			if debug {
				fmt.Printf("Using flag value from Lagoon environment variable %s\n", fmt.Sprintf("LAGOON_FEATURE_FLAG_%s", key))
			}
			return lVar.Value
		}
	}
	// return default
	if value, ok := os.LookupEnv(fmt.Sprintf("LAGOON_FEATURE_FLAG_DEFAULT_%s", key)); ok {
		if debug {
			fmt.Printf("Using default flag value from build variable %s\n", fmt.Sprintf("LAGOON_FEATURE_FLAG_DEFAULT_%s", key))
		}
		return value
	}
	// otherwise nothing
	return ""
}

func CheckAdminFeatureFlag(key string, debug bool) string {
	if value, ok := os.LookupEnv(fmt.Sprintf("ADMIN_LAGOON_FEATURE_FLAG_%s", key)); ok {
		if debug {
			fmt.Printf("Using admin feature flag value from build variable %s\n", fmt.Sprintf("ADMIN_LAGOON_FEATURE_FLAG_%s", key))
		}
		return value
	}
	return ""
}

func ValidateResourceQuantity(s string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New(fmt.Sprint(x))
			}
		}
	}()
	resource.MustParse(s)
	return nil
}

func ValidateResourceSize(size string) (int64, error) {
	volQ, err := resource.ParseQuantity(size)
	if err != nil {
		return 0, err
	}
	volS, _ := volQ.AsInt64()
	return volS, nil
}

// ContainsRegistry checks if a string slice contains a specific string regex match.
func ContainsRegistry(regex []ContainerRegistry, match string) bool {
	for _, v := range regex {
		m, _ := regexp.MatchString(v.URL, match)
		if m {
			return true
		}
	}
	return false
}

func checkDuplicateCronjobs(cronjobs []lagoon.Cronjob) error {
	var unique []lagoon.Cronjob
	var duplicates []lagoon.Cronjob
	for _, v := range cronjobs {
		skip := false
		for _, u := range unique {
			if v.Name == u.Name {
				skip = true
				duplicates = append(duplicates, v)
				break
			}
		}
		if !skip {
			unique = append(unique, v)
		}
	}
	var uniqueDuplicates []lagoon.Cronjob
	for _, d := range duplicates {
		for _, u := range unique {
			if d.Name == u.Name {
				uniqueDuplicates = append(uniqueDuplicates, u)
			}
		}
	}
	// join the two together
	result := append(duplicates, uniqueDuplicates...)
	if result != nil {
		b, _ := json.Marshal(result)
		return fmt.Errorf("duplicate named cronjobs detected: %v", string(b))
	}
	return nil
}

// getDBaasEnvironment will check the dbaas provider to see if an environment exists or not
func getDBaasEnvironment(
	buildValues *BuildValues,
	dbaasEnvironment *string,
	lagoonOverrideName,
	lagoonType string,
) (bool, error) {
	if buildValues.DBaaSEnvironmentTypeOverrides != nil {
		dbaasEnvironmentTypeSplit := strings.Split(buildValues.DBaaSEnvironmentTypeOverrides.Value, ",")
		for _, sType := range dbaasEnvironmentTypeSplit {
			sTypeSplit := strings.Split(sType, ":")
			if sTypeSplit[0] == lagoonOverrideName {
				*dbaasEnvironment = sTypeSplit[1]
			}
		}
	}
	exists, err := buildValues.DBaaSClient.CheckProvider(buildValues.DBaaSOperatorEndpoint, lagoonType, *dbaasEnvironment)
	if err != nil {
		return exists, fmt.Errorf("there was an error checking DBaaS endpoint %s: %v", buildValues.DBaaSOperatorEndpoint, err)
	}
	return exists, nil
}
