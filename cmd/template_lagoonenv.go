package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	generator "github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	servicestemplates "github.com/uselagoon/build-deploy-tool/internal/templating"
)

type DBaaSCredRefs []map[string]string

var lagoonEnvGeneration = &cobra.Command{
	Use:     "lagoon-env",
	Aliases: []string{"le"},
	Short:   "Generate the lagoon-env secret template for a Lagoon build",
	RunE: func(cmd *cobra.Command, args []string) error {
		generator, err := GenerateInput(*rootCmd, false)
		if err != nil {
			return err
		}
		routes, err := cmd.Flags().GetString("routes")
		if err != nil {
			return fmt.Errorf("error reading routes flag: %v", err)
		}
		secretName, err := cmd.Flags().GetString("secret-name")
		if err != nil {
			return fmt.Errorf("error reading secret-name flag: %v", err)
		}
		dbaasCreds, err := rootCmd.PersistentFlags().GetString("dbaas-creds")
		if err != nil {
			return fmt.Errorf("error reading dbaas creds flag: %v", err)
		}
		configMapVars, err := cmd.Flags().GetString("configmap-vars")
		if err != nil {
			return fmt.Errorf("error reading configmap variables flag: %v", err)
		}
		dbaasCredRefs, err := loadCredsFromFile(dbaasCreds)
		if err != nil {
			return err
		}
		cmVars := map[string]string{}
		if cmd.Flags().Lookup("configmap-vars").Changed {
			if err := json.Unmarshal([]byte(configMapVars), &cmVars); err != nil {
				return fmt.Errorf("error unmarshalling lagoon-env configmap variables payload: %v", err)
			}
		}
		generator.ConfigMapVars = cmVars
		dbCreds := map[string]string{}
		for _, v := range *dbaasCredRefs {
			for k, v1 := range v {
				dbCreds[k] = v1
			}
		}
		generator.DBaaSVariables = dbCreds
		return LagoonEnvTemplateGeneration(secretName, generator, routes)
	},
}

func loadCredsFromFile(file string) (*DBaaSCredRefs, error) {
	dbaasCredRefs := &DBaaSCredRefs{}
	dbaasCredJSON, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("couldn't read file %v: %v", file, err)
	}
	if err := json.Unmarshal(dbaasCredJSON, dbaasCredRefs); err != nil {
		return nil, fmt.Errorf("error unmarshalling dbaas creds payload: %v", err)
	}
	return dbaasCredRefs, nil
}

// LagoonEnvTemplateGeneration .
func LagoonEnvTemplateGeneration(
	name string,
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
	cm, err := servicestemplates.GenerateLagoonEnvSecret(name, *lagoonBuild.BuildValues)
	if err != nil {
		return fmt.Errorf("couldn't generate template: %v", err)
	}
	templateBytes, err := servicestemplates.TemplateSecret(cm)
	if err != nil {
		return fmt.Errorf("couldn't generate template: %v", err)
	}
	if len(templateBytes) > 0 {
		if g.Debug {
			fmt.Printf("Templating lagoon-env secret %s\n", fmt.Sprintf("%s/%s-secret.yaml", savedTemplates, name))
		}
		helpers.WriteTemplateFile(fmt.Sprintf("%s/%s-secret.yaml", savedTemplates, name), templateBytes)
	}
	return nil
}

func init() {
	templateCmd.AddCommand(lagoonEnvGeneration)
	lagoonEnvGeneration.Flags().StringP("routes", "R", "",
		"The routes from the environment")
	lagoonEnvGeneration.Flags().StringP("secret-name", "S", "",
		"The name of the secret")
	lagoonEnvGeneration.Flags().StringP("configmap-vars", "N", "",
		"Any variables from the legacy configmap that need to be retained")
}
