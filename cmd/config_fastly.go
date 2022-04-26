package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/uselagoon/lagoon-routegen/internal/helpers"
	"github.com/uselagoon/lagoon-routegen/internal/lagoon"
)

var domainName string

var fastlyConfigGeneration = &cobra.Command{
	Use:     "fastly",
	Aliases: []string{"f"},
	Short:   "Generate fastly configuration for a specific ingress domain",
	RunE: func(cmd *cobra.Command, args []string) error {
		// environment variables will override what is provided by flags
		lagoonFastlyCacheNoCahce = helpers.GetEnv("LAGOON_FASTLY_NOCACHE_SERVICE_ID", lagoonFastlyCacheNoCahce, true)
		lagoonFastlyServiceID = helpers.GetEnv("ROUTE_FASTLY_SERVICE_ID", lagoonFastlyServiceID, true)

		// get the project and environment variables
		projectVariables = helpers.GetEnv("LAGOON_PROJECT_VARIABLES", projectVariables, true)
		environmentVariables = helpers.GetEnv("LAGOON_ENVIRONMENT_VARIABLES", environmentVariables, true)

		// unmarshal and then merge the two so there is only 1 set of variables to iterate over
		projectVars := []lagoon.EnvironmentVariable{}
		envVars := []lagoon.EnvironmentVariable{}
		json.Unmarshal([]byte(projectVariables), &projectVars)
		json.Unmarshal([]byte(environmentVariables), &envVars)
		lagoonEnvVars := lagoon.MergeVariables(projectVars, envVars)

		// generate the fastly configuration from the provided flags/variables
		f, err := lagoon.GenerateFastlyConfiguration(lagoonFastlyCacheNoCahce, lagoonFastlyServiceID, domainName, lagoonEnvVars)
		if err != nil {
			return err
		}
		strF, _ := json.Marshal(f)
		fmt.Println(string(strF))
		return nil
	},
}

func init() {
	configCmd.AddCommand(fastlyConfigGeneration)
	fastlyConfigGeneration.Flags().StringVarP(&domainName, "domain", "D", "",
		"The .lagoon.yml file to read")
	fastlyConfigGeneration.Flags().StringVarP(&projectVariables, "project-variables", "v", "",
		"The projects environment variables JSON payload")
	fastlyConfigGeneration.Flags().StringVarP(&environmentVariables, "environment-variables", "V", "",
		"The environments environment variables JSON payload")
	fastlyConfigGeneration.Flags().StringVarP(&lagoonFastlyCacheNoCahce, "fastly-cache-no-cache-id", "F", "",
		"The fastly cache no cache service ID to use")
	fastlyConfigGeneration.Flags().StringVarP(&lagoonFastlyServiceID, "fastly-service-id", "f", "",
		"The fastly service ID to use")
}
