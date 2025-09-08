package cleanup

import (
	"context"
	"fmt"

	"github.com/uselagoon/build-deploy-tool/internal/collector"
	"github.com/uselagoon/build-deploy-tool/internal/generator"
	"github.com/uselagoon/build-deploy-tool/internal/identify"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

func RunCleanup(c *collector.Collector, gen generator.GeneratorInput, performDeletion bool) ([]string, []string, []string, []string, []string, []string, error) {
	_, mariadbDelete, mongodbDelete, postgresqlDelete, depDelete, volDelete, servDelete, state, err := identify.GetCurrentState(c, gen)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	if len(mariadbDelete) > 0 || len(mongodbDelete) > 0 || len(postgresqlDelete) > 0 || len(depDelete) > 0 || len(volDelete) > 0 || len(servDelete) > 0 {
		fmt.Println(`>> Lagoon detected services or volumes that have been removed from the docker-compose file`)
		if !performDeletion {
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

		var mariaDBToDelete, mongoDBToDelete, postgresToDelete, volumesToDelete, servicesToDelete, deploymentsToDelete []string
		ctx := context.Background()
		for _, i := range depDelete {
			deploymentsToDelete = append(deploymentsToDelete, i.Name)
			if performDeletion {
				fmt.Printf(">> Removing deployment %s\n", i.Name)
				if err := c.Client.Delete(ctx, &i); err != nil {
					return nil, nil, nil, nil, nil, nil, err
				}
			} else {
				fmt.Printf(">> Would remove deployment %s\n", i.Name)
			}
		}
		for _, i := range volDelete {
			volumesToDelete = append(volumesToDelete, i.Name)
			if performDeletion {
				fmt.Printf(">> Removing volume %s\n", i.Name)
				if err := c.Client.Delete(ctx, &i); err != nil {
					return nil, nil, nil, nil, nil, nil, err
				}
			} else {
				fmt.Printf(">> Would remove volume %s\n", i.Name)
			}
		}
		for _, i := range servDelete {
			servicesToDelete = append(servicesToDelete, i.Name)
			if performDeletion {
				fmt.Printf(">> Removing service %s\n", i.Name)
				if err := c.Client.Delete(ctx, &i); err != nil {
					return nil, nil, nil, nil, nil, nil, err
				}
			} else {
				fmt.Printf(">> Would remove service %s\n", i.Name)
			}
		}
		for _, i := range mariadbDelete {
			mariaDBToDelete = append(mariaDBToDelete, i.Name)
			if performDeletion {
				fmt.Printf(">> Removing mariadb consumer %s\n", i.Name)
				if err := c.Client.Delete(ctx, &i); err != nil {
					return nil, nil, nil, nil, nil, nil, err
				}
				err := removePreBackupPod(ctx, c.Client, state, i.Name)
				if err != nil {
					return nil, nil, nil, nil, nil, nil, err
				}
			} else {
				fmt.Printf(">> Would remove mariadb consumer %s and associated components\n", i.Name)
			}
		}
		for _, i := range mongodbDelete {
			mongoDBToDelete = append(mongoDBToDelete, i.Name)
			if performDeletion {
				fmt.Printf(">> Removing mongodb consumer %s\n", i.Name)
				if err := c.Client.Delete(ctx, &i); err != nil {
					return nil, nil, nil, nil, nil, nil, err
				}
				err := removePreBackupPod(ctx, c.Client, state, i.Name)
				if err != nil {
					return nil, nil, nil, nil, nil, nil, err
				}
			} else {
				fmt.Printf(">> Would remove mongodb consumer %s and associated components\n", i.Name)
			}
		}
		for _, i := range postgresqlDelete {
			postgresToDelete = append(postgresToDelete, i.Name)
			if performDeletion {
				fmt.Printf(">> Removing postgresql consumer %s\n", i.Name)
				if err := c.Client.Delete(ctx, &i); err != nil {
					return nil, nil, nil, nil, nil, nil, err
				}
				err := removePreBackupPod(ctx, c.Client, state, i.Name)
				if err != nil {
					return nil, nil, nil, nil, nil, nil, err
				}
			} else {
				fmt.Printf(">> Would remove postgresql consumer %s and associated components\n", i.Name)
			}
		}
		return mariaDBToDelete, mongoDBToDelete, postgresToDelete, deploymentsToDelete, volumesToDelete, servicesToDelete, nil
	} else {
		return nil, nil, nil, nil, nil, nil, nil
	}
}

func removePreBackupPod(ctx context.Context, c client.Client, state *collector.LagoonEnvState, name string) error {
	for _, pbp := range state.PreBackupPodsV1.Items {
		if pbp.Name == fmt.Sprintf("%s-prebackuppod", name) {
			fmt.Printf(">> Removing mariadb prebackuppod %s\n", name)
			if err := c.Delete(ctx, &pbp); err != nil {
				return err
			}
		}
	}
	for _, pbp := range state.PreBackupPodsV1Alpha1.Items {
		if pbp.Name == fmt.Sprintf("%s-prebackuppod", name) {
			fmt.Printf(">> Removing mariadb prebackuppod %s\n", name)
			if err := c.Delete(ctx, &pbp); err != nil {
				return err
			}
		}
	}
	return nil
}
