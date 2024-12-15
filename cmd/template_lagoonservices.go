package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	generator "github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	servicestemplates "github.com/uselagoon/build-deploy-tool/internal/templating"
	"sigs.k8s.io/yaml"
)

type ImageReferences struct {
	Images map[string]string `json:"images"`
}

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
		imageRefs, err := loadImagesFromFile(images)
		if err != nil {
			return err
		}
		gen.ImageReferences = imageRefs.Images
		return LagoonServiceTemplateGeneration(gen)
	},
}

func loadImagesFromFile(file string) (*ImageReferences, error) {
	imageRefs := &ImageReferences{}
	imageYAML, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("couldn't read file %v: %v", file, err)
	}
	if err := yaml.Unmarshal(imageYAML, imageRefs); err != nil {
		return nil, fmt.Errorf("error unmarshalling images payload: %v", err)
	}
	return imageRefs, nil
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
	secrets, err := servicestemplates.GenerateRegistrySecretTemplate(*lagoonBuild.BuildValues)
	if err != nil {
		return fmt.Errorf("couldn't generate template: %v", err)
	}
	for _, secret := range secrets {
		templateBytes, err := servicestemplates.TemplateSecret(secret)
		if err != nil {
			return fmt.Errorf("couldn't generate template: %v", err)
		}
		if g.Debug {
			fmt.Printf("Templating registry secret manifests %s\n", fmt.Sprintf("%s/%s.yaml", savedTemplates, secret.Name))
		}
		helpers.WriteTemplateFile(fmt.Sprintf("%s/%s.yaml", savedTemplates, secret.Name), templateBytes)
	}
	services, err := servicestemplates.GenerateServiceTemplate(*lagoonBuild.BuildValues)
	if err != nil {
		return fmt.Errorf("couldn't generate template: %v", err)
	}
	for _, d := range services {
		templateBytes, err := servicestemplates.TemplateService(d)
		if err != nil {
			return fmt.Errorf("couldn't generate template: %v", err)
		}
		if g.Debug {
			fmt.Printf("Templating service manifests %s\n", fmt.Sprintf("%s/service-%s.yaml", savedTemplates, d.Name))
		}
		helpers.WriteTemplateFile(fmt.Sprintf("%s/service-%s.yaml", savedTemplates, d.Name), templateBytes)
	}
	pvcs, err := servicestemplates.GeneratePVCTemplate(*lagoonBuild.BuildValues)
	if err != nil {
		return fmt.Errorf("couldn't generate template: %v", err)
	}
	for _, d := range pvcs {
		templateBytes, err := servicestemplates.TemplatePVC(d)
		if err != nil {
			return fmt.Errorf("couldn't generate template: %v", err)
		}
		if g.Debug {
			fmt.Printf("Templating pvc manifests %s\n", fmt.Sprintf("%s/pvc-%s.yaml", savedTemplates, d.Name))
		}
		helpers.WriteTemplateFile(fmt.Sprintf("%s/pvc-%s.yaml", savedTemplates, d.Name), templateBytes)
	}
	deployments, err := servicestemplates.GenerateDeploymentTemplate(*lagoonBuild.BuildValues)
	if err != nil {
		return fmt.Errorf("couldn't generate template: %v", err)
	}
	for _, d := range deployments {
		templateBytes, err := servicestemplates.TemplateDeployment(d)
		if err != nil {
			return fmt.Errorf("couldn't generate template: %v", err)
		}
		if g.Debug {
			fmt.Printf("Templating deployment manifests %s\n", fmt.Sprintf("%s/deployment-%s.yaml", savedTemplates, d.Name))
		}
		helpers.WriteTemplateFile(fmt.Sprintf("%s/deployment-%s.yaml", savedTemplates, d.Name), templateBytes)
	}
	cronjobs, err := servicestemplates.GenerateCronjobTemplate(*lagoonBuild.BuildValues)
	if err != nil {
		return fmt.Errorf("couldn't generate template: %v", err)
	}
	for _, d := range cronjobs {
		templateBytes, err := servicestemplates.TemplateCronjobs(d)
		if err != nil {
			return fmt.Errorf("couldn't generate template: %v", err)
		}
		if g.Debug {
			fmt.Printf("Templating cronjob manifests %s\n", fmt.Sprintf("%s/cronjob-%s.yaml", savedTemplates, d.Name))
		}
		helpers.WriteTemplateFile(fmt.Sprintf("%s/cronjob-%s.yaml", savedTemplates, d.Name), templateBytes)
	}
	if lagoonBuild.BuildValues.IsolationNetworkPolicy {
		// if isolation network policies are enabled, template that here
		np, err := servicestemplates.GenerateNetworkPolicy(*lagoonBuild.BuildValues)
		if err != nil {
			return fmt.Errorf("couldn't generate template: %v", err)
		}
		templateBytes, err := servicestemplates.TemplateNetworkPolicy(np)
		if err != nil {
			return fmt.Errorf("couldn't generate template: %v", err)
		}
		if g.Debug {
			fmt.Printf("Templating networkpolicy manifest %s\n", fmt.Sprintf("%s/isolation-network-policy.yaml", savedTemplates))
		}
		helpers.WriteTemplateFile(fmt.Sprintf("%s/isolation-network-policy.yaml", savedTemplates), templateBytes)
	}
	return nil
}

func init() {
	templateCmd.AddCommand(lagoonServiceGeneration)
}
