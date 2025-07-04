package cmd

import (
	"context"
	"fmt"
	"strings"

	mariadbv1 "github.com/amazeeio/dbaas-operator/apis/mariadb/v1"
	mongodbv1 "github.com/amazeeio/dbaas-operator/apis/mongodb/v1"
	postgresv1 "github.com/amazeeio/dbaas-operator/apis/postgres/v1"
	"github.com/spf13/cobra"
	"github.com/uselagoon/build-deploy-tool/internal/collector"
	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/k8s"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var cleanupCmd = &cobra.Command{
	Use:     "cleanup",
	Aliases: []string{"clean", "cu", "c"},
	Short:   "Cleanup old services",
	RunE: func(cmd *cobra.Command, args []string) error {
		deleteServices, err := cmd.Flags().GetBool("delete")
		if err != nil {
			return fmt.Errorf("error reading domain flag: %v", err)
		}
		client, err := k8s.NewClient()
		if err != nil {
			return err
		}
		// create a collector
		col := collector.NewCollector(client)
		gen, err := generator.GenerateInput(*rootCmd, false)
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
		namespace := helpers.GetEnv("NAMESPACE", "", false)
		namespace, err = helpers.GetNamespace(namespace, "/var/run/secrets/kubernetes.io/serviceaccount/namespace")
		if err != nil {
			return err
		}
		if namespace == "" {
			return fmt.Errorf("unable to detect namespace")
		}
		gen.Namespace = namespace
		gen.ImageReferences = imageRefs.Images
		_, _, _, _, _, _, err = RunCleanup(col, gen, deleteServices)
		if err != nil {
			return err
		}
		return nil
	},
}

