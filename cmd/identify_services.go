package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	generator "github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
)

type identifyServices struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

var servicesIdentify = &cobra.Command{
	Use:     "services",
	Aliases: []string{"s"},
	Short:   "Identify services that this build would create",
	RunE: func(cmd *cobra.Command, args []string) error {
		generator, err := generator.GenerateInput(*rootCmd, false)
		if err != nil {
			return err
		}
		ret, _, err := IdentifyServices(generator)
		if err != nil {
			return err
		}
		retJSON, _ := json.Marshal(ret)
		fmt.Println(string(retJSON))
		return nil
	},
}

// IdentifyServices identifies services that this build would create
func IdentifyServices(g generator.GeneratorInput) ([]string, []identifyServices, error) {
	lagoonBuild, err := generator.NewGenerator(
		g,
	)
	if err != nil {
		return nil, nil, err
	}

	services := []string{}
	serviceTypes := []identifyServices{}
	for _, service := range lagoonBuild.BuildValues.Services {
		if service.Type != "" {
			services = helpers.AppendIfMissing(services, service.OverrideName)
			serviceTypes = AppendIfMissing(serviceTypes, identifyServices{
				Name: service.OverrideName,
				Type: service.Type,
			})
		}
	}
	return services, serviceTypes, nil
}

func init() {
	identifyCmd.AddCommand(servicesIdentify)
}

func AppendIfMissing(slice []identifyServices, i identifyServices) []identifyServices {
	for _, ele := range slice {
		if ele.Name == i.Name {
			return slice
		}
	}
	return append(slice, i)
}
