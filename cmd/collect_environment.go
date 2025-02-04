package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/uselagoon/build-deploy-tool/internal/collector"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	"github.com/uselagoon/build-deploy-tool/internal/k8s"
)

func init() {
	collectCmd.AddCommand(collectEnvironment)
}

var collectEnvironment = &cobra.Command{
	Use:     "environment",
	Aliases: []string{"e"},
	Short:   "Collect seed information about the environment",
	RunE: func(cmd *cobra.Command, args []string) error {
		// get a k8s client
		client, err := k8s.NewClient()
		if err != nil {
			return err
		}
		// create a collector
		col := collector.NewCollector(client)
		namespace := helpers.GetEnv("NAMESPACE", "", false)
		namespace, err = helpers.GetNamespace(namespace, "/var/run/secrets/kubernetes.io/serviceaccount/namespace")
		if err != nil {
			return err
		}
		if namespace == "" {
			return fmt.Errorf("unable to detect namespace")
		}
		// collect the environment
		data, err := CollectEnvironment(col, namespace)
		if err != nil {
			return err
		}
		env, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(env))
		return nil
	},
}

// CollectEnvironment .
func CollectEnvironment(c *collector.Collector, namespace string) (*collector.LagoonEnvState, error) {
	state, err := c.Collect(context.Background(), namespace)
	if err != nil {
		return nil, err
	}
	return state, nil
}
