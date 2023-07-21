package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	generator "github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	servicestemplates "github.com/uselagoon/build-deploy-tool/internal/templating/services"
)

var lagoonServiceGeneration = &cobra.Command{
	Use:     "lagoon-services",
	Aliases: []string{"ls"},
	Short:   "Generate the lagoon service templates for a Lagoon build",
	RunE: func(cmd *cobra.Command, args []string) error {
		gen, err := generatorInput(true)
		if err != nil {
			return err
		}
		images, err := rootCmd.PersistentFlags().GetString("images")
		if err != nil {
			return fmt.Errorf("error reading images flag: %v", err)
		}
		var imageRefs struct {
			Images map[string]string `json:"images"`
		}
		imagesStr, _ := base64.StdEncoding.DecodeString(images)
		json.Unmarshal(imagesStr, &imageRefs)
		gen.ImageReferences = imageRefs.Images
		return LagoonServiceTemplateGeneration(gen)
	},
}

// LagoonServiceTemplateGeneration .
func LagoonServiceTemplateGeneration(g generator.GeneratorInput) error {
	lagoonBuild, err := generator.NewGenerator(
		g,
	)
	if err != nil {
		return err
	}
	savedTemplates := g.SavedTemplatesPath

	// generate the templates
	if g.Debug {
		fmt.Println(fmt.Sprintf("Templating service manifests %s", fmt.Sprintf("%s/%s.yaml", savedTemplates, "services")))
	}
	serviceTemplateYAML, err := servicestemplates.GenerateServiceTemplate(*lagoonBuild.BuildValues)
	if err != nil {
		return fmt.Errorf("couldn't generate template: %v", err)
	}
	helpers.WriteTemplateFile(fmt.Sprintf("%s/%s.yaml", savedTemplates, "services"), serviceTemplateYAML)
	if g.Debug {
		fmt.Println(fmt.Sprintf("Templating pvc manifests %s", fmt.Sprintf("%s/%s.yaml", savedTemplates, "pvcs")))
	}
	pvcTemplateYAML, err := servicestemplates.GeneratePVCTemplate(*lagoonBuild.BuildValues)
	if err != nil {
		return fmt.Errorf("couldn't generate template: %v", err)
	}
	helpers.WriteTemplateFile(fmt.Sprintf("%s/%s.yaml", savedTemplates, "pvcs"), pvcTemplateYAML)
	if g.Debug {
		fmt.Println(fmt.Sprintf("Templating deployment manifest %s", fmt.Sprintf("%s/%s.yaml", savedTemplates, "deployments")))
	}
	deploymentTemplateYAML, err := servicestemplates.GenerateDeploymentTemplate(*lagoonBuild.BuildValues)
	if err != nil {
		return fmt.Errorf("couldn't generate template: %v", err)
	}
	helpers.WriteTemplateFile(fmt.Sprintf("%s/%s.yaml", savedTemplates, "deployments"), deploymentTemplateYAML)
	return nil
}

func init() {
	templateCmd.AddCommand(lagoonServiceGeneration)
}
