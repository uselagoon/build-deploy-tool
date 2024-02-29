package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	generator "github.com/uselagoon/build-deploy-tool/internal/generator"
)

// this is an intermediate helper command while transitioning from bash to go
// eventually this won't be required
var dbaasIdentify = &cobra.Command{
	Use:     "dbaas",
	Aliases: []string{"db", "d"},
	Short:   "Identify if any dbaas consumers are created",
	RunE: func(cmd *cobra.Command, args []string) error {
		generator, err := generator.GenerateInput(*rootCmd, false)
		if err != nil {
			return err
		}
		dbaasConsumers, err := IdentifyDBaaSConsumers(generator)
		if err != nil {
			return err
		}
		for _, dbc := range dbaasConsumers {
			fmt.Println(dbc)
		}
		return nil
	},
}

func IdentifyDBaaSConsumers(g generator.GeneratorInput) ([]string, error) {
	lagoonBuild, err := generator.NewGenerator(
		g,
	)
	if err != nil {
		return nil, err
	}
	ret := []string{}
	for _, svc := range lagoonBuild.BuildValues.Services {
		if svc.IsDBaaS {
			ret = append(ret, fmt.Sprintf("%s:%s", svc.Name, svc.Type))
		}
	}
	return ret, nil
}
