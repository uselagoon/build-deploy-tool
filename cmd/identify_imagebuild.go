package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	generator "github.com/uselagoon/build-deploy-tool/internal/generator"
)

var imageBuildIdentify = &cobra.Command{
	Use:     "image-builds",
	Aliases: []string{"image-build", "img-build", "ib"},
	Short:   "Identify the configuration for building images for a Lagoon build",
	RunE: func(cmd *cobra.Command, args []string) error {
		gen, err := generator.GenerateInput(*rootCmd, false)
		if err != nil {
			return err
		}
		out, err := ImageBuildConfigurationIdentification(gen)
		if err != nil {
			return err
		}
		bc, err := json.Marshal(out)
		if err != nil {
			return err
		}
		fmt.Println(string(bc))
		return nil
	},
}

type imageBuild struct {
	BuildKit            *bool                         `json:"buildKit"`
	Images              []imageBuilds                 `json:"images"`
	BuildArguments      map[string]string             `json:"buildArguments"`
	ContainerRegistries []generator.ContainerRegistry `json:"containerRegistries,omitempty"`
}

type imageBuilds struct {
	Name       string               `json:"name"`
	ImageBuild generator.ImageBuild `json:"imageBuild"`
}

// ImageBuildConfigurationIdentification takes the output of the generator and turns it into a JSON payload
// that can be used by the legacy bash to build container images. This payload contains the buildkit flag if it was as part of the build
// but it also provides all the container contexts and dockerfile paths that can be passed to the build command
// this includes the:
// * temporary image name (namespace-service)
// * the push image name (registry/project/environment/service:latest)
// * the pull through image, if there is no dockerfile
// * eventually other information like build args etc
func ImageBuildConfigurationIdentification(g generator.GeneratorInput) (imageBuild, error) {

	lServices := imageBuild{}
	lagoonBuild, err := generator.NewGenerator(
		g,
	)
	if err != nil {
		return lServices, err
	}
	lServices.BuildKit = lagoonBuild.BuildValues.DockerBuildKit
	lServices.BuildArguments = lagoonBuild.BuildValues.ImageBuildArguments
	for _, service := range lagoonBuild.BuildValues.Services {
		if service.ImageBuild != nil {
			lServices.Images = append(lServices.Images, imageBuilds{
				Name:       service.Name,
				ImageBuild: *service.ImageBuild,
			})
		}
	}
	lServices.ContainerRegistries = lagoonBuild.BuildValues.ContainerRegistry
	return lServices, nil
}

func init() {
	identifyCmd.AddCommand(imageBuildIdentify)
}
