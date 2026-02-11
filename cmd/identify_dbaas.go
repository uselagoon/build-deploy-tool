package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/uselagoon/build-deploy-tool/internal/identify"
)

// this is an intermediate helper command while transitioning from bash to go
// eventually this won't be required
var dbaasIdentify = &cobra.Command{
	Use:     "dbaas",
	Aliases: []string{"db", "d"},
	Short:   "Identify if any dbaas consumers are created",
	RunE: func(cmd *cobra.Command, args []string) error {
		generator, err := GenerateInput(*rootCmd, false)
		if err != nil {
			return err
		}
		dbaasConsumers, err := identify.IdentifyDBaaSConsumers(generator)
		if err != nil {
			return err
		}
		for _, dbc := range dbaasConsumers {
			fmt.Println(dbc)
		}
		return nil
	},
}
