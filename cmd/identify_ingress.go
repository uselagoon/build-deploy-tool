package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	generator "github.com/uselagoon/build-deploy-tool/internal/generator"
)

type ingressIdentifyJSON struct {
	Primary       string   `json:"primary"`
	Secondary     []string `json:"secondary"`
	Autogenerated []string `json:"autogenerated"`
}

var primaryIngressIdentify = &cobra.Command{
	Use:     "primary-ingress",
	Aliases: []string{"pi"},
	Short:   "Identify the primary ingress for a specific environment",
	RunE: func(cmd *cobra.Command, args []string) error {
		primary, _, _, err := IdentifyPrimaryIngress(generatorInput(true))
		if err != nil {
			return err
		}
		fmt.Println(primary)
		return nil
	},
}

var ingressIdentify = &cobra.Command{
	Use:     "ingress",
	Aliases: []string{"i"},
	Short:   "Identify all ingress for a specific environment",
	RunE: func(cmd *cobra.Command, args []string) error {
		primary, secondary, autogen, err := IdentifyPrimaryIngress(generatorInput(true))
		if err != nil {
			return err
		}
		ret := ingressIdentifyJSON{
			Primary:       primary,
			Secondary:     secondary,
			Autogenerated: autogen,
		}
		retJSON, _ := json.Marshal(ret)
		fmt.Println(string(retJSON))
		return nil
	},
}

// IdentifyPrimaryIngress .
func IdentifyPrimaryIngress(g generator.GeneratorInput) (string, []string, []string, error) {
	lagoonBuild, err := generator.NewGenerator(
		g,
	)
	if err != nil {
		return "", nil, nil, err
	}

	return lagoonBuild.BuildValues.Route, lagoonBuild.BuildValues.Routes, lagoonBuild.BuildValues.AutogeneratedRoutes, nil
}

func init() {
	identifyCmd.AddCommand(primaryIngressIdentify)
	identifyCmd.AddCommand(ingressIdentify)
}
