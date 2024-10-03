package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	generator "github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/templating/lagoonenv"
	"sigs.k8s.io/yaml"
)

type DBaaSCredRefs []map[string]string

var lagoonEnvGeneration = &cobra.Command{
	Use:     "lagoon-env",
	Aliases: []string{"le"},
	Short:   "Generate the lagoon-env configmap template for a Lagoon build",
	RunE: func(cmd *cobra.Command, args []string) error {
		generator, err := generator.GenerateInput(*rootCmd, true)
		if err != nil {
			return err
		}
		routes, err := cmd.Flags().GetString("routes")
		if err != nil {
			return fmt.Errorf("error reading routes flag: %v", err)
		}
		dbaasCreds, err := rootCmd.PersistentFlags().GetString("dbaas-creds")
		if err != nil {
			return fmt.Errorf("error reading images flag: %v", err)
		}
		dbaasCredRefs, err := loadCredsFromFile(dbaasCreds)
		if err != nil {
			return err
		}
		dbCreds := map[string]string{}
		for _, v := range *dbaasCredRefs {
			for k, v1 := range v {
				dbCreds[k] = v1
			}
		}
		generator.DBaaSVariables = dbCreds
		return LagoonEnvTemplateGeneration(generator, routes)
	},
}

func loadCredsFromFile(file string) (*DBaaSCredRefs, error) {
	dbaasCredRefs := &DBaaSCredRefs{}
	dbaasCredJSON, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("couldn't read file %v: %v", file, err)
	}
	if err := json.Unmarshal(dbaasCredJSON, dbaasCredRefs); err != nil {
		return nil, fmt.Errorf("error unmarshalling images payload: %v", err)
	}
	return dbaasCredRefs, nil
}

// LagoonEnvTemplateGeneration .
func LagoonEnvTemplateGeneration(
	g generator.GeneratorInput,
	routes string,
) error {
	lagoonBuild, err := generator.NewGenerator(
		g,
	)
	if err != nil {
		return err
	}
	savedTemplates := g.SavedTemplatesPath
	// if the routes have been passed from the command line, use them instead. we do this since lagoon currently doesn't enforce route state to match
	// what is in the `.lagoon.yml` file, so there may be items that exist in the cluster that don't exist in yaml
	// eventually once route state enforcement is enforced, or the tool can reconcile what is in the cluster itself rather than in bash
	// then this can be removed
	// https://github.com/uselagoon/build-deploy-tool/blob/f527a89ad5efb46e19a2f59d9ff3ffbff541e2a2/legacy/build-deploy-docker-compose.sh#L1090
	if routes != "" {
		lagoonBuild.BuildValues.Routes = strings.Split(routes, ",")
	}
	cm, err := lagoonenv.GenerateLagoonEnvConfigMap(*lagoonBuild.BuildValues)
	if err != nil {
		return fmt.Errorf("couldn't generate template: %v", err)
	}
	cmBytes, err := yaml.Marshal(cm)
	if err != nil {
		return fmt.Errorf("couldn't generate template: %v", err)
	}
	if len(cmBytes) > 0 {
		if g.Debug {
			fmt.Printf("Templating lagoon-env configmap %s\n", fmt.Sprintf("%s/%s.yaml", savedTemplates, "lagoon-env-configmap"))
		}
		helpers.WriteTemplateFile(fmt.Sprintf("%s/%s.yaml", savedTemplates, "lagoon-env-configmap"), cmBytes)
	}
	return nil
}

func init() {
	templateCmd.AddCommand(lagoonEnvGeneration)
	lagoonEnvGeneration.Flags().StringP("routes", "R", "",
		"The routes from the environment")
}
