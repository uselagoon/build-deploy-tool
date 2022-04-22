package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/uselagoon/lagoon-routegen/internal/generator"
)

var ingressDomain string

var fastlyConfigGeneration = &cobra.Command{
	Use:     "fastly-config",
	Aliases: []string{"f", "fc"},
	Short:   "Generate fastly configuration for a specific ingress domain",
	RunE: func(cmd *cobra.Command, args []string) error {
		lagoonFastlyCacheNoCahce = getEnv("LAGOON_FASTLY_NOCACHE_SERVICE_ID", lagoonFastlyCacheNoCahce)

		// get the project and environment variables
		projectVariables = getEnv("LAGOON_PROJECT_VARIABLES", projectVariables)
		environmentVariables = getEnv("LAGOON_ENVIRONMENT_VARIABLES", environmentVariables)

		// unmarshal and then merge the two so there is only 1 set of variables to iterate over
		projectVars := []generator.LagoonEnvironmentVariable{}
		envVars := []generator.LagoonEnvironmentVariable{}
		json.Unmarshal([]byte(projectVariables), &projectVars)
		json.Unmarshal([]byte(environmentVariables), &envVars)
		lagoonEnvVars := generator.MergeVariables(projectVars, envVars)

		f, err := generator.GenerateFastlyConfiguration(lagoonFastlyCacheNoCahce, "", ingressDomain, lagoonEnvVars)
		if err != nil {
			return err
		}
		strF, _ := json.Marshal(f)
		fmt.Println(string(strF))
		return nil
	},
}

func init() {
	fastlyConfigGeneration.Flags().StringVarP(&ingressDomain, "ingress-domain", "I", "",
		"The .lagoon.yml file to read")
}
