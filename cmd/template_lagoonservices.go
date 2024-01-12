package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	generator "github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	servicestemplates "github.com/uselagoon/build-deploy-tool/internal/templating/services"
	"sigs.k8s.io/yaml"
)

var lagoonServiceGeneration = &cobra.Command{
	Use:     "lagoon-services",
	Aliases: []string{"ls"},
	Short:   "Generate the lagoon service templates for a Lagoon build",
	RunE: func(cmd *cobra.Command, args []string) error {
		gen, err := generator.GenerateInput(*rootCmd, true)
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
		imagesStr, err := base64.StdEncoding.DecodeString(images)
		if err != nil {
			return fmt.Errorf("error decoding images payload: %v", err)
		}
		if err := json.Unmarshal(imagesStr, &imageRefs); err != nil {
			return fmt.Errorf("error unmarshalling images payload: %v", err)
		}
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
	services, err := servicestemplates.GenerateServiceTemplate(*lagoonBuild.BuildValues)
	if err != nil {
		return fmt.Errorf("couldn't generate template: %v", err)
	}
	if services != nil {
		for _, d := range services {
			serviceBytes, err := yaml.Marshal(d)
			if err != nil {
				return fmt.Errorf("couldn't generate template: %v", err)
			}
			separator := []byte("---\n")
			restoreResult := append(separator[:], serviceBytes[:]...)
			if g.Debug {
				fmt.Println(fmt.Sprintf("Templating service manifests %s", fmt.Sprintf("%s/service-%s.yaml", savedTemplates, d.Name)))
			}
			helpers.WriteTemplateFile(fmt.Sprintf("%s/service-%s.yaml", savedTemplates, d.Name), restoreResult)
		}
	}
	pvcs, err := servicestemplates.GeneratePVCTemplate(*lagoonBuild.BuildValues)
	if err != nil {
		return fmt.Errorf("couldn't generate template: %v", err)
	}
	if pvcs != nil {
		for _, d := range pvcs {
			serviceBytes, err := yaml.Marshal(d)
			if err != nil {
				return fmt.Errorf("couldn't generate template: %v", err)
			}
			separator := []byte("---\n")
			restoreResult := append(separator[:], serviceBytes[:]...)
			if g.Debug {
				fmt.Println(fmt.Sprintf("Templating pvc manifests %s", fmt.Sprintf("%s/pvc-%s.yaml", savedTemplates, d.Name)))
			}
			helpers.WriteTemplateFile(fmt.Sprintf("%s/pvc-%s.yaml", savedTemplates, d.Name), restoreResult)
		}
	}
	deployments, err := servicestemplates.GenerateDeploymentTemplate(*lagoonBuild.BuildValues)
	if err != nil {
		return fmt.Errorf("couldn't generate template: %v", err)
	}
	if deployments != nil {
		for _, d := range deployments {
			deploymentBytes, err := yaml.Marshal(d)
			if err != nil {
				return fmt.Errorf("couldn't generate template: %v", err)
			}
			separator := []byte("---\n")
			restoreResult := append(separator[:], deploymentBytes[:]...)
			if g.Debug {
				fmt.Println(fmt.Sprintf("Templating deployment manifests %s", fmt.Sprintf("%s/deployment-%s.yaml", savedTemplates, d.Name)))
			}
			helpers.WriteTemplateFile(fmt.Sprintf("%s/deployment-%s.yaml", savedTemplates, d.Name), restoreResult)
		}
	}
	cronjobs, err := servicestemplates.GenerateCronjobTemplate(*lagoonBuild.BuildValues)
	if err != nil {
		return fmt.Errorf("couldn't generate template: %v", err)
	}
	if cronjobs != nil {
		for _, d := range cronjobs {
			deploymentBytes, err := yaml.Marshal(d)
			if err != nil {
				return fmt.Errorf("couldn't generate template: %v", err)
			}
			separator := []byte("---\n")
			restoreResult := append(separator[:], deploymentBytes[:]...)
			if g.Debug {
				fmt.Println(fmt.Sprintf("Templating cronjob manifests %s", fmt.Sprintf("%s/cronjob-%s.yaml", savedTemplates, d.Name)))
			}
			helpers.WriteTemplateFile(fmt.Sprintf("%s/cronjob-%s.yaml", savedTemplates, d.Name), restoreResult)
		}
	}
	return nil
}

func init() {
	templateCmd.AddCommand(lagoonServiceGeneration)
}