func RunCleanup(c *collector.Collector, gen generator.GeneratorInput, deleteServices bool) ([]string, []string, []string, []string, []string, []string, error) {
	out, err := LagoonServiceTemplateIdentification(gen)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	dbaas, err := IdentifyDBaaSConsumers(gen)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	state, err := c.Collect(context.Background(), gen.Namespace)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	mariadbMatch := false
	var mariadbDelete []mariadbv1.MariaDBConsumer
	var mariadbDeleteS []string
	for _, exist := range state.MariaDBConsumers.Items {
		for _, prov := range dbaas {
			sp := strings.Split(prov, ":")
			if strings.Contains(sp[1], "mariadb-dbaas") {
				if exist.Name == sp[0] {
					mariadbMatch = true
					continue
				}
			}
		}
		if !mariadbMatch {
			mariadbDelete = append(mariadbDelete, exist)
			mariadbDeleteS = append(mariadbDeleteS, exist.Name)
		}
		mariadbMatch = false
	}

	mongodbMatch := false
	var mongodbDelete []mongodbv1.MongoDBConsumer
	var mongodbDeleteS []string
	for _, exist := range state.MongoDBConsumers.Items {
		for _, prov := range dbaas {
			sp := strings.Split(prov, ":")
			if strings.Contains(sp[1], "mongodb-dbaas") {
				if exist.Name == sp[0] {
					mongodbMatch = true
					continue
				}
			}
		}
		if !mongodbMatch {
			mongodbDelete = append(mongodbDelete, exist)
			mongodbDeleteS = append(mongodbDeleteS, exist.Name)
		}
		mongodbMatch = false
	}

	postgresqlMatch := false
	var postgresqlDelete []postgresv1.PostgreSQLConsumer
	var postgresqlDeleteS []string
	for _, exist := range state.PostgreSQLConsumers.Items {
		for _, prov := range dbaas {
			sp := strings.Split(prov, ":")
			if strings.Contains(sp[1], "postgres-dbaas") {
				if exist.Name == sp[0] {
					postgresqlMatch = true
					continue
				}
			}
		}
		if !postgresqlMatch {
			postgresqlDelete = append(postgresqlDelete, exist)
			postgresqlDeleteS = append(postgresqlDeleteS, exist.Name)
		}
		postgresqlMatch = false
	}

	depMatch := false
	var depDelete []appsv1.Deployment
	var depDeleteS []string
	for _, exist := range state.Deployments.Items {
		for _, prov := range out.Deployments {
			if exist.Name == prov {
				depMatch = true
				continue
			}
		}
		if !depMatch {
			depDelete = append(depDelete, exist)
			depDeleteS = append(depDeleteS, exist.Name)
		}
		depMatch = false
	}

	volMatch := false
	var volDelete []corev1.PersistentVolumeClaim
	var volDeleteS []string
	for _, exist := range state.PVCs.Items {
		for _, prov := range out.Volumes {
			if exist.Name == prov {
				volMatch = true
				continue
			}
		}
		if !volMatch {
			volDelete = append(volDelete, exist)
			volDeleteS = append(volDeleteS, exist.Name)
		}
		volMatch = false
	}

	servMatch := false
	var servDelete []corev1.Service
	var servDeleteS []string
	for _, exist := range state.Services.Items {
		for _, prov := range out.Services {
			if exist.Name == prov {
				servMatch = true
				continue
			}
		}
		if !servMatch {
			servDelete = append(servDelete, exist)
			servDeleteS = append(servDeleteS, exist.Name)
		}
		servMatch = false
	}

	if len(mariadbDeleteS) > 0 || len(mongodbDeleteS) > 0 || len(postgresqlDeleteS) > 0 || len(depDeleteS) > 0 || len(volDeleteS) > 0 || len(servDeleteS) > 0 {
		fmt.Println(`>> Lagoon detected services or volumes that have been removed from the docker-compose file`)
		if !deleteServices {
			fmt.Println(`> If you no longer need these services, you can instruct Lagoon to remove it from the environment by setting the following variable
  'LAGOON_FEATURE_FLAG_CLEANUP_REMOVED_LAGOON_SERVICES=enabled' as a GLOBAL scoped variable to this environment or project.
  Removing unused resources will result in the services and any data they had being deleted.
  Ensure your application is no longer configured to use these resources before removing them.
  If you're not sure, contact your support team with a link to this build.`)
		} else {
			fmt.Println(`> The flag 'LAGOON_FEATURE_FLAG_CLEANUP_REMOVED_LAGOON_SERVICES' is enabled.
  Resources that were removed from the docker-compose file will now be removed from the environment.
  The services and any data they had will be deleted.
  You should remove this variable if you don't want services to be removed automatically in the future.`)
		}
		fmt.Println(`> Future releases of Lagoon may remove services automatically, you should ensure that your services are up always up to date if you see this warning."`)

		for _, i := range depDelete {
			if deleteServices {
				fmt.Printf(">> Removing deployment %s\n", i.Name)
				if err := c.Client.Delete(context.Background(), &i); err != nil {
					return nil, nil, nil, nil, nil, nil, err
				}
			} else {
				fmt.Printf(">> Would remove deployment %s\n", i.Name)
			}
		}
		for _, i := range volDelete {
			if deleteServices {
				fmt.Printf(">> Removing volume %s\n", i.Name)
				if err := c.Client.Delete(context.Background(), &i); err != nil {
					return nil, nil, nil, nil, nil, nil, err
				}
			} else {
				fmt.Printf(">> Would remove volume %s\n", i.Name)
			}
		}
		for _, i := range servDelete {
			if deleteServices {
				fmt.Printf(">> Removing service %s\n", i.Name)
				if err := c.Client.Delete(context.Background(), &i); err != nil {
					return nil, nil, nil, nil, nil, nil, err
				}
			} else {
				fmt.Printf(">> Would remove service %s\n", i.Name)
			}
		}
		for _, i := range mariadbDelete {
			if deleteServices {
				fmt.Printf(">> Removing mariadb consumer %s\n", i.Name)
				if err := c.Client.Delete(context.Background(), &i); err != nil {
					return nil, nil, nil, nil, nil, nil, err
				}
			} else {
				fmt.Printf(">> Would remove mariadb consumer %s\n", i.Name)
			}
		}
		for _, i := range mongodbDelete {
			if deleteServices {
				fmt.Printf(">> Removing mongodb consumer %s\n", i.Name)
				if err := c.Client.Delete(context.Background(), &i); err != nil {
					return nil, nil, nil, nil, nil, nil, err
				}
			} else {
				fmt.Printf(">> Would remove mongodb consumer %s\n", i.Name)
			}
		}
		for _, i := range postgresqlDelete {
			if deleteServices {
				fmt.Printf(">> Removing postgresql consumer %s\n", i.Name)
				if err := c.Client.Delete(context.Background(), &i); err != nil {
					return nil, nil, nil, nil, nil, nil, err
				}
			} else {
				fmt.Printf(">> Would remove postgresql consumer %s\n", i.Name)
			}
		}
		return mariadbDeleteS, mongodbDeleteS, postgresqlDeleteS, depDeleteS, volDeleteS, servDeleteS, nil
	} else {
		return nil, nil, nil, nil, nil, nil, nil
	}
}

func init() {
	runCmd.AddCommand(cleanupCmd)
	cleanupCmd.Flags().Bool("delete", false, "flag to actually delete services")
}
