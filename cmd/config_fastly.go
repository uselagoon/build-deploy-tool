package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/lagoon"
)

var (
	domainName string
)

var fastlyConfigGeneration = &cobra.Command{
	Use:     "fastly",
	Aliases: []string{"f"},
	Short:   "Generate fastly configuration for a specific ingress domain",
	RunE: func(cmd *cobra.Command, args []string) error {
		// generate the fastly configuration from the provided flags/variables
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
	fastlyCacheNoCahce = helpers.GetEnv("LAGOON_FASTLY_NOCACHE_SERVICE_ID", fastlyCacheNoCahce, debug)
	fastlyServiceID = helpers.GetEnv("ROUTE_FASTLY_SERVICE_ID", fastlyServiceID, debug)
	fastlyAPISecretPrefix = helpers.GetEnv("FASTLY_API_SECRET_PREFIX", fastlyAPISecretPrefix, debug)

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
	f, err := lagoon.GenerateFastlyConfiguration(fastlyCacheNoCahce, fastlyServiceID, domain, fastlyAPISecretPrefix, lagoonEnvVars)
	if err != nil {
		return lagoon.Fastly{}, err
	}
	return f, nil
}

func init() {
	configCmd.AddCommand(fastlyConfigGeneration)
	fastlyConfigGeneration.Flags().StringVarP(&domainName, "domain", "D", "",
		"The .lagoon.yml file to read")
	fastlyConfigGeneration.Flags().StringVarP(&projectVariables, "project-variables", "v", "",
		"The projects environment variables JSON payload")
	fastlyConfigGeneration.Flags().StringVarP(&environmentVariables, "environment-variables", "V", "",
		"The environments environment variables JSON payload")
	fastlyConfigGeneration.Flags().StringVarP(&fastlyCacheNoCahce, "fastly-cache-no-cache-id", "F", "",
		"The fastly cache no cache service ID to use")
	fastlyConfigGeneration.Flags().StringVarP(&fastlyServiceID, "fastly-service-id", "f", "",
		"The fastly service ID to use")
}
