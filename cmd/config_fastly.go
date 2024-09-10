package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
)

var fastlyConfigGeneration = &cobra.Command{
	Use:     "fastly",
	Aliases: []string{"f"},
	Short:   "Generate fastly configuration for a specific ingress domain",
	RunE: func(cmd *cobra.Command, args []string) error {
		// generate the fastly configuration from the provided flags/variables
		domainName, err := cmd.Flags().GetString("domain")
		if err != nil {
			return fmt.Errorf("error reading domain flag: %v", err)
		}
		f, err := FastlyConfigGeneration(false, domainName)
		if err != nil {
			return err
		}
		strF, _ := json.Marshal(f)
		fmt.Println(string(strF))
		return nil
	},
}

// FastlyConfigGeneration .
func FastlyConfigGeneration(debug bool, domain string) (lagoon.Fastly, error) {
	// environment variables will override what is provided by flags
	fastlyCacheNoCahce, err := rootCmd.PersistentFlags().GetString("fastly-cache-no-cache-id")
	if err != nil {
		return lagoon.Fastly{}, fmt.Errorf("error reading fastly-cache-no-cache-id flag: %v", err)
	}
	fastlyServiceID, err := rootCmd.PersistentFlags().GetString("fastly-service-id")
	if err != nil {
		return lagoon.Fastly{}, fmt.Errorf("error reading fastly-service-id flag: %v", err)
	}
	projectVariables, err := rootCmd.PersistentFlags().GetString("project-variables")
	if err != nil {
		return lagoon.Fastly{}, fmt.Errorf("error reading project-variables flag: %v", err)
	}
	environmentVariables, err := rootCmd.PersistentFlags().GetString("environment-variables")
	if err != nil {
		return lagoon.Fastly{}, fmt.Errorf("error reading environment-variables flag: %v", err)
	}

	fastlyCacheNoCahce = helpers.GetEnv("LAGOON_FASTLY_NOCACHE_SERVICE_ID", fastlyCacheNoCahce, debug)
	fastlyServiceID = helpers.GetEnv("ROUTE_FASTLY_SERVICE_ID", fastlyServiceID, debug)

	// get the project and environment variables
	projectVariables = helpers.GetEnv("LAGOON_PROJECT_VARIABLES", projectVariables, debug)
	environmentVariables = helpers.GetEnv("LAGOON_ENVIRONMENT_VARIABLES", environmentVariables, debug)

	// unmarshal and then merge the two so there is only 1 set of variables to iterate over
	projectVars := []lagoon.EnvironmentVariable{}
	envVars := []lagoon.EnvironmentVariable{}
	json.Unmarshal([]byte(projectVariables), &projectVars)
	json.Unmarshal([]byte(environmentVariables), &envVars)
	lagoonEnvVars := lagoon.MergeVariables(projectVars, envVars)

	// generate the fastly configuration from the provided flags/variables
	f := &lagoon.Fastly{}
	err = lagoon.GenerateFastlyConfiguration(f, fastlyCacheNoCahce, fastlyServiceID, domain, lagoonEnvVars)
	if err != nil {
		return lagoon.Fastly{}, err
	}
	return *f, nil
}

func init() {
	configCmd.AddCommand(fastlyConfigGeneration)
	fastlyConfigGeneration.Flags().StringP("domain", "D", "",
		"The .lagoon.yml file to read")
}
